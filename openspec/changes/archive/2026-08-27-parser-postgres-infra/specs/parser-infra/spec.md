## Purpose

Provides PostgreSQL-backed repository implementations for the parser worker: `PGConfigRepository` resolves `ParsingConfig` by `(source_id, doc_type)`, and `PGAdvertsRepository` persists and retrieves `Document` rows. Used by the `ParseSearchPage` and `ParseAdvertContent` Temporal activities.

## ADDED Requirements

### Requirement: PGConfigRepository.GetConfig returns config by source and doc type

`PGConfigRepository.GetConfig(ctx, sourceID, docType)` SHALL return the `ParsingConfig` whose row in `parsing_configs` has matching `source_id` and `doc_type` columns.

#### Scenario: Row exists for given source and doc type

- **WHEN** `GetConfig(ctx, "siteapi", "search_page")` is called and a row exists
- **THEN** the returned `ParsingConfig` has `SourceID = "siteapi"`, `DocumentType = "search_page"`, and `Params` populated from the `config` JSONB column

#### Scenario: Row exists returns ID and Name from row

- **WHEN** `GetConfig(ctx, "siteapi", "search_page")` is called and a row exists with `id=42, name="search-config"`
- **THEN** the returned `ParsingConfig` has `ID = 42` and `Name = "search-config"`

#### Scenario: No row exists for given source and doc type

- **WHEN** `GetConfig(ctx, "unknown", "search_page")` is called and no row exists
- **THEN** the call returns `ErrNotFound`

#### Scenario: config JSONB is malformed

- **WHEN** `GetConfig(ctx, "siteapi", "search_page")` is called and the `config` JSONB column contains invalid JSON
- **THEN** the call returns a wrapped error matching `ErrUnmarshalConfig`

### Requirement: parsing_configs.config JSONB stores params array

The `parsing_configs.config` column SHALL store a JSON array of objects, each having `name` (string), `jmespath` (string), and `default` (string) fields.

#### Scenario: Valid params JSONB

- **WHEN** a `parsing_configs` row has `config = [{"name":"external_url","jmespath":"urls[*]","default":""}]`
- **THEN** `GetConfig` returns `Params` with one `ParsingParam{Name:"external_url", JMESPath:"urls[*]", Default:""}`

#### Scenario: Empty params array

- **WHEN** a `parsing_configs` row has `config = []`
- **THEN** `GetConfig` returns `Params` as a non-nil empty `[]ParsingParam`

### Requirement: PGAdvertsRepository.GetDocument retrieves by composite PK

`PGAdvertsRepository.GetDocument(ctx, sdocID, sourceID, docType)` SHALL return the `Document` whose row in `documents` has matching `(sdoc_id, source_id, doc_type)`.

#### Scenario: Document exists

- **WHEN** `GetDocument(ctx, "abc123", "siteapi", "search_page")` is called and a row exists
- **THEN** the returned `Document` has `SdocID = "abc123"`, `SourceID = "siteapi"`, `Type = "search_page"`, and `Body` as `[]byte`

#### Scenario: Document does not exist

- **WHEN** `GetDocument(ctx, "notfound", "siteapi", "search_page")` is called and no row exists
- **THEN** the call returns `ErrNotFound`

### Requirement: PGAdvertsRepository.SaveDocument upserts by composite PK

`PGAdvertsRepository.SaveDocument(ctx, doc)` SHALL insert a new row or update an existing row in `documents` with the same `(sdoc_id, source_id, doc_type)` composite key.

#### Scenario: Insert new document

- **WHEN** `SaveDocument(ctx, doc)` is called where no row exists for `(doc.SdocID, doc.SourceID, doc.Type)`
- **THEN** a new row is inserted with all fields from `doc`

#### Scenario: Update existing document preserves created_at

- **WHEN** `SaveDocument(ctx, doc)` is called where a row already exists for `(doc.SdocID, doc.SourceID, doc.Type)`
- **THEN** the existing row is updated: `external_url`, `body`, `updated_at`, and `update_interval_sec` are set to the values from `doc`; `created_at` is preserved from the existing row

#### Scenario: Upserted document body is preserved on read

- **WHEN** a document is saved with `Body = []byte("{\"title\":\"test\"}")`
- **THEN** a subsequent `GetDocument` returns the same body bytes

### Requirement: documents table conforms to documents spec

The `documents` table SHALL conform to `openspec/specs/documents/spec.md`. The repository SHALL NOT redefine or extend those requirements.

### Requirement: Test database isolation

Integration tests SHALL create an isolated database per test run with schema applied from `crawler/parser/db/schema.sql`, and tear it down after the test run completes.

#### Scenario: Integration test creates isolated database

- **WHEN** integration tests are run with `//go:build integration`
- **THEN** a new database named `parser-{pid}` is created before tests, schema is applied, and the database is dropped after tests complete
