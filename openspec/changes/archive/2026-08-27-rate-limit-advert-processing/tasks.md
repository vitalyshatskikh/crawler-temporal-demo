## 1. Workflow edit

- [x] 1.1 Edit `crawler/surfer/application/workflows/process_search_page.py`: add `import datetime as dt` to the stdlib import block (after `hashlib`)
- [x] 1.2 Edit the `for doc_meta in documents_meta:` loop: add `processed = 0` and `skipped = 0` before the loop; flip the skip condition to `if doc_meta.updated_at + dt.timedelta(seconds=doc_meta.update_interval_sec) > workflow.now(): skipped += 1; continue`; increment `processed += 1` after `start_child_workflow`
- [x] 1.3 Add post-loop `workflow.logger.info` with `extra={"surf_config_name": ..., "page_url": ..., "processed": processed, "skipped": skipped, "total": len(documents_meta)}`

## 2. Test additions

- [x] 2.1 Edit `crawler/surfer/tests/unit/application/workflows/test_process_search_page.py`: add `import datetime as dt` to the stdlib import block
- [x] 2.2 Append async test `test_run__when_doc_meta_recently_updated__then_skips_advert_children`:
  - `recent_ts = dt.datetime.now(tz=dt.UTC)`
  - `recent_1 = DocMetaFactory(sdoc_id='1', created_at=recent_ts, updated_at=recent_ts)`
  - `recent_2 = DocMetaFactory(sdoc_id='2', created_at=recent_ts, updated_at=recent_ts)`
  - Mock `DOWNLOAD_SEARCH_PAGE → {}`, `PARSE_SEARCH_PAGE → [recent_1, recent_2]`, `PROCESS_ADVERT → None`
  - Run `ProcessSearchPage` with id `test-psp-recent-skip`
  - Assert `len(get_mock_calls(PROCESS_ADVERT)) == 0`

## 3. Verification

- [x] 3.1 Run `cd crawler/surfer && poetry run ruff check .` and fix any issues
- [x] 3.2 Run `cd crawler/surfer && poetry run mypy surfer` and fix any type errors
- [x] 3.3 Run `cd crawler/surfer && poetry run pytest tests/unit/application/workflows/test_process_search_page.py -v` — all 5 tests (4 existing + 1 new) must pass
- [x] 3.4 Run `cd crawler/surfer && poetry run pytest -m "not linting and not integration"` — full unit suite must pass
