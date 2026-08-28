## Context

`ProcessSearchPage` (surfer) and `WebDownloader` (downloader) together move a document through its lifecycle. The surfer produces a `SURFED_ADVERT` `DocumentMeta` per advert found on a search page; the downloader fetches the body and is expected to persist a `DOWNLOADED_ADVERT` row; the surfer then hands that `DOWNLOADED_ADVERT` to the parser via `PARSE_ADVERT_CONTENT`. This change fixes four bugs that broke that flow end-to-end and adds the `ProcessSearchPage` per-page stats report to the workflow output. See `proposal.md` for the bug narrative.

## Goals / Non-Goals

**Goals:**
- Make the `WebDownloader` activity the single point at which `created_at`, `updated_at`, and `type` are stamped onto the persisted document, using the actual moment of the download.
- Make `ProcessSearchPage` correctly handle the very first observation of a document (the `created_at == updated_at` case), so brand-new surfaced adverts are never silently dropped by the rate-limit check.
- Surface the per-page processed/skipped counts as workflow output so callers can drive retries and dashboards off them.
- Annotate the `PARSE_SEARCH_PAGE` activity call with a concrete `result_type` for replay-correctness.

**Non-Goals:**
- No DB schema change.
- No new Temporal activity or workflow.
- No change to the parser contract — the Go parser already expects `DOWNLOADED_ADVERT` for `ParseAdvertContent` and `SEARCH_PAGE` for `ParseSearchPage` (see `parser-activities` spec).
- No change to the surf config, parsing config, or worker bootstrap.

## Decisions

### Decision: WebDownloader overrides timestamps and type at the activity boundary

**Chosen:** Inside `WebDownloader.download_to_repo`, construct the `adverts.Document` by:
1. Excluding `created_at`, `updated_at`, `type` from `doc_meta.model_dump(...)`.
2. Filling them in explicitly: `created_at=now`, `updated_at=now`, `type=_downloaded_type(doc_meta.type)`, where `now = dt.datetime.now(tz=dt.UTC)`.

**Rationale:** The downloader is the only code path that actually performs an HTTP fetch. It is the natural single source of truth for the *download* timestamps and the post-fetch `type`. Routing that responsibility through the call-site (`ProcessAdvert`) would scatter the logic and force every caller (including any future ones) to remember to stamp the document.

**Alternatives considered:**
- *Map the type on the surfer side before passing to `ProcessAdvert`'s downloader call*: splits the rule across two packages; loses the ability to call `_downloaded_type` in tests of the downloader.
- *Have the downloader repo do the mapping*: `IDocumentRepository.save` would need to know the lifecycle, breaking single-responsibility.

### Decision: Type transition table implemented as a single match expression

**Chosen:** A module-level helper `_downloaded_type(in_type)` using a Python `match` statement:

```python
def _downloaded_type(in_type: adverts.DocumentType) -> adverts.DocumentType:
    out_type = in_type
    match in_type:
        case adverts.DocumentType.SEARCH_PAGE:
            out_type = adverts.DocumentType.SEARCH_PAGE
        case adverts.DocumentType.SURFED_ADVERT:
            out_type = adverts.DocumentType.DOWNLOADED_ADVERT
        case adverts.DocumentType.DOWNLOADED_ADVERT:
            out_type = adverts.DocumentType.DOWNLOADED_ADVERT
        case adverts.DocumentType.PARSED_ADVERT:
            out_type = adverts.DocumentType.DOWNLOADED_ADVERT
    return out_type
```

**Rationale:** Explicit, exhaustive, easy to extend when a new `DocumentType` is added. Lives in the same module as the activity that uses it, so the unit test for the activity is the natural place to assert the mapping. The default branch (no match) is unreachable today (`DocumentType` is a closed `enum.StrEnum`); leaving `out_type = in_type` as the fallback keeps the helper total without raising.

**Alternatives considered:**
- *Dict mapping*: equivalent semantics, less self-documenting than a `match`.
- *Pydantic validator on `DocumentMeta`*: would conflate validation with lifecycle semantics; lifecycle belongs in the activity, not the model.

### Decision: Just-created bypass expressed as `not just_created and need_update`

**Chosen:** The skip condition in `ProcessSearchPage.run` is:

```python
just_created = doc_meta.updated_at == doc_meta.created_at
need_update = doc_meta.updated_at + dt.timedelta(seconds=doc_meta.update_interval_sec) > workflow.now()
if not just_created and need_update:
    skipped += 1
    continue
```

**Rationale:** The previous condition (`if updated_at + interval > now: skip`) silently dropped brand-new docs because, at the moment of first observation, `updated_at + interval` is always in the future. The new condition only skips when the doc has *been seen before* (`updated_at > created_at`) and is *still fresh*. This matches the producer side: a parser producing a fresh surfed advert sets `created_at == updated_at`, which is a clear "first time" signal.

**Alternatives considered:**
- *Track "first time" with an explicit flag on `DocumentMeta`*: would require a model change and a parser-side update. The implicit `created_at == updated_at` signal is already present and load-bearing in the rate-limit logic — leveraging it avoids new state.
- *Always start the child workflow on first observation, deferring the rate-limit check to inside `ProcessAdvert`*: changes the temporal semantics (child may run before we know whether to skip it) and spreads the policy.

### Decision: `result_type` annotation on the activity call

**Chosen:** Add `result_type=list[adverts.DocumentMeta]` to the `workflow.execute_activity(consts.ActivityName.PARSE_SEARCH_PAGE, ...)` call.

**Rationale:** `temporalio.workflow.execute_activity` is generic in `result_type`; without the explicit hint the workflow replayer treats the result as `Any`, which obscures the type for the loop body. The new spec captures this as a requirement so the contract is documented.

**Alternatives considered:**
- *Cast after the call (`tp.cast(list[adverts.DocumentMeta], ...)`)*: hides the change from a spec reader; doesn't help mypy on the replayer.
- *Use the activity's typed handle*: requires switching to `execute_activity_method`-style calls; bigger refactor.

### Decision: ProcessAdvert sends a type-rewritten copy to the parser

**Chosen:** `ProcessAdvert.run` invokes:

```python
in_.doc_meta.model_copy(update={'type': adverts.DocumentType.DOWNLOADED_ADVERT})
```

as the activity argument, instead of the raw `in_.doc_meta`. The local `in_.doc_meta` is left untouched for logging.

**Rationale:** The surfer-owned `doc_meta` is what the parse activity needs to *consume* — and by the time `ProcessAdvert` calls `PARSE_ADVERT_CONTENT`, the downloader has already saved a row with `type = DOWNLOADED_ADVERT` (per the `WebDownloader` decision above). The parser contract (`parser-activities` spec) requires the input `DocumentMeta.Type` to be `DocumentTypeDownloadedAdvert` for `ParseAdvertContent`. Rewriting the copy at the call site keeps the surfer's own view of the doc (`SURFED_ADVERT`) intact and makes the data flow explicit.

**Alternatives considered:**
- *Update the surfer's `doc_meta` in place*: would silently change behavior elsewhere in the workflow (e.g., the `wf_id` derived from `sdoc_id` and the `process_advert.ProcessAdvertIn` payload).
- *Have the parser accept both types*: violates the existing parser contract for no benefit; we'd still need a different invocation per type.

### Decision: Workflow output becomes the stats dict

**Chosen:** `ProcessSearchPage.run` returns the `stat` dict that was already being constructed and passed to the logger. The signature changes from `async def run(self, in_) -> None` to `async def run(self, in_) -> dict[str, tp.Any]`.

**Rationale:** The values are already computed; surfacing them as workflow output gives callers (and tests) a typed, queryable report without duplicating the loop logic. The change to the return type is the only consumer-visible workflow signature change in this fix; the workflow has no parent that relies on the previous `None` return (`ProcessSearchBranch` is the only caller and does not capture the value).

**Alternatives considered:**
- *Return a typed `ProcessSearchPageReport` pydantic model*: more precise but adds a new exported type just for an internal report. A `dict[str, tp.Any]` is sufficient because the structure is documented in the spec.
- *Expose counters only via Temporal's search attributes (side-channel)*: would require server-side setup and wouldn't help tests.

### Decision: freezegun for downloader activity tests, not for workflow tests

**Chosen:** Add `freezegun = "^1.5.5"` to `crawler/pyproject.toml` and decorate the success-path and 404-path unit tests in `crawler/downloader/tests/unit/application/activities/test_web_downloader.py` with `@freezegun.freeze_time(_FREEZE_TS_STR)`. Do *not* decorate the workflow tests in `crawler/surfer/tests/unit/application/workflows/test_process_search_page.py`.

