## Purpose

Defines the domain layer for the Go parser service: document identity models, per-source+doctype extraction configuration, JMESPath-based content extraction, and the parser service operations used by Temporal activities on the `parsing` queue.

## ADDED Requirements

### Requirement: Document model validation

The `DocumentMeta` type SHALL enforce the following invariants:
- `SourceID` is non-empty.
- `SdocID` is non-empty.
- `ExternalURL` is non-empty.
- `CreatedAt` is not the zero value.
- `UpdatedAt` is not the zero value.
- `UpdatedAt` is not earlier than `CreatedAt`.

#### Scenario: Valid DocumentMeta passes validation
- **WHEN** `DocumentMeta{SourceID: "src1", SdocID: "abc", ExternalURL: "https://example.com", CreatedAt: T, UpdatedAt: T}` is validated
- **THEN** validation succeeds with no error

#### Scenario: Empty SourceID fails validation
- **WHEN** `DocumentMeta{SourceID: "", SdocID: "abc", ExternalURL: "https://example.com"}` is validated
- **THEN** validation returns `ErrValidation`

#### Scenario: Empty SdocID fails validation
- **WHEN** `DocumentMeta{SourceID: "src1", SdocID: "", ExternalURL: "https://example.com"}` is validated
- **THEN** validation returns `ErrValidation`

#### Scenario: Empty ExternalURL fails validation
- **WHEN** `DocumentMeta{SourceID: "src1", SdocID: "abc", ExternalURL: ""}` is validated
- **THEN** validation returns `ErrValidation`

#### Scenario: UpdatedAt before CreatedAt fails validation
- **WHEN** `DocumentMeta{SourceID: "src1", SdocID: "abc", ExternalURL: "https://example.com", CreatedAt: T+1s, UpdatedAt: T}` is validated
- **THEN** validation returns `ErrValidation`

#### Scenario: Zero CreatedAt fails validation
- **WHEN** `DocumentMeta{SourceID: "src1", SdocID: "abc", ExternalURL: "https://example.com", CreatedAt: zero, UpdatedAt: T}` is validated
- **THEN** validation returns `ErrValidation`

#### Scenario: Zero UpdatedAt fails validation
- **WHEN** `DocumentMeta{SourceID: "src1", SdocID: "abc", ExternalURL: "https://example.com", CreatedAt: T, UpdatedAt: zero}` is validated
- **THEN** validation returns `ErrValidation`

### Requirement: SdocID is deterministically derived from ExternalURL

The system SHALL compute `SdocID` as `md5_hex(external_url)` (lower-case hex, 32 characters) so that the same URL always produces the same ID.

#### Scenario: Same URL produces same SdocID
- **WHEN** `SdocIDForURL("https://example.com/1")` is called twice
- **THEN** both calls return the same 32-character lower-case hex string

#### Scenario: Different URLs produce different SdocIDs
- **WHEN** `SdocIDForURL("https://example.com/1")` and `SdocIDForURL("https://example.com/2")` are called
- **THEN** the two results are different

#### Scenario: URLs differing only in trailing slash produce same SdocID
- **WHEN** `SdocIDForURL("https://example.com/page/")` and `SdocIDForURL("https://example.com/page")` are called
- **THEN** both calls return the same SdocID

#### Scenario: URLs differing only in query parameter order produce same SdocID
- **WHEN** `SdocIDForURL("https://example.com/search?a=1&b=2")` and `SdocIDForURL("https://example.com/search?b=2&a=1")` are called
- **THEN** both calls return the same SdocID

#### Scenario: URLs with fragment produce same SdocID as without fragment
- **WHEN** `SdocIDForURL("https://example.com/page#section")` and `SdocIDForURL("https://example.com/page")` are called
- **THEN** both calls return the same SdocID

#### Scenario: Invalid URL returns ErrValidation
- **WHEN** `SdocIDForURL("not a url")` is called
- **THEN** the error wraps `ErrValidation`

### Requirement: Parsing config validation

The `ParsingConfig` type SHALL enforce:
- `SourceID` is non-empty.
- `DocumentType` is non-empty.
- `Params` contains at least one element.
- If `DocumentType == DocumentTypeSearchPage`, at least one `ParsingParam` has `Name` equal to `PropExternalURL` (`"external_url"`). This is not required for `DocumentTypeDownloadedAdvert` configs because the URL is already known via `meta.ExternalURL`.

#### Scenario: Valid config passes validation
- **WHEN** `ParsingConfig{SourceID: "src1", DocumentType: "search_page", Params: [{Name: "external_url", JMESPath: "items[*].url", Default: ""}]}` is validated
- **THEN** validation succeeds

