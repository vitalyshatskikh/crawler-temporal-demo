## Why

The parser's error surface is fragmented across two layers and three
near-duplicate sentinels (`domain.ErrUnmarshalBody`,
`repositories.ErrUnmarshalConfig`, plus the generic
`domain.ErrParsingFailed`). All three describe "could not parse the
input", but callers cannot treat them uniformly because they live in
different packages. At the same time, the activity layer does not
distinguish permanent failures (bad input, missing data, malformed
JSON/JMES) from transient ones (DB blip), so when these activities
become Temporal activities they will retry forever on errors that
retrying cannot fix. We want one canonical parse-error sentinel,
domain-owned errors throughout, and explicit non-retryable marking
in the activities.

## What Changes

- Consolidate `domain.ErrUnmarshalBody` and
  `repositories.ErrUnmarshalConfig` into the existing
  `domain.ErrParsingFailed`. Remove both duplicates.
- Update `domain/jmes.go`, `infrastructure/repositories/pg_config_repo.go`,
  and all related tests to reference `domain.ErrParsingFailed`.
- Add `go.temporal.io/sdk` to `crawler/go.mod` so the activity layer
  can call `temporal.NewNonRetryableError`.
- In `application/activities/parser.go`, wrap activity return errors
  with `temporal.NewNonRetryableError` whenever the wrapped chain
  contains `domain.ErrValidation`, `domain.ErrNotFound`, or
  `domain.ErrParsingFailed`. Transient errors (raw DB errors,
  `context.*`) pass through unchanged so Temporal retries them.
- Update activity tests to assert non-retryable wrapping via both
  `temporal.IsNonRetryableError` and `errors.Is(err, domain.Err*)`.
  Add new tests proving transient errors are NOT marked non-retryable.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `parser-activities`: Activities must mark errors wrapping
  `ErrValidation`, `ErrNotFound`, or `ErrParsingFailed` as
  Temporal non-retryable. Transient errors (e.g. raw DB errors,
  context cancellation) remain retryable. Activity tests must
  assert both `temporal.IsNonRetryableError` and the underlying
  domain error.

## Impact

- **Code**
  - `crawler/parser/internal/domain/errors.go` (delete one line)
  - `crawler/parser/internal/domain/jmes.go` (replace sentinel)
  - `crawler/parser/internal/domain/jmes_test.go` (rename + assertion)
  - `crawler/parser/internal/infrastructure/repositories/pg_config_repo.go` (delete one var, replace sentinel)
  - `crawler/parser/internal/infrastructure/repositories/pg_config_repo_integration_test.go` (assertion)
  - `crawler/parser/internal/application/activities/parser.go` (add helper, wrap returns)
  - `crawler/parser/internal/application/activities/parser_test.go` (assertions + new tests)
- **Dependencies**
  - `crawler/go.mod`: add `go.temporal.io/sdk` (latest stable)
- **External behavior**: none beyond retry semantics when activities
  are wired into a Temporal worker.
- **Backwards compatibility**: any out-of-tree caller importing
  `domain.ErrUnmarshalBody` or `repositories.ErrUnmarshalConfig`
  will break. Both are package-private sentinel errors with no
  documented API contract, so we treat this as **BREAKING** for
  internal callers only.
