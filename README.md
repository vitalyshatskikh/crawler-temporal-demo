Crawler, Python code:
[![CI Crawler Python](https://github.com/vitalyshatskikh/crawler-temporal-demo/actions/workflows/ci-crawler-py.yml/badge.svg)](https://github.com/vitalyshatskikh/crawler-temporal-demo/actions/workflows/ci-crawler-py.yml)
[![codecov](https://codecov.io/gh/vitalyshatskikh/crawler-temporal-demo/graph/badge.svg?token=3HS6Z0KAR5&flag=crawler-py)](https://codecov.io/gh/vitalyshatskikh/crawler-temporal-demo)

Crawler, Go code:
[![CI Crawler Go](https://github.com/vitalyshatskikh/crawler-temporal-demo/actions/workflows/ci-crawler-go.yml/badge.svg)](https://github.com/vitalyshatskikh/crawler-temporal-demo/actions/workflows/ci-crawler-go.yml)
[![codecov](https://codecov.io/gh/vitalyshatskikh/crawler-temporal-demo/graph/badge.svg?token=3HS6Z0KAR5&flag=crawler-go)](https://codecov.io/gh/vitalyshatskikh/crawler-temporal-demo)

Example site:
[![CI Example Site](https://github.com/vitalyshatskikh/crawler-temporal-demo/actions/workflows/ci-example-site.yml/badge.svg)](https://github.com/vitalyshatskikh/crawler-temporal-demo/actions/workflows/ci-example-site.yml)
[![codecov](https://codecov.io/gh/vitalyshatskikh/crawler-temporal-demo/graph/badge.svg?token=3HS6Z0KAR5&flag=example-site)](https://codecov.io/gh/vitalyshatskikh/crawler-temporal-demo)

---

# Crawler

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