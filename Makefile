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

.PHONY: crawler-py-fmt crawler-py-lint crawler-py-test-unit crawler-py-test-integration

crawler-py-fmt:
	cd crawler && poetry run ruff check --fix .

crawler-py-lint:
	cd crawler && poetry run pytest -m linting -v

crawler-py-test-unit:
	cd crawler && poetry run pytest \
		--cov=shared --cov=surfer --cov=downloader \
		--cov-report=xml:coverage-py.xml \
		-m 'not linting and not integration' -v

crawler-py-test-integration: export POSTGRES__HOSTS=localhost:5534
crawler-py-test-integration:
	cd crawler \
		&& $(DOCKER_COMPOSE) up -d --wait postgres \
		&& poetry run pytest \
			--cov=shared --cov=surfer --cov=downloader \
			--cov-report=xml:coverage-py-int.xml \
			-m 'integration' -v

crawler-py-qa: crawler-py-lint crawler-py-test-unit crawler-py-test-integration

# ----------
# Migrations


crawler-py-migrate: export POSTGRES__HOSTS=localhost:5534
crawler-py-migrate:
	cd crawler && poetry run alembic upgrade head

crawler-py-migrate-new:
	cd crawler && poetry run alembic revision -m "$(name)"

crawler-py-migrate-history:
	cd crawler && poetry run alembic history --verbose

# ----------
# Surfer

.PHONY: run-surfer

run-surfer:
	cd crawler && poetry run python surfer

# ----------
# Downloader

.PHONY: run-downloader

run-downloader:
	cd crawler && poetry run python downloader

# ----------
# Crawler all services

.PHONY: crawler crawler-down
crawler:
	cd crawler && $(DOCKER_COMPOSE) up -d --build && $(DOCKER_COMPOSE) ps

crawler-down:
	cd crawler && $(DOCKER_COMPOSE) down -v
