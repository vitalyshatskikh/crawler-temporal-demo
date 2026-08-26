## Why

`ProcessSearchPage` currently starts a `ProcessAdvert` child workflow for every document returned by the parse activity, regardless of how recently each document was last updated. Now that `DocumentMeta` carries `update_interval_sec`, we should skip starting a child when the document's refresh interval has not yet elapsed — preventing redundant processing of still-fresh documents.

## What Changes

- `ProcessSearchPage.run` loop: skip starting a `ProcessAdvert` child for any `doc_meta` where `updated_at + update_interval_sec > workflow.now()` (document is still within its refresh window)
- Track `processed` and `skipped` counters in the loop; log a structured stats line with `extra={"processed": N, "skipped": M, "total": L}` after the loop
- Add `import datetime as dt` to `process_search_page.py` (the `dt.timedelta` reference on line 68 has no corresponding import)
- Add unit test `test_run__when_doc_meta_recently_updated__then_skips_advert_children` covering the skip path; existing tests remain unchanged

## Capabilities

### New Capabilities

- `process-search-page`: Covers the `ProcessSearchPage` workflow's loop behavior: rate-limiting advert-child starts based on per-document `update_interval_sec`, and structured stats logging of processed/skipped/total counts.

### Modified Capabilities

*(none)*

## Impact

- **File modified**: `crawler/surfer/application/workflows/process_search_page.py` — add `import datetime as dt`, flip skip condition to `>`, add counters + post-loop `extra`-dict log
- **Test added**: `crawler/surfer/tests/unit/application/workflows/test_process_search_page.py` — one new async test for the skip scenario
- **No DB or API changes**
