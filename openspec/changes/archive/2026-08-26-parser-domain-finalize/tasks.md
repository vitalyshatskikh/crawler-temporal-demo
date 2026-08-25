## 1. Setup and Dependencies

- [x] 1.1 Add `github.com/jmespath/go-jmespath v0.4.0` to `crawler/go.mod` (`go mod edit -require`)
- [x] 1.2 Run `go mod tidy` in `crawler/` to verify all dependencies resolve
- [x] 1.3 Verify `go build ./crawler/parser/...` shows the existing compilation errors (baseline)

## 2. Domain Models — adverts package

- [x] 2.1 Add `SdocIDForURL(url string) SdocID` helper in `adverts/models.go` using `crypto/md5` (hex encode, lowercase)
- [x] 2.2 Add `Validate() error` method to `DocumentMeta` in `adverts/models.go` — validates non-empty SourceID, SdocID, ExternalURL and UpdatedAt >= CreatedAt
- [x] 2.3 Add `ErrValidation = errors.New("validation failed")` sentinel to `parsing/errors.go` (single sentinel for all domain validation)

## 3. Domain Models — parsing package

- [x] 3.1 Add `Validate() error` method to `ParsingConfig` in `parsing/config.go` — validates non-empty SourceID, non-empty DocumentType, len(Params) >= 1, and at least one ParsingParam has Name == PropExternalURL
- [x] 3.2 Add `ErrUnmarshalBody` sentinel error to `parsing/errors.go`

## 4. ConfigRepository Interface

- [x] 4.1 Change `ConfigRepository.GetConfig` signature in `crawler/parser/internal/domain/repositories.go` from `(ctx context.Context, id SourceID) (ParsingConfig, error)` to `(ctx context.Context, sourceID SourceID, docType DocumentType) (ParsingConfig, error)` — `AdvertsRepository.GetDocument` and `SaveDocument` are also defined in the same file; the interface change only affects `ConfigRepository`.

## 5. JMESParser — parsing/jmes.go

- [x] 5.1 In `JMESParser.Parse(ctx, body []byte)`, add `json.Unmarshal(body, &data)` before running JMESPath expressions; return `ErrUnmarshalBody` on failure
- [x] 5.2 Change `prop.expr.Search(body)` to `prop.expr.Search(data)` (search the unmarshaled data, not the raw string)
- [x] 5.3 Rename `jmesProp.nvl` to `jmesProp.defaultVal` throughout `jmes.go` (lines 12, 30, 53) for readability
- [x] 5.4 Verify all error wraps use the correct sentinel (`ErrParsingFailed`, `ErrUnmarshalBody`, `ErrValidation`)

## 6. ParsingService — parsing/parser.go

- [x] 6.1 Fix `NewParsingService(confRepo ConfigRepository, docRepo Repository) (*ParsingService, error)` — save both repos, validate non-nil, return `ErrValidation` on nil, initialize `jmesParsers` map with `make(map[parserKey]*JMESParser)`
- [x] 6.2 Fix `ParseAdvertContent` — remove undefined `i` variable and `[%d]` format specifier (single document, no index)
- [x] 6.3 In `ParseSearchPage`: call `meta.Validate()` at entry, wrap as `ErrValidation`
- [x] 6.4 In `ParseSearchPage`: compute `snipMeta.SdocID = SdocIDForURL(snippetURL)` instead of reusing parent SdocID
- [x] 6.5 In `ParseSearchPage`: use single `now := time.Now()` for both `CreatedAt` and `UpdatedAt`
- [x] 6.6 In `ParseAdvertContent`: call `meta.Validate()` at entry, wrap as `ErrValidation`
- [x] 6.7 In `ParseAdvertContent`: use single `now := time.Now()` for both timestamps
- [x] 6.8 In `getJMESParser`: add `golang.org/x/sync/singleflight.Group` field to `ParsingService` struct; use `sf.Group.Do` keyed by `parserKey{sourceID, docType}` for slow-path config fetch; add `select { case <-ctx.Done(): return nil, ctx.Err(); default: }` at the start of the slow path before `sf.Group.Do`
- [x] 6.9 In `getJMESParser`: after fetching config, call `cnf.Validate()`, return `ErrValidation` on failure
- [x] 6.10 Verify `go build ./crawler/parser/...` compiles clean

