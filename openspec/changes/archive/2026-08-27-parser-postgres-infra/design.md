## Context

See proposal.md. The parser worker (`crawler/parser`) has domain interfaces (`ConfigRepository`, `AdvertsRepository`) but no PostgreSQL implementation. The existing Python services use Alembic migrations to manage `parsing_configs` and `documents` tables. The Go parser must share the same schema. The `example-site` service already has a working `PGAdvertsRepo` using `pgx/v5` and `go-lib`'s `postgres.NewPGXPool` — use it as the pattern to follow.

## Goals / Non-Goals

**Goals:**
- Implement `PGConfigRepository` and `PGAdvertsRepository` using `pgx/v5` with a connection pool from `go-lib`.
- Use `sqlc` for type-safe SQL code generation (not hand-rolled SQL strings).
- Apply the database schema via a single `schema.sql` file in tests (no golang-migrate).
- Provide integration tests that create and tear down an isolated test database per test run.
- Commit generated sqlc code at `crawler/parser/db/gen/`.

**Non-Goals:**
- Wiring the repositories into `cmd/parser/main.go` (left for a follow-up).
- Changing `parser-activities` or `documents` spec behavior.
- Supporting connection pooling configurations beyond what `go-lib` provides.
- Using an ORM or query builder (e.g., no `goqu` like example-site uses).

## Decisions

### Decision: sqlc over hand-rolled SQL or query builder

**Chosen:** sqlc generates type-safe Go from SQL queries.

**Rationale:** The user explicitly requested sqlc. sqlc catches type mismatches at compile time and produces idiomatic Go code. Compared to `goqu` (used in example-site), sqlc keeps SQL visible and explicit rather than hiding it behind a DSL. Compared to raw `pgx` calls, sqlc eliminates boilerplate row-scanning code.

**Alternatives considered:**
- `goqu`: A SQL builder that composes queries in Go. More flexible but hides SQL semantics and requires runtime query building. Example-site uses it, but user requested sqlc.
- Raw `pgx`: Direct `Conn.Query()` / `Conn.Exec()` calls. Fully flexible but verbose and type-unsafe. sqlc's generated code is more concise.

### Decision: Generated code at `db/gen`, committed to git

**Chosen:** Generated sqlc output lives at `crawler/parser/db/gen/` and is committed.

**Rationale:** The generated code is deterministic given the SQL source files. Committing it means:
- No sqlc binary required to build (just to regenerate after editing queries).
- CI does not need to run `sqlc generate`.
- `git diff` on query changes shows both the `.sql` file and the resulting generated `.go` file, aiding review.
- sqlc's `package: "queries"` config is set in `sqlc.yaml`, so generated code lives at `crawler/parser/db/gen` with package name `queries` (no import alias needed).

**Alternatives considered:**
- Gitignore generated code, run `go generate` on every build: Forces sqlc installation on all developers and in CI, slows initial build.
- Generate on demand via `go generate` only: Commits the responsibility to developers to regenerate after editing queries; easy to forget.

### Decision: JSONB config column stores just `[]ParsingParam`

**Chosen:** `parsing_configs.config` JSONB stores an array of `{name, jmespath, default}` objects only.

**Rationale:** The structured columns (`id`, `source_id`, `doc_type`, `name`) are stored in their own SQL columns. The JSONB column holds only the dynamic `params` array. This avoids duplicating `source_id`/`doc_type` in both the SQL schema and the JSONB, and keeps the payload minimal.

**Example:**
```json
[{"name": "external_url", "jmespath": "urls[*]", "default": ""}]
```

**Alternatives considered:**
- Store the entire `ParsingConfig` (including `id`, `source_id`, `doc_type`, `name`) in JSONB alongside the structured columns: Creates two sources of truth for the same data.
- Wrap in `{"params": [...]}`: Adds a nesting level for a single field.

### Decision: `SaveDocument` is an upsert via `ON CONFLICT DO UPDATE`

**Chosen:** `INSERT INTO documents ... ON CONFLICT (sdoc_id, source_id, doc_type) DO UPDATE SET external_url = EXCLUDED.external_url, body = EXCLUDED.body, updated_at = EXCLUDED.updated_at, update_interval_sec = EXCLUDED.update_interval_sec`.

