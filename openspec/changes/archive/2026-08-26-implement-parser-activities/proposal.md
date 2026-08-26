## Why

The Temporal parser service in `crawler/parser/internal/application/activities/parser.go` ships with two no-op activity stubs (`ParseSearchPage`, `ParseAdvertContent`) and an empty `NewParser` constructor. The downstream `ParsingService`, `AdvertsRepository` interface, generated mocks, and domain factories are all in place but unused. Without a working implementation the parser worker cannot turn downloaded search pages into surfaced adverts or downloaded adverts into parsed adverts — the pipeline halts here. The field type on `Parser` is also broken (`*domain.AdvertsRepository`, no such concrete type), so even the current stub does not compile against the intended interface.

## What Changes

- Rewrite `ParseSearchPage(ctx, meta)` to fetch the parent search page from `AdvertsRepository.GetDocument`, call `ParsingService.ParseSearchPage`, persist each resulting surfed advert via `AdvertsRepository.SaveDocument`, and return the saved `DocumentMeta` slice so the workflow can fan out child workflows.
- Rewrite `ParseAdvertContent(ctx, meta)` to fetch the downloaded advert document, call `ParsingService.ParseAdvertContent`, persist the parsed advert via `SaveDocument`, and return nil on success.
- Replace `NewParser()` with `NewParser(svc *domain.ParsingService, repo AdvertsRepository) (*Parser, error)` that validates both dependencies and returns `domain.ErrValidation` when either is nil (mirrors `NewParsingService`).
- Fix the broken `Parser.repo` field type (`*domain.AdvertsRepository` → `application.AdvertsRepository`).
- Add `internal/application/testutil/factories.go` with activity-specific builders (`MustSearchPageDocument`, `MustDownloadedAdvertDocument`, `MustSurfedAdvertMeta`).
- Add `internal/application/activities/parser_test.go` covering constructor validation and every error/happy path of both activities, using `MockAdvertsRepository` + real `ParsingService` wired with `MockConfigRepository`.

No **BREAKING** changes: the only existing caller (`cmd/parser/main.go` is still a stub) has nothing to update. Public method signatures of `ParseSearchPage` and `ParseAdvertContent` are unchanged.

## Capabilities

### New Capabilities

- `parser-activities`: Temporal activities exposed by the parser worker — `ParseSearchPage` and `ParseAdvertContent` — that fetch, parse, and persist documents via the existing domain service and adverts repository.

### Modified Capabilities

(none — first spec for this capability)

## Impact

- Code:
  - rewrite `crawler/parser/internal/application/activities/parser.go`
  - new `crawler/parser/internal/application/testutil/factories.go`
  - new `crawler/parser/internal/application/activities/parser_test.go`
- API: `NewParser` signature changes from `NewParser() (*Parser, error)` to `NewParser(*domain.ParsingService, AdvertsRepository) (*Parser, error)`. No existing callers.
- Dependencies: none new — `testify`, `MockConfigRepository`, `MockAdvertsRepository`, `go-faker` are all already vendored.
- Out of scope: Temporal SDK registration (`activity.Defn`), `cmd/parser/main.go` worker bootstrap, repository interface expansion, `ParsingService` changes.
