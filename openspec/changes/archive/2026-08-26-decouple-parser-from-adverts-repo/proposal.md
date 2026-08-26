## Why

The `ParsingService` in `crawler/parser/internal/domain/` currently depends on `AdvertsRepository` to fetch the `Document` body before parsing. This couples the parser domain to a persistence concern that belongs at the application layer. The Temporal activities (application layer) will fetch documents from PostgreSQL; the parser should only be responsible for JMESPath extraction given a already-loaded `Document`.

Removing this dependency simplifies the service and makes it a pure function: `Document in → parsed Documents out`.

## What Changes

- **BREAKING** `ParsingService` constructor: `NewParsingService(confRepo ConfigRepository)` — drops `AdvertsRepository` parameter
- **BREAKING** `ParseSearchPage(ctx, doc Document)` — replaces `ParseSearchPage(ctx, meta DocumentMeta)`; caller is responsible for fetching the document
- **BREAKING** `ParseAdvertContent(ctx, doc Document)` — replaces `ParseAdvertContent(ctx, meta DocumentMeta)`; caller is responsible for fetching the document
- Move `AdvertsRepository` interface from `domain/repositories.go` to `application/repositories.go` (the application layer owns the document fetch contract)
- Update `.mockery.yml` to reflect new interface location
- Delete `domain/testutil/mock_adverts_repository.go` (no longer needed by parser tests)
- Update all `ParsingService` tests to pass `Document` instead of `DocumentMeta`, dropping all `mockDocRepo` setups

## Capabilities

This is a pure refactor that modifies existing spec-level requirements without adding new ones. The parser's behavior is unchanged — it still accepts a `Document` with a body and returns parsed results. Only the call site responsibility shifts (document fetch moves from parser domain to application layer).

## Impact

- **crawler/parser/internal/domain/parser.go**: `ParsingService` struct, `NewParsingService`, `ParseSearchPage`, `ParseAdvertContent`
- **crawler/parser/internal/domain/repositories.go**: `AdvertsRepository` interface removed from domain (moved to application)
- **crawler/parser/internal/domain/parser_test.go**: all tests using `mockDocRepo` must be updated
- **crawler/parser/.mockery.yml**: `AdvertsRepository` entry moves from `domain` to `application` package
- **crawler/parser/internal/domain/testutil/mock_adverts_repository.go**: deleted
- **crawler/parser/internal/application/repositories.go**: new file with `AdvertsRepository` interface
