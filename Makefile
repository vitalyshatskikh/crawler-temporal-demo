.PHONY: setup-dev

setup-dev:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

.PHONY: example-site-fmt example-site-lint example-site-test-unit example-site-qa

example-site-generate:
	cd example-site && go generate ./...

example-site-fmt:
	cd example-site && go fmt ./... && go fix ./...

example-site-lint:
	cd example-site && golangci-lint run ./...

example-site-test-unit:
	cd example-site && go tool gotestsum -- -v -race -coverprofile=coverage.out ./...

example-site-qa: example-site-fmt example-site-lint example-site-test-unit
