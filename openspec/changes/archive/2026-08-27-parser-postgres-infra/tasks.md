## 1. Setup

- [x] 1.1 Add `github.com/jackc/pgx/v5`, `github.com/vitalyshatskikh/go-lib`, `go.uber.org/zap` to `crawler/go.mod` require block
- [x] 1.2 Add `github.com/sqlc-dev/sqlc/cmd/sqlc` to the `tool` directive in `crawler/go.mod`
- [x] 1.3 Run `cd crawler && go mod tidy` to populate `go.sum`
- [x] 1.4 Verify `sqlc` binary is available (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` if not)

## 2. Schema and SQLC configuration

- [x] 2.1 Create `crawler/parser/db/schema.sql` with `CREATE TABLE parsing_configs (...)` and `CREATE TABLE documents (...)` matching the existing Alembic definitions verbatim:
  - `parsing_configs`: `id BIGSERIAL PRIMARY KEY`, `source_id TEXT NOT NULL`, `doc_type TEXT NOT NULL`, `name TEXT NOT NULL`, `config JSONB NOT NULL DEFAULT '{}'::jsonb`
  - `parsing_configs`: index `idx_parsing_configs_source_id ON parsing_configs (source_id)`
  - `parsing_configs`: unique constraint `uq_parsing_configs_source_doc_type UNIQUE (source_id, doc_type)`
  - `documents`: `sdoc_id TEXT NOT NULL`, `source_id TEXT NOT NULL`, `doc_type TEXT NOT NULL`, `external_url TEXT NOT NULL`, `body TEXT NOT NULL`, `created_at TIMESTAMPTZ NOT NULL`, `updated_at TIMESTAMPTZ NOT NULL`, `update_interval_sec INTEGER NOT NULL DEFAULT 86400`
  - `documents`: composite PK `pk_documents PRIMARY KEY (sdoc_id, source_id, doc_type)`
  - `documents`: check constraint `ck_documents_updated_at CHECK (updated_at >= created_at)`
  - `documents`: indexes `idx_documents_source_id ON documents (source_id)` and `idx_documents_doc_type ON documents (doc_type)`
- [x] 2.2 Create `crawler/parser/db/queries/parsing_configs.sql` with a `-- name: GetParsingConfig :one` query selecting by `source_id` and `doc_type`, returning all columns
- [x] 2.3 Create `crawler/parser/db/queries/documents.sql` with:
  - `-- name: GetDocument :one` selecting by `(sdoc_id, source_id, doc_type)` composite PK
  - `-- name: UpsertDocument :exec` doing `INSERT INTO documents ... ON CONFLICT (sdoc_id, source_id, doc_type) DO UPDATE SET external_url = EXCLUDED.external_url, body = EXCLUDED.body, updated_at = EXCLUDED.updated_at, update_interval_sec = EXCLUDED.update_interval_sec`
- [x] 2.4 Create `crawler/parser/db/sqlc.yaml` with:
  - `engine: postgresql`
  - `sql_package: "pgx/v5"`
  - `emit_json_tags: true`
  - `emit_empty_slices: true`
  - `output: "gen"`
  - `package: "queries"`
  - `schema: "schema.sql"`
  - `queries: "queries/"`
- [x] 2.5 Run `cd crawler/parser/db && sqlc generate` and verify `db/gen/` is produced with `db.go`, `models.go`, `parsing_configs.sql.go`, `documents.sql.go`

## 3. Repository implementations

- [x] 3.1 Create `crawler/parser/internal/infrastructure/repositories/pg_config_repo.go`:
  - `NewPGConfigRepo(pool *pgxpool.Pool) *PGConfigRepo`
  - Implement `GetConfig(ctx, sourceID, docType)` calling sqlc's `GetParsingConfig`, scanning `config` JSONB into `[]byte`, unmarshaling into `[]domain.ParsingParam`, returning `domain.ParsingConfig`
  - Define `var ErrUnmarshalConfig = errors.New("unmarshal parsing_configs.config")`; return `fmt.Errorf("%w: %w", ErrUnmarshalConfig, err)` on JSON decode failure
  - Compile-time assertion: `_ domain.ConfigRepository = (*PGConfigRepo)(nil)`
- [x] 3.2 Create `crawler/parser/internal/infrastructure/repositories/pg_adverts_repo.go`:
  - `NewPGAdvertsRepo(pool *pgxpool.Pool) *PGAdvertsRepo`
  - Implement `GetDocument(ctx, sdocID, sourceID, docType)` calling sqlc's `GetDocument`, converting `Body string` from sqlc row to `[]byte`, mapping result to `domain.Document`
  - Implement `SaveDocument(ctx, doc)` calling sqlc's `UpsertDocument`, converting `doc.Body []byte` to `string`
  - Compile-time assertion: `_ application.AdvertsRepository = (*PGAdvertsRepo)(nil)`

## 4. Test helper

- [x] 4.1 Create `crawler/parser/internal/infrastructure/repositories/testutil/pg_test_helper.go`:
  - `//go:build integration` build tag
  - Expose `TestPool *pgxpool.Pool` variable
  - `Setup()` function: `config.Load()` (from go-lib) → admin pool → `CREATE DATABASE parser-{pid}` → pool to test DB → read `../../../../db/schema.sql` via `runtime.Caller` and `os.ReadFile` → `pool.Exec` the schema → return nil
  - `Teardown()` function: close pool → `SELECT pg_terminate_backend` → `DROP DATABASE`
- [x] 4.2 Create `crawler/parser/internal/infrastructure/repositories/integration_test.go`:
  - `//go:build integration` build tag
  - Single `TestMain(m *testing.M)` that calls `testutil.Setup()`, then `os.Exit(m.Run())`, deferring `testutil.Teardown()`
  - This file provides the test entry point; `pg_config_repo_integration_test.go` and `pg_adverts_repo_integration_test.go` contain only test functions and helpers

## 5. Integration tests

- [x] 5.1 Create `crawler/parser/internal/infrastructure/repositories/pg_config_repo_integration_test.go`:
  - `//go:build integration`
  - Raw SQL helper: `insertParsingConfig(t, pool, id, sourceID, docType, name, paramsJSON)`
  - Tests: `WhenRowExists`, `WhenNoRowForDocType`, `WhenSameSourceDifferentDocType`
- [x] 5.2 Create `crawler/parser/internal/infrastructure/repositories/pg_adverts_repo_integration_test.go`:
  - `//go:build integration`
  - Raw SQL helper: `insertDocument(t, pool, doc)`
  - Tests: `GetDocument_WhenFound`, `GetDocument_WhenNotFound`, `SaveDocument_WhenNew`, `SaveDocument_WhenExisting`, `SaveDocument_WhenOnlyBodyDiffers`

## 6. Verification

- [x] 6.1 Run `cd crawler && go vet ./parser/...` — no errors
- [x] 6.2 Run `cd crawler && go build ./parser/...` — builds cleanly
- [x] 6.3 Run `cd crawler && go test ./parser/internal/infrastructure/repositories/...` (unit-only, no integration tag) — should skip
- [x] 6.4 Run `make crawler-go-test-integration` — all integration tests pass
- [x] 6.5 Verify `golangci-lint run ./parser/...` passes (no new lints introduced)
