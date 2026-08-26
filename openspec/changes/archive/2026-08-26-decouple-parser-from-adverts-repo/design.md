## Context

`ParsingService` in `crawler/parser/internal/domain/parser.go` has a dependency on `AdvertsRepository`, which it uses to call `GetDocument` before parsing. This means the domain layer is responsible for knowing how to fetch a document — a persistence concern that belongs at the application layer (Temporal activities). The application will hold the `AdvertsRepository` implementation (backed by PostgreSQL) and is responsible for fetching; the parser should only transform `Document → []Document`.

**Current method signatures:**
```go
func NewParsingService(confRepo ConfigRepository, docRepo AdvertsRepository) (*ParsingService, error)
func (s *ParsingService) ParseSearchPage(ctx context.Context, meta DocumentMeta) ([]Document, error)
func (s *ParsingService) ParseAdvertContent(ctx context.Context, meta DocumentMeta) (Document, error)
```

**Target method signatures:**
```go
func NewParsingService(confRepo ConfigRepository) (*ParsingService, error)
func (s *ParsingService) ParseSearchPage(ctx context.Context, doc Document) ([]Document, error)
func (s *ParsingService) ParseAdvertContent(ctx context.Context, doc Document) (Document, error)
```

See proposal.md for full motivation.

## Goals / Non-Goals

**Goals:**
- Parser domain is a pure transformation: `Document in → parsed Documents out`
- `AdvertsRepository` interface lives at the application layer where document persistence is a concern
- No changes to `JMESParser` or `ParsingConfig` logic
- All existing tests updated; all tests still pass

**Non-Goals:**
- Implementing the Temporal activities that call `ParseSearchPage`/`ParseAdvertContent` — deferred to a later change
- Updating the archived `parser-domain-finalize` spec — accepted as drift
- Changing `DocumentMeta` validation logic (remains the caller's responsibility before constructing `Document`)

## Decisions

### 1. Remove `AdvertsRepository` from domain/repositories.go

**Decision:** Delete `AdvertsRepository` from `crawler/parser/internal/domain/repositories.go`. The domain no longer defines a persistence interface.

**Rationale:** The domain should not know how documents are loaded. If the parser held an interface, the application would be forced to implement it. By removing the interface from the domain entirely, the application is free to own the `AdvertsRepository` contract.

**Alternative considered:** Keep a minimal `DocumentProvider` interface in the domain. Rejected — the application already knows it will own document fetching; no need to encode it in the domain's interface contract.

### 2. Move `AdvertsRepository` to application layer

**Decision:** Create `crawler/parser/internal/application/repositories.go` containing the `AdvertsRepository` interface with `GetDocument` and `SaveDocument`.

**Rationale:** The application layer (Temporal activities) orchestrates workflow steps: fetch document → call parser → save result. It needs the `AdvertsRepository` interface to do this. Placing it at `application/` level keeps it close to where the Temporal activities live.

**Alternative considered:** Define in `crawler/parser/internal/infrastructure/`. Rejected — infrastructure is for concrete implementations (PostgreSQL, Redis), not interface definitions. The application layer is the right home for orchestration interfaces.

### 3. Method signatures take `Document` not `DocumentMeta`

**Decision:** Replace `meta DocumentMeta` parameter with `doc Document` in both `ParseSearchPage` and `ParseAdvertContent`. The `Document` struct embeds `DocumentMeta`, so callers set `doc.SourceID`, `doc.SdocID`, `doc.Type`, `doc.ExternalURL` and pass `doc.Body` containing the raw bytes to parse.

**Rationale:** Since the application fetches the document, it constructs the full `Document` including the body. The parser receives everything it needs and returns parsed output. No domain code ever needs to call `GetDocument` internally.

**Trade-off:** This is a breaking API change. All call sites in the application layer must be updated to construct `Document` instead of `DocumentMeta`. The benefit of a stateless, dependency-free parser outweighs this cost.

### 4. Retain `meta.Validate()` call at method entry

**Decision:** Both `ParseSearchPage` and `ParseAdvertContent` still call `doc.Validate()` (which validates the embedded `DocumentMeta`) at entry.

**Rationale:** The parser cannot parse a malformed input meaningfully. Validating at the boundary catches programmer errors early before JMESPath processing. The validation does not require any repository.

### 5. Update `.mockery.yml` to reflect new interface location

**Decision:** Remove `AdvertsRepository` from the `domain` package entry; add `AdvertsRepository` under a new `application` package entry in `crawler/parser/.mockery.yml`.

**Rationale:** Mocks for `application.AdvertsRepository` will be needed for Temporal activity tests in a later change. Generating them now keeps the mock configuration consistent.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Breaking API change — all call sites must update | This change is scoped to the parser module; the application layer doesn't exist yet, so no call sites to update. Future activity implementations will use the new signatures. |
| `mock_adverts_repository.go` is deleted but the mock might be needed for future app-layer tests | The mock is auto-generated from the interface definition. When the application layer is implemented, running `mockery --dir crawler/parser/internal/application/` will regenerate it. |
| Archived parser-domain spec has stale scenarios ("document repository returns ErrNotFound") | Accepted drift per user decision. The archived spec is historical; future spec work will correct it. |

## Open Questions

None. All decisions are made; implementation is fully specified.
