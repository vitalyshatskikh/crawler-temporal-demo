## Context

Today the Python `ProcessSearchPage` workflow calls `PARSE_SEARCH_PAGE` (a Temporal activity backed by Go's `Parser.ParseSearchPage`) and consumes its result as `list[adverts.SdocID]`. The Go activity actually returns `[]domain.DocumentMeta` and persists each parsed advert document before returning; see `openspec/specs/parser-activities/spec.md` — "ParseSearchPage fetches, parses, and persists". To recover the `DocumentMeta`s the workflow needs for child `ProcessAdvert` invocations, it then schedules a second local activity (`AdvertsRepo.get_documents_meta` → `GET_DOCUMENTS_META`) on the surfing task queue, which in turn calls `PGAdvertsRepo.get_documents_meta_by_sdoc_id` to read the same rows back.

The mismatch forces a chain of Python-only plumbing (port interface, dummy, PG impl, activity class, const, worker wiring, integration test) that exists solely to translate `SdocID → DocumentMeta` a few milliseconds after the parser already produced it. Removing the round-trip aligns the workflow with the existing parser contract and deletes everything that exists only to bridge it.

See `proposal.md` for motivation and `proposal.md` → "Impact" for the full file inventory.

## Goals / Non-Goals

**Goals:**
- Workflow consumes `list[adverts.DocumentMeta]` from `PARSE_SEARCH_PAGE` directly.
- Delete the Python `AdvertsRepo` activity, the `IAdvertsRepository` port, its `DummyAdvertsRepository`, and its `PGAdvertsRepo` impl.
- Update `dummy_parse_search_page` to match the real activity's return type.
- Keep all `SurferConfig` fields (audit confirms each is still referenced by `SearchAdverts` / `ProcessAdvert` / `ProcessSearchPage`).
- Sync `SurferConfigFactory` with the real `SurferConfig` so unit tests stop silently running with default workflow-level timeouts and missing retry configs.
- Tests pass: `ruff check`, `mypy surfer`, `pytest -m "not linting and not integration"`, `go test ./...` (parser).

**Non-Goals:**
- No Go-side changes.
- No migration of in-flight workflows (no in-flight `ProcessSearchPage` workflows depend on the old contract in any environment we ship).
- No new capability spec — the `parser-activities` spec already encodes the correct `ParseSearchPage` return shape; the workflow change brings code into compliance with it.
- No rename of `ProcessAdvert` workflow id pattern (`sdocid/{sdoc_id}`); the value is the same, just sourced from the parser rather than from a re-fetched row.

## Decisions

### Decision 1: Iterate the parser result directly — no shape change in the workflow

After `PARSE_SEARCH_PAGE` returns `documents_meta: list[adverts.DocumentMeta]`, the workflow's loop is:

```python
for doc_meta in documents_meta:
    wf_id = f"{consts.WorkflowName.PROCESS_ADVERT}/{in_.surf_params.name}/sdocid/{doc_meta.sdoc_id}"
    await workflow.start_child_workflow(
        process_advert.ProcessAdvert.run,
        process_advert.ProcessAdvertIn(
            surfer_config=in_.surfer_config,
            surf_params=in_.surf_params,
            doc_meta=doc_meta,
        ),
        id=wf_id,
        ...
    )
```

This matches the existing loop body, except it iterates a `list` instead of `dict.items()`.

**Behavioral note:** iterating a slice preserves parser-write order; iterating `dict.items()` previously preserved DB-read order. With this change, child-workflow start order becomes parser-write order (slice index). This is observable in Temporal event history and is a deliberate improvement — the parser is the source of truth for ordering, and DB read order is incidental.

**Alternatives considered:**
- *Keep a dict for stable lookup.* Rejected — the parser returns a slice in document order; we iterate in order to start children in order, and order is observable in Temporal event history. A dict adds nothing.
- *Wrap with a `dict[SdocID, DocumentMeta]` for backwards compatibility.* Rejected — no callers consume the dict shape; this is a single in-workflow consumer.

### Decision 2: Update `dummy_parse_search_page` signature to match the real activity

The dummy's current body is `return list(map(adverts.SdocID, ("1", "2", "3")))`. The dummy is only registered for the `PARSING` worker and is never invoked by any unit test (every `ProcessSearchPage` test overrides via `set_mock`). After this change, the dummy matches the real Go `Parser.ParseSearchPage(ctx, meta domain.DocumentMeta) ([]domain.DocumentMeta, error)` contract:

```python
@activity.defn(name=consts.ActivityName.PARSE_SEARCH_PAGE)
async def dummy_parse_search_page(meta: adverts.DocumentMeta) -> list[adverts.DocumentMeta]:
    activity.logger.info("parse search page %s", meta.external_url)
    return []
```

**Alternatives considered:**
- *Fabricate stub `DocumentMeta`s.* Rejected — the stub is unused by tests, and a fabricated list could mislead a future test that forgets to mock. Returning `[]` is honest.
- *Keep the `url: str` parameter and only change the return type.* Rejected — the real activity takes a `DocumentMeta`, the call site passes a `DocumentMeta`, and the dummy is the only place where the parameter type contradicts both. Hygiene fix is cheap.
- *Compute `SdocID` from the URL via `hashlib.md5` to match `process_search_page.py`'s own computation.* Rejected — diverges from the real parser's behavior (which hashes via `SdocIDForURL` in Go) and adds noise.

The dummy is annotated as a TODO to be replaced by a real Go-side registration on a parser worker. The signature alignment here is purely so that the contract between caller and dummy matches reality; once the Go worker is wired, the dummy disappears entirely.

### Decision 3: `dummy_parse_advert_content` is unaffected

`PARSE_ADVERT_CONTENT` already takes `DocumentMeta` and returns `None`. No stub change.

### Decision 4: Test mocks for `PARSE_SEARCH_PAGE` use `DocMetaFactory`

In `test_run__when_parse_returns_sdocs__then_starts_advert_children`, the mock result becomes:

```python
doc_meta_1 = DocMetaFactory(sdoc_id='1')
doc_meta_2 = DocMetaFactory(sdoc_id='2')
set_mock(consts.ActivityName.PARSE_SEARCH_PAGE, result=[doc_meta_1, doc_meta_2])
```

`DocMetaFactory` (already in `tests/_factories.py`) produces valid `DocumentMeta` with the `SURFED_ADVERT` type — the same type the real parser produces.

### Decision 5: Worker & entry-point surgery

- `SurfingWorker.__init__` drops `adverts_repo: adverts.IAdvertsRepository` and its `self._adverts_repo` setup; the `activities=[...]` list shrinks to just `surf_config_repo.get_surf_params`.
- `__main__.py` stops constructing `PGAdvertsRepo` and stops passing it to `SurfingWorker`.
- `infrastructure/repositories/__init__.py` no longer exports `PGAdvertsRepo`.
- `domain/adverts/__init__.py` no longer exports `IAdvertsRepository`.
- The infrastructure file `infrastructure/repositories/adverts_repo.py` and the domain file `domain/adverts/repositories.py` are deleted outright.

### Decision 6: Delete `PGAdvertsRepo` integration test, prune its fixtures

`tests/integration/infrastructure/repositories/test_pg_adverts_repo_integration.py` is deleted (no repo to test). `tests/integration/conftest.py` loses the `pg_adverts_repo` fixture and the `insert_document` helper (audited to be used only by the deleted file).

### Decision 7: Go `DocumentMeta` gets snake_case JSON tags

To make the Go → Python wire format symmetric (input is already snake_case-friendly via the existing Python pydantic converter setup at `__main__.py`), add `json:"..."` struct tags to `crawler/parser/internal/domain/models.go:DocumentMeta`:

```go
type DocumentMeta struct {
    SdocID            SdocID            `json:"sdoc_id"`
    CreatedAt         time.Time         `json:"created_at"`
    UpdatedAt         time.Time         `json:"updated_at"`
    SourceID          SourceID          `json:"source_id"`
    Type              DocumentType      `json:"type"`
    ExternalURL       string            `json:"external_url"`
    UpdateIntervalSec int               `json:"update_interval_sec"`
}
```

Add focused unit tests in `internal/domain/models_test.go` covering `Marshal` → snake_case keys and `Unmarshal` of snake_case JSON → all fields populated.

**Alternatives considered:**
- *Add a pydantic `alias_generator=to_camel` to Python's `DocumentMeta` instead.* Rejected — every other Go → Python payload in this codebase would benefit from the same treatment; fixing it at the Go source is a single point of change and keeps the Python model idiomatic snake_case.
- *Leave Go as-is and rely on Temporal's default JSON encoding (PascalCase keys).* Rejected — the first live run of `ProcessSearchPage` would fail pydantic validation on every returned `DocumentMeta` field. Worth fixing at the wire boundary, not at runtime.
- *Skip the unit tests, since the parser is "obviously correct."* Rejected — the JSON tags are the only behavioral invariant this change creates on the Go side; a 10-line test is cheap insurance.

## Risks / Trade-offs

- **[Temporal serialization of `DocumentMeta`]** → The Python workflow now receives a richer payload from the Go parser. The `pydantic_data_converter` is already configured at the client (`__main__.py` line 29), so pydantic v2 ↔ JSON handles the encoding. With the JSON tags added in Decision 7, the wire format uses snake_case keys (`sdoc_id`, `created_at`, etc.) matching Python's pydantic field names exactly — no alias generator needed. **Mitigation:** focused Go unit tests (Decision 7, task 0.2) assert `Marshal`/`Unmarshal` round-trip; if a future field is added without a JSON tag, those tests fail immediately.

- **[Order of child starts]** → Iterating a slice preserves parser order; iterating `dict.items()` did not guarantee DB-read order. In practice Python 3.7+ `dict` iteration is insertion-ordered and the parser wrote them in order, so behavior is unchanged in steady state — but the source of order is now parser-write order, not DB-read order. **Mitigation:** the existing test for "starts advert children" still asserts `sdoc_id == '1'` first then `'2'`; the assertion stays green. The shift in order source is acknowledged in Decision 1 and called out in `proposal.md` → "Out of Scope" so it can be promoted to a spec in a follow-up if it becomes load-bearing.

- **[Misplaced `DocumentType` on returned metas]** → If the parser ever returns a meta with a non-`SURFED_ADVERT` type, the workflow will start a `ProcessAdvert` child with a surprising `meta.type`. The parser today always produces `SURFED_ADVERT` per `internal/application/testutil/factories.go` → `MustSurfedAdvertMeta`. **Mitigation:** out of scope; no change to type filtering. Worth a follow-up spec for "ProcessAdvert must only receive SURFED_ADVERT metas".

- **[`SurferConfig` cleanup missed]** → If any field becomes orphaned post-change, the linter/typechecker will flag the unused field via `ruff`/`mypy`. The audit shows no orphan fields, so this is a non-issue; if an audit mistake is later discovered, it surfaces immediately.

- **[In-flight workflows referencing `GET_DOCUMENTS_META`]** → If any historical `ProcessSearchPage` workflow is replayed after deploy, the missing activity will fail replay. **Mitigation:** confirmed by the author that no environment (dev / staging / prod) currently runs `ProcessSearchPage`; there are no in-flight workflows. Mentioned in the proposal's "Backwards compatibility" line.

## Migration Plan

1. Land all file changes in one commit (atomic refactor).
2. Re-run lint, typecheck, and unit tests:
   - `cd crawler/parser && go test ./...`
   - `poetry run ruff check .`
   - `poetry run mypy surfer`
   - `poetry run pytest -m "not linting and not integration"`.
3. No deploy / rollout step required: this is a single worker (`__main__.py` runs `SurfingWorker` + `AdvertsWorker` + `ParsingWorker`), and the parser worker is a separate Go process. The Python worker is the only consumer of the deleted types/activity/const. Restart the Python worker to pick up the new build.
4. Rollback: revert the commit; restart the Python worker. No DB schema migration.

## Open Questions

None. All material ambiguities were resolved in the planning conversation: `SurferConfig` audit confirmed no orphans; workflow id construction is unchanged; dummy activity contract is the minimal `[]` stub; failure semantics are preserved.
