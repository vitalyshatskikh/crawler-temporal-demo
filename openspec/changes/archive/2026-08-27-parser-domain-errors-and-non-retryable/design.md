## Context

The parser package (`crawler/parser/internal/...`) currently has three
overlapping sentinels for "could not parse the input":
`domain.ErrUnmarshalBody`, `domain.ErrParsingFailed`, and the
package-local `repositories.ErrUnmarshalConfig`. None of them
differentiate permanent from transient failure, and the activity layer
returns them as plain Go errors. When these functions become Temporal
activities, every failure will be retried by default.

This design covers two changes: (1) consolidating the parse-error
sentinels so all layers funnel into `domain.ErrParsingFailed`, and
(2) marking the three domain sentinels
(`ErrValidation`, `ErrNotFound`, `ErrParsingFailed`) as
non-retryable in the activities via
`temporal.NewNonRetryableError`. Raw DB errors and `context.*`
errors are left untouched so Temporal retries them.

## Goals / Non-Goals

**Goals**
- One canonical parse-error sentinel in `domain` (`ErrParsingFailed`).
- Repos return `domain.*` errors (already true for not-found; the
  unmarshal site is converted too).
- Activity errors wrapping `ErrValidation`, `ErrNotFound`, or
  `ErrParsingFailed` are marked non-retryable.
- Transient errors (raw DB errors, `context.*`) pass through unchanged.
- Unit tests assert both `temporal.IsNonRetryableError` and the
  underlying domain sentinel via `errors.Is`.

**Non-Goals**
- Wrapping raw `pgx` errors in a new `domain.ErrInternal` — they stay
  as-is and Temporal retries them.
- Configuring Temporal retry policies or activity options — left to
  the worker bootstrap when activities get registered.
- Touching the example-site package or any Python surfers.

## Decisions

### Single sentinel for all parse failures

Use `domain.ErrParsingFailed` for body unmarshal, config unmarshal,
and JMES search failures. Rationale: they all describe the same
externally-visible outcome ("input was not parseable") and downstream
callers should treat them identically. Alternatives:

- Split into `ErrInvalidJSON`, `ErrInvalidJMES`, `ErrInvalidConfig` —
  rejected because callers cannot usefully act differently on them
  and the surfacing cost (more sentinels to keep in sync) outweighs
  the benefit.
- Keep three separate sentinels with explicit names — rejected for
  the same reason.

### Drop `domain.ErrUnmarshalBody` and `repositories.ErrUnmarshalConfig`

Both are absorbed into `ErrParsingFailed`. `domain.ErrUnmarshalBody`
lives in `domain/errors.go` and is removed at the source.
`repositories.ErrUnmarshalConfig` is a package-level var in
`pg_config_repo.go`; removing it does not affect any other package
since the only references are the definition site, the use site in
the same file, and one integration test in the same package.

### Mark via `temporal.NewNonRetryableError` in a small helper

Add a single `wrapErr(err error) error` helper in
`activities/parser.go` (or a sibling `errors.go` in the same
package). It uses `errors.Is` to check for the three domain
sentinels and wraps with
`temporal.NewNonRetryableError(message, errTypeParsing, err)` where
`errTypeParsing = "ParsingError"` is a package-level const. Every
existing return in `ParseSearchPage` and `ParseAdvertContent` is
funneled through it.

```
err -> wrapErr(err)
```

Rationale: a single helper keeps the policy in one place. Alternatives:

- Mark at construction site (wrap each `fmt.Errorf` at the repo /
  service layer) — rejected because the application/activities
  layer is the Temporal boundary, and the rule "domain errors are
  non-retryable" is an activity-layer concern, not a domain concern.
- Define a `NonRetryable` interface in `domain` that activities
  detect via `errors.As` — rejected as over-engineering; the
  three-class policy is simple enough for a helper.
- Configure non-retryable via Temporal activity options (`RetryPolicy`
  with `NonRetryableErrorTypes`) — rejected because it requires the
  worker bootstrap that doesn't exist yet and would couple retry
  semantics to registration order. Per-error marking is more local
  and testable.

### `go.temporal.io/sdk` added to `crawler/go.mod`

This is a new direct dependency for the parser module. It is needed
in `activities/parser.go` (for `temporal.NewNonRetryableError`) and
in `activities/parser_test.go` (for `temporal.IsNonRetryableError`).
`go mod tidy` resolves the latest stable version.

### Activity tests use both assertions

Every existing test that asserts a domain sentinel
(`ErrValidation`, `ErrNotFound`) gains an
`assert.True(t, temporal.IsNonRetryableError(err))` line. New tests
are added for the transient path: raw `errors.New(...)` from the repo
and `SaveDocument` must NOT be marked non-retryable.

## Risks / Trade-offs

- **[Risk] Adding `go.temporal.io/sdk` is a sizable dependency.**
  → Mitigation: it is the canonical Go SDK and is already the
  intended runtime for this package. The cost is one-time; the
  alternative (custom marker interface) is more code and less
  expressive.

- **[Risk] `temporal.IsNonRetryableError` only checks the top-level
  error type, not the wrapped chain.** Activity returns are wrapped
  via `fmt.Errorf("%w: ...", err)`, so the non-retryable error from
  `temporal.NewNonRetryableError` is the top wrapper and the test
  passes. But if a future refactor wraps the activity result again,
  the assertion would silently pass-through (since the wrapped
  `*temporal.NonRetryableError` is no longer at the top).
  → Mitigation: tests also assert `errors.Is(err, domain.Err*)`
  so the underlying sentinel is still verifiable. Add a comment in
  the helper reminding callers not to wrap the returned value again.

- **[Risk] `domain.ErrParsingFailed` is now used for three distinct
  internal failures (body unmarshal, config unmarshal, JMES search
  failure), making error messages less specific.**
  → Mitigation: existing `fmt.Errorf("%w: ...", ...)` wrapping at each
  site keeps the messages informative
  (e.g. `failed to parse prop <src>/<type>/<name>: <jmes err>`).
  Callers relying on `errors.Is(err, domain.ErrParsingFailed)` still
  get a single switch.

- **[Risk] Removing `domain.ErrUnmarshalBody` and
  `repositories.ErrUnmarshalConfig` is breaking for any out-of-tree
  caller.** → Mitigation: both are package-local sentinel errors
  with no documented export contract; the parser module is
  `internal/` and only this repo imports it.

## Migration Plan

No data migration. Code rollout:

1. Land the error-sentinel consolidation and test renames in one
   commit. `go vet ./... && go test ./...` from the parser module.
2. Add `go.temporal.io/sdk` to `crawler/go.mod` and run `go mod tidy`.
3. Land the `wrapErr` helper, the activity wrapping, and the
   updated/new activity tests in a second commit. Re-run vet + unit
   tests.
4. Roll back by reverting either commit; no runtime data is touched.

## Open Questions

None. All five clarifying decisions from the planning conversation
were resolved:

1. `ErrParsingFailed` (typo correction).
2. Add `go.temporal.io/sdk` and use `temporal.NewNonRetryableError`.
3. Don't wrap raw DB errors.
4. Non-retryable classes: `ErrValidation`, `ErrNotFound`,
   `ErrParsingFailed` only.
5. Activity tests assert via both `temporal.IsNonRetryableError` and
   `errors.Is`.