**Rationale:** Parser re-runs may revisit the same `(sdoc_id, source_id, doc_type)` tuple when re-parsing the same URL. Idempotent upserts handle this gracefully without requiring the caller to check for existence first. `updated_at` is always refreshed on update. The `created_at` is preserved from the original insert (not overwritten on update). The example-site `UpsertAdvert` uses the same pattern.

**Alternatives considered:**
- Strict insert, error on conflict: Forces upstream code to handle "already exists" cases, brittle for re-runs.
- Update-only: Not valid for new documents.

### Decision: Test helper applies `schema.sql` directly (no golang-migrate)

**Chosen:** The test helper reads `crawler/parser/db/schema.sql` and calls `pool.Exec()` to run it.

**Rationale:** The user explicitly requested manual application of `schema.sql`. golang-migrate brings in a second migration system alongside the existing Alembic migrations in `crawler/migrations/versions/`. Since the schema is canonical in `db/schema.sql` (which mirrors the Alembic definitions), applying it directly in tests is simpler and avoids migration-tooling conflicts.

**Alternatives considered:**
- golang-migrate in tests: Adds a second migration tool. The schema already exists in Alembic form; a parallel golang-migrate setup would require keeping two sets of migrations in sync.
- Run Alembic in tests: Would require the Python runtime and Alembic to be available during Go integration tests.

### Decision: `body` column stored as PostgreSQL `text`, decoded to `[]byte` in Go

**Chosen:** `documents.body` is PostgreSQL `text`. sqlc generates Go `string` for the column. The repository converts `string` to `[]byte` on read and `string(doc.Body)` on write.

**Rationale:** Matches the Alembic migration (`0004_create_documents.py`) which uses `sa.Text()`. The parser only produces HTML/JSON bodies, which are valid UTF-8. `[]byte(string)` is safe for these content types. If the parser ever needs to store arbitrary binary data, a migration to `bytea` would be required.

### Decision: ErrUnmarshalConfig sentinel for JSON decode errors

**Chosen:** Define `var ErrUnmarshalConfig = errors.New("unmarshal parsing_configs.config")` in the repository package. On JSON decode failure in `GetConfig`, return `fmt.Errorf("%w: %w", ErrUnmarshalConfig, err)`.

**Rationale:** Callers can use `errors.Is(err, ErrUnmarshalConfig)` rather than string-matching on an error message. Keeps error handling testable and refactor-safe.

## Risks / Trade-offs

- **Schema drift between Alembic and `schema.sql`**: The Go tests use `db/schema.sql` which must be manually kept in sync with `crawler/migrations/versions/`. If the Alembic migrations change, `schema.sql` must be updated manually.
  - **Mitigation**: `schema.sql` mirrors exactly what the Alembic up migrations produce (enumerated in tasks.md 2.1). Keep both in sync when modifying the schema.

- **sqlc JSONB handling**: sqlc with `pgx/v5` returns JSONB as `[]byte`. Unmarshaling to `[]ParsingParam` in the repo layer is extra work but keeps the sqlc mapping simple.
  - **Mitigation**: Repo code unmarshals `[]byte` into `[]domain.ParsingParam` and surfaces decode errors via `ErrUnmarshalConfig`.

- **Test database isolation**: Integration tests create a DB per process (`parser-{pid}`). If tests run in parallel from multiple processes, DB names could collide.
  - **Mitigation**: PID-based naming is sufficient for the current single-process test runner. If parallel package testing is needed, add a random suffix.

- **Generated code size**: sqlc generates several files per query file. Committing them inflates the git history slightly.
  - **Mitigation**: The generated code is small (~200 lines total). The benefit of reproducible builds outweighs this.

- **`body` column type assumption**: `documents.body` is `text` in PostgreSQL, which stores UTF-8 strings. `[]byte(string)` is safe for HTML/JSON content but would corrupt arbitrary binary data.
  - **Mitigation**: The parser only produces HTML/JSON bodies. If binary content is needed in the future, migrate `body` to `bytea`.
