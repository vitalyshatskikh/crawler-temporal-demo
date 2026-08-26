## Context

The parser worker (`crawler/parser`) already provides everything the two parser activities need except their bodies:
- `domain.ParsingService` (real, fully tested) handles JMESPath-based document parsing.
- `application.AdvertsRepository` (interface) defines `GetDocument` / `SaveDocument`.
- A generated `MockAdvertsRepository` (testify template) and `MockConfigRepository` live under `internal/{application,domain}/testutil/`.
- `internal/domain/testutil/factories.go` builds valid `DocumentMeta` and `Document` values.

The current `parser.go` has two problems beyond the no-op bodies: it declares `repo *domain.AdvertsRepository` (no such concrete type — only the `application.AdvertsRepository` interface exists), and the constructor takes no dependencies. Both are fixed in this change.

## Goals / Non-Goals

**Goals**
- Wire `ParseSearchPage` and `ParseAdvertContent` through the existing `ParsingService` and `AdvertsRepository`.
- Make the activity dependencies explicit via a validated constructor.
- Cover every happy and error path with unit tests using generated mocks and existing factories.

**Non-Goals**
- Temporal SDK registration (`activity.Defn`) — the module does not currently import Temporal; that belongs in `cmd/parser/main.go`.
- Expanding the `AdvertsRepository` interface (e.g., transactional save).
- Changing `ParsingService` or its config repository.
- Integration / e2e tests.

## Decisions

### Decision: activity validates and uses `meta.Type` to identify which document to fetch

`DocumentMeta.Type` already carries `DocumentTypeSearchPage` for the parent of `ParseSearchPage` and `DocumentTypeDownloadedAdvert` for the parent of `ParseAdvertContent`. Before any I/O, the activity asserts `meta.Type` matches the expected `DocumentType*`; mismatch returns `domain.ErrValidation` so workflow mis-routing fails fast rather than as a wrapped "not found" from the repository layer. Passing the validated `meta.Type` into `repo.GetDocument` then removes a hard-coded constant from the activity and keeps the activity self-describing.

Alternatives considered: hard-code `DocumentTypeSearchPage` / `DocumentTypeDownloadedAdvert` in each method (rejected — duplicates caller intent, silently misbehaves on wrong meta); skip validation entirely (rejected — wrong-type meta surfaces as a wrapped "not found" error that is harder to debug than a clear `ErrValidation` at the activity boundary).

### Decision: `NewParser` returns `(*Parser, error)` and validates deps

Matches the `NewParsingService(confRepo ConfigRepository) (*ParsingService, error)` pattern already established in the domain layer. Both deps must be non-nil; nil returns `domain.ErrValidation`. The returned error is typed (sentinel) so callers can branch with `errors.Is`.

### Decision: real `ParsingService` in activity tests, mock only the repository

`MockConfigRepository` + real `ParsingService` are sufficient to exercise every code path of both activities. Mocking `ParsingService` would require a second interface and add no real coverage — the parsing logic is already covered by `domain/parser_test.go`. Mock only at the seam you control: `AdvertsRepository`.

### Decision: stop saving on first failure in `ParseSearchPage`

Surfacing the error to the workflow preserves Temporal's retry semantics. Earlier successful saves remain — the persistence layer treats them as idempotent by `SdocID`, so a retry will overwrite them. Aggregating errors or rolling back is rejected because `AdvertsRepository` has no transactional API and Temporal retries are the natural compensation mechanism.

### Decision: test factories live under `application/testutil/factories.go`

Mirrors the existing `domain/testutil/factories.go`. Add only the activity-specific builders we actually use; do not re-export the domain factories — Go's `internal/` rules mean test code already has direct access when needed.

## Architecture / Data Flow

```
ParseSearchPage(meta: DocumentMeta)
  │
  ├─► repo.GetDocument(ctx, meta.SdocID, meta.SourceID, meta.Type)
  │       │
  │       ▼
  │   Document (search_page)
  │       │
  │       ▼
  ├─► svc.ParseSearchPage(ctx, doc)
  │       │
  │       ▼
  │   []Document (surf adverts)
  │       │
  │       ├─► repo.SaveDocument(ctx, d[0]) ─┐
  │       ├─► repo.SaveDocument(ctx, d[1]) ─┤
  │       └─► ...                          │
  │                                        ▼
  └─► return []DocumentMeta{d[0].DocumentMeta, ...}, nil

ParseAdvertContent(meta: DocumentMeta)
  │
  ├─► repo.GetDocument(ctx, meta.SdocID, meta.SourceID, meta.Type)
  │       │
  │       ▼
  │   Document (downloaded_advert)
  │       │
  │       ▼
  ├─► svc.ParseAdvertContent(ctx, doc)
  │       │
  │       ▼
  │   Document (parsed_advert)
  │       │
  │       ▼
  └─► repo.SaveDocument(ctx, parsed) → return nil
```

Both activities wrap every downstream error with `fmt.Errorf("... %s/%s: %w", sourceID, sdocID, err)` so failures can be traced back to the specific document being processed.

## Risks / Trade-offs

- **`Parser.repo` field type change**: only existing reference is the broken declaration in the file we're rewriting; no other code references this struct today. No external breakage.
- **No temporal decorators yet**: methods remain plain functions. If the Temporal SDK is added later, `activity.Defn` decorators wrap these methods and `NewParser` still controls their dependencies.
- **Partial saves**: a mid-loop save failure leaves previously-saved documents persisted. Acceptable given idempotent SdocID-based persistence and Temporal retries.

## Open Questions

None — all material ambiguities (test layout, constructor validation, partial-save behavior) were resolved during exploration.