## 7. Mockery Setup

- [x] 7.1 Create `crawler/parser/.mockery.yml` with `dir: "{{.InterfaceDir}}/testutil"`, `filename: "mock_{{.InterfaceName | snakecase}}.go"`, `pkgname: testutil`, `formatter: goimports`
- [x] 7.2 Add `MockConfigRepository` generation: `packages: github.com/.../crawler/parser/internal/domain` + `interfaces: ConfigRepository`
- [x] 7.3 Add `MockAdvertsRepository` generation: `packages: github.com/.../crawler/parser/internal/domain` + `interfaces: Repository`
- [x] 7.4 Run `mockery --dir crawler/parser/` to generate both mocks
- [x] 7.5 Verify generated files compile

## 8. Test Factories

- [x] 8.1 Create `crawler/parser/internal/domain/testutil/factories.go` with:
  - `DocumentMetaFactory()` — faker-based, optional fields (use for "any valid meta" cases)
  - `MustDocumentMeta(url, sdocID SdocID, sourceID SourceID) DocumentMeta` — deterministic, for tests asserting exact SdocID or URL values
  - `DocumentFactory()`, `SdocIDFactory()`
  - `ConfigFactory()` — faker-based, valid defaults (use for "any valid config" cases)
  - `MustSearchPageConfig(sourceID SourceID, params []ParsingParam) ParsingConfig` — deterministic, for tests asserting exact config values
  - `MustAdvertConfig(sourceID SourceID, params []ParsingParam) ParsingConfig` — deterministic
  - `ParamFactory()`, `ValidSearchPageConfigFactory()`, `ValidAdvertConfigFactory()`

## 9. Unit Tests — adverts/models

- [x] 9.1 `SdocIDForURL_WhenSameURL_ThenSameID` — assert stability across two calls
- [x] 9.2 `SdocIDForURL_WhenDifferentURLs_ThenDifferentIDs` — assert inequality
- [x] 9.3 `SdocIDForURL_WhenURLsDifferOnlyInTrailingSlash_ThenSameID` — `https://x.com/a/` vs `https://x.com/a`
- [x] 9.4 `SdocIDForURL_WhenURLsDifferOnlyInQueryOrder_ThenSameID` — `?a=1&b=2` vs `?b=2&a=1`
- [x] 9.5 `SdocIDForURL_WhenURLsDifferOnlyInFragment_ThenSameID` — `https://x.com/a#s1` vs `https://x.com/a`
- [x] 9.6 `SdocIDForURL_WhenInvalidURL_ThenErrValidation` — non-URL string returns error wrapping `ErrValidation`
- [x] 9.7 `DocumentMeta.Validate_WhenValid_ThenNil` — valid meta passes
- [x] 9.8 `DocumentMeta.Validate_WhenEmptySourceID_ThenErrValidation`
- [x] 9.9 `DocumentMeta.Validate_WhenEmptySdocID_ThenErrValidation`
- [x] 9.10 `DocumentMeta.Validate_WhenEmptyExternalURL_ThenErrValidation`
- [x] 9.11 `DocumentMeta.Validate_WhenUpdatedAtBeforeCreatedAt_ThenErrValidation`

## 10. Unit Tests — parsing/jmes

- [x] 10.1 `NewJMESParser_WhenInvalidJMESPath_ThenErrValidation` — table-driven: bad expressions
- [x] 10.2 `NewJMESParser_WhenEmptyConfig_ThenErrValidation`
- [x] 10.3 `JMESParser.Parse_WhenValidJSONWithSliceResult_ThenCorrectMap` — body `{"urls": ["a","b"]}`, expr `urls`, result `["a","b"]`
- [x] 10.4 `JMESParser.Parse_WhenNonJSONBody_ThenErrUnmarshalBody`
- [x] 10.5 `JMESParser.Parse_WhenMissingPath_ThenUsesDefault` — config with default `"none"`
- [x] 10.6 `JMESParser.Parse_WhenNilJMESResult_ThenUsesDefault`
- [x] 10.7 `JMESParser.Parse_WhenScalarResult_ThenWrappedInSlice`
- [x] 10.8 `JMESParser.Parse_WhenContextCancelled_ThenCtxErr`

