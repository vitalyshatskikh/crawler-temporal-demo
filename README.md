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