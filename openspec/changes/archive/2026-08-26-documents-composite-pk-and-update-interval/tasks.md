## 1. Migrations (in-place edit, rollback to base and reapply)

- [x] 1.1 Edit `crawler/migrations/versions/0001_create_surf_configs.py`: add `update_interval_sec INTEGER NOT NULL DEFAULT 86400` column to `surf_configs` table definition; also update `downgrade()` to drop this column
- [x] 1.2 Edit `crawler/migrations/versions/0004_create_documents.py`: replace sole `sdoc_id` PK with `PrimaryKeyConstraint("sdoc_id", "source_id", "doc_type", name="pk_documents")`; remove `op.create_index("idx_documents_source_id", ...)` and `op.create_index("idx_documents_doc_type", ...)`; add `update_interval_sec INTEGER NOT NULL DEFAULT 86400` column; also update `downgrade()` to: drop the new column, drop the composite PK constraint, recreate the sole-`sdoc_id` primary key, and recreate the two dropped single-column indexes
- [x] 1.3 Edit `crawler/migrations/versions/0005_seed_surf_configs.py`: add `update_interval_sec=600` to `siteapi-local-debug` row; add `update_interval_sec=7200` to `siteapi-demo-fresh` and `siteapi-demo-all` rows

## 2. ORM models (Python)

- [x] 2.1 Edit `crawler/surfer/infrastructure/db/orm/surf_configs.py`: add `update_interval_sec: Mapped[int] = mapped_column(sa.Integer, server_default=text("86400"))`
- [x] 2.2 Edit `crawler/shared/py/db/orm/documents.py`: set `primary_key=True` on `sdoc_id`, `source_id`, and `doc_type` columns; remove `index=True` from `source_id` and `doc_type`; add `update_interval_sec: Mapped[int] = mapped_column(sa.Integer, server_default=text("86400"))`

## 3. Mappers (Python)

- [x] 3.1 Edit `crawler/shared/py/db/mappers.py`: add `update_interval_sec=row.update_interval_sec` to `document_to_meta`; add `update_interval_sec=doc.update_interval_sec` to `document_to_orm`
- [x] 3.2 Edit `crawler/surfer/infrastructure/db/mappers.py`: add `update_interval_sec=row.update_interval_sec` to `surf_config_to_params`; add `update_interval_sec=params.update_interval_sec` to `params_to_surf_config`

## 4. Domain models (Python)

- [x] 4.1 Edit `crawler/surfer/domain/adverts/models.py`: add `update_interval_sec: int = pydantic.Field(..., gt=0)` to `DocumentMeta`
- [x] 4.2 Edit `crawler/surfer/domain/surfing/models.py`: add `update_interval_sec: int = pydantic.Field(..., gt=0)` to `Params`

## 5. Domain models (Go)

- [x] 5.1 Edit `crawler/parser/internal/domain/models.go`: add `UpdateIntervalSec int` field to `DocumentMeta` struct; add `if d.UpdateIntervalSec <= 0 { return fmt.Errorf("%w: UpdateIntervalSec must be > 0", ErrValidation) }` to `Validate()`

## 6. Go parser propagation

- [x] 6.1 Edit `crawler/parser/internal/domain/parser.go` (`ParseSearchPage`): copy `doc.UpdateIntervalSec` into each `DocumentMeta` in `parsedDocs` slice
- [x] 6.2 Edit `crawler/parser/internal/domain/parser.go` (`ParseAdvertContent`): copy `doc.UpdateIntervalSec` into the returned `DocumentMeta`

## 7. Workflows (Python)

- [x] 7.1 Edit `crawler/surfer/application/workflows/process_search_page.py`: add `update_interval_sec=in_.surf_params.update_interval_sec` to `page_doc_meta` construction

## 8. Repository upsert (Python)

- [x] 8.1 Edit `crawler/downloader/infrastructure/repositories/document_repo.py`: change `index_elements=[orm.DocumentORM.sdoc_id]` to `index_elements=[orm.DocumentORM.sdoc_id, orm.DocumentORM.source_id, orm.DocumentORM.doc_type]`; add `"update_interval_sec": stmt.excluded.update_interval_sec` to `set_=` so the interval is refreshed on every upsert

## 9. Go test factories and test data

