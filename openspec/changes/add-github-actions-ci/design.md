## Context

There is no automated CI pipeline. Developers run `make` targets locally, which is error-prone. Each of the three components (example-site, crawler-py, crawler-go) has its own `make` targets for fmt, lint, test-unit, and test-integration with a local docker-compose postgres setup. GitHub Actions service containers provide equivalent postgres infrastructure in CI.

## Goals / Non-Goals

**Goals:**
- Automated validation on every PR and push to main for all three components
- Build → lint → test-unit → test-integration pipeline for each component
- Parallel execution across components
- Codecov coverage tracking per component with unit/integration separation
- Path-based filtering so only affected components run on a given change

**Non-Goals:**
- Deploying on merge (this is CI-only, not CD)
- Running e2e tests
- Multi-platform matrix testing (Linux only, as per current dev environment)
- Adding `CODECOV_TOKEN` secret (requires human action in GitHub UI)

## Decisions

### Decision: Three separate workflow files

Each component gets its own `.github/workflows/ci-<name>.yml` file (ci-example-site.yml, ci-crawler-py.yml, ci-crawler-go.yml). They run in parallel as independent GitHub Actions jobs.

**Rationale:** User preference for maximum independence. Each workflow can be modified, disabled, or extended independently. GitHub UI shows three separate status checks per commit.

**Alternatives considered:**
- Single workflow file with three jobs: Cleaner to manage but less independent. Rejected per user preference.
- Reusable workflow (`_component-ci.yml` called three times): Most DRY but adds indirection. Rejected per user preference.

### Decision: Use `go-version-file` in setup-go

```yaml
- uses: actions/setup-go@v5
  with:
    go-version-file: example-site/go.mod  # or crawler/go.mod
    cache: true
```

**Rationale:** Reads the Go version directly from the module's go.mod file. No hardcoded version drift when go.mod is updated.

**Alternatives considered:**
- Hardcode `go 1.26`: Explicit but drifts when go.mod updates. Rejected.
- Read from `.go-version` file: Requires extra file to maintain. Rejected.

### Decision: Install Poetry via pip

```yaml
- name: Install Poetry
  run: python -m pip install poetry
- name: Install dependencies
  run: poetry install --no-interaction
```

**Rationale:** Standard pip install is simple and works in all CI environments without extra third-party actions.

**Alternatives considered:**
- `snok/install-poetry@v1`: More features (version pinning, virtualenv creation) but adds another third-party dependency. Rejected.
- `pipx install poetry`: Equivalent but pip is more universally available. pip chosen for simplicity.

### Decision: golangci-lint v2.12.2 via go install

```yaml
- name: Install golangci-lint
  run: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
```

**Rationale:** Matches the version used in the project's Makefile (`v2.12.2`). Installing via `go install` ensures the same version in CI as locally.

**Alternatives considered:**
- `golangci-lint/golangci-lint-action`: Bundled installation, version less visible. go install chosen for consistency with Makefile.

### Decision: Use GitHub Actions service container for postgres

```yaml
services:
  postgres:
    image: postgres:18-alpine
    env:
      POSTGRES_DB: <db-name>
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      5432:5432
    options: >-
      --health-cmd pg_isready
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5
```

**Rationale:** GitHub Actions native postgres service is faster and cleaner than docker-compose in CI. Uses the same `postgres:18-alpine` image as local development.

**Alternatives considered:**
- `docker compose up -d postgres`: Mirrors local dev exactly but adds docker-in-docker overhead and complexity in Actions. Rejected.
- `services: postgres` with custom healthcheck: Equivalent to chosen approach but more verbose. Rejected.

### Decision: Codecov with per-component flags and per-tier components

codecov.yml defines:
- 3 flags: `example-site`, `crawler-py`, `crawler-go`
- 6 components: `{component}-unit` and `{component}-integration` for each

**Rationale:** Flags give independent status checks in GitHub branch protection. Components give per-tier (unit vs integration) coverage breakdown in Codecov UI.

### Decision: Path filters per workflow

Each workflow has `paths:` filter matching only its relevant files plus its own workflow file.

**Rationale:** Prevents unnecessary workflow runs when unrelated components change. Running only affected components is faster and produces cleaner CI feedback.

### Decision: Concurrency cancel-in-progress

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

**Rationale:** Standard practice to prevent resource waste from stale runs. A new push to a branch cancels any in-progress run on that same ref.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| `CODECOV_TOKEN` secret not set, uploads silently fail | Document prerequisite in tasks. Workflow will still complete; codecov step will show as success but upload will fail silently. |
| golangci-lint version drift between local and CI | `go install` pins version; Makefile also pins. Keep them in sync when upgrading. |
| golangci-lint binary not on PATH after `go install` | Add `echo "$GOPATH/bin" >> "$GITHUB_PATH"` step after installing, or use full path in lint step. |
| Poetry version not pinned | `pip install poetry` gets latest stable. Consider pinning if behavior diverges across runs. |
| crawler migrations path overlap | Both crawler-py and crawler-go use `crawler/migrations/` but alembic (Python) vs golang-migrate (Go) are incompatible. Ensure path filters correctly separate: crawler-py triggers on `crawler/migrations/**` (alembic INI references it), crawler-go triggers on `crawler/parser/**` which has its own go.mod. |
| Python 3.13 availability in GitHub Actions | `setup-python` defaults to latest. Explicitly specify `python-version: '3.13'` to match pyproject.toml. |
| crawler-go integration tests do not exist yet | test-integration stage is a no-op that completes immediately; no postgres service or migrations are run for crawler-go. Integration tests will be added in a future change. |
| POSTGRES_HOSTS env var not set | Export `POSTGRES_HOSTS=localhost:5432` (example-site, crawler-go) and `POSTGRES__HOSTS=localhost:5432` (crawler-py) before running test commands. |

## Open Questions

No open questions that affect implementation. All technical decisions are made.