**Rationale:** The downloader test exercises the activity directly (no Temporal sandbox), so it can use real-time mocking. The surfer workflow tests already use `temporalio.testing.WorkflowEnvironment.start_time_skipping()`, and the project AGENTS.md explicitly notes "Time: temporalio.testing.WorkflowEnvironment.start_time_skipping() (NOT freezegun; Temporal simulates time)" for workflows. Mixing freezegun into workflow tests would fight the time-skipping environment.

**Alternatives considered:**
- *Use a custom clock injected into the activity*: requires changing the production signature, larger surface.
- *Mock `dt.datetime.now` per-test*: freezegun does exactly this, with less boilerplate.

## Risks / Trade-offs

[Risk] `ProcessSearchPage.run` return type changes from `None` to `dict` — running workflows serialized with the old `None` return type will fail on replay.
→ **Mitigation:** Treat workflow history as ephemeral; wipe the Temporal namespace and restart workers on deploy. This is the same posture taken by the `add-content-url-column` and `rate-limit-advert-processing` changes.

[Risk] `WebDownloader` overwrites `created_at` even on re-download — the original first-observation timestamp is lost on the second fetch.
→ **Mitigation:** This matches the agreed semantic ("created_at = time of download"). If we ever need to preserve the original first-observation timestamp, that becomes a separate change adding a distinct `first_seen_at` field.

[Risk] `freezegun` global state could leak across tests if a fixture is misused.
→ **Mitigation:** All decorations are scoped to individual test functions; the global module-level `_FREEZE_TS` constant is only read inside those tests. No autouse fixture, no module-level decorator.

[Risk] `_downloaded_type` defaults to `out_type = in_type` on the no-match branch, which silently passes through any future `DocumentType` that is added to the enum.
→ **Mitigation:** When a new `DocumentType` is added, the matching case in `_downloaded_type` must be added alongside it. To make this enforced, the `match` could be turned into a `match` with a wildcard branch that raises. Acceptable trade-off for now; tracked in Open Questions.

[Risk] `doc_meta.updated_at == doc_meta.created_at` is a weak signal for "just created" — anything that explicitly sets both to the same value (e.g., a future re-surfacing flow) would also bypass the rate limit.
→ **Mitigation:** Document the assumption in the requirement text. If the assumption breaks, switch to an explicit `first_observed` flag.

## Migration Plan

1. Update `crawler/downloader/application/activities/web_downloader.py` (timestamp + type override + `_downloaded_type`).
2. Update `crawler/surfer/application/workflows/process_advert.py` (`model_copy` with `type=DOWNLOADED_ADVERT`).
3. Update `crawler/surfer/application/workflows/process_search_page.py` (skip condition, `result_type` annotation, return type and value).
4. Update `crawler/downloader/tests/unit/application/activities/test_web_downloader.py` (freezegun decoration, frozen `now` assertions, `type=DOWNLOADED_ADVERT` assertions).
5. Update `crawler/surfer/tests/unit/application/workflows/test_process_search_page.py` (assert returned stats in existing tests; add `test_run__when_doc_meta_just_created__then_starts_advert_children`; adjust the skip test so `updated_at != created_at`).
6. Add `freezegun = "^1.5.5"` to `crawler/pyproject.toml` and re-lock.
7. Run `poetry run ruff check --fix .`, `poetry run mypy surfer downloader`, and `poetry run pytest -m "not linting and not integration"` from `crawler/`.
8. Wipe the Temporal namespace on deploy and restart all workers.

No DB migration, no Go changes, no docker-compose change.

## Open Questions

1. *Should `_downloaded_type` raise on an unrecognized future `DocumentType`?* Today the helper silently passes through. If a new enum value is added without updating this helper, the downloader will persist the wrong `type` for that input. Trade-off: stricter = safer but breaks the moment a new type is added. Decision deferred — the current explicit `match` is at least *visible* at code review time.
2. *Should the `ProcessSearchPage` stats report include a per-document breakdown (e.g., processed `sdoc_id`s)?* Not needed today. The current counters are sufficient for dashboards and retry decisions. If detailed per-doc reporting is needed later, it should be added to the workflow's `result_type` and to this spec.
