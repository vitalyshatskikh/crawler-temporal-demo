## Why

The Python `ProcessSearchPage` workflow currently treats the Go `ParseSearchPage` activity as if it returned `list[SdocID]`, then performs a second lookup through a local `AdvertsRepo.get_documents_meta` activity to fetch `DocumentMeta` for each id. The parser already saves every parsed advert document and returns its `DocumentMeta` slice, so the round-trip is redundant: the same metadata is read back from the repository a few milliseconds later. The mismatch between the parser's actual contract and the workflow's expectation forces an extra activity, an extra repository call, an extra task-queue hop, and a pile of plumbing (`IAdvertsRepository`, `PGAdvertsRepo`, `DummyAdvertsRepository`, `AdvertsRepo` activity class, the `GET_DOCUMENTS_META` const) that exists only to bridge the gap.

## What Changes

- **`ProcessSearchPage` workflow** consumes `list[adverts.DocumentMeta]` returned by the `PARSE_SEARCH_PAGE` activity and iterates it directly to start `ProcessAdvert` children — no second lookup.
- **Remove the redundant repo bridge**:
  - delete `surfer.application.activities.AdvertsRepo` and its `get_documents_meta` activity
  - delete `surfer.domain.adverts.IAdvertsRepository` and `DummyAdvertsRepository`
  - delete `surfer.infrastructure.repositories.PGAdvertsRepo`
  - remove `ActivityName.GET_DOCUMENTS_META`
  - drop the `adverts_repo` dependency from `SurfingWorker` and `__main__.py`
- **Update the Python parse-activity stub** (`dummy_parse_search_page`): change parameter from `url: str` to `meta: adverts.DocumentMeta` and change return type from `list[adverts.SdocID]` to `list[adverts.DocumentMeta]`. The new signature matches both the real Go `Parser.ParseSearchPage(ctx, meta DocumentMeta) ([]DocumentMeta, error)` and the call site at `application/workflows/process_search_page.py:59-65`, which already passes a `DocumentMeta`.
- **Add Go JSON tags to `DocumentMeta`** (`crawler/parser/internal/domain/models.go:29`) so PascalCase Go fields serialize to snake_case JSON keys, matching the Python pydantic field names. This makes the Go → Python wire format of `[]DocumentMeta` cross-language round-trip safe without a pydantic alias generator.
- **Drop related tests and fixtures** (`test_pg_adverts_repo_integration.py`, `pg_adverts_repo` / `insert_document` fixtures, `AdvertsRepo` registrations in unit workflow tests).
- **Sync `SurferConfigFactory`** (`crawler/surfer/tests/_factories.py:68-79`) with the real `SurferConfig` (`application/config.py`): rename the five `*_timeout` fields whose real names have the `_wf_` infix (`process_branch_wf_timeout`, `process_search_page_wf_timeout`, `process_advert_wf_timeout`, `download_search_page_wf_timeout`, `download_advert_content_wf_timeout`); add `search_adverts_timeout`, `repo_request_retry`, `parse_search_page_retry`, `parse_advert_content_retry`. Existing fields keep their values.
- **Refresh `crawler/surfer/AGENTS.md`** to reflect the removed port, activity, and dummy repo.

The only Go-side change is the JSON-tag addition noted above. The activity contract itself is unchanged: `Parser.ParseSearchPage` already returns `[]domain.DocumentMeta` and already persists each parsed document via `SaveDocument` (see `openspec/specs/parser-activities/spec.md` — "ParseSearchPage fetches, parses, and persists"). The Python side is being aligned with that existing contract; the JSON tags make the cross-language wire format symmetric (input and output both snake_case).

`SurferConfig` is unchanged: `repo_request_timeout` / `repo_request_retry` are still used by `SearchAdverts` (for the `GET_SURF_CONFIG` local activity) and by `ProcessAdvert` (passed to `DownloadIn` as `config_request_timeout`). No config field becomes orphaned. The `SurferConfigFactory` drift fix above brings test-side instantiation into line with the real model so unit tests stop silently running with default timeouts.

## Out of Scope

