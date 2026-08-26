## 1. Domain Layer — Remove docRepo Dependency

- [x] 1.1 In `parser.go`: remove `docRepo AdvertsRepository` field from `ParsingService` struct
- [x] 1.2 In `parser.go`: update `NewParsingService(confRepo ConfigRepository)` — drop `docRepo` parameter and `docRepo == nil` validation
- [x] 1.3 In `parser.go`: update `ParseSearchPage(ctx, meta DocumentMeta)` → `ParseSearchPage(ctx, doc Document)` — replace all `meta.X` with `doc.X`, remove `s.docRepo.GetDocument(...)` call, use `doc.Body` directly; `doc.Validate()` resolves to embedded `DocumentMeta.Validate()` via Go method promotion (no logic change)
- [x] 1.4 In `parser.go`: update `ParseAdvertContent(ctx, meta DocumentMeta)` → `ParseAdvertContent(ctx, doc Document)` — same treatment; `doc.Validate()` resolves to embedded `DocumentMeta.Validate()` via Go method promotion (no logic change)
- [x] 1.5 In `repositories.go`: remove `AdvertsRepository` interface (moved to application layer in step 2)

## 2. Application Layer — Add AdvertsRepository Interface

- [x] 2.1 Create `crawler/parser/internal/application/repositories.go` with `//go:generate go tool mockery` and `AdvertsRepository` interface (GetDocument + SaveDocument methods, same signatures as before)
- [x] 2.2 Update `crawler/parser/.mockery.yml` — remove `AdvertsRepository` from `domain` package entry; add new `application` package entry with `AdvertsRepository` interface (same dir/template pattern as domain)

## 3. Test Utilities — Remove mock_adverts_repository

- [x] 3.1 Delete `crawler/parser/internal/domain/testutil/mock_adverts_repository.go`

## 4. Tests — Update parser_test.go

- [x] 4.1 Remove `TestNewParsingServiceWhenNilDocRepo_ThenErrValidation` (no longer applicable)
- [x] 4.2 Remove `TestService_ParseSearchPage_WhenDocRepoError_ThenWrapsErr`
- [x] 4.3 Remove `TestService_ParseAdvertContent_WhenDocRepoError_ThenWrapsErr`
- [x] 4.4 Update all `NewParsingService(confRepo, docRepo)` calls → `NewParsingService(confRepo)` (affects ~17 remaining tests)
- [x] 4.5 Update all test helper setups that create `mockDocRepo` — remove `NewMockAdvertsRepository(t)` and `mockDocRepo` variable declarations and `mockDocRepo := testutil.NewMockAdvertsRepository(t)` assignments
- [x] 4.6 For each test that calls `svc.ParseXxx(ctx, meta)`: construct `doc domain.Document` using the existing `meta` values (set `doc.DocumentMeta = meta`) and add the `Body` field that was previously returned by `mockDocRepo.GetDocument(...)`
- [x] 4.7 Remove all `mockDocRepo.EXPECT().GetDocument(...)` call chains from test setups
- [x] 4.8 Update assertions that reference `meta.SdocID` or `meta.SourceID` → use `doc.SdocID`, `doc.SourceID` (values are identical; just the accessor path changes)
- [x] 4.9 Verify `TestService_ParseSearchPage_WhenCtxCancelled_ThenReturnsCtxErr` — ctx cancellation still works via `getJMESParser`'s pre-flight select; no `mockDocRepo` setup was needed in this test

## 5. Verification

- [x] 5.1 `go vet ./crawler/parser/...` — clean
- [x] 5.2 `gofmt -l crawler/parser/` — no files listed
- [x] 5.3 `go test -race ./crawler/parser/internal/domain/...` — all tests green
- [x] 5.4 `go build ./crawler/parser/...` — compiles clean
- [x] 5.5 `mockery --dir crawler/parser/internal/application/` — generates `mock_adverts_repository.go` in application/testutil
