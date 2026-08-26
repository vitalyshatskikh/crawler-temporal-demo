## 1. Prerequisite

- [ ] 1.1 Add `CODECOV_TOKEN` as a repository secret in GitHub UI (Settings → Secrets and variables → Actions). The codecov-action upload step will fail silently without it; coverage reports will still be generated and the workflow will pass.

## 2. Codecov Configuration

- [x] 2.1 Create `codecov.yml` at repo root with 3 flags (example-site, crawler-py, crawler-go) and 6 components (unit + integration per flag)

## 3. ci-example-site.yml

- [x] 3.1 Create `.github/workflows/ci-example-site.yml`
- [x] 3.2 Configure trigger: push to main, pull_request targeting main, workflow_dispatch; paths filter: `example-site/**`, `go.work`, `go.work.sum`, `.github/workflows/ci-example-site.yml`
- [x] 3.3 Add concurrency: cancel-in-progress
- [x] 3.4 Add postgres service container (postgres:18-alpine, POSTGRES_DB=example-site)
- [x] 3.5 Add steps: checkout, setup-go (go-version-file: example-site/go.mod, cache), export POSTGRES_HOSTS=localhost:5432, build (go build ./...), lint (golangci-lint v2.12.2, add $GOPATH/bin to PATH via `echo "$GOPATH/bin" >> "$GITHUB_PATH"`), test-unit (gotestsum -race -coverprofile=coverage.out), test-integration (migrate up, gotestsum -tags=integration -race -coverprofile=coverage-int.out)
- [x] 3.6 Add codecov upload step (files: coverage.out, coverage-int.out; flag: example-site)

## 4. ci-crawler-py.yml

- [x] 4.1 Create `.github/workflows/ci-crawler-py.yml`
- [x] 4.2 Configure trigger: push to main, pull_request targeting main, workflow_dispatch; paths filter: `crawler/**` excluding `crawler/parser/**`, plus `crawler/pyproject.toml`, `crawler/poetry.lock`, `crawler/alembic.ini`, `crawler/Dockerfile.migrate`, `crawler/migrations/**`, `.github/workflows/ci-crawler-py.yml`
- [x] 4.3 Add concurrency: cancel-in-progress
- [x] 4.4 Add postgres service container (postgres:18-alpine, POSTGRES_DB=crawler)
- [x] 4.5 Add steps: checkout, setup-python (3.13), install poetry (pip install poetry), export POSTGRES__HOSTS=localhost:5432, poetry install, lint (poetry run pytest -m linting -v), test-unit (pytest --cov=shared --cov=surfer --cov=downloader --cov-report=xml:coverage-py.xml -m 'not linting and not integration'), test-integration (alembic upgrade head, pytest --cov=shared --cov=surfer --cov=downloader --cov-report=xml:coverage-py-int.xml -m integration)
- [x] 4.6 Add codecov upload step (files: coverage-py.xml, coverage-py-int.xml; flag: crawler-py)

## 5. ci-crawler-go.yml

- [x] 5.1 Create `.github/workflows/ci-crawler-go.yml`
- [x] 5.2 Configure trigger: push to main, pull_request targeting main, workflow_dispatch; paths filter: `crawler/parser/**`, `crawler/go.mod`, `go.work`, `go.work.sum`, `.github/workflows/ci-crawler-go.yml`
- [x] 5.3 Add concurrency: cancel-in-progress
- [x] 5.4 (No postgres service container for crawler-go — integration tests do not exist yet)
- [x] 5.5 Add steps: checkout, setup-go (go-version-file: crawler/go.mod, cache), build (go build ./...), lint (golangci-lint v2.12.2, add $GOPATH/bin to PATH via `echo "$GOPATH/bin" >> "$GITHUB_PATH"`), test-unit (gotestsum -race -coverprofile=coverage.out), test-integration (no-op: `echo 'no integration tests for crawler-go yet' && true`)
- [x] 5.6 Add codecov upload step (files: coverage.out; flag: crawler-go)
