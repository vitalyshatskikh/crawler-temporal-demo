## Context

See proposal.md for motivation. The crawler stack runs three services (surfer, downloader, parser) that need to be containerized alongside the existing Temporal infrastructure and example-site.

Existing patterns in the project:
- `example-site/Dockerfile` — golang:1.26 builder + scratch final (Go multistage)
- `crawler/Dockerfile.migrate` — python:3.13-slim + Poetry (Python single-stage)
- `infra/docker/docker-compose.yml` — defines `temporal-network` as named external bridge
- `example-site/docker-compose.yml` — defines `example-site-network` as named external bridge
- `crawler/docker-compose.yml` — defines `crawler-network` as named bridge, only postgres + migrate

## Goals / Non-Goals

**Goals:**
- Containerize surfer, downloader, and parser services
- Follow existing project conventions for Dockerfiles and compose
- Maintain startup ordering (migrate must complete before workers start)
- Workers connect to Temporal via `temporal-network`

**Non-Goals:**
- Changing service behavior (code stays the same)
- Adding healthchecks (Temporal handles worker lifecycle)
- Altering local dev workflow (poetry run / go run still work)

## Decisions

### 1. Single Python image for surfer + downloader

Surfer and downloader share the same Python dependencies (via `pyproject.toml` and `shared/` package). Rather than two images with identical dependencies, build one `crawler-surfer:latest` image and differentiate at runtime via `command` override.

**Alternatives considered:**
- Separate Dockerfiles: duplicates dependency installation, doubles build time and image count
- Entrypoint scripts with CMD args: adds complexity for a simple override

### 2. Go multistage builder (golang:1.26 → scratch)

Follows the established `example-site/Dockerfile` pattern. Builder stage uses `golang:1.26` (Debian-based by default). Final stage is `scratch` with only the binary + CA certificates for HTTPS outbound calls.

### 3. External network references in crawler/docker-compose.yml

`temporal-network` and `example-site-network` are managed by other compose files. Attaching to them requires `external: true` with explicit `name` to avoid compose creating prefixed copies.

### 4. Namespace alignment (infra fix)

`infra/docker/docker-compose.yml` creates namespace `github.com/vitalyshatskikh/crawler-temporal-demo/crawler`. Worker defaults are `crawler` namespace. Align by changing infra's `DEFAULT_NAMESPACE` to `crawler`.

## Risks / Trade-offs

- **Network dependency**: crawler-compose cannot start without temporal-network running. This is by design — workers need Temporal to register.
- **Parser build context**: Since `crawler/` is the build context and `go.mod` lives at `crawler/go.mod`, the parser Dockerfile `COPY go.mod go.sum ./` and `COPY . .` paths are relative to `crawler/`.
- **No healthchecks**: Temporal SDK handles worker registration/heartbeating. If a worker crashes, Temporal marks it unhealthy after ~10s. Docker healthcheck is not needed.