#### Scenario: Missing external_url param on search_page config fails validation
- **WHEN** `ParsingConfig{SourceID: "src1", DocumentType: "search_page", Params: [{Name: "title", JMESPath: "items[*].title", Default: ""}]}` is validated
- **THEN** validation returns `ErrValidation`

#### Scenario: Advert config without external_url param is valid
- **WHEN** `ParsingConfig{SourceID: "src1", DocumentType: "downloaded_advert", Params: [{Name: "title", JMESPath: "title", Default: ""}]}` is validated
- **THEN** validation succeeds

#### Scenario: Empty SourceID fails validation
- **WHEN** `ParsingConfig{SourceID: "", DocumentType: "search_page", Params: [{Name: "external_url", JMESPath: "x", Default: ""}]}` is validated
- **THEN** validation returns `ErrValidation`

#### Scenario: Empty Params fails validation
- **WHEN** `ParsingConfig{SourceID: "src1", DocumentType: "search_page", Params: []}` is validated
- **THEN** validation returns `ErrValidation`

### Requirement: JMESParser unmarshals body before searching

When `JMESParser.Parse(ctx, body)` is called, the `body` string SHALL be unmarshaled as JSON before any JMESPath expression is evaluated.

#### Scenario: Valid JSON body returns parsed results
- **WHEN** `JMESParser.Parse` is called with body `"{\"items\": [{\"url\": \"https://a.com\"}, {\"url\": \"https://b.com\"}]}"` and config with `external_url` param `items[*].url`
- **THEN** the result contains `{"external_url": [{"url": "https://a.com"}, {"url": "https://b.com"}]}`

#### Scenario: Non-JSON body returns ErrUnmarshalBody
- **WHEN** `JMESParser.Parse` is called with body `"not json <html>"`
- **THEN** the error wraps `ErrUnmarshalBody`

#### Scenario: Missing JMESPath in body returns default value
- **WHEN** `JMESParser.Parse` is called with body `"{}"` and config with `external_url` param `items[*].url` with default `"none"`
- **THEN** the result contains `{"external_url": ["none"]}`

#### Scenario: Nil JMESPath result uses default value
- **WHEN** `JMESParser.Parse` is called with body `"{\"external_url\": null}"` and config with `external_url` param `external_url` with default `"missing"`
- **THEN** the result contains `{"external_url": ["missing"]}`

#### Scenario: Scalar JMESPath result is wrapped in slice
- **WHEN** `JMESParser.Parse` is called with body `"{\"title\": \"My Title\"}"` and config with `title` param `title`
- **THEN** the result contains `{"title": ["My Title"]}`

#### Scenario: Context cancellation returns context error
- **WHEN** a cancelled context is passed to `JMESParser.Parse`
- **THEN** the error is `context.Canceled` or `context.DeadlineExceeded`

### Requirement: ParseSearchPage returns one Document per extracted URL

`ParsingService.ParseSearchPage` SHALL:
1. Validate `meta` via `meta.Validate()` — if `meta.Type != DocumentTypeSearchPage` or any field is invalid, return an error wrapping `ErrValidation`.
2. Fetch the search page `Document` for `(meta.SdocID, meta.SourceID, DocumentTypeSearchPage)` from the document repository — if not found, return an error wrapping `ErrNotFound`.
3. Fetch the `ParsingConfig` for `(sourceID, DocumentTypeSearchPage)` from the config repository — if not found, return an error wrapping `ErrNotFound`.
4. Run `JMESParser.Parse` on the document body.
5. For each extracted `external_url`, produce a `Document` with:
   - `SdocID = SdocIDForURL(external_url)` (unique per snippet).
   - `Type = DocumentTypeSurfedAdvert`.
   - `CreatedAt` and `UpdatedAt` set to the same timestamp.
   - `ExternalURL` set to the extracted URL.
   - `Body` set to the JSON-encoded map of all extracted fields for that snippet index.

#### Scenario: ParseSearchPage returns unique SdocIDs per snippet
- **WHEN** `ParseSearchPage` is called with a search page containing two advert URLs
- **THEN** the two returned Documents have different `SdocID` values derived from their respective URLs

#### Scenario: ParseSearchPage with zero snippets returns empty slice
- **WHEN** `ParseSearchPage` is called with a search page body containing no matching URLs
- **THEN** the returned slice is empty with no error

#### Scenario: ParseSearchPage validates meta before processing
- **WHEN** `ParseSearchPage` is called with invalid `DocumentMeta` (empty SourceID)
- **THEN** the error wraps `ErrValidation`