- [x] 9.1 Edit `crawler/parser/internal/domain/testutil/factories.go`: add `UpdateIntervalSec: 86400` to `DocumentMetaFactory`, `MustDocumentMeta`, and `DocumentFactory`
- [x] 9.2 Edit `crawler/parser/internal/application/testutil/factories.go`: add `UpdateIntervalSec: 86400` to `MustSurfedAdvertMeta`
- [x] 9.3 Edit `crawler/parser/internal/domain/models_test.go`: add `UpdateIntervalSec: 86400` to every inline `DocumentMeta{...}` literal (10 sites); add `TestDocumentMeta_WhenZeroOrNegativeUpdateIntervalSec_ThenErrValidation` function covering 0 and -1
- [x] 9.4 Edit `crawler/parser/internal/domain/parser_test.go`: add `UpdateIntervalSec: 86400` to every inline `DocumentMeta` literal
- [x] 9.5 Edit `crawler/parser/internal/application/activities/parser_test.go`: add `UpdateIntervalSec: 86400` to every inline `DocumentMeta` literal

## 10. Python test factories

- [x] 10.1 Edit `crawler/surfer/tests/_factories.py`: add `update_interval_sec = 86400` to `SurfParamsFactory` and `DocMetaFactory`
- [x] 10.2 Edit `crawler/downloader/tests/_factories.py`: add `update_interval_sec = 86400` to `DownloaderDocMetaFactory`

## 11. Python unit tests

- [x] 11.1 Edit `crawler/surfer/tests/unit/domain/surfing/test_models_validation.py`: add table-driven test for `WhenNonPositiveUpdateIntervalSec_ThenErrValidation` covering 0 and -1 on `Params.update_interval_sec`
- [x] 11.2 Edit `crawler/surfer/tests/unit/infrastructure/db/test_surf_mappers.py`: add `update_interval_sec=86400` to `Params(...)` literal and to `SurfConfigORM(...)` row
- [x] 11.3 Edit `crawler/surfer/tests/unit/infrastructure/db/test_surf_mappers.py`: add `update_interval_sec=86400` to `Document(...)` constructor and to `DocumentORM(...)` row
- [x] 11.4 Edit `crawler/surfer/tests/unit/application/workflows/test_process_search_page.py`: extend both `meta.model_dump()` assertions to include `"update_interval_sec": in_.surf_params.update_interval_sec` (date fields still excluded)
- [x] 11.5 Edit `crawler/downloader/tests/unit/application/activities/test_web_downloader.py`: add `update_interval_sec=86400` to inline `DocumentMeta` literal at line 108

## 12. Python integration test fixtures

- [x] 12.1 Edit `crawler/surfer/tests/integration/conftest.py` (`insert_document`): add `update_interval_sec=86400` to the `sa.insert(shared_orm.DocumentORM).values(...)` call
- [x] 12.2 Edit `crawler/surfer/tests/integration/infrastructure/repositories/test_pg_config_repo_integration.py`: add `update_interval_sec=86400` to both `sa.insert(surfer_orm.SurfConfigORM).values(...)` calls; assert `result.update_interval_sec == 86400`; also insert a row with a non-default value (e.g. `update_interval_sec=300`) and assert it round-trips correctly
- [x] 12.3 Edit `crawler/downloader/tests/integration/conftest.py` (`insert_document`): add `update_interval_sec=86400` to the `sa.insert(shared_orm.DocumentORM).values(...)` call
- [x] 12.4 Edit `crawler/downloader/tests/integration/infrastructure/repositories/test_pg_document_repo_integration.py`: add `update_interval_sec=86400` to all three `adverts.Document(...)` constructors
- [x] 12.5 Edit `crawler/downloader/tests/integration/infrastructure/repositories/test_pg_document_repo_integration.py`: add a test that (a) inserts two `documents` rows sharing the same `sdoc_id` but differing in `source_id` (e.g. `"src-a"` vs `"src-b"`), then queries by `(sdoc_id, source_id, doc_type)` and asserts both rows persist independently; (b) calls `PGDocumentRepository.save` with an existing composite key and asserts only the matching row is updated while the other is untouched

## 13. Migration execution and verification

- [x] 13.1 Run `cd crawler && poetry run alembic downgrade base && poetry run alembic upgrade head`
- [x] 13.2 Run `make crawler-py-lint` and fix any issues
- [x] 13.3 Run `make crawler-py-test-unit` and ensure all tests pass
- [x] 13.4 Run `make crawler-go-fmt` and `make crawler-go-lint` and fix any issues
- [x] 13.5 Run `make crawler-go-test-unit` and ensure all tests pass
