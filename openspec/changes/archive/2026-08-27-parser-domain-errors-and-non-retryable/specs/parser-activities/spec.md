## ADDED Requirements

### Requirement: Domain errors are marked non-retryable

When a parser activity returns an error whose wrapped chain contains
`domain.ErrValidation`, `domain.ErrNotFound`, or
`domain.ErrParsingFailed`, the activity MUST return that error as a
non-retryable error so the Temporal worker does not retry it.
Transient errors (for example, raw database errors and context
cancellation) MUST remain retryable.

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
