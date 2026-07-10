.PHONY: setup-dev

setup-dev:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: external-site-fmt external-site-lint external-site-test-unit external-site-qa

external-site-generate:
	cd external-site && go generate ./...

external-site-fmt:
	cd external-site && go fmt ./... && go fix ./...

external-site-lint:
	cd external-site && golangci-lint run ./...

external-site-test-unit:
	cd external-site && go test -v -race ./...

external-site-qa: external-site-fmt external-site-lint external-site-test-unit
