## MODIFIED Requirements

### Requirement: ParseSearchPage takes a pre-fetched Document

**Breaking change from `parser-domain-finalize`:** `ParseSearchPage` no longer accepts `DocumentMeta` and does not fetch the document from a repository. The caller is responsible for fetching the `Document` and passing it in. The `AdvertsRepository` dependency is removed from `ParsingService`.

`ParsingService.ParseSearchPage(ctx, doc Document)` SHALL:
1. Validate `doc` via `doc.Validate()` — if `doc.Type != DocumentTypeSearchPage` or the embedded `DocumentMeta` is invalid, return an error wrapping `ErrValidation`.
2. Fetch the `ParsingConfig` for `(doc.SourceID, DocumentTypeSearchPage)` from the config repository — if not found, return an error wrapping `ErrNotFound`.
3. Run `JMESParser.Parse` on `doc.Body`.
4. For each extracted `external_url`, produce a `Document` with:
   - `SdocID = SdocIDForURL(external_url)` (unique per snippet).
   - `SourceID = doc.SourceID`.
   - `Type = DocumentTypeSurfedAdvert`.
   - `CreatedAt` and `UpdatedAt` set to the same timestamp.
   - `ExternalURL` set to the extracted URL.
   - `Body` set to the JSON-encoded map of all extracted fields for that snippet index.

#### Scenario: ParseSearchPage returns unique SdocIDs per snippet
- **WHEN** `ParseSearchPage` is called with a search page Document containing two advert URLs
- **THEN** the two returned Documents have different `SdocID` values derived from their respective URLs

#### Scenario: ParseSearchPage with zero snippets returns empty slice
- **WHEN** `ParseSearchPage` is called with a search page body containing no matching URLs
- **THEN** the returned slice is empty with no error

#### Scenario: ParseSearchPage validates doc before processing
- **WHEN** `ParseSearchPage` is called with invalid `DocumentMeta` (empty SourceID)
- **THEN** the error wraps `ErrValidation`

#### Scenario: ParseSearchPage returns correct document type
- **WHEN** `ParseSearchPage` is called with a valid search page Document
- **THEN** each returned Document has `Type == DocumentTypeSurfedAdvert`

#### Scenario: ParseSearchPage with wrong doc.Type returns ErrValidation
- **WHEN** `ParseSearchPage` is called with `doc.Type == DocumentTypeDownloadedAdvert`
- **THEN** the error wraps `ErrValidation`

#### Scenario: ParseSearchPage with missing config returns ErrNotFound
- **WHEN** `ParseSearchPage` is called and config repository returns `ErrNotFound`
- **THEN** the error wraps `ErrNotFound`

#### Scenario: ParseSearchPage with cancelled context returns context error
- **WHEN** a cancelled context is passed to `ParseSearchPage`
- **THEN** the error is `context.Canceled` or `context.DeadlineExceeded`

### Requirement: ParseAdvertContent takes a pre-fetched Document

**Breaking change from `parser-domain-finalize`:** `ParseAdvertContent` no longer accepts `DocumentMeta` and does not fetch the document from a repository. The caller is responsible for fetching the `Document` and passing it in.

`ParsingService.ParseAdvertContent(ctx, doc Document)` SHALL:
1. Validate `doc` via `doc.Validate()` — if `doc.Type != DocumentTypeDownloadedAdvert` or the embedded `DocumentMeta` is invalid, return an error wrapping `ErrValidation`.
2. Fetch the `ParsingConfig` for `(doc.SourceID, DocumentTypeDownloadedAdvert)` from the config repository — if not found, return an error wrapping `ErrNotFound`.
3. Run `JMESParser.Parse` on `doc.Body`.
4. Return a `Document` with:
   - `SdocID = doc.SdocID`.
   - `SourceID = doc.SourceID`.
   - `Type = DocumentTypeParsedAdvert`.
   - `ExternalURL = doc.ExternalURL`.
   - `CreatedAt` and `UpdatedAt` set to the same timestamp.
   - `Body` set to the JSON-encoded parsed map.

#### Scenario: ParseAdvertContent returns parsed document with correct type
- **WHEN** `ParseAdvertContent` is called with a valid advert Document
- **THEN** the returned Document has `Type == DocumentTypeParsedAdvert`

#### Scenario: ParseAdvertContent with missing config returns error wrapping ErrNotFound
- **WHEN** `ParseAdvertContent` is called and config repository returns `ErrNotFound`
- **THEN** the error wraps `ErrNotFound`

#### Scenario: ParseAdvertContent with wrong doc.Type returns ErrValidation
- **WHEN** `ParseAdvertContent` is called with `doc.Type == DocumentTypeSearchPage`
- **THEN** the error wraps `ErrValidation`

### Requirement: ParsingService constructor takes only ConfigRepository

**Breaking change from `parser-domain-finalize`:** `NewParsingService` no longer takes an `AdvertsRepository`. Document fetching is the caller's responsibility.

`NewParsingService(confRepo ConfigRepository)` SHALL return `ErrValidation` when `confRepo` is `nil`.

#### Scenario: Nil confRepo returns ErrValidation
- **WHEN** `NewParsingService(nil)` is called
- **THEN** the error is `ErrValidation`

#### Scenario: Non-nil confRepo returns service
- **WHEN** `NewParsingService(mockConfRepo)` is called
- **THEN** a non-nil `*ParsingService` is returned with the config repo wired

#### Scenario: Service has no document repository dependency
- **WHEN** a `*ParsingService` is constructed with a valid `ConfigRepository`
- **THEN** the service has no `AdvertsRepository` field
