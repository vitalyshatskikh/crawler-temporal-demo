## 1. Consolidate parse-error sentinels

- [x] 1.1 Remove `ErrUnmarshalBody` from `crawler/parser/internal/domain/errors.go`
- [x] 1.2 In `crawler/parser/internal/domain/jmes.go:42`, replace `ErrUnmarshalBody` with `ErrParsingFailed`
- [x] 1.3 Remove the package-level `var ErrUnmarshalConfig` from `crawler/parser/internal/infrastructure/repositories/pg_config_repo.go:18`
- [x] 1.4 In `crawler/parser/internal/infrastructure/repositories/pg_config_repo.go:42`, wrap the JSON unmarshal failure with `domain.ErrParsingFailed` instead of `ErrUnmarshalConfig`
- [x] 1.5 Rename `TestJMESParser_WhenNonJSONBody_ThenErrUnmarshalBody` to `TestJMESParser_WhenNonJSONBody_ThenErrParsingFailed` and switch its assertion to `domain.ErrParsingFailed` in `crawler/parser/internal/domain/jmes_test.go`
- [x] 1.6 Rename `TestJMESParser_WhenEmptyBody_ThenErrUnmarshalBody` to `TestJMESParser_WhenEmptyBody_ThenErrParsingFailed` and switch its assertion to `domain.ErrParsingFailed` in `crawler/parser/internal/domain/jmes_test.go`
- [x] 1.7 In `crawler/parser/internal/infrastructure/repositories/pg_config_repo_integration_test.go:99`, replace `ErrUnmarshalConfig` with `domain.ErrParsingFailed`
- [x] 1.8 Run `go vet ./...` and `go test ./crawler/parser/internal/...` from `crawler/` to confirm green

## 2. Add Temporal SDK dependency

- [x] 2.1 Add `go.temporal.io/sdk` to `crawler/go.mod` direct requires
- [x] 2.2 Run `go mod tidy` from `crawler/` to resolve transitive deps

## 3. Mark domain errors non-retryable in activities

- [x] 3.1 In `crawler/parser/internal/application/activities/parser.go`, add imports `errors` and `go.temporal.io/sdk/temporal`
- [x] 3.2 Add `const errTypeParsing = "ParsingError"` at package scope in `activities/parser.go`. Then add a `wrapErr(err error) error` helper in the same package that wraps with `temporal.NewNonRetryableError(message, errTypeParsing, err)` when `errors.Is` matches `domain.ErrValidation`, `domain.ErrNotFound`, or `domain.ErrParsingFailed`, and passes through otherwise
- [x] 3.3 In `ParseSearchPage`, funnel every error return (validation, get document, parse, save) through `wrapErr`
- [x] 3.4 In `ParseAdvertContent`, funnel every error return (validation, get document, parse, save) through `wrapErr`
- [x] 3.5 Add a short comment on the helper reminding callers not to re-wrap its result

## 4. Update activity tests

- [x] 4.1 Add `go.temporal.io/sdk/temporal` to imports in `crawler/parser/internal/application/activities/parser_test.go`
- [x] 4.2 For every existing test that asserts `domain.ErrValidation` or `domain.ErrNotFound`, add `assert.True(t, temporal.IsNonRetryableError(err))`
- [x] 4.3 Add `TestParser_ParseSearchPage_WhenRepoGetFails_ThenRetryable` asserting `!temporal.IsNonRetryableError(err)` for a raw `errors.New("repo get failed")`
- [x] 4.4 Add `TestParser_ParseAdvertContent_WhenRepoGetFails_ThenRetryable` with the same shape
- [x] 4.5 Add `TestParser_ParseSearchPage_WhenSaveFails_ThenRetryable` asserting the save error is NOT non-retryable
- [x] 4.6 Add `TestParser_ParseAdvertContent_WhenSaveFails_ThenRetryable` with the same shape

## 5. Verification

- [x] 5.1 Run `go mod tidy` from `crawler/`
- [x] 5.2 Run `go vet ./...` from `crawler/`
- [x] 5.3 Run `go test ./crawler/parser/internal/domain/...` from `crawler/`
- [x] 5.4 Run `go test ./crawler/parser/internal/application/...` from `crawler/`
- [x] 5.5 Run `go test -tags integration ./crawler/parser/internal/infrastructure/...` from `crawler/` (requires live Postgres; skip if not available)
