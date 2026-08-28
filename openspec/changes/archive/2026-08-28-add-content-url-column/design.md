## Context

The `documents` table currently has `external_url TEXT NOT NULL`. The Python downloader (`WebDownloader`) fetches from `doc_meta.external_url`. The Go parser writes only `external_url` when creating documents — there is no mechanism to store or use an alternate fetch URL. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Add `content_url` column to `documents` table with `""` as the default sentinel.
- Surface `content_url` on `DocumentMeta` in both Python and Go stacks.
- Update `WebDownloader` to use `content_url` when set, falling back to `external_url`.
- Preserve all existing behavior when `content_url` is empty.

**Non-Goals:**
- Changing the parser to populate `content_url` — that is deferred to a future change.
- Removing or renaming `external_url` — it remains the canonical URL.
- Adding validation that `content_url` must be a valid URL when non-empty — deferred.

## Decisions

### Column position and nullability

**Decision:** `content_url TEXT NOT NULL DEFAULT ''` placed after `external_url` in both schemas.

**Rationale:** Using `NOT NULL DEFAULT ''` (option b from clarification) avoids nullable-column pitfalls in the ORM and Go sqlc generated code. Empty string is the unambiguous sentinel meaning "no override". Placing it after `external_url` keeps related URL columns adjacent.

**Alternatives considered:**
- Nullable (`content_url TEXT`): Introduces NULL-checking overhead in mappers and Go code; Python ORMs treat NULL and empty string differently in queries.
- Separate table: Over-engineered for a single optional field.

### Python model field

**Decision:** `content_url: str = pydantic.Field(default="")` on `DocumentMeta`. No `min_length` constraint.

**Rationale:** Pydantic's default `""` means callers can omit the field. No `min_length=1` since empty is valid. The Go parser (which does not set this field yet) will produce documents with `ContentURL = ""` (Go zero value), which serializes as `""` in JSON and round-trips correctly.

### Python downloader fallback

**Decision:** `actual_url = doc_meta.content_url or doc_meta.external_url` — Python's `or` treats empty string as falsy, so `"" or external_url` returns `external_url`.

**Rationale:** Idiomatic, single line. No explicit `if` needed.

**Alternatives considered:**
- Explicit `if doc_meta.content_url:` — more verbose, same semantics.
- Ternary — more verbose.

### Upsert behavior for content_url

**Decision:** `content_url` is included in the upsert `set_` dict — the input document's `content_url` always overwrites any stored value.

**Rationale:** Consistent with `external_url` and `body`, which are also unconditionally overwritten on re-download. If a subsequent download sets `content_url=""` (empty sentinel), the stored override is cleared. This is intentional — the input document is the source of truth for the current fetch URL.

### json tag for Go ContentURL field

**Decision:** `ContentURL string \`json:"content_url"\`` in `DocumentMeta`.

**Rationale:** Matches the existing convention for `ExternalURL` (`json:"external_url"`). sqlc generates `ContentUrl` (non-acronym) from the DB column; the domain struct uses `ContentURL` (acronym) for consistency with the existing model.

### Go sqlc regeneration

**Decision:** Add `content_url` to `documents.sql`, then regenerate `models.go` and `documents.sql.go` via `sqlc generate`.

**Rationale:** sqlc is the source of truth for the generated Go DB layer. Editing generated files directly would be overwritten on next `sqlc generate`. The `db/gen/` files are committed to the repo, so both schema change and `sqlc generate` must be applied in the same commit.

## Risks / Trade-offs

[Risk] **Running workflow history incompatibility**
→ The change adds `content_url` to `DocumentMeta` serialized over Temporal. Running workflows serialized before this change will fail on replay. Treat workflow history as ephemeral: wipe Temporal namespace and restart workers when deploying. Not a blocker for this project.

[Risk] **Go parser does not yet populate `content_url`**
→ Parser-produced documents will have `ContentURL = ""`. Downloader will fall back to `external_url`, preserving current behavior. This is the agreed deferral.

[Risk] **Integration tests use raw INSERT with explicit column lists**
→ Integration tests in `pg_adverts_repo_integration_test.go` and `test_pg_document_repo_integration.py` must add `content_url` to INSERT statements. Missed updates cause column-not-found errors. Mitigated by updating test helper functions first.

## Migration Plan

1. Edit `migrations/versions/0004_create_documents.py` and `parser/db/schema.sql`.
2. Wipe DB (per user confirmation — `docker compose down -v` or equivalent).
3. Run `poetry run alembic upgrade head` (Python) and `sqlc generate` (Go) to apply schema.
4. Update Python models, ORM, mappers, repository, downloader.
5. Update Go models, queries, repository.
6. Update test factories and helpers.
7. Run `ruff check --fix . && mypy surfer downloader` (Python), `go fmt && go vet && go test ./...` (Go).
8. Restart all workers.

No data migration needed — existing documents will have `content_url = ''` after schema reapply.

## Open Questions

1. **Parser `content_url` population**: The Go parser currently writes only `external_url` and never populates `content_url`. If a future parser needs to extract a CDN/mirror URL from a search result and surface it as `content_url`, a separate change will be needed to add that field to the parser's output. This change lays the groundwork but does not implement that path.
