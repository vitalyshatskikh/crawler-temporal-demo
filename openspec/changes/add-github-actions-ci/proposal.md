## Why

Currently there is no automated validation of code pushed to `main` or opened as pull requests. Contributors must manually run `make` targets locally, which is error-prone and creates risk of breaking code reaching `main`. We need CI pipelines that run automatically on every MR and on every push to `main` to catch issues before they reach the default branch.

## What Changes

- Add `.github/workflows/ci-example-site.yml` — GitHub Actions workflow for the `example-site` Go service (external-site)
- Add `.github/workflows/ci-crawler-py.yml` — GitHub Actions workflow for the Python crawler packages (`surfer`, `downloader`, `shared`)
- Add `.github/workflows/ci-crawler-go.yml` — GitHub Actions workflow for the Go parser service (`crawler/parser`)
- Add `codecov.yml` at repo root — Codecov configuration with per-component flags and component coverage breakdowns
- Each workflow runs the pipeline: **build → lint → test-unit → test-integration** in sequence, with parallel execution across the three components
- Test jobs upload coverage reports to Codecov for the project's coverage dashboard

## Capabilities

### New Capabilities
- `ci-pipeline`: Introduces automated CI/CD pipeline capability for validating code on MR and push to `main`. Defines requirements for trigger events, job structure, coverage reporting, and parallelism across components.

## Impact

- New files: `.github/workflows/ci-example-site.yml`, `.github/workflows/ci-crawler-py.yml`, `.github/workflows/ci-crawler-go.yml`, `codecov.yml`
- New external dependency: Codecov (requires `CODECOV_TOKEN` secret to be added to the repository settings by a human)
- No changes to existing source code, APIs, or runtime behavior
- Go modules (`crawler`, `example-site`) and Python packages (`surfer`, `downloader`, `shared`) are unchanged
- Note: `ci-crawler-go.yml` runs build, lint, and test-unit stages for `crawler/parser`; the test-integration stage is a no-op until an infrastructure layer with postgres-backed repositories exists
