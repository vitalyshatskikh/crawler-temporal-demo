## Why

The current `documents` table only stores `external_url` — the canonical URL of the document source. Some document sources require fetching content from a different URL (e.g., CDN, S3 presigned URL, mirror) than the canonical URL shown in search results or listings. Adding an optional `content_url` column allows this override to be stored and used by the downloader, while keeping `external_url` as the canonical/reference URL.

## What Changes

- **documents table**: Add nullable `content_url TEXT NOT NULL DEFAULT ''` column. Empty string means "use external_url".
- **Python models** (surfer): Add `content_url: str = pydantic.Field(default="")` to `DocumentMeta`. No min_length constraint — empty is the valid sentinel.
- **Go models** (parser): Add `ContentURL string` to `DocumentMeta`. No validation change — empty is valid.
- **Python mappers**: Add `content_url` mapping in `document_to_meta` and `document_to_orm`.
- **Python repository**: Add `content_url` to upsert conflict resolution set.
- **Python downloader**: Use `content_url or external_url` as the fetch URL; log the actual URL fetched.
- **Go sqlc**: Add `content_url` to schema, queries, and regenerated models.
- **Go repository**: Add `ContentURL` mapping in `GetDocument` and `SaveDocument`.
- **Tests**: Update all affected test factories and cases.
- **OpenSpec**: Add delta spec to `documents` capability.

## Capabilities

### New Capabilities

*(none — this change does not introduce a new capability, only modifies an existing one)*

### Modified Capabilities

- `documents`: Add `content_url` column and override behavior in downloader. Three new requirements: `content_url` column, `content_url` field on `DocumentMeta`, and downloader fallback logic.

## Impact

- **DB**: `documents` table schema changes. `external_url` remains required; `content_url` is optional with `""` default.
- **Python stack**: `DocumentMeta`, `DocumentORM`, mappers (`shared/py/db/mappers.py`), `PGDocumentRepository`, `WebDownloader` activity.
- **Go stack**: `DocumentMeta` in parser domain, `queries/documents.sql`, sqlc-generated `models.go` and `documents.sql.go`, `PGAdvertsRepo`.
- **Parser flow unchanged**: Go parser writes only `ExternalURL`; `ContentURL` stays empty for parser-produced documents.
- **Workflow history**: Running workflows serialized with the old `DocumentMeta` shape (without `content_url`) may fail on replay. Treat workflow history as ephemeral — wipe DB and restart workers when deploying.
