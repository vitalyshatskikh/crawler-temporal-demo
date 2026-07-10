.PHONY: setup-dev

setup-dev:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

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
