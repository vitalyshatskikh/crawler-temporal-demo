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

Distributed web crawling/scraping app built on Temporal for workflow orchestration.
Python surfer and downloader components, Go parser component, and a Go example
classifieds site that acts as the scrape target.

## Architecture

```mermaid
flowchart LR
    subgraph example-site
        direction LR
        es-api["siteapi"]
        es-db[("(postgres)")]
        es-sellers["sellers"]
    end

    subgraph temporal
        direction LR
        t-server["server"]
        t-db[("(postgres)")]
        t-ui["ui"]
    end

    subgraph crawler
        direction LR
        c-surfer["surfer\n(python)"]
        c-downloader["downloader\n(python)"]
        c-parser["parser\n(go)"]
        c-db[("(postgres)")]
    end

    es-sellers --> es-api
    es-api <--> es-db
    c-downloader -.->|HTTP| es-api

    c-surfer --> t-server
    c-downloader --> t-server
    c-parser --> t-server
    t-ui --> t-server
    t-server <--> t-db

    c-surfer <--> c-db
    c-downloader <--> c-db
    c-parser <--> c-db
```

## SearchAdverts workflow

```mermaid
sequenceDiagram
    autonumber
    participant Sched as Temporal Schedule
    participant Surfer
    participant TS as Temporal Server
    participant Downloader
    participant Parser
    participant CDB as Crawler DB
    participant SAPI as siteapi

    Sched->>TS: start SearchAdverts
    TS->>TS: persist workflow state

    Surfer->>Surfer: GetSurfParams (local activity)
    Surfer->>CDB: load surf config
    CDB-->>Surfer: surf_params

    par for each branch
        Surfer->>TS: start ProcessSearchBranch
        TS->>Surfer: dispatch to surfing queue
        loop for each page in branch
            Surfer->>TS: start ProcessSearchPage
            TS->>Surfer: dispatch to surfing queue

            par download + parse search page
                Surfer->>TS: start DownloadSearchPage
                TS->>Downloader: dispatch to downloading queue
                Downloader->>CDB: load download config
                Downloader->>SAPI: GET search page
                SAPI-->>Downloader: response
                Downloader->>CDB: save document (search_page)

                Surfer->>TS: schedule ParseSearchPage activity
                TS->>Parser: dispatch to parsing queue
                Parser->>CDB: get document (search_page)
                Parser->>CDB: save documents (surfed_advert metas)
                Parser-->>Surfer: list of DocumentMeta
            end

            loop for each advert meta
                Surfer->>TS: start ProcessAdvert (fire & forget)
                TS->>Surfer: dispatch to advert-processing queue

                par download + parse advert content
                    Surfer->>TS: start DownloadAdvertContent
                    TS->>Downloader: dispatch to downloading queue
                    Downloader->>CDB: load download config
                    Downloader->>SAPI: GET advert page
                    SAPI-->>Downloader: response
                    Downloader->>CDB: save document (downloaded_advert)

                    Surfer->>TS: schedule ParseAdvertContent activity
                    TS->>Parser: dispatch to parsing queue
                    Parser->>CDB: get document (downloaded_advert)
                    Parser->>CDB: save document (parsed_advert)
                end
            end
        end
    end
```

## Prerequisites

- docker
- docker compose

## Run

```bash
make infra
make example-site
make crawler
```

`make infra` starts the Temporal stack — server, postgres, opensearch, UI on
http://localhost:8080 (namespace `crawler` is created automatically).

`make example-site` starts the target site — `siteapi` on http://localhost:8090
with Swagger UI at `/docs`, plus three sellers continuously creating and
deleting adverts.

`make crawler` starts the crawler workers (surfer, downloader, parser) — no
exposed HTTP ports, just Temporal workers polling their queues.

Open http://localhost:8080 to watch `SearchAdverts` and child workflows run.

## Stop

```bash
make crawler-down
make example-site-down
make infra-down
```

## Develop

### Prerequisites

- docker, docker compose
- python
- go

```bash
make setup-dev
make infra
make example-site
(cd crawler && docker-compose up -d postgres migrate)
```

## Links

- [samples-server/compose](https://github.com/temporalio/samples-server/tree/main/compose) - Run Temporal services using docker-compose
