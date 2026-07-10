DOCKER_COMPOSE := $(shell command -v docker-compose 2>/dev/null || echo 'docker compose')

.PHONY: setup-dev

setup-dev:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
	go install tool

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

.PHONY: example-site-run
example-site-run: export POSTGRES_HOSTS=localhost:5433
example-site-run: export API_PORT=8090
example-site-run: export METRICS_PORT=8091
example-site-run:
	cd example-site \
		&& $(DOCKER_COMPOSE) up -d --wait postgres \
		&& go run ./cmd/siteapi/main.go
