## Why

The parser worker (`crawler/parser`) needs to persist documents and lookup parsing configuration, but currently has no implementation of the repository interfaces defined in `domain.ConfigRepository` and `application.AdvertsRepository`. The Go-based parser must have a PostgreSQL-backed infrastructure layer to integrate with the existing database schema used by the Python services.

## What Changes

- Add `PGConfigRepository` implementing `domain.ConfigRepository` — looks up `parsing_configs` rows by `(source_id, doc_type)`.
- Add `PGAdvertsRepository` implementing `application.AdvertsRepository` — gets and upserts `documents` rows.
- Use `pgx/v5` connection pool via `github.com/vitalyshatskikh/go-lib` (`postgres.NewPGXPool`).
- Use `sqlc` for type-safe SQL access with `pgx/v5` driver.
- Generated sqlc code lives at `crawler/parser/db/gen/` and is committed (not gitignored).
- Schema applied manually via `crawler/parser/db/schema.sql` (no golang-migrate in tests).
- Integration tests with raw SQL test helpers following the `example-site` pattern.
- No changes to `cmd/parser/main.go`.

## Capabilities

### New Capabilities

- `parser-infra`: PostgreSQL-backed repository implementations for the parser worker. Defines how `PGConfigRepository` resolves `ParsingConfig` by `(source_id, doc_type)` from `parsing_configs`, how `PGAdvertsRepository` persists and retrieves `Document` rows from `documents`, and the JSONB shape for the `config` column (just `[]ParsingParam`).

### Modified Capabilities

<!-- No spec-level requirement changes. parser-activities and documents specs define behavior that this implementation satisfies. -->

## Impact

- New Go dependencies in `crawler/go.mod`: `github.com/jackc/pgx/v5`, `github.com/vitalyshatskikh/go-lib`, `go.uber.org/zap`.
- New tool dependency: `github.com/sqlc-dev/sqlc/cmd/sqlc` in the `tool` directive.
- New directory structure: `crawler/parser/db/` (schema, sqlc config, queries, generated code).
- New package: `crawler/parser/internal/infrastructure/repositories/` with two repo implementations and a test helper.
- No API or behavioral changes to existing specs.
