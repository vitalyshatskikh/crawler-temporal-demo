## Why

The parser domain layer (`crawler/parser/internal/domain/`) is in draft state with multiple bugs that prevent it from compiling and working correctly. Beyond the build errors, the draft has missing validation, incorrect snippet identity (SdocID reuse causing PK collisions), no per-(source, docType) config, and no unit tests. The domain must be finalized before the application (Temporal activities) and infrastructure (PG repos) layers can be built.

## What Changes

- **Bug fixes**: `NewParsingService()` discards its arguments (returns empty struct); `undefined: i` in `ParseAdvertContent`; `JMESParser.Parse` passes raw string to `jmespath.Search` instead of unmarshaled JSON; missing `github.com/jmespath/go-jmespath` dependency.
- **Per-(source, docType) config**: `ConfigRepository.GetConfig` signature changes from `(ctx, sourceID)` to `(ctx, sourceID, docType)` — aligns with the cache key already including `docType`.
- **Unique snippet identity**: `ParseSearchPage` now generates each snippet's `SdocID` as `md5(external_url)` (via `SdocIDForURL` helper) instead of reusing the parent search page's SdocID.
- **Input validation**: `DocumentMeta.Validate()` (non-empty SourceID, SdocID, ExternalURL, non-zero CreatedAt and UpdatedAt; UpdatedAt ≥ CreatedAt) and `ParsingConfig.Validate()` (non-empty SourceID, DocumentType, ≥1 ParsingParam; for `DocumentTypeSearchPage`, at least one param must be `PropExternalURL`).
- **Concurrency**: `getJMESParser` slow path uses `singleflight.Group` to collapse concurrent first-time config fetches.
- **JMESPath body handling**: `Parse` unmarshals `body []byte` into `any` before running compiled JMESPath expressions; proper error mapping (`ErrUnmarshalBody`, `ErrParsingFailed`, `ErrValidation`).
- **Single timestamp**: `CreatedAt` and `UpdatedAt` share one `time.Now()` call.
- **Unit tests**: Comprehensive table-driven tests for `JMESParser`, `ParsingService.ParseSearchPage`, `ParsingService.ParseAdvertContent`, models validation, and mock contracts; generated via mockery.

## Capabilities

### NewParsingService Capabilities

- `parser-domain`: Domain layer for the Go parser service — defines document models (`SdocID`, `SourceID`, `DocumentMeta`, `Document`, `DocumentType`), parsing config (`ParsingConfig`, `ParsingParam`), JMESPath-based extraction (`JMESParser`), and the parser service (`ParsingService`) with its two operations. Covers the domain logic, validation rules, error taxonomy, and test contracts.

### Modified Capabilities

- _(none — no existing spec-level behavior is being changed)_

## Impact

- **Code**: `crawler/parser/internal/domain/` and `crawler/parser/internal/domain/` — all domain files modified or created fresh.
- **API**: `ConfigRepository.GetConfig(ctx, sourceID)` → `GetConfig(ctx, sourceID, docType)` (breaking change to interface; concrete PG implementation deferred to infra change).
- **Dependencies**: Add `github.com/jmespath/go-jmespath v0.4.0` to `crawler/go.mod`.
- **Tests**: NewParsingService test files in `crawler/parser/internal/domain/` and `crawler/parser/internal/domain/`; new `testutil` packages with mockery-generated mocks and factory helpers.
- **Out of scope**: Application layer (Temporal activities, worker bootstrap) and infrastructure layer (PG repos, sqlc) — each deferred to its own change.
