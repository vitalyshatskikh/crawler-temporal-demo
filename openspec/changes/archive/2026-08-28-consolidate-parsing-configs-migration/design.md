## Context

The `parsing_configs` table is currently created in two Alembic migrations: `0003_create_parsing_configs.py` creates the table with 5 columns; `0007_add_parsing_config_url_fields.py` adds 3 URL-related columns via `ALTER TABLE`. The end state matches `parser/db/schema.sql`.

Since DB recreation is acceptable during local development, consolidating into a single `CREATE TABLE` with all 6 columns is safe and reduces migration chain complexity.

## Goals / Non-Goals

**Goals:**
- Single migration step for `parsing_configs` table creation
- Final schema identical to current production state
- No behavior change for any application component

**Non-Goals:**
- No migration history preservation (0007 is deleted)
- No spec or requirement changes

## Decisions

### Decision: Inline columns into 0003 instead of keeping 0007 as a no-op

**Choice:** Edit `0003_create_parsing_configs.py` to add the three columns directly in `create_table`, delete `0007_add_parsing_config_url_fields.py`.

**Rationale:** A single `CREATE TABLE` with all columns is cleaner than a create + alter pair. Downgrade (`DROP TABLE`) already removes all columns atomically — no change needed to `downgrade()`.

**Alternatives considered:**
- Keep 0007 as a stub/empty migration — adds file with no value during local dev phase.
- Keep both migrations but stamp 0007's columns in 0003 — leaves orphaned migration file that no longer reflects applied state on a fresh DB.

### Decision: Reference `parser/db/schema.sql` as the target schema

**Choice:** The consolidated `parsing_configs` definition matches `parser/db/schema.sql` exactly.

**Rationale:** `parser/db/schema.sql` is the canonical Go reference schema. Aligning Python migration with it ensures consistency across the codebase.

## Risks / Trade-offs

- **Risk:** 0007 deletion loses migration history. **Mitigation:** Only acceptable during local dev phase (per user confirmation). Production environments should never run this change without a proper migration path.
- **Risk:** CI integration tests may have migrated past 0007. **Mitigation:** Tests use `docker compose up -d --wait postgres` which creates a fresh volume; they will run all migrations from base including the consolidated 0003.
