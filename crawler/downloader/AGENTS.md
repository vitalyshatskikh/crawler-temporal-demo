# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## ! Important !

**Never** mention agent's name, model and vendor in commit messages or generated code or any other materials

## Package

**Package manager:** Poetry

### Setup

```bash
# Install dependencies
poetry sync
```

### Test

```bash
# Run all unit tests
poetry run pytest -m "not linting and not integration"

# Run a single test file
poetry run pytest tests/unit/application/workflows/test_download_search_page.py

# Run a single test function
poetry run pytest tests/unit/application/activities/test_web_downloader.py::test_download_to_repo__when_response_4xx_excluding_404__then_raises_validation_error
```

### Lint/Format

```bash
# Lint and fix
poetry run ruff check --fix .

# Type check
poetry run mypy surfer downloader
```

---

## Architecture Overview

**Framework:** Temporal Python SDK for workflow orchestration

The downloader is a separate Temporal worker application responsible for downloading web page content.

```
domain/
├── errors.py              # DownloaderError, NotFoundError, ValidationError
└── downloading/
    ├── models.py          # Params (pydantic) — download config for a source
    └── repositories.py    # IDownloadingRepository, IDocumentRepository + Dummy* stubs

application/
├── config.py              # DownloaderConfig (timeouts, retry, aiohttp settings)
├── consts.py             # ActivityName enum
├── activities/
│   ├── downloading_repo.py  # DownloadingRepo (GetDownloadingConfig activity)
│   └── web_downloader.py   # WebDownloader (DownloadToRepo activity, aiohttp)
└── workflows/
    ├── download_search_page.py    # DownloadSearchPage workflow
    └── download_advert_content.py # DownloadAdvertContent workflow

infrastructure/
├── db/                   # SQLAlchemy ORM mappers and model registry
├── repositories/
│   ├── config_repo.py     # PGDownloadingRepository (ISurfingRepository impl)
│   └── document_repo.py  # PGDocumentRepository (IDocumentRepository impl)
└── workers.py            # DownloadingWorker (Temporal worker bootstrap)

__main__.py               # Entry point: connect Temporal → run DownloadingWorker
```

**Cross-package dependency:** Downloader imports from `surfer.domain.adverts` (`SourceID`, `DocumentType`, `Document`, `DocumentMeta`) and `surfer.application.consts` (`QueueName`, `WorkflowName`). This is intentional — surfer is the main package defining the interface between components.

---

## Code Style

### Formatting

- Max line length: **120 characters**
- Indentation: **4 spaces**
- Quotes: **single** for inline strings, **double** for docstrings
- Line endings: LF, UTF-8, final newline, no trailing whitespace

### Imports

- Group in order: **stdlib → third-party → internal**. Sorted alphabetically within groups.
- Import modules not symbols
- Use `import typing as tp` and `import datetime as dt` aliases

```python
import abc
import datetime as dt
import typing as tp

import aiohttp
import pydantic
import temporalio

from downloader.application import activities, consts
from surfer.domain import adverts
```

### Naming

| Entity | Convention | Example |
|---|---|---|
| Classes | `PascalCase` | `DownloadSearchPageWorkflow` |
| Interfaces (ABCs) | `I` prefix | `IDownloadingRepository` |
| Functions / methods | `snake_case` | `download_to_repo`, `make_params` |
| Private methods | `_snake_case` | `_raise_for_status` |
| Constants | `UPPER_SNAKE_CASE` | `ActivityName.DOWNLOAD_TO_REPO` |
| Test methods | `test_<method>__when_<condition>__then_<outcome>` | `test_download_to_repo__when_response_4xx__then_raises` |
| Files | `snake_case.py` | `web_downloader.py`, `config_repo.py` |

### Type Annotations

- Mandatory on all functions (`disallow_untyped_defs = true` in mypy)
- Union types: `X | Y` — never `Optional[X]` or `Union[X, Y]`
- Built-in generics: `list[str]`, `dict[str, int]` — never bare `list`, `dict`
- Use `tp.cast()` to narrow types mypy cannot infer
- Use `# type: ignore` sparingly, always with a comment explaining why

---

## Testing Conventions

