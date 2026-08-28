# documents Delta

## ADDED Requirements

### Requirement: Parser may populate content_url via ContentURLTemplate

When a parser activity uses a `ParsingConfig` with a non-empty `ContentURLTemplate`, the produced surfed advert document's `ContentURL` field SHALL be set to the rendered template result. When `ContentURLTemplate` is empty, the `ContentURL` field SHALL be set to the empty string.

#### Scenario: content_url populated by parser via ContentURLTemplate

- **WHEN** `ParseSearchPage` produces a surfed advert with `ContentURL = "https://cdn.example.com/item1"` (from a non-empty `ContentURLTemplate`)
- **THEN** the saved document has `content_url = "https://cdn.example.com/item1"` in the database

#### Scenario: content_url empty when template not set

- **WHEN** `ParseSearchPage` produces a surfed advert with `ContentURL = ""` (empty `ContentURLTemplate`)
- **THEN** the saved document has `content_url = ""` in the database
