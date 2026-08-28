## 1. DB Schema

- [x] 1.1 Edit `migrations/versions/0004_create_documents.py` — add `content_url TEXT NOT NULL DEFAULT ''` column after `external_url`
- [x] 1.2 Edit `parser/db/schema.sql` — add `content_url TEXT NOT NULL DEFAULT ''` column after `external_url`
- [ ] 1.3 Wipe DB and re-apply migrations: `docker compose down -v` then `make crawler-migrate` (Python alembic) and `sqlc generate` (Go)
- [ ] 1.4 Verify DB schema: connect to postgres and confirm `content_url` column exists with correct default

## 2. Python Models (surfer package)

- [x] 2.1 Edit `surfer/domain/adverts/models.py` — add `content_url: str = pydantic.Field(default="")` to `DocumentMeta`
- [x] 2.2 Edit `shared/py/db/orm/documents.py` — add `content_url: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text, server_default=text(""))` to `DocumentORM`
- [x] 2.3 Edit `shared/py/db/mappers.py` — add `content_url=row.content_url` in `document_to_meta`, add `"content_url": doc.content_url` in `document_to_orm`
- [x] 2.4 Run `poetry run ruff check --fix . && poetry run mypy surfer shared` — fix any type errors

## 3. Python Downloader

- [x] 3.1 Edit `downloader/infrastructure/repositories/document_repo.py` — add `"content_url": stmt.excluded.content_url` to upsert `set_` dict
- [x] 3.2 Edit `downloader/application/activities/web_downloader.py` — replace `url=doc_meta.external_url` with `actual_url = doc_meta.content_url or doc_meta.external_url; url=actual_url`. Update log lines to use `actual_url` instead of `doc_meta.external_url`
- [x] 3.3 Run `poetry run ruff check --fix . && poetry run mypy downloader` — fix any type errors

## 4. Go sqlc + Parser

- [x] 4.1 Edit `parser/db/queries/documents.sql` — add `content_url` to `GetDocument` SELECT, add to `UpsertDocument` INSERT column list and ON CONFLICT SET
- [x] 4.2 Run `sqlc generate` from `crawler/parser/db/` — verify `db/gen/models.go` and `db/gen/documents.sql.go` have `ContentUrl` field
- [x] 4.3 Edit `parser/internal/domain/models.go` — add `ContentURL string \`json:"content_url"\`` to `DocumentMeta`
- [x] 4.4 Edit `parser/internal/infrastructure/repositories/pg_adverts_repo.go` — add `ContentURL: row.ContentUrl` in `GetDocument` return, add `ContentUrl: doc.ContentURL` in `SaveDocument` params
- [x] 4.5 Run `go fmt ./... && go vet ./...` — fix any errors

## 5. Python Tests

Note: `surfer/tests/_factories.py` (used by surfer workflow tests) requires no changes — `content_url` defaults to `""` via the model field default, so existing `make_doc_meta` calls continue to work without updates.

- [x] 5.1 Edit `downloader/tests/_factories.py` — add `content_url = ""` to `DownloaderDocMetaFactory`
- [x] 5.2 Edit `surfer/tests/unit/infrastructure/db/test_surf_mappers.py` — add `content_url=""` to ORM rows and doc objects in `TestDocumentToMeta` and `TestDocumentToOrm`, assert round-trip
- [x] 5.3 Edit `downloader/tests/integration/conftest.py` — add `content_url: str | None = None` param to `insert_document`, default to `""`
- [x] 5.4 Add new test `test_download_to_repo__when_content_url_set__then_fetches_from_content_url` in `tests/unit/application/activities/test_web_downloader.py` — mock `content_url` and assert GET target
- [x] 5.5 Run `poetry run pytest -m 'not linting and not integration' -v` — all unit tests pass

## 6. Go Tests

- [x] 6.1 Edit `parser/internal/infrastructure/repositories/pg_adverts_repo_integration_test.go` — update `insertDocument` helper to include `content_url` in INSERT and params; pass `""` on all existing call sites
- [x] 6.2 Edit `parser/internal/domain/models_test.go` — update JSON marshal/unmarshal expected strings to include `"content_url":""`; add a round-trip subtest asserting that `ContentURL = "https://cdn.example.com/x"` survives a marshal → unmarshal cycle
- [x] 6.3 Edit `parser/internal/application/testutil/factories.go` — add `ContentURL: ""` to `MustSearchPageDocument`, `MustDownloadedAdvertDocument`, `MustSurfedAdvertMeta`
- [x] 6.4 Edit `parser/internal/domain/testutil/factories.go` — add `ContentURL: ""` to `DocumentMetaFactory`, `MustDocumentMeta`
- [x] 6.5 Edit `parser/internal/application/activities/parser_test.go` — add `ContentURL: ""` to all `DocumentMeta{}` literals
- [x] 6.6 Run `go test ./...` (unit only, no integration) — all tests pass

## 7. Final Verification

- [x] 7.1 Python: `poetry run ruff check --fix . && poetry run mypy surfer downloader shared`
- [x] 7.2 Go: `go fmt ./... && go vet ./...`
- [x] 7.3 Python unit tests: `poetry run pytest -m 'not linting and not integration' -v`
- [x] 7.4 Go unit tests: `go test -tags='!integration' ./...`
- [x] 7.5 OpenSpec: `openspec validate --change add-content-url-column`
