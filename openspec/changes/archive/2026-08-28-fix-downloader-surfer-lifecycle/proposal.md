## Why

While wiring up the surfer/downloader/parser flow end-to-end, four bugs surfaced that together broke the document lifecycle. The `ProcessSearchPage` rate-limit skip rule (`updated_at + update_interval_sec > now`) wrongly skips documents that have just been created (`updated_at == created_at`) — i.e., a brand-new advert surfaced from a fresh search page parse is never downloaded, defeating the entire pipeline. The `WebDownloader` activity persists the downloaded document with the *input* `created_at`/`updated_at`/`type` from `doc_meta` (which reflect when the *surfaced* entry was created, not when the download happened) and does not advance the document type to `DOWNLOADED_ADVERT` after a successful fetch. `ProcessAdvert` then calls `PARSE_ADVERT_CONTENT` with the stale `SURFED_ADVERT` type, which the parser rejects. Finally, the `PARSE_SEARCH_PAGE` activity call in `ProcessSearchPage` is missing a `result_type` annotation, and the workflow swallows the per-page processed/skipped counts instead of returning them as workflow output.

## What Changes

- **`WebDownloader` activity** (`crawler/downloader/application/activities/web_downloader.py`): When saving the downloaded `Document`, override `created_at` and `updated_at` with the current UTC timestamp (`dt.datetime.now(tz=dt.UTC)`) and remap `type` via a new helper `_downloaded_type`:
  - `SEARCH_PAGE` → `SEARCH_PAGE`
  - `SURFED_ADVERT` / `PARSED_ADVERT` → `DOWNLOADED_ADVERT`
  - `DOWNLOADED_ADVERT` → `DOWNLOADED_ADVERT` (idempotent re-download)
- **`ProcessAdvert` workflow** (`crawler/surfer/application/workflows/process_advert.py`): Pass `in_.doc_meta.model_copy(update={'type': adverts.DocumentType.DOWNLOADED_ADVERT})` to `PARSE_ADVERT_CONTENT` so the parser sees the post-download type, not the stale `SURFED_ADVERT`.
- **`ProcessSearchPage` workflow** (`crawler/surfer/application/workflows/process_search_page.py`):
  - Add `result_type=list[adverts.DocumentMeta]` to the `PARSE_SEARCH_PAGE` activity call so the temporalio workflow replayer has a concrete type.
  - Replace the skip rule with `if not just_created and need_update: skip`, where `just_created = doc_meta.updated_at == doc_meta.created_at`. Brand-new documents (`created_at == updated_at`) are always processed; only documents that have been seen before and are still within their refresh window are skipped.
  - Return the existing `stat` dict (`surf_config_name`, `page_url`, `processed`, `skipped`, `total`) as the workflow output instead of discarding it.
- **Tests**:
  - `crawler/downloader/tests/unit/application/activities/test_web_downloader.py`: add `freezegun` decoration to the success-path and 404-path tests so the frozen `now` matches `_FREEZE_TS`; assert that the saved document has fresh `created_at`/`updated_at` and `type=DOWNLOADED_ADVERT`.
  - `crawler/surfer/tests/unit/application/workflows/test_process_search_page.py`: assert the returned stats dict; add a new test `test_run__when_doc_meta_just_created__then_starts_advert_children` covering the just-created path; update the existing skip test to use `updated_at > created_at` (e.g., `+1s`).
- **Dependencies**: add `freezegun = "^1.5.5"` to `crawler/pyproject.toml` (and the matching `poetry.lock` entry). Used only by the downloader activity tests where Temporal's time-skipping environment is not active.

## Capabilities

### New Capabilities

*(none — all changes modify behavior covered by existing capabilities)*

### Modified Capabilities

- `documents`: The downloader's responsibility to advance document type and refresh timestamps on a successful fetch is now a documented behavior. The downloader no longer passes `doc_meta` through verbatim — it stamps the document with the time of the actual download and transitions the type to the `DOWNLOADED_ADVERT` state for advert-shaped inputs.
- `process-search-page`: The skip rule now treats `created_at == updated_at` as a "just created" signal that always bypasses the rate-limit check. The workflow also returns a structured stats report (`processed`/`skipped`/`total`) instead of returning `None`. The `PARSE_SEARCH_PAGE` activity call gets an explicit `result_type` for replay-correctness.

## Impact

- **Python — surfer**: `process_search_page.py` (skip condition, return type, log/return dict, `result_type` annotation), `process_advert.py` (model_copy with type override).
- **Python — downloader**: `web_downloader.py` (timestamp/type override, `_downloaded_type` helper).
- **Tests**: 2 test files updated; 1 new test added. Pyproject adds `freezegun` as a dev dep.
- **No DB schema change.** No new Temporal activity or workflow.
- **No Go-side changes.** The Go parser is unchanged; it already expects `DOWNLOADED_ADVERT` input to `ParseAdvertContent` and `SEARCH_PAGE` input to `ParseSearchPage` per the existing `parser-activities` spec.
- **Replay compatibility**: Changing the workflow output type of `ProcessSearchPage` from `None` to `dict` is a workflow signature change for in-flight executions. Treat workflow history as ephemeral — restart workers on a clean Temporal namespace after deploy (same posture as prior changes).