- **Renaming `repo_request_timeout` / `repo_request_retry`** — after this change the fields are no longer used by any advert-repo code path; the names become misnomers. Deferred to a follow-up to keep this change atomic.
- **Wiring `crawler/parser/cmd/parser/main.go`** — the Go parser binary is still empty (`func main() {}`); this change does not register `Parser.ParseSearchPage` on a Go worker. The Python `dummy_parse_search_page` registered on `ParsingWorker` continues to serve the `PARSING` task queue. Real Go-parser wiring is a separate concern.
- **Encoding "ProcessAdvert must only receive `SURFED_ADVERT` metas" as a spec** — the parser today always emits `SURFED_ADVERT` per `internal/application/testutil/factories.go` → `MustSurfedAdvertMeta`, but no spec currently asserts this. Worth a follow-up.
- **Iteration-order guarantee on child-workflow starts** — the workflow now starts `ProcessAdvert` children in parser-write-order (slice) rather than DB-read-order (`dict.items()`). This is observable in Temporal event history but not asserted by any spec today.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None.

This is a pure refactor: the externally observable behavior of the system (the set of `ProcessAdvert` children started and the `doc_meta` they receive) does not change. The `parser-activities` spec already requires `ParseSearchPage` to return `[]domain.DocumentMeta`. The Python side is simply catching up to that contract. Because no spec-level requirement is changing, the change declares `skip_specs: true` in `.openspec.yaml`.

## Impact

- **Code (Python, surfer):**
  - `application/workflows/process_search_page.py` — drop second lookup, iterate `list[DocumentMeta]`
  - `application/activities.py` — remove `AdvertsRepo`; update `dummy_parse_search_page` signature
  - `application/consts.py` — remove `ActivityName.GET_DOCUMENTS_META`
  - `application/config.py` — unchanged (audit confirms all fields still used)
  - `domain/adverts/repositories.py` — deleted
  - `domain/adverts/__init__.py` — drop `IAdvertsRepository` export
  - `infrastructure/repositories/adverts_repo.py` — deleted
  - `infrastructure/repositories/__init__.py` — drop `PGAdvertsRepo` export
  - `infrastructure/workers.py` — drop `adverts_repo` from `SurfingWorker.__init__` and from its activities list
  - `__main__.py` — stop constructing `PGAdvertsRepo` and stop passing it to `SurfingWorker`
- **Tests (Python, surfer):**
  - `tests/unit/application/workflows/test_process_search_page.py` — drop `AdvertsRepo` registrations, switch `set_mock(PARSE_SEARCH_PAGE, ...)` to return `list[DocumentMeta]`, build expectations via `DocMetaFactory`
  - `tests/unit/application/workflows/test_search_adverts.py` — drop `AdvertsRepo` registrations
  - `tests/integration/infrastructure/repositories/test_pg_adverts_repo_integration.py` — deleted
  - `tests/integration/conftest.py` — drop `pg_adverts_repo` fixture and `insert_document` helper (only used by the deleted test)
- **Docs:** `crawler/surfer/AGENTS.md` — drop mentions of `IAdvertsRepository`, `AdvertsRepo` activity, `DummyAdvertsRepository`, and the `PGAdvertsRepo` stub
- **Code (Go, parser):** add JSON struct tags to `DocumentMeta` (`crawler/parser/internal/domain/models.go:29`) for snake_case serialization (`json:"sdoc_id"`, `json:"created_at"`, `json:"updated_at"`, `json:"source_id"`, `json:"type"`, `json:"external_url"`, `json:"update_interval_sec"`). Add a focused unit test in `crawler/parser/internal/domain/models_test.go` asserting `json.Marshal(meta)` produces snake_case keys and that `json.Unmarshal` of snake_case JSON round-trips. No behavioral or contract change to `ParseSearchPage` itself.
- **Wire contract:** Temporal activity `ParseSearchPage` now returns `list[DocumentMeta]` (Python-side) instead of `list[SdocID]`. With the Go JSON-tag addition, the payload keys are snake_case, matching the Python pydantic `DocumentMeta` field names — no alias generator needed. This is an internal cross-language contract change that aligns the Python side with the existing Go-side spec; the parser's return shape was already `[]DocumentMeta`.
- **Backwards compatibility:** none required — no in-flight workflows or external clients depend on the old `GET_DOCUMENTS_META` activity or the old `ParseSearchPage` return shape.
