## Context

The `documents` table currently uses `sdoc_id` alone as the primary key. The `surf_configs` table lacks an update interval field. Both require migration to support the new composite key and per-document refresh intervals.

See `proposal.md` for motivation.

## Goals / Non-Goals

**Goals:**
- Implement composite primary key `(sdoc_id, source_id, doc_type)` on `documents` — `sdoc_id` is an MD5 hash of the normalized URL, not a global identifier; the composite key establishes unique document identity across sources and document types
- Add `update_interval_sec` to both `documents` and `surf_configs` tables with DB-level default of 86400
- Propagate `update_interval_sec` from surf config → search page doc → child docs through the parsing pipeline
- Make `update_interval_sec` a required, positive integer in both Python and Go domain models

**Non-Goals:**
- Preserving existing data during migration (full wipe and replay)
- Fixing `PGAdvertsRepo.get_documents_meta_by_sdoc_id` collision behavior (marked for removal)
- Changing activity signatures for parser or downloader
- Adding `update_interval_sec` to the `parser-activities` spec (no externally observable behavior change)

## Decisions

### Decision: Composite PK `(sdoc_id, source_id, doc_type)` instead of sdoc_id alone

**Rationale:** The same URL content may be fetched from different sources (e.g., mirror sites) or processed as different document types. Each distinct `(sdoc_id, source_id, doc_type)` tuple represents a unique document occurrence.

**Alternative considered:** Keep `sdoc_id` as sole PK and add `(source_id, doc_type)` as nullable columns. Rejected because it conflates identity with metadata and makes uniqueness ambiguous.

### Decision: DB-level `server_default="86400"` for `update_interval_sec`

**Rationale:** Both tables get the same default. The DB-level default means bare inserts (e.g., in integration test fixtures) work without explicit values. ORM `server_default` ensures the ORM model does not need to supply the value either.

**Alternative considered:** No server default, require application to always supply the value. Rejected because it would require updating every test fixture and integration conftest insert.

### Decision: `index_elements` for upsert uses composite PK columns

**Rationale:** `PGDocumentRepository.save` uses `ON CONFLICT DO UPDATE`. The conflict target must reference the new composite primary key constraint columns `(sdoc_id, source_id, doc_type)`. Using only `sdoc_id` would fail because there is no unique index on that column alone.

**Alternative considered:** Create a separate unique index on just `sdoc_id`. Rejected because it would allow duplicate `sdoc_id` rows, contradicting the composite PK design.

### Decision: `update_interval_sec` required in domain models, validated `> 0`

**Rationale:** A value of 0 or negative would be nonsensical ("refresh every 0 seconds" or "refresh every -1 seconds"). Making it required forces callers to be explicit; using pydantic `Field(..., gt=0)` enforces it at construction time in Python.

**Alternative considered:** Optional with DB default. Rejected because downstream scheduling logic would need to handle `None` everywhere.

### Decision: Propagation via model_dump spread in Python downloader

**Rationale:** `WebDownloader.download_to_repo` uses `**doc_meta.model_dump()` to construct the `Document`. Any field present in `DocumentMeta` (including `update_interval_sec`) is automatically passed through without explicit naming.

**Alternative considered:** Explicit field assignment. Rejected as unnecessary boilerplate given the spread approach.

## Risks / Trade-offs

[Risk] `PGAdvertsRepo.get_documents_meta_by_sdoc_id` silently keeps only the last row when multiple rows share the same `sdoc_id`
→ **Mitigation:** The method is already marked for removal. Left as-is for now.

[Risk] Many test fixtures and integration conftest inserts must include `update_interval_sec`
→ **Mitigation:** All sites identified and enumerated in tasks.md. Default of 86400 is straightforward to add everywhere.

[Risk] Go parser `DocumentMeta.Validate()` now checks `UpdateIntervalSec > 0`; existing test data may fail
→ **Mitigation:** All Go test literals updated to include `UpdateIntervalSec: 86400`.

[Risk] Migration wipe discards all existing data
→ **Mitigation:** Explicit non-goal. Data is seeded by migrations or re-created by crawler runs.

## Migration Plan

1. Edit `0001_create_surf_configs.py` to include `update_interval_sec INTEGER NOT NULL DEFAULT 86400`
2. Edit `0004_create_documents.py` to use composite PK and add `update_interval_sec`
3. Edit `0005_seed_surf_configs.py` to set per-row `update_interval_sec` values
4. Run `cd crawler && poetry run alembic downgrade base && poetry run alembic upgrade head`
5. Verify: `make crawler-py-lint`, `make crawler-py-test-unit`, `make crawler-go-fmt`, `make crawler-go-lint`, `make crawler-go-test-unit`

## Open Questions

None. All decisions resolved during proposal phase.
