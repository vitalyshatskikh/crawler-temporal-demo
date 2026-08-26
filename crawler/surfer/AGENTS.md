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
poetry run pytest tests/unit/application/workflows/test_search_adverts.py

# Run a single test function (module-level async tests, NOT class-wrapped)
poetry run pytest tests/unit/application/workflows/test_search_adverts.py::test_run__when_three_branches__then_starts_three_children_with_sequential_ids
```

### Lint/Format

```bash
# Lint
poetry run ruff check --quiet .

# Lint and fix
poetry run ruff check --fix .

# Type check
poetry run mypy surfer
```

---

## Architecture Overview

**Framework:** Temporal Python SDK for workflow orchestration

```
domain/
├── adverts/           # Adverts domain models
├── parsing/          # Parser contract (IParser) + DummyParser (currently unused)
└── surfing/          # Surfing Params/TemplateContext models, URLGenerator, ISurfingRepository
application/          # Temporal workflow defns, activity classes, config, consts
infrastructure/       # Worker bootstrap (3 worker classes), PGConfigRepository stub
```

**`application/` contents:**
- Temporal workflow defns: SearchAdverts, ProcessSearchBranch, ProcessSearchPage, ProcessAdvert
- Port-bound activity classes: SurfConfigRepo (ISurfingRepository → GetSurfParams)
- Module-level dummy parse activities: dummy_parse_search_page, dummy_parse_advert_content (marked TODO to extract into separate app)
- SurferConfig (timing/retry knobs); RetryConfig lives in shared/py/settings.py
- Consts: QueueName (6), ActivityName (3), WorkflowName (6 = 4 surfer + 2 downloader-owned) — all in application/consts.py

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
from uuid import UUID

import pydantic
import temporalio

from surfer.domain.surfing import models
```

### Naming

| Entity | Convention | Example |
|---|---|---|
| Classes | `PascalCase` | `SearchAdvertsWorkflow` |
| Interfaces (ABCs) | `I` prefix | `ISurfingRepository`, `IDownloader` |
| Functions / methods | `snake_case` | `run_surfing_worker`, `make_doc_meta` |
| Private methods | `_snake_case` | `_repo_retry_from` |
| Constants | `UPPER_SNAKE_CASE` | `QueueName.SURFING_TASK`, `ActivityName.GET_SURF_CONFIG` |
| Test methods | `test_<method>__when_<condition>__then_<outcome>` | `test_run__when_three_branches__then_three_children_started` |
| Files | `snake_case.py` | `workflows.py`, `url_generator.py` |

### Type Annotations

- Mandatory on all functions (`disallow_untyped_defs = true` in mypy)
- Union types: `X | Y` — never `Optional[X]` or `Union[X, Y]`
- Built-in generics: `list[str]`, `dict[str, int]`, `tuple[int, ...]` — never `List`, `Dict`
- Use `tp.cast()` to narrow types mypy cannot infer
- Use `# type: ignore` sparingly, always with a comment explaining why

---

## Testing Conventions

- **Framework:** pytest + pytest-asyncio (`asyncio_mode = "auto"`)
- **Time:** `temporalio.testing.WorkflowEnvironment.start_time_skipping()` (NOT freezegun; Temporal simulates time)
- **Test doubles:** port dummies (`DummyConfigRepository` in `domain/surfing/repositories.py`) and module-level dummy parse activities (`dummy_parse_search_page`, `dummy_parse_advert_content` in `application/activities.py`); override pytest fixtures in the test module to inject raising variants
- **Test data builders:** live in `tests/_factories.py`; `conftest.py` only holds fixtures
- **Test interceptors:** `tests/_interceptors.py` (`WorkflowMockInterceptor`, `set_mock`/`clear_mocks`/`get_mock_calls`, `MockHandle`) — mocks child workflows and remote activities; `tests/linting/test_linting.py` runs ruff + mypy under the `linting` marker
- **Test layout:** `tests/unit/application/workflows/` for workflow tests, `tests/unit/domain/surfing/` for domain-model/URL-generator tests; module-level async functions (NOT class-wrapped); `test_<method>__when_<cond>__then_<outcome>` naming still applies
- **Workflow-failure assertions:** use `_factories.assert_workflow_failure_message(excinfo, ...)` after `with pytest.raises(WorkflowFailureError) as excinfo:` — `str(WorkflowFailureError)` is just `'Workflow execution failed'`, the wrapped message lives on `.cause`

---

## Error Handling

- Use standard Python exceptions with meaningful messages
- Define custom exceptions in `domain/errors.py`
- Return errors from functions rather than logging them
- Use `%s` formatting (never f-strings) in logging

---

## Verification

After making changes, always run:

```bash
poetry run ruff check .
poetry run mypy surfer
poetry run pytest -m "not linting and not integration"
```

Do not run integration/e2e tests unless the user explicitly asks.
