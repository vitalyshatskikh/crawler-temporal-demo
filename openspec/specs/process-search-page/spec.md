# process-search-page Specification

## Purpose

Controls how `ProcessSearchPage` decides whether to start a `ProcessAdvert` child workflow for each document returned by the parse activity — rate-limiting based on per-document `update_interval_sec` and logging structured processing stats.

## Requirements

### Requirement: Rate-limit advert-child starts by update_interval_sec

`ProcessSearchPage` SHALL skip starting a `ProcessAdvert` child workflow for any `doc_meta` where `doc_meta.updated_at == doc_meta.created_at` is FALSE AND `doc_meta.updated_at + timedelta(seconds=doc_meta.update_interval_sec) > workflow.now()`.

A document whose `created_at == updated_at` is treated as just-created: the rate-limit check SHALL be bypassed and the child workflow SHALL be started. Only documents that have been seen before (`updated_at > created_at`) AND are still within their refresh window SHALL be skipped.

#### Scenario: Document is still within its refresh window — skip

- **WHEN** `doc_meta.updated_at > doc_meta.created_at` AND `doc_meta.updated_at + timedelta(seconds=doc_meta.update_interval_sec) > workflow.now()`
- **THEN** `ProcessAdvert` child workflow is NOT started for this `doc_meta`

#### Scenario: Document is past its refresh window — process

- **WHEN** `doc_meta.updated_at + timedelta(seconds=doc_meta.update_interval_sec) <= workflow.now()`
- **THEN** `ProcessAdvert` child workflow IS started for this `doc_meta`

#### Scenario: Just-created document is always processed

- **WHEN** `doc_meta.updated_at == doc_meta.created_at` (regardless of `update_interval_sec`)
- **THEN** `ProcessAdvert` child workflow IS started for this `doc_meta`, even if `updated_at + update_interval_sec` is far in the future

#### Scenario: Empty documents_meta list — no children started

- **WHEN** `documents_meta` returned by `PARSE_SEARCH_PAGE` is empty
- **THEN** no `ProcessAdvert` child workflows are started

### Requirement: Structured stats logged after loop

After processing all `documents_meta`, `ProcessSearchPage` SHALL log a structured line with `extra={"processed": N, "skipped": M, "total": L}` where N is the count of docs that had a child started, M is the count skipped, and L is `len(documents_meta)`.

#### Scenario: All docs processed — zero skipped

- **WHEN** all `documents_meta` entries pass the refresh-window check
- **THEN** the log line reports `processed=N, skipped=0, total=N`

#### Scenario: Some docs skipped

- **WHEN** K of N `documents_meta` entries fail the refresh-window check
- **THEN** the log line reports `processed=N-K, skipped=K, total=N`

#### Scenario: All docs skipped

- **WHEN** all `documents_meta` entries fail the refresh-window check
- **THEN** the log line reports `processed=0, skipped=N, total=N` and no child workflows are started

### Requirement: ProcessSearchPage returns a stats report

`ProcessSearchPage.run` SHALL return a `dict[str, tp.Any]` containing the surf config name, the page URL, the counts `processed`, `skipped`, and `total`. The returned value SHALL match the structured log line emitted at the end of the workflow.

#### Scenario: Empty documents_meta — zero counts returned

- **WHEN** `ProcessSearchPage.run` is invoked and `PARSE_SEARCH_PAGE` returns an empty slice
- **THEN** the workflow returns `{"surf_config_name": ..., "page_url": ..., "processed": 0, "skipped": 0, "total": 0}`

#### Scenario: All documents processed — counts reflect the full list

- **WHEN** `ProcessSearchPage.run` is invoked and all returned `doc_meta` entries trigger child workflows
- **THEN** the workflow returns `{"surf_config_name": ..., "page_url": ..., "processed": N, "skipped": 0, "total": N}`

#### Scenario: Mix of processed and skipped — counts split correctly

- **WHEN** `ProcessSearchPage.run` is invoked with `K` skipped and `N-K` processed entries
- **THEN** the workflow returns `{"surf_config_name": ..., "page_url": ..., "processed": N-K, "skipped": K, "total": N}`

#### Scenario: All documents skipped

- **WHEN** `ProcessSearchPage.run` is invoked and all returned `doc_meta` entries fail the refresh-window check
- **THEN** the workflow returns `{"surf_config_name": ..., "page_url": ..., "processed": 0, "skipped": N, "total": N}`

### Requirement: ProcessSearchPage advertises the activity return type

`ProcessSearchPage` SHALL call `workflow.execute_activity(PARSE_SEARCH_PAGE, ...)` with an explicit `result_type=list[adverts.DocumentMeta]` so that the temporalio workflow replayer treats the activity's return as a `list[adverts.DocumentMeta]`. The runtime value of the activity's return SHALL be the slice of `DocumentMeta` produced by the parser.

#### Scenario: Activity call uses explicit result_type

- **WHEN** `ProcessSearchPage` calls `PARSE_SEARCH_PAGE`
- **THEN** the call passes `result_type=list[adverts.DocumentMeta]`

#### Scenario: Activity return is treated as a list of DocumentMeta

- **WHEN** the parser activity returns N `DocumentMeta` entries
- **THEN** the workflow iterates the returned value as `list[adverts.DocumentMeta]` of length N
