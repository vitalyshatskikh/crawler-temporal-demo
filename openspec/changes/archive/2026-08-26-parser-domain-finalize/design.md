## Context

The parser domain layer (`crawler/parser/internal/domain/`) is a draft that does not compile. Key issues:

1. `NewParsingService()` discards its constructor arguments — every method NPEs.
2. `ParseAdvertContent` references an undefined variable `i`.
3. `JMESParser.Parse` passes a raw `string` to `jmespath.Search`, but the library expects decoded Go data (`map[string]any` from `json.Unmarshal`).
4. `github.com/jmespath/go-jmespath` is not declared in `crawler/go.mod`.
5. `ParseSearchPage` reuses the parent page's `SdocID` for every extracted snippet — all surfed advert documents collide on PK.
6. `ConfigRepository.GetConfig(ctx, sourceID)` ignores `docType`, but the cache key and the `ParsingConfig.DocumentType` field are both present, suggesting per-(source, docType) config is needed.
7. No input validation on `DocumentMeta` or `ParsingConfig`.
8. No unit tests.

The downstream (application layer with Temporal activities) and infrastructure layer (PG repos) are not yet built; this change only finalizes the domain.

See `docs/superpowers/specs/2026-08-24-parser-app-design.md` for the broader parser-service design. This change does not implement that full design — it finalizes only the domain layer as a prerequisite.

## Goals / Non-Goals

**Goals:**
- Domain code compiles and is fully unit-tested.
- All bugs in the current draft are fixed.
- Interface signatures reflect the per-(source, docType) config model.
- Snippet identity is stable and unique per URL (md5_hex).
- Validation is explicit and testable.
- Mocking strategy is established via mockery for the test layer.

**Non-Goals:**
- Application layer (Temporal `ParseSearchPage` / `ParseAdvertContent` activities) — deferred to next change.
- Infrastructure layer (PG config repo, PG document store, sqlc) — deferred to its own change.
- Alignment with the long design doc's `SnippetJMESPath` + `AdvertProps` + `AdvertExternalURLTemplate` (mustache) split — deferred if needed.
- Integration tests — deferred to infrastructure change.
- `ParsingConfig.ID` and `ParsingConfig.Name` — populated by the infrastructure layer (PG repo); the domain layer does not validate or reference them.
- Parser cache invalidation in `getJMESParser` — the cache is never invalidated during the lifetime of a `ParsingService` instance. Any config update requires a new `ParsingService` instance.

## Decisions

### Decision: `ConfigRepository.GetConfig(ctx, sourceID, docType)` — three-argument signature

The draft has `GetConfig(ctx, sourceID)` but the cache key in `getJMESParser` already includes `docType`, and the `ParsingConfig` struct has a `DocumentType` field. This was a latent inconsistency.

**Alternatives considered:**
- Keep `(ctx, sourceID)` and filter config by `docType` in the service — pushes complexity into the service and requires all configs to be loaded at once.
- Add a separate `GetDocumentTypeConfig` method — doubles the interface surface for no benefit.

**Chosen:** Change signature to `(ctx, sourceID, docType)`. The concrete PG repo will be updated in the infrastructure change. This is a breaking interface change.

#### Sub-decision: `PropExternalURL` is only required for `search_page` configs

`ParsingConfig.Validate()` only requires at least one param named `PropExternalURL` when `DocumentType == DocumentTypeSearchPage`. For `DocumentTypeDownloadedAdvert`, the URL is already known via `meta.ExternalURL` (passed by the caller), so no `PropExternalURL` param is needed. This was not explicitly called out in the draft and is enforced in the implementation.

### Decision: `SdocID = md5_hex(external_url)` for snippet identity

Each search page snippet (a surfed advert) must have a stable, globally unique `SdocID`. Using the parent search page's `SdocID` for all snippets causes primary-key collisions in storage and breaks the downstream `GetDocumentsMeta` lookup by sdoc_ids.

**Alternatives considered:**
- Sequential counter per search page — not globally unique, requires coordination.
- UUID v4 — random, no collision risk, but not deterministic (same URL produces different IDs on re-scrape).
- Parent SdocID + index — unique per page but changes if page ordering shifts on re-scrape.

**Chosen:** `md5_hex(url)`. Deterministic, globally unique, matches what the Python workflow expects (`SdocID(hashlib.md5(url.encode()).hexdigest()`).