- **Framework:** pytest + pytest-asyncio (`asyncio_mode = "auto"`)
- **Time:** `temporalio.testing.WorkflowEnvironment.start_time_skipping()` (NOT freezegun; Temporal simulates time)
- **Test doubles:** `DummyConfigRepository`, `DummyDocumentRepository` in `domain/downloading/repositories.py`; `set_mock`/`clear_mocks`/`get_mock_calls` via `tests/_interceptors.py`
- **Test data builders:** `tests/_factories.py` (`make_params`, `make_doc_meta`)
- **Test interceptors:** `tests/_interceptors.py` (`WorkflowMockInterceptor`, `MockHandle`) — mocks child workflows and remote/local activities
- **Test layout:** `tests/unit/application/activities/` for activity tests, `tests/unit/application/workflows/` for workflow tests, `tests/unit/infrastructure/` for infrastructure tests; module-level async functions (NOT class-wrapped)
- **Workflow-failure assertions:** use `_factories.assert_workflow_failure_message(excinfo, ...)` after `with pytest.raises(WorkflowFailureError) as excinfo:`

---

## Error Handling

- **HTTP 404 is a success case for scraping** — `WebDownloader.download_to_repo` saves the document regardless of 404; body may be empty or contain a 404 page. Log at info level.
- **HTTP 4xx (except 404):** raise `ValidationError("bad request")` — the source rejected the request as malformed
- **HTTP 5xx:** raise `DownloaderError("internal error")` — server-side failure
- **NotFoundError:** raised by `PGDownloadingRepository.get_download_config` when no config row exists for (source_id, doc_type). Non-retryable.
- **ValidationError:** raised for bad user input / malformed requests. Non-retryable.
- Use `%s` formatting (never f-strings) in logging

---

## Configuration

`DownloaderConfig` in `application/config.py`:

| Field | Default | Description |
|---|---|---|
| `download_timeout` | 5 min | Activity timeout for `DownloadToRepo` |
| `config_request_timeout` | 5 min | Activity timeout for `GetDownloadingConfig` |
| `config_request_retry` | 3 attempts | Retry for config fetch |
| `download_retry` | 1 attempt | Retry for download (default: no retries) |
| `http_total_timeout` | 30 s | aiohttp total timeout |
| `http_connect_timeout` | 10 s | aiohttp connect timeout |
| `http_connector_limit` | 100 | Max concurrent connections |
| `http_proxy` | None | Optional HTTP proxy URL |

---

## Cross-package Sync Contract

The downloader depends on types and constants defined in `surfer`. Changes to those shared definitions must be propagated to `crawler/parser` as well.

### Types imported from `surfer.domain.adverts`

| Type | Used for | Notes |
|---|---|---|
| `SourceID` | `Params.source_id`, `DocumentMeta.source_id` | Identifies a scraping source |
| `DocumentType` | `Params.doc_type`, `DocumentMeta.type` | Values: `search_page`, `surfed_advert`, `downloaded_advert`, `parsed_advert` |
| `Document` | Return type of `WebDownloader.download_to_repo` | `DocumentMeta` + body str |
| `DocumentMeta` | Activity input (`DownloadSearchPageIn`, `DownloadAdvertContentIn`) | Common metadata across all document stages |

### Constants imported from `surfer.application.consts`

| Constant | Used for | Notes |
|---|---|---|
| `QueueName.DOWNLOADING` | Task queue for downloader worker | `"downloading"` |
| `WorkflowName.DOWNLOAD_SEARCH_PAGE` | Child workflow name | `"DownloadSearchPage"` |
| `WorkflowName.DOWNLOAD_ADVERT_CONTENT` | Child workflow name | `"DownloadAdvertContent"` |

### DocumentMeta field semantics (shared with surfer)
- `external_url` — canonical URL used as **identity** (stable across re-downloads)
- `content_url` — URL to **fetch body from** (may differ from external_url)
- `update_interval_sec` — advisory crawl interval

When changing `surfer.domain.adverts` or `surfer.application.consts`, update `crawler/parser` in the **same commit**.

---

## Verification

After making changes, always run:

```bash
poetry run ruff check --fix .
poetry run mypy surfer downloader
poetry run pytest -m "not linting and not integration" -v
```

Do not run integration/e2e tests unless the user explicitly asks.
