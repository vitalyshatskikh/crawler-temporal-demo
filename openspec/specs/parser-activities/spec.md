# parser-activities

## Purpose

Define the behavior of the Temporal activities exposed by the parser worker — `ParseSearchPage` and `ParseAdvertContent` — that integrate the existing `domain.ParsingService` with the application-layer `AdvertsRepository`.

## Requirements

### Requirement: Parser activity construction

The system MUST provide a `NewParser` constructor accepting a non-nil `*domain.ParsingService` and a non-nil `application.AdvertsRepository`. When either dependency is nil, the constructor MUST return `nil` and `domain.ErrValidation`. When both are non-nil, it MUST return a non-nil `*Parser` and a nil error.

#### Scenario: nil service rejected

- **WHEN** `NewParser(nil, repo)` is called with a valid repository
- **THEN** the constructor returns `nil` and an error matching `domain.ErrValidation`

#### Scenario: nil repository rejected

- **WHEN** `NewParser(svc, nil)` is called with a valid service
- **THEN** the constructor returns `nil` and an error matching `domain.ErrValidation`

#### Scenario: valid dependencies accepted

- **WHEN** `NewParser(svc, repo)` is called with both dependencies non-nil
- **THEN** the constructor returns a non-nil `*Parser` and no error

### Requirement: ParseSearchPage fetches, parses, and persists

The `ParseSearchPage` activity MUST, given a `domain.DocumentMeta`:
0. Verify `meta.Type == DocumentTypeSearchPage`; if not, return `nil` and an error matching `domain.ErrValidation` without calling the repository.
1. Call `repo.GetDocument(ctx, meta.SdocID, meta.SourceID, meta.Type)`.
2. Pass the fetched `Document` to `svc.ParseSearchPage(ctx, doc)`.
3. For each `Document` returned, call `repo.SaveDocument(ctx, doc)`.
4. Return the saved `DocumentMeta` slice (one entry per parsed document) and a nil error.

If step 0 fails (invalid meta type), the activity MUST return `nil` and an error matching `domain.ErrValidation`. If step 1 fails, the activity MUST return `nil` and a wrapped error. If step 2 fails, the activity MUST return `nil` and a wrapped error without any save call. If any save in step 3 fails, the activity MUST stop persisting further documents and return `nil` and a wrapped error.

When the parsed list is empty, the activity MUST make no save calls and MUST return a non-nil, empty `[]domain.DocumentMeta`.

#### Scenario: search page with multiple advert URLs

- **WHEN** the repo returns a search page document and the parser produces N surfed adverts
- **THEN** `SaveDocument` is invoked exactly N times with the parsed documents, and the activity returns a slice of N `DocumentMeta` in order

#### Scenario: search page with no advert URLs

- **WHEN** the parser produces an empty list of surfed adverts
- **THEN** no `SaveDocument` calls occur and the activity returns an empty slice

#### Scenario: get-document error

- **WHEN** `repo.GetDocument` returns an error
- **THEN** the activity returns `nil` and a wrapped error; no `SaveDocument` calls occur

#### Scenario: parse error

- **WHEN** `svc.ParseSearchPage` returns an error
- **THEN** the activity returns `nil` and a wrapped error; no `SaveDocument` calls occur

#### Scenario: save error mid-loop

- **WHEN** at least one `repo.SaveDocument` call returns an error during the loop
- **THEN** the activity returns `nil` and a wrapped error, and no further saves are attempted

#### Scenario: invalid meta type rejected

- **WHEN** the activity is invoked with a `DocumentMeta` whose `Type` is not `DocumentTypeSearchPage`
- **THEN** the activity returns `nil` and an error matching `domain.ErrValidation`; no `GetDocument` or `SaveDocument` calls occur

### Requirement: ParseAdvertContent fetches, parses, and persists

The `ParseAdvertContent` activity MUST, given a `domain.DocumentMeta`:
0. Verify `meta.Type == DocumentTypeDownloadedAdvert`; if not, return `nil` and an error matching `domain.ErrValidation` without calling the repository.
1. Call `repo.GetDocument(ctx, meta.SdocID, meta.SourceID, meta.Type)`.
2. Pass the fetched `Document` to `svc.ParseAdvertContent(ctx, doc)`.
3. Call `repo.SaveDocument(ctx, parsed)`.
4. Return a nil error.

If step 0 fails (invalid meta type), the activity MUST return `nil` and an error matching `domain.ErrValidation`. If any subsequent step fails, the activity MUST return a wrapped error and MUST NOT call subsequent steps.

#### Scenario: successful parse and save

- **WHEN** the repo returns a downloaded advert document and the parser produces a parsed advert
- **THEN** `SaveDocument` is invoked exactly once with the parsed document and the activity returns nil

#### Scenario: get-document error

- **WHEN** `repo.GetDocument` returns an error
- **THEN** the activity returns a wrapped error and no save call occurs

#### Scenario: parse error

- **WHEN** `svc.ParseAdvertContent` returns an error
- **THEN** the activity returns a wrapped error and no save call occurs

#### Scenario: save error

- **WHEN** `repo.SaveDocument` returns an error
- **THEN** the activity returns a wrapped error

#### Scenario: invalid meta type rejected

- **WHEN** the activity is invoked with a `DocumentMeta` whose `Type` is not `DocumentTypeDownloadedAdvert`
- **THEN** the activity returns `nil` and an error matching `domain.ErrValidation`; no `GetDocument` or `SaveDocument` calls occur

### Requirement: Domain errors are marked non-retryable

When a parser activity returns an error whose wrapped chain contains `domain.ErrValidation`, `domain.ErrNotFound`, or `domain.ErrParsingFailed`, the activity MUST return that error as a non-retryable error so the Temporal worker does not retry it. Transient errors (for example, raw database errors and context cancellation) MUST remain retryable.

#### Scenario: validation error is non-retryable

- **WHEN** an activity fails validation (for example, invalid `meta.Type` or a `Document` that fails `Validate`)
- **THEN** the returned error is marked non-retryable and wraps `domain.ErrValidation`

#### Scenario: missing config is non-retryable

- **WHEN** `repo.GetDocument` or `ConfigRepository.GetConfig` returns `domain.ErrNotFound`
- **THEN** the returned activity error is marked non-retryable and wraps `domain.ErrNotFound`

#### Scenario: parsing failure is non-retryable

- **WHEN** the JMES parser fails to unmarshal the body or fails an `expr.Search`
- **THEN** the returned activity error is marked non-retryable and wraps `domain.ErrParsingFailed`

#### Scenario: transient repository error remains retryable

- **WHEN** `repo.GetDocument` or `repo.SaveDocument` returns an error that is not `domain.ErrValidation`, `domain.ErrNotFound`, or `domain.ErrParsingFailed` (for example, a raw database connectivity error)
- **THEN** the returned activity error is NOT marked non-retryable

#### Scenario: context cancellation remains retryable

- **WHEN** the activity fails with `context.Canceled` or `context.DeadlineExceeded` propagated from `domain.ParsingService` or `repo.GetDocument`
- **THEN** the returned activity error is NOT marked non-retryable
