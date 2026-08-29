## Why

The crawler services (surfer, downloader, parser) currently run only as local processes during development. Adding Docker Compose support enables containerized deployment, aligning with the project's existing pattern (example-site already uses Dockerfile + docker-compose).

## What Changes

- Add `Dockerfile.surfer` — single-stage Python image for surfer + downloader
- Add `Dockerfile.parser` — multistage Go image (golang:1.26 builder → scratch final)
- Add `.dockerignore` to keep build contexts small
- Update `crawler/docker-compose.yml` with surfer, downloader, and parser services
- Update `infra/docker/docker-compose.yml` — fix Temporal namespace from full path to `crawler`

## Capabilities

This change is purely infrastructure / deployment. No spec-level behavior changes. The surfer, downloader, and parser remain unchanged — only their deployment method changes.

### New Capabilities

(none — tooling/deployment only)

### Modified Capabilities

(none — no requirement changes)

## Impact

- **New files**: `Dockerfile.surfer`, `Dockerfile.parser`, `.dockerignore`
- **Modified files**: `crawler/docker-compose.yml`, `infra/docker/docker-compose.yml`
- **Networks**: crawler services attach to existing `temporal-network` and `example-site-network`
- **Dependencies**: No new runtime dependencies; Poetry and Go toolchains unchanged
