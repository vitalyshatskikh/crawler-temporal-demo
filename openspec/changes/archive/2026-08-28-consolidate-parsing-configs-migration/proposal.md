## Why

The `parsing_configs` table was created across two Alembic migrations: `0003_create_parsing_configs.py` (creates the table with 5 columns) and `0007_add_parsing_config_url_fields.py` (adds 3 URL-related columns). During local development where database recreation is acceptable, this two-step creation is unnecessary complexity in the migration chain.

## What Changes

- Edit `0003_create_parsing_configs.py` to define `parsing_configs` with all 6 columns in a single `CREATE TABLE` statement.
- Delete `0007_add_parsing_config_url_fields.py` entirely.
- The migration downgrade remains unchanged — `DROP TABLE` removes all columns atomically.
- No change to `0004_create_documents.py` (its `down_revision = "0003"` is already correct).
- No change to `downloader/infrastructure/db/orm/parsing_configs.py` (ORM already declares all 6 columns).
- No change to `parser/db/schema.sql` (already the reference schema with all 6 columns).

## Capabilities

This is a pure refactoring with no behavior changes and no new or modified capability requirements.

## Impact

- **Alembic migrations**: Migration chain reduced from 7 → 6 files; 0007 removed.
- **Local dev workflow**: `alembic upgrade head` produces the same final schema in one step instead of two.
- **No breaking changes**: Downgrade still drops the table cleanly; upgrade produces identical schema.
- **CI**: Integration tests spin up fresh databases via `docker compose` and run all migrations from base — they will pick up the consolidated migration automatically.
