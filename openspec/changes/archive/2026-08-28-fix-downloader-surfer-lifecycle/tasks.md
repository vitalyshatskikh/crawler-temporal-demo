## 1. Downloader activity

- [x] 1.1 Edit `crawler/downloader/application/activities/web_downloader.py`: add `import datetime as dt`; in `download_to_repo`, build the `adverts.Document` with `**doc_meta.model_dump(exclude={'created_at', 'updated_at', 'type'})`, then explicitly set `created_at=now`, `updated_at=now`, `type=_downloaded_type(doc_meta.type)`, where `now = dt.datetime.now(tz=dt.UTC)`.
- [x] 1.2 Add module-level helper `_downloaded_type(in_type: adverts.DocumentType) -> adverts.DocumentType` with the `match` table: `SEARCH_PAGE→SEARCH_PAGE`, `SURFED_ADVERT→DOWNLOADED_ADVERT`, `DOWNLOADED_ADVERT→DOWNLOADED_ADVERT`, `PARSED_ADVERT→DOWNLOADED_ADVERT`.
- [x] 1.3 Run `poetry run ruff check --fix downloader` and `poetry run mypy surfer downloader` from `crawler/` — fix any issues.

## 2. ProcessAdvert workflow

- [x] 2.1 Edit `crawler/surfer/application/workflows/process_advert.py`: change the `PARSE_ADVERT_CONTENT` activity argument from `in_.doc_meta` to `in_.doc_meta.model_copy(update={'type': adverts.DocumentType.DOWNLOADED_ADVERT})`.
- [x] 2.2 Run `poetry run ruff check --fix .` and `poetry run mypy surfer` from `crawler/` — fix any issues.

## 3. ProcessSearchPage workflow

- [x] 3.1 Edit `crawler/surfer/application/workflows/process_search_page.py`: add `import typing as tp` to the stdlib imports.
- [x] 3.2 Change the `run` return annotation from `-> None` to `-> dict[str, tp.Any]`.
- [x] 3.3 On the `workflow.execute_activity(consts.ActivityName.PARSE_SEARCH_PAGE, ...)` call, add `result_type=list[adverts.DocumentMeta]`.
- [x] 3.4 Replace the skip condition with `just_created = doc_meta.updated_at == doc_meta.created_at; need_update = doc_meta.updated_at + dt.timedelta(seconds=doc_meta.update_interval_sec) > workflow.now(); if not just_created and need_update: skipped += 1; continue`.
- [x] 3.5 Build the `stat` dict (already partially constructed) and return it from `run` instead of discarding it.
- [x] 3.6 Run `poetry run ruff check --fix .` and `poetry run mypy surfer` from `crawler/` — fix any issues.

## 4. Test updates

- [x] 4.1 Edit `crawler/downloader/tests/unit/application/activities/test_web_downloader.py`: add `import freezegun`; declare module-level `_FREEZE_TS` and `_FREEZE_TS_STR`; decorate the success-path and 404-path tests with `@freezegun.freeze_time(_FREEZE_TS_STR)`; update assertions to use `_FREEZE_TS` for `created_at`/`updated_at` and `adverts_models.DocumentType.DOWNLOADED_ADVERT` for `type`.
- [x] 4.2 Edit `crawler/surfer/tests/unit/application/workflows/test_process_search_page.py`: in the existing tests, capture the return value from `env.client.execute_workflow(...)` and assert it equals the expected `stat` dict.
- [x] 4.3 Add a new test `test_run__when_doc_meta_just_created__then_starts_advert_children` that constructs `recent_1` and `recent_2` with `created_at == updated_at == recent_ts`, runs the workflow, and asserts `processed=2, skipped=0, total=2` plus two `PROCESS_ADVERT` calls.
- [x] 4.4 Update `test_run__when_doc_meta_recently_updated__then_skips_advert_children` so the `recent_*` fixtures have `updated_at = recent_ts + dt.timedelta(seconds=1)` (so `updated_at != created_at` while still being within the refresh window).
- [x] 4.5 Run `poetry run pytest -m "not linting and not integration" -v` from `crawler/` — all unit tests pass.

## 5. Dependencies

- [x] 5.1 Edit `crawler/pyproject.toml`: add `freezegun = "^1.5.5"` to the dev dependency group.
- [x] 5.2 Re-lock with `poetry lock` from `crawler/` to update `crawler/poetry.lock`.

## 6. Final verification

- [x] 6.1 `poetry run ruff check --fix .` from `crawler/` — clean.
- [x] 6.2 `poetry run mypy surfer downloader` from `crawler/` — clean.
- [x] 6.3 `poetry run pytest -m "not linting and not integration" -v` from `crawler/` — all unit tests pass.
- [x] 6.4 Wipe the Temporal namespace and restart all workers on deploy (workflow return-type change makes in-flight history incompatible).
- [x] 6.5 `openspec validate --change fix-downloader-surfer-lifecycle` — passes.
