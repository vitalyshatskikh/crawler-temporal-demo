## Why

The `ParsingConfig` domain model currently encodes URL extraction via a convention: a `ParsingParam` with `Name == PropExternalURL` (`"external_url"`) drives how `ParseSearchPage` extracts advert URLs from a search page. This was a pragmatic early choice but creates two problems: (1) the validation rule that "at least one param must be named `external_url`" is implicit and easy to misconfigure, and (2) there is no mechanism to transform the extracted URL into a canonical form or to populate the `content_url` field on produced documents — a gap left open when `add-content-url-column` deferred parser-side `content_url` population.

## What Changes

- Add three new top-level fields to `ParsingConfig`: `ExternalURLJMESPath` (required for `search_page`), `ExternalURLTemplate` (optional mustache), `ContentURLTemplate` (optional mustache).
- Remove the `PropExternalURL` constant and the implicit "must have a param named `external_url`" validation rule.
- `ParseSearchPage` uses `ExternalURLJMESPath` directly (no more magic param name); `JMESParser` is extended to compile and expose the URL result under the reserved key `_external_url`. Renders `ExternalURLTemplate` and `ContentURLTemplate` with the full parsed snippet as mustache context (using `_external_url` for the raw URL) to produce `ExternalURL` and `ContentURL` on each output document. The `_external_url` key is included in the output `Document.Body` alongside other extracted fields.
- DB schemas (Python migration, Go `schema.sql`) gain three new `TEXT NOT NULL DEFAULT ''` columns on `parsing_configs`.
- `PGConfigRepo` reads and returns the three new columns.
- All Go test fixtures and factory helpers that reference `PropExternalURL` are updated to use `ExternalURLJMESPath`.

## Capabilities

### New Capabilities

- `parser-config`: Adds `ExternalURLJMESPath`, `ExternalURLTemplate`, and `ContentURLTemplate` fields to `ParsingConfig`, replacing the `PropExternalURL`-as-magic-param convention for URL extraction. The mustache context uses `_external_url` as the reserved key for the raw URL; the `_external_url` key is included in the output `Document.Body`. Enables parser-side `content_url` population via mustache template rendering over the parsed snippet.

### Modified Capabilities

- `documents`: The requirement that "parser-produced documents have `content_url = \"\"`" is updated to allow `content_url` to be populated by the parser via `ContentURLTemplate`. The `content_url` column and downloader fallback behavior already exist; this change enables the parser path.

## Impact

- **Go package `crawler/parser/internal/domain`**: `config.go`, `parser.go` change; new `mustache.go` added (simple `RenderTemplate` function, no pre-compiled template cache).
- **Go package `crawler/parser/internal/infrastructure/repositories`**: `pg_config_repo.go` updated.
- **Go package `crawler/parser/db`**: `schema.sql`, `queries/parsing_configs.sql`, and sqlc-generated files updated.
- **Python package `downloader/infrastructure/db/orm`**: `parsing_configs.py` ORM updated.
- **Migration**: New Alembic migration for the three `parsing_configs` columns.
- **Dependencies**: `github.com/cbroglie/mustache` added to `crawler/go.mod`.
- **Test fixtures**: All `domain.PropExternalURL` references in test factories and test cases replaced with `ExternalURLJMESPath`.
- **Body composition**: `ParseSearchPage` output `Document.Body` is constructed from the parsed result map (URLs and params) zipped at snippet index; the `_external_url` key is included in the body alongside other extracted fields.