#### Scenario: ParseSearchPage returns correct document type
- **WHEN** `ParseSearchPage` is called with valid input
- **THEN** each returned Document has `Type == DocumentTypeSurfedAdvert`

#### Scenario: ParseSearchPage with wrong meta.Type returns ErrValidation
- **WHEN** `ParseSearchPage` is called with `meta.Type == DocumentTypeDownloadedAdvert`
- **THEN** the error wraps `ErrValidation`

#### Scenario: ParseSearchPage with missing config returns ErrNotFound
- **WHEN** `ParseSearchPage` is called and config repository returns `ErrNotFound`
- **THEN** the error wraps `ErrNotFound`

#### Scenario: ParseSearchPage with cancelled context returns context error
- **WHEN** a cancelled context is passed to `ParseSearchPage`
- **THEN** the error is `context.Canceled` or `context.DeadlineExceeded`

#### Scenario: ParseSearchPage with missing document returns error wrapping ErrNotFound
- **WHEN** `ParseSearchPage` is called and the document repository returns `ErrNotFound`
- **THEN** the error wraps `ErrNotFound`

### Requirement: ParseAdvertContent returns a single parsed Document

`ParsingService.ParseAdvertContent` SHALL:
1. Validate `meta` via `meta.Validate()` — if `meta.Type != DocumentTypeDownloadedAdvert` or any field is invalid, return an error wrapping `ErrValidation`.
2. Fetch the advert `Document` for `(meta.SdocID, meta.SourceID, DocumentTypeDownloadedAdvert)` from the document repository — if not found, return an error wrapping `ErrNotFound`.
3. Fetch the `ParsingConfig` for `(sourceID, DocumentTypeDownloadedAdvert)` from the config repository — if not found, return an error wrapping `ErrNotFound`.
4. Run `JMESParser.Parse` on the document body.
5. Return a `Document` with:
   - `SdocID = meta.SdocID`.
   - `Type = DocumentTypeParsedAdvert`.
   - `CreatedAt` and `UpdatedAt` set to the same timestamp.
   - `ExternalURL` set to `meta.ExternalURL`.
   - `Body` set to the JSON-encoded parsed map.

#### Scenario: ParseAdvertContent returns parsed document with correct type
- **WHEN** `ParseAdvertContent` is called with valid meta
- **THEN** the returned Document has `Type == DocumentTypeParsedAdvert`

#### Scenario: ParseAdvertContent with missing config returns error
- **WHEN** `ParseAdvertContent` is called and config repository returns `ErrNotFound`
- **THEN** the error wraps the not-found error from the repository

#### Scenario: ParseAdvertContent with wrong meta.Type returns ErrValidation
- **WHEN** `ParseAdvertContent` is called with `meta.Type == DocumentTypeSearchPage`
- **THEN** the error wraps `ErrValidation`

#### Scenario: ParseAdvertContent with missing document returns error wrapping ErrNotFound
- **WHEN** `ParseAdvertContent` is called and the document repository returns `ErrNotFound`
- **THEN** the error wraps `ErrNotFound`

### Requirement: ParsingService constructor validates dependencies

`NewParsingService(confRepo, docRepo)` SHALL return `ErrValidation` when either argument is `nil`.

#### Scenario: Nil confRepo returns ErrValidation
- **WHEN** `NewParsingService(nil, mockDocRepo)` is called
- **THEN** the error is `ErrValidation`

#### Scenario: Nil docRepo returns ErrValidation
- **WHEN** `NewParsingService(mockConfRepo, nil)` is called
- **THEN** the error is `ErrValidation`

#### Scenario: Both non-nil returns service
- **WHEN** `NewParsingService(mockConfRepo, mockDocRepo)` is called
- **THEN** a non-nil `*ParsingService` is returned with the repos wired

### Requirement: JMESParser cache returns same instance on repeated calls

`ParsingService.getJMESParser` SHALL return the same `*JMESParser` instance on repeated calls with the same `(sourceID, docType)` key, without calling `ConfigRepository` again.

#### Scenario: Second call with same key returns cached parser
- **WHEN** `getJMESParser(ctx, "src1", "search_page")` is called twice in sequence
- **THEN** the second call does not invoke `ConfigRepository`

#### Scenario: Concurrent calls for same key collapse to single config fetch
- **WHEN** `getJMESParser` is called concurrently with the same `(sourceID, docType)`
- **THEN** `ConfigRepository.GetConfig` is invoked exactly once

#### Scenario: getJMESParser with cancelled context returns context error without invoking repository
- **WHEN** a cancelled context is passed to `getJMESParser`
- **THEN** the error is `context.Canceled` or `context.DeadlineExceeded`, and `ConfigRepository` is not called
