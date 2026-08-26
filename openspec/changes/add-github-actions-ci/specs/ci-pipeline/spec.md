## Purpose

Automates validation of code changes by running build, lint, and test pipelines on every pull request and push to main. Ensures code quality and test coverage are maintained as the codebase evolves.

## ADDED Requirements

### Requirement: Workflow trigger events

The CI pipeline SHALL trigger on push to the `main` branch and on pull requests targeting the `main` branch. The pipeline SHALL also support manual triggering via workflow_dispatch.

#### Scenario: Push to main branch
- **WHEN** a commit is pushed directly to the `main` branch
- **THEN** all three component workflows SHALL run in parallel

#### Scenario: Pull request opened against main
- **WHEN** a pull request is opened targeting the `main` branch
- **THEN** all three component workflows SHALL run in parallel

#### Scenario: Manual workflow dispatch
- **WHEN** a user triggers the workflow via GitHub UI workflow_dispatch
- **THEN** all three component workflows SHALL run in parallel

### Requirement: Sequential pipeline stages

Each component workflow SHALL execute stages in strict order: build, then lint, then test-unit, then test-integration. A stage SHALL NOT execute if any preceding stage has failed.

#### Scenario: Build failure prevents subsequent stages
- **WHEN** the build stage fails for a component
- **THEN** lint, test-unit, and test-integration SHALL NOT run for that component

#### Scenario: Successful build allows lint to run
- **WHEN** the build stage succeeds for a component
- **THEN** the lint stage SHALL execute

#### Scenario: Successful lint allows test-unit to run
- **WHEN** the lint stage succeeds for a component
- **THEN** the test-unit stage SHALL execute

#### Scenario: Successful test-unit allows test-integration to run
- **WHEN** the test-unit stage succeeds for a component
- **THEN** the test-integration stage SHALL execute

### Requirement: Component job parallelism

The three component workflows (example-site, crawler-py, crawler-go) SHALL run in parallel without inter-job dependencies. Failure in one component SHALL NOT affect the execution of other components.

#### Scenario: One component fails while others succeed
- **WHEN** crawler-go fails but example-site and crawler-py succeed
- **THEN** example-site and crawler-py SHALL complete successfully and report their results independently

### Requirement: Integration test database

The test-integration stage for each component SHALL use a postgres service container. The container SHALL use the `postgres:18-alpine` image and be accessible at `localhost:5432`.

#### Scenario: Crawler-py integration tests with postgres
- **WHEN** the crawler-py test-integration stage runs
- **THEN** a postgres:18-alpine service container SHALL be started on port 5432
- **AND** alembic migrations SHALL be applied before running integration tests

#### Scenario: Crawler-go integration tests with postgres
- **WHEN** the crawler-go test-integration stage runs
- **THEN** the stage SHALL complete as a no-op because `crawler/parser/internal/infrastructure/repositories/` contains no integration tests
- **NOTE**: Integration tests for crawler-go are deferred until an infrastructure layer with a postgres-backed repository exists

#### Scenario: Example-site integration tests with postgres
- **WHEN** the example-site test-integration stage runs
- **THEN** a postgres:18-alpine service container SHALL be started on port 5432
- **AND** golang-migrate migrations SHALL be applied before running integration tests

### Requirement: Coverage report generation

The test stages SHALL produce coverage reports in codecov-compatible formats. Go test stages SHALL produce coverage files at `example-site/coverage.out` and `example-site/coverage-int.out` for example-site, and `crawler/coverage.out` for crawler-go. Python pytest stages SHALL produce coverage files at `crawler/coverage-py.xml` and `crawler/coverage-py-int.xml`.

#### Scenario: Go unit tests produce coverage
- **WHEN** the test-unit stage runs for example-site
- **THEN** a file named `example-site/coverage.out` SHALL be produced with Go coverage profile data
- **WHEN** the test-unit stage runs for crawler-go
- **THEN** a file named `crawler/coverage.out` SHALL be produced with Go coverage profile data

#### Scenario: Go integration tests produce separate coverage
- **WHEN** the test-integration stage runs for example-site
- **THEN** a file named `example-site/coverage-int.out` SHALL be produced with Go coverage profile data
- **NOTE**: crawler-go has no integration tests; no `coverage-int.out` is produced for crawler-go

#### Scenario: Python unit tests produce coverage XML
- **WHEN** the test-unit stage runs for crawler-py
- **THEN** a file named `coverage-py.xml` SHALL be produced with pytest-cov XML data

#### Scenario: Python integration tests produce separate coverage XML
- **WHEN** the test-integration stage runs for crawler-py
- **THEN** a file named `coverage-py-int.xml` SHALL be produced with pytest-cov XML data

### Requirement: Codecov upload

Each component workflow SHALL upload its coverage reports to Codecov using the official codecov-action. Uploads SHALL include a flag identifying the component and SHALL use the latest stable codecov-action version.

#### Scenario: Codecov upload on example-site workflow
- **WHEN** the example-site workflow completes test stages
- **THEN** coverage files `example-site/coverage.out` and `example-site/coverage-int.out` SHALL be uploaded to Codecov with flag `example-site`

#### Scenario: Codecov upload on crawler-py workflow
- **WHEN** the crawler-py workflow completes test stages
- **THEN** coverage files `crawler/coverage-py.xml` and `crawler/coverage-py-int.xml` SHALL be uploaded to Codecov with flag `crawler-py`

#### Scenario: Codecov upload on crawler-go workflow
- **WHEN** the crawler-go workflow completes test stages
- **THEN** coverage file `crawler/coverage.out` SHALL be uploaded to Codecov with flag `crawler-go`
- **NOTE**: crawler-go has no integration tests; only unit coverage (`coverage.out`) is uploaded

### Requirement: Path-based job filtering

Each workflow SHALL only run when files relevant to that component are changed, or when the workflow file itself changes.

#### Scenario: Example-site workflow runs on relevant changes
- **WHEN** files under `example-site/` or `go.work` or `go.work.sum` are changed
- **THEN** the example-site workflow SHALL run

#### Scenario: Crawler-py workflow runs on relevant changes
- **WHEN** files under `crawler/` excluding `crawler/parser/` are changed
- **OR WHEN** `crawler/pyproject.toml`, `crawler/poetry.lock`, `crawler/alembic.ini`, `crawler/Dockerfile.migrate`, or `crawler/migrations/` are changed
- **THEN** the crawler-py workflow SHALL run

#### Scenario: Crawler-go workflow runs on relevant changes
- **WHEN** files under `crawler/parser/` are changed
- **OR WHEN** `crawler/go.mod` or `go.work` or `go.work.sum` are changed
- **THEN** the crawler-go workflow SHALL run

### Requirement: Codecov configuration

The repository root SHALL contain a `codecov.yml` file that defines separate flags for each component and separate components for unit vs integration coverage.

#### Scenario: Codecov flags are defined
- **WHEN** Codecov processes uploads from this repository
- **THEN** three flags SHALL exist: `example-site`, `crawler-py`, `crawler-go`
- **AND** six components SHALL exist distinguishing unit from integration coverage per flag

### Requirement: Workflow concurrency control

Each workflow SHALL cancel in-progress runs when a new run is triggered for the same ref, preventing redundant CI consumption.

#### Scenario: New push cancels previous run
- **WHEN** a new push arrives for a branch with an existing in-progress CI run
- **THEN** the previous run SHALL be cancelled and the new run SHALL execute
