# Tasks

## 1. Rewrite `crawler/parser/internal/application/activities/parser.go`

- [x] 1.1 Change `Parser.repo` field type from `*domain.AdvertsRepository` to `application.AdvertsRepository`.
- [x] 1.2 Rewrite `NewParser(svc *domain.ParsingService, repo application.AdvertsRepository) (*Parser, error)`; return `domain.ErrValidation` when either dep is nil.
- [x] 1.3 Implement `ParseSearchPage(ctx, meta)`:
  - validate `meta.Type == DocumentTypeSearchPage`; return `ErrValidation` if not;
  - fetch document from repo using `meta.SdocID, meta.SourceID, meta.Type`;
  - pass to `svc.ParseSearchPage`;
  - save each result document and return the corresponding `DocumentMeta` slice.
- [x] 1.4 Implement `ParseAdvertContent(ctx, meta)`:
  - validate `meta.Type == DocumentTypeDownloadedAdvert`; return `ErrValidation` if not;
  - fetch document from repo using `meta.SdocID, meta.SourceID, meta.Type`;
  - pass to `svc.ParseAdvertContent`;
  - save the parsed document; return nil.
- [x] 1.5 Wrap every downstream error with `fmt.Errorf("... %s/%s: %w", sourceID, sdocID, err)`.

## 2. Add activity-level test factories

- [x] 2.1 Create `crawler/parser/internal/application/testutil/factories.go` with:
  - `MustSearchPageDocument(sourceID, url string) domain.Document` (type = `DocumentTypeSearchPage`);
  - `MustDownloadedAdvertDocument(sourceID domain.SourceID, sdocID domain.SdocID, url string) domain.Document` (type = `DocumentTypeDownloadedAdvert`);
  - `MustSurfedAdvertMeta(sourceID domain.SourceID, sdocID domain.SdocID, url string) domain.DocumentMeta` (type = `DocumentTypeSurfedAdvert`).

## 3. Add unit tests

- [x] 3.1 Create `crawler/parser/internal/application/activities/parser_test.go` (package `activities_test`) importing:
  - `domain`, `domain/testutil` (factories + `MockConfigRepository`),
  - `application`, `application/testutil` (`MockAdvertsRepository`, new factories).
- [x] 3.2 Constructor tests:
  - `TestNewParser_WhenNilSvc_ThenErrValidation`
  - `TestNewParser_WhenNilRepo_ThenErrValidation`
  - `TestNewParser_WhenValid_ThenNonNilParser`
- [x] 3.3 `ParseSearchPage` tests:
  - `TestParser_ParseSearchPage_WhenInvalidMetaType_ThenErrValidation`
  - `TestParser_ParseSearchPage_WhenRepoGetFails_ThenWrappedError`
  - `TestParser_ParseSearchPage_WhenServiceFails_ThenWrappedError`
  - `TestParser_ParseSearchPage_WhenValid_ThenSavesAllAndReturnsMetas`
  - `TestParser_ParseSearchPage_WhenEmptyURLs_ThenReturnsNonNilEmptySliceAndNoSaves`
  - `TestParser_ParseSearchPage_WhenSaveFails_ThenReturnsErrorAndStops`
- [x] 3.4 `ParseAdvertContent` tests:
  - `TestParser_ParseAdvertContent_WhenInvalidMetaType_ThenErrValidation`
  - `TestParser_ParseAdvertContent_WhenRepoGetFails_ThenWrappedError`
  - `TestParser_ParseAdvertContent_WhenServiceFails_ThenWrappedError`
  - `TestParser_ParseAdvertContent_WhenSaveFails_ThenWrappedError`
  - `TestParser_ParseAdvertContent_WhenValid_ThenSavesAndReturnsNil`

## 4. Verification

- [x] 4.1 Run `gofmt -l crawler/parser/...` — no diff expected.
- [x] 4.2 Run `go vet ./crawler/parser/...` — clean.
- [x] 4.3 Run `go test ./crawler/parser/...` — all tests pass.
