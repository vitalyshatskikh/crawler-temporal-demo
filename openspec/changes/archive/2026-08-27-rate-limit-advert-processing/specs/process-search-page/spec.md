## Purpose

Controls how `ProcessSearchPage` decides whether to start a `ProcessAdvert` child workflow for each document returned by the parse activity — rate-limiting based on per-document `update_interval_sec` and logging structured processing stats.

## ADDED Requirements

### Requirement: Rate-limit advert-child starts by update_interval_sec

`ProcessSearchPage` SHALL skip starting a `ProcessAdvert` child workflow for any `doc_meta` where `doc_meta.updated_at + timedelta(seconds=doc_meta.update_interval_sec) > workflow.now()`.

#### Scenario: Document is still within its refresh window — skip

- **WHEN** `doc_meta.updated_at + timedelta(seconds=doc_meta.update_interval_sec) > workflow.now()`
- **THEN** `ProcessAdvert` child workflow is NOT started for this `doc_meta`

#### Scenario: Document is past its refresh window — process

- **WHEN** `doc_meta.updated_at + timedelta(seconds=doc_meta.update_interval_sec) <= workflow.now()`
- **THEN** `ProcessAdvert` child workflow IS started for this `doc_meta`

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