#### Sub-decision: URL normalization before hashing

URLs that differ only in trailing slash, query-parameter ordering, or fragment identifier must produce the same `SdocID`, otherwise the same logical advert scraped at different times or via different redirect paths creates duplicate database rows.

**Algorithm:**
1. Parse URL via `net/url.Parse`.
2. If parsing fails, return `ErrValidation` — the input is invalid.
3. Lowercase the scheme and host (`u.Scheme`, `u.Host`).
4. Strip the fragment (`u.Fragment = ""`).
5. Sort query parameters by key and reconstruct — `url.Values` orders keys on `Encode()`, achieving stable order.
6. Trim a trailing slash from the path only if the path is not the root `/` (`strings.TrimSuffix(u.Path, "/")` when `u.Path != "/"`).
7. Serialize the normalized URL and compute `md5_hex`.

**Alternatives considered:**
- Normalize path trailing slash to `/` for all URLs — changes the identity of root-page URLs.
- Sort query params by insertion order (unstable) — same URL can produce different hashes.
- Strip all query params — loses critical identity for sites where the query string identifies the advert.

**Chosen:** Algorithm above. Acceptable for v1; avoids re-crawl fragmentation creating duplicate `SdocID`s from trivially-different URLs.

### Decision: `JMESParser.Parse` unmarshals body internally

The jmespath library operates on decoded JSON data structures (`map[string]any`, slices, etc.), not raw strings. The draft passes `body []byte` directly to `expr.Search(body)`, which would treat the string itself as the JMESPath expression's result (or fail silently).

**Alternatives considered:**
- Require caller to unmarshal — shifts responsibility to caller; service would need to hold a pool of decoded buffers.
- Unmarshal in `ParsingService` methods before calling `Parse` — leaks JSON handling into the service layer.

**Chosen:** Unmarshal inside `JMESParser.Parse` at the start of each call. Keeps the interface clean (`body []byte → map[string][]any`), is consistent with how the jmespath library is used in its own examples, and isolates JSON handling to where it belongs.

### Decision: `singleflight.Group` for parser cache population

The `getJMESParser` slow path has a race: multiple goroutines for the same `(sourceID, docType)` can all pass the fast-path lock, then all call `ConfigRepository.GetConfig`, build a parser, and race to write the map. Only one result wins but work is wasted.

**Alternatives considered:**
- `sync.Map` — eliminates the write race but still allows multiple config fetches.
- Hold write lock during config fetch — blocks all other keys during one slow I/O; bad for a service with many sources.

**Chosen:** `singleflight.Group` keyed by `parserKey{sourceID, docType}`. All concurrent requests for the same key collapse to a single `ConfigRepository.GetConfig` call; subsequent cache hits are lock-free. Simple, correct, no external dependency (already using `golang.org/x/sync` in go.mod).

### Decision: Mockery for test doubles

Follows the pattern established in `example-site/` where mockery generates `*Mock*` types into `testutil` packages.

**Alternatives considered:**
- Hand-rolled stubs — more boilerplate, not generated, risk of going stale.
- `gomock` — similar capability, but mockery is already used in the Go portion of this repo.

**Chosen:** Mockery + `testify/mock`. Add `crawler/parser/.mockery.yml` following the example-site pattern.

## Risks / Trade-offs

- **Breaking interface change** (`ConfigRepository.GetConfig` signature) means any existing caller must be updated. All callers are in this change's scope (application layer, deferred), so the risk is contained.
- **`md5` for SdocID** is not cryptographically secure, but SdocID is not a security-sensitive function — it's a stable identifier. MD5 is fast and widely available in stdlib.
- **`json.Unmarshal` into `any`** will coerce all JSON numbers to `float64` (the standard Go behavior). Downstream consumers (e.g., advert property extractors) must handle this. Not a new constraint — any JSON unmarshaling in Go has this behavior.
- **Test coverage of the PG layer** is deferred to infrastructure change. The domain tests use mocks, so they test logic only; integration tests for the actual PG repo are out of scope here.
- **Singleflight + cancellation propagation** — if the first goroutine in a collapsing call is cancelled (context deadline/expiry), the cancellation is propagated to the in-flight `ConfigRepository.GetConfig` call, so all waiting goroutines receive the context error. Callers must treat `context.Canceled` as a retryable condition. Acceptable for v1; callers should retry on `context.Canceled`.
