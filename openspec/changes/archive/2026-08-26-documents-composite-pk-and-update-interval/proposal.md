## Why

The `documents` table currently uses `sdoc_id` alone as its primary key. This assumes each content hash is globally unique across all sources and document types. In practice, the same URL content may be fetched from different sources or processed as different document types, requiring distinct rows. Additionally, there is no per-document update interval field, making it impossible to express "how often this document should be refreshed" at the document level.

## What Changes

- **documents table**: Change primary key from `sdoc_id` alone to composite `(sdoc_id, source_id, doc_type)`. Drop the standalone `idx_documents_source_id` and `idx_documents_doc_type` indexes since the composite PK covers them. Note: `sdoc_id` is an MD5 hash of the normalized URL (`SdocIDForURL` in `crawler/parser/internal/domain/models.go`); the composite key establishes unique document identity across sources and types.
- **documents table**: Add `update_interval_sec INTEGER NOT NULL DEFAULT 86400` column.
- **surf_configs table**: Add `update_interval_sec INTEGER NOT NULL DEFAULT 86400` column.
- **DocumentMeta (Python, surfer)**: Add required `update_interval_sec: int` field, validated `> 0`.
- **DocumentMeta (Go, parser)**: Add `UpdateIntervalSec int` field, validated `> 0` in `Validate()`.
- **surfing.Params (Python)**: Add required `update_interval_sec: int` field, validated `> 0`.
- **PGDocumentRepository.save**: Update upsert conflict target from `sdoc_id` alone to `(sdoc_id, source_id, doc_type)` to match the new PK.
- **ParseSearchPage / ParseAdvertContent (Go)**: Propagate `UpdateIntervalSec` from input document to each output document.
- **Workflow process_search_page (Python)**: Set `update_interval_sec` on search page DocumentMeta from `surf_params.update_interval_sec`.
- **Migrations**: Roll back to base and re-apply all migrations with updated schemas.

## Capabilities

### New Capabilities

- `documents`: Covers the `documents` table schema (composite PK, `update_interval_sec` column) and the `DocumentMeta` model in both Python and Go stacks. This is the core document storage capability.
- `surf-configs`: Covers the `surf_configs` table schema (`update_interval_sec` column) and the `surfing.Params` model. This enables per-source update interval configuration.

### Modified Capabilities

*(none — `parser-activities` carries `UpdateIntervalSec` propagation but produces no externally observable spec delta; implementation changes are covered by tasks 5 and 6)*

## Impact

- **DB**: `documents` and `surf_configs` table schemas change. All data wiped on migration replay.
- **DB**: Downstream queries that assumed `sdoc_id` uniqueness (e.g., `PGAdvertsRepo.get_documents_meta_by_sdoc_id`) may now return multiple rows per `sdoc_id`. That method is already marked for removal and should not be relied upon.
- **ORM**: `DocumentORM` and `SurfConfigORM` updated.
- **Mappers**: `document_to_meta`, `document_to_orm`, `surf_config_to_params`, `params_to_surf_config` updated.
- **Go parser**: `DocumentMeta.Validate()` now checks `UpdateIntervalSec > 0`. `ParseSearchPage` and `ParseAdvertContent` propagate the field.
- **Tests**: All test factories and inline literals constructing `DocumentMeta`, `Document`, `Params`, or `SurfConfigORM` must include the new field.