## 11. Unit Tests — parsing/parser

- [x] 11.1 `Service_New_WhenNilConfRepo_ThenErrValidation`
- [x] 11.2 `Service_New_WhenNilDocRepo_ThenErrValidation`
- [x] 11.3 `Service_New_WhenBothValid_ThenNonNilService`
- [x] 11.4 `Service_ParseSearchPage_WhenInvalidMeta_ThenErrValidation`
- [x] 11.5 `Service_ParseSearchPage_WhenDocRepoError_ThenWrapsErr`
- [x] 11.6 `Service_ParseSearchPage_WhenConfigMissing_ThenErrNotFound`
- [x] 11.7 `Service_ParseSearchPage_WhenNoExternalURLParam_ThenErrValidation`
- [x] 11.8 `Service_ParseSearchPage_WhenValidInput_ThenUniqueSdocIDsAndCorrectType` — assert each doc has unique SdocID (md5 of its URL) and Type == DocumentTypeSurfedAdvert
- [x] 11.9 `Service_ParseSearchPage_WhenEmptySnippets_ThenEmptySlice`
- [x] 11.10 `Service_ParseAdvertContent_WhenInvalidMeta_ThenErrValidation`
- [x] 11.11 `Service_ParseAdvertContent_WhenDocRepoError_ThenWrapsErr`
- [x] 11.12 `Service_ParseAdvertContent_WhenConfigNotFound_ThenErrNotFound`
- [x] 11.13 `Service_ParseAdvertContent_WhenValid_ThenCorrectTypeAndBody`
- [x] 11.14 `Service_ParseSearchPage_WhenWrongMetaType_ThenErrValidation` — `meta.Type != DocumentTypeSearchPage` returns `ErrValidation`
- [x] 11.15 `Service_ParseAdvertContent_WhenWrongMetaType_ThenErrValidation` — `meta.Type != DocumentTypeDownloadedAdvert` returns `ErrValidation`

## 12. Verification

- [x] 12.1 `go vet ./crawler/parser/...` — clean
- [x] 12.2 `gofmt -l crawler/parser/` — no files listed
- [x] 12.3 `go test -race ./crawler/parser/internal/domain/...` — all tests green
- [x] 12.4 `mockery --dir crawler/parser/` — generates without error

## Notes

The following test cases were added during implementation beyond the initial task list:

- **adverts/models**: `TestSdocIDForURL_WhenUppercaseScheme_ThenLowercased`, `TestSdocIDForURL_WhenUppercaseHost_ThenLowercased`, `TestSdocIDForURL_WhenEmptyString_ThenErrValidation`, `TestDocumentMeta_WhenZeroCreatedAt_ThenErrValidation`, `TestDocumentMeta_WhenZeroUpdatedAt_ThenErrValidation`, `TestDocumentMeta_WhenBothZeroTimestamps_ThenErrValidation`, `TestDocumentMeta_WhenCreatedAtEqualsUpdatedAt_ThenValid`
- **parsing/jmes**: `TestJMESParser_WhenEmptyBody_ThenErrUnmarshalBody`, `TestJMESParser_WhenBooleanResult_ThenWrappedInSlice`, `TestJMESParser_WhenNumericResult_ThenWrappedInSlice`
- **parsing/parser**: `TestService_ParseSearchPage_WhenPropertyHasFewerValuesThanURLs_ThenNoPanicAndSkipsProperty`, `TestService_ParseSearchPage_WhenSecondCall_ThenConfigRepoNotCalledAgain`, `TestService_ParseSearchPage_WhenConcurrentFirstLoad_ThenConfigLoadedOnce`, `TestService_GetJMESParser_WhenCtxCancelledBefore_ThenReturnsCtxErr`
