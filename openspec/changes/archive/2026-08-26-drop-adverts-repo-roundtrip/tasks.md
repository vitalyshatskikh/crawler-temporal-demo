## 0. Add Go JSON tags to `DocumentMeta`

- [x] 0.1 In `crawler/parser/internal/domain/models.go:29-37`, add `json:"..."` struct tags to every field of `DocumentMeta`: `SdocID → json:"sdoc_id"`, `CreatedAt → json:"created_at"`, `UpdatedAt → json:"updated_at"`, `SourceID → json:"source_id"`, `Type → json:"type"`, `ExternalURL → json:"external_url"`, `UpdateIntervalSec → json:"update_interval_sec"`.
- [x] 0.2 In `crawler/parser/internal/domain/models_test.go`, add `TestDocumentMeta_WhenMarshaledToJSON_ThenKeysAreSnakeCase` and `TestDocumentMeta_WhenUnmarshaledFromSnakeCaseJSON_ThenFieldsPopulated` using the standard `encoding/json` package. Use `assert.JSONEq` and direct field comparisons.
- [x] 0.3 Run `cd crawler/parser && go test ./...` and confirm green.

## 1. Remove the activity const

- [x] 1.1 In `crawler/surfer/application/consts.py`, remove `GET_DOCUMENTS_META = "GetDocumentsMeta"` from `ActivityName`.

## 2. Refactor the workflow to consume DocumentMeta directly

- [x] 2.1 In `crawler/surfer/application/workflows/process_search_page.py`, change the `PARSE_SEARCH_PAGE` activity result variable from `sdoc_ids` to `documents_meta: list[adverts.DocumentMeta]`.
- [x] 2.2 Delete the entire `execute_local_activity_method(activities.AdvertsRepo.get_documents_meta, ...)` block.
- [x] 2.3 Change the `for sdoc_id, doc_meta in documents_meta.items():` loop to `for doc_meta in documents_meta:` and update the workflow id to use `doc_meta.sdoc_id` directly.
- [x] 2.4 Confirm `repo_request_timeout` / `repo_request_retry` are no longer referenced inside this file.

## 3. Remove the AdvertsRepo activity and update the dummy parse activity

- [x] 3.1 In `crawler/surfer/application/activities.py`, delete the `AdvertsRepo` class (constructor + `get_documents_meta`).
- [x] 3.2 Change `dummy_parse_search_page`'s parameter from `url: str` to `meta: adverts.DocumentMeta` so the dummy signature matches both the real Go `Parser.ParseSearchPage(ctx, meta domain.DocumentMeta) ([]domain.DocumentMeta, error)` and the call site in `process_search_page.py:59-65`. Update the log line to use `meta.external_url` instead of `url`.
- [x] 3.3 Change `dummy_parse_search_page`'s return type from `list[adverts.SdocID]` to `list[adverts.DocumentMeta]`; body returns `[]`.
- [x] 3.4 Verify the `adverts` import is still needed for `dummy_parse_advert_content` (it is).

## 4. Delete the port and its implementations

- [x] 4.1 Delete `crawler/surfer/domain/adverts/repositories.py`.
- [x] 4.2 In `crawler/surfer/domain/adverts/__init__.py`, drop `IAdvertsRepository` from the `from .repositories import ...` line and from `__all__`.
- [x] 4.3 Delete `crawler/surfer/infrastructure/repositories/adverts_repo.py`.
- [x] 4.4 In `crawler/surfer/infrastructure/repositories/__init__.py`, drop the `PGAdvertsRepo` import line and `__all__` entry.

## 5. Rewire the worker and entry point

- [x] 5.1 In `crawler/surfer/infrastructure/workers.py`, drop `adverts_repo: adverts.IAdvertsRepository` from `SurfingWorker.__init__`, drop `self._adverts_repo = activities.AdvertsRepo(adverts_repo)`, and drop `self._adverts_repo.get_documents_meta` from the `activities=[...]` list. Keep `dummy_parse_search_page` and `dummy_parse_advert_content` registered on `ParsingWorker`.
- [x] 5.2 In `crawler/surfer/__main__.py`, delete the `adverts_repo = repositories.PGAdvertsRepo(sessionmaker)` line and remove the `adverts_repo=adverts_repo` kwarg from `SurfingWorker(...)`.

## 6. Update unit workflow tests

