## Context

`ProcessSearchPage` iterates over `documents_meta` returned by `PARSE_SEARCH_PAGE` and starts a `ProcessAdvert` child workflow for each entry without checking whether the document has been processed too recently. `DocumentMeta` now carries `update_interval_sec`, making per-document rate-limiting possible.

See `proposal.md` for motivation.

## Goals / Non-Goals

**Goals:**
- Prevent starting redundant `ProcessAdvert` children for documents whose refresh window has not elapsed
- Add structured logging of processed/skipped counts for observability
- Minimum code changes: one new import, one added condition, two counters, one log line, and one new test

**Non-Goals:**
- Changing `PARSE_SEARCH_PAGE` activity behavior
- Changing `ProcessAdvert` workflow
- Adding any DB schema or migration
- Rate-limiting across the full `ProcessSearchBranch` or `SearchAdverts` tree — only individual advert doc_meta in `ProcessSearchPage`

## Decisions

### Decision: Skip condition uses `>` comparison (rate-limiting)

**Chosen:** `if doc_meta.updated_at + dt.timedelta(seconds=doc_meta.update_interval_sec) > workflow.now(): skip`

**Rationale:** This skips documents that are still within their refresh window (`now < updated_at + interval`). This is the conventional rate-limiting pattern — do not re-process until the interval has fully elapsed.

**Alternative considered:** `if doc_meta.updated_at + dt.timedelta(seconds=doc_meta.update_interval_sec) < workflow.now(): skip` — this would skip stale documents instead of fresh ones, which is backwards for this use case.

### Decision: Counters are simple integers, incremented before continue

**Chosen:** `processed = 0; skipped = 0` with increment before the `continue` for skipped.

**Rationale:** Straightforward, no need for a dataclass or namedtuple. Increment happens immediately before the `continue` to keep the skip branch minimal.

### Decision: Stats logged as `workflow.logger.info` with `extra` dict

**Chosen:**
```python
workflow.logger.info(
    "documents processed",
    extra={
        "surf_config_name": in_.surf_params.name,
        "page_url": in_.page_url,
        "processed": processed,
        "skipped": skipped,
        "total": len(documents_meta),
    },
)
```

**Rationale:** Matches the existing logging pattern in this workflow (log message as format string, structured data in `extra`). The human-readable message is `"documents processed"` — a summary noun phrase, consistent with the `workflow.logger.info("starting %s", ...)` call at the top of the workflow.

### Decision: `import datetime as dt` added for `dt.timedelta`

**Chosen:** `import datetime as dt` in the stdlib import block.

**Rationale:** Per AGENTS.md import conventions: `import datetime as dt` is the project standard. Placed alphabetically after `hashlib`.

### Decision: New test uses `dt.datetime.now(tz=dt.UTC)` for recent timestamp

**Chosen:** `recent_ts = dt.datetime.now(tz=dt.UTC)` — used for both `created_at` and `updated_at` on fresh doc_meta fixtures.

**Rationale:** With `update_interval_sec=86400` (factory default), `updated_at + 86400s` is firmly in the future relative to the test env clock, reliably triggering the skip. No need to mock time or use `workflow.now()` outside the workflow.

## Risks / Trade-offs

[Risk] `dt.datetime.now(tz=dt.UTC)` in test vs. `workflow.now()` in workflow — small clock offset if test setup takes real-time between fixture construction and workflow execution
→ **Mitigation:** The offset is on the order of milliseconds. With `update_interval_sec=86400`, the skip condition margin is ~86400s. The risk of flakiness is negligible.

[Risk] Existing tests use factory default `updated_at=2024-01-01` which is stale relative to 2026 test env — after the flip to `>`, these docs are processed (not skipped), so existing tests pass unchanged
→ **Mitigation:** Verified: `2024-01-02 > 2026-something` is `False` → no skip → child workflows start as before.

[Risk] Two counters (`processed`, `skipped`) but only `processed` is incremented after `start_child_workflow` — if the child workflow raises synchronously before returning, the counter still increments
→ **Mitigation:** `start_child_workflow` with `ParentClosePolicy.ABANDON` is fire-and-forget; incrementing after the `await` is the correct point. The increment is best-effort but acceptable for observability.

## Migration Plan

No migration needed — pure in-memory workflow change. No DB schema, no data, no deployment order.

Verification steps after implementation:
1. `poetry run ruff check .`
2. `poetry run mypy surfer`
3. `poetry run pytest tests/unit/application/workflows/test_process_search_page.py -v`
4. Full unit suite: `poetry run pytest -m "not linting and not integration"`

## Open Questions

None.
