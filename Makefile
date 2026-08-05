DOCKER_COMPOSE := $(shell command -v docker-compose 2>/dev/null || echo 'docker compose')

.PHONY: setup-dev

setup-dev:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
	go install tool
	(cd crawler && poetry sync)

# ----------
# Infra

.PHONY: infra infra-down

infra:
	cd infra/docker && $(DOCKER_COMPOSE) up -d && $(DOCKER_COMPOSE) ps

infra-down:
	cd infra/docker && $(DOCKER_COMPOSE) down -v

# ----------
# Example site

.PHONY: example-site-fmt example-site-lint example-site-test-unit example-site-test-integration example-site-qa

example-site-generate:
	cd example-site && go generate ./...

example-site-fmt:
	cd example-site && go fmt ./... && go fix ./...

example-site-lint:
	cd example-site && golangci-lint run ./...

example-site-test-unit:
	cd example-site && go tool gotestsum -- -v -race -coverprofile=coverage.out ./...

example-site-test-integration: export POSTGRES_HOSTS=localhost:5433
example-site-test-integration:
	cd example-site \
		&& $(DOCKER_COMPOSE) up -d --wait postgres \
		&& go tool gotestsum -- -tags=integration -v -race -coverprofile=coverage-int.out ./internal/infrastructure/repositories/...

example-site-qa: example-site-fmt example-site-lint example-site-test-unit example-site-test-integration

.PHONY: example-site example-site-down
example-site:
	cd example-site && $(DOCKER_COMPOSE) up -d --build && $(DOCKER_COMPOSE) ps

example-site-down:
	cd example-site && $(DOCKER_COMPOSE) down -v

# ----------
# Surfer (and downloader)

.PHONY: crawler-py-fmt crawler-py-lint crawler-py-test-unit

crawler-py-fmt:
	cd crawler && poetry run ruff check --fix .

crawler-py-lint:
	cd crawler && poetry run pytest -m linting -v

crawler-py-test-unit:
	cd crawler && poetry run pytest -m 'not linting and not integration' -v

crawler-py-qa: crawler-py-lint crawler-py-test-unit

# ----------