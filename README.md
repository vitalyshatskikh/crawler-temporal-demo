# Crawler

[![CI Crawler Python](https://github.com/vitalyshatskikh/crawler-temporal-demo/actions/workflows/ci-crawler-py.yml/badge.svg)](https://github.com/vitalyshatskikh/crawler-temporal-demo/actions/workflows/ci-crawler-py.yml)
[![CI Crawler Go](https://github.com/vitalyshatskikh/crawler-temporal-demo/actions/workflows/ci-crawler-go.yml/badge.svg)](https://github.com/vitalyshatskikh/crawler-temporal-demo/actions/workflows/ci-crawler-go.yml)
[![CI Example Site](https://github.com/vitalyshatskikh/crawler-temporal-demo/actions/workflows/ci-example-site.yml/badge.svg)](https://github.com/vitalyshatskikh/crawler-temporal-demo/actions/workflows/ci-example-site.yml)

Sample web crawing/scraping app

```mermaid
sequenceDiagram
```

## Prerequisites

- docker, docker-compose
- python
- go
- golangci-lint

## Run

### Temporal services

From project's root

```bash
(cd infra/docker && docker-compose up -d)
```

or 

```bash
(cd infra/docker && docker compose up -d)
```

The Temporal Web UI will be available at http://localhost:8080

<details>
<summary>To clear:</summary>

```bash
(cd infra/docker && docker-compose down -v)
```

or

```bash
(cd infra/docker && docker compose down -v)
```

</details>

## Develop

TBD

## Links

- [samples-server/compose](https://github.com/temporalio/samples-server/tree/main/compose) - Run Temporal services using docker-compose