- [x] 6.1 In `crawler/surfer/tests/unit/application/workflows/test_process_search_page.py`, remove the `activities.AdvertsRepo(DummyAdvertsRepository(...)).get_documents_meta` activity registration from all four `WorkerSpec`s. Drop the `from surfer.domain.adverts import repositories as adverts_repo` import.
- [x] 6.2 In `test_run__when_parse_returns_empty__then_no_advert_children`, leave `set_mock(PARSE_SEARCH_PAGE, result=[])` as-is (empty list still valid).
- [x] 6.3 In `test_run__when_parse_returns_sdocs__then_starts_advert_children`, replace the `sdoc_ids = [SdocID('1'), SdocID('2')]` + repo dict + `documents_meta` setup with `doc_meta_1 = DocMetaFactory(sdoc_id='1')`, `doc_meta_2 = DocMetaFactory(sdoc_id='2')`, and `set_mock(PARSE_SEARCH_PAGE, result=[doc_meta_1, doc_meta_2])`. Update the `assert advert_calls[i].args[0].doc_meta.sdoc_id == ...` lines to read directly from the returned list.
- [x] 6.4 In `test_run__when_download_fails__then_propagates_error` and `test_run__when_parse_fails__then_propagates_error`, no mock-data changes are needed; just confirm the `WorkerSpec.activities` lists are empty now.
- [x] 6.5 In `crawler/surfer/tests/unit/application/workflows/test_search_adverts.py`, remove the `activities.AdvertsRepo(DummyAdvertsRepository(result={})).get_documents_meta` registration from all five `WorkerSpec`s. Drop the `from surfer.domain.adverts import repositories as adverts_repo` import.

## 7. Delete the integration test and prune its fixtures

- [x] 7.1 Delete `crawler/surfer/tests/integration/infrastructure/repositories/test_pg_adverts_repo_integration.py`.
- [x] 7.2 In `crawler/surfer/tests/integration/conftest.py`, remove the `from surfer.infrastructure.repositories.adverts_repo import PGAdvertsRepo` import, the `pg_adverts_repo` fixture, and the `insert_document` helper.

## 8. Refresh surfer AGENTS.md

- [x] 8.1 In `crawler/surfer/AGENTS.md`, drop the `IAdvertsRepository`/`AdvertsRepo` mentions from the architecture overview ("4 worker classes" → "3", remove the `AdvertsRepo` line from the "Port-bound activity classes" bullet, drop `DummyAdvertsRepository` from the "Test doubles" bullet).
- [x] 8.1.b In `crawler/surfer/AGENTS.md`, change `ActivityName (4)` → `ActivityName (3)` in the "Consts" line (the fourth activity, `GET_DOCUMENTS_META`, is deleted).
- [x] 8.2 Optionally note that the `ParseSearchPage` activity now returns `list[DocumentMeta]` end-to-end.

## 8a. Sync `SurferConfigFactory` with real `SurferConfig`

- [x] 8a.1 In `crawler/surfer/tests/_factories.py:68-79`, rename the five workflow-level timeout fields to match `application/config.py`:
  - `process_branch_timeout` → `process_branch_wf_timeout`
  - `process_search_page_timeout` → `process_search_page_wf_timeout`
  - `process_advert_timeout` → `process_advert_wf_timeout`
  - `download_search_page_timeout` → `download_search_page_wf_timeout`
  - `download_advert_content_timeout` → `download_advert_content_wf_timeout`
- [x] 8a.2 Add `search_adverts_timeout = dt.timedelta(minutes=20)`.
- [x] 8a.3 Add the three missing `RetryConfig` defaults: `repo_request_retry = RetryConfig()`, `parse_search_page_retry = RetryConfig()`, `parse_advert_content_retry = RetryConfig()`. Add `from shared.py.settings import RetryConfig` to imports.
- [x] 8a.4 Keep the four existing non-`RetryConfig` fields unchanged: `download_search_page_timeout`, `download_advert_content_timeout`, `repo_request_timeout`, `parse_search_page_timeout`, `parse_advert_content_timeout`.

## 9. Verification

- [x] 9.1 Run `poetry run ruff check .` and resolve any unused imports / undefined names.
- [x] 9.2 Run `poetry run mypy surfer` and resolve any type errors (in particular, the workflow's new `list[DocumentMeta]` typing).
- [x] 9.3 Run `poetry run pytest -m "not linting and not integration"` and confirm all unit tests pass.
- [x] 9.4 Manually grep the repo for residual references: `IAdvertsRepository`, `AdvertsRepo`, `DummyAdvertsRepository`, `PGAdvertsRepo`, `get_documents_meta`, `GET_DOCUMENTS_META`. Every hit should be inside this change's documentation (proposal/design/tasks) and nothing else.
