# documents Specification

## Purpose

Stores crawled web documents with a composite primary key `(sdoc_id, source_id, doc_type)` and a `update_interval_sec` field with DB default 86400, expressing how often the document should be refreshed.

## Requirements

### Requirement: Documents table primary key

The `documents` table SHALL use a composite primary key of `(sdoc_id, source_id, doc_type)`.

#### Scenario: Composite PK constraint

- **WHEN** two rows with the same `(sdoc_id, source_id, doc_type)` are inserted
- **THEN** the second insert fails with a primary key violation

#### Scenario: Same sdoc_id with different source_id

- **WHEN** two rows share the same `sdoc_id` but differ in `source_id`
- **THEN** both rows are stored as distinct entries

#### Scenario: Same sdoc_id with different doc_type

- **WHEN** two rows share the same `sdoc_id` but differ in `doc_type`
- **THEN** both rows are stored as distinct entries

### Requirement: update_interval_sec field

The `documents` table SHALL include an `update_interval_sec INTEGER NOT NULL DEFAULT 86400` column.

#### Scenario: Default value

- **WHEN** a row is inserted without specifying `update_interval_sec`
- **THEN** the column value is 86400

#### Scenario: Custom value

- **WHEN** a row is inserted with `update_interval_sec = 3600`
- **THEN** the stored value is 3600

### Requirement: DocumentMeta model requires update_interval_sec > 0

The `DocumentMeta` type in both Python (surfer) and Go (parser) SHALL include an `update_interval_sec` integer field validated as greater than zero.

#### Scenario: Zero value rejected

- **WHEN** a `DocumentMeta` is constructed with `update_interval_sec = 0`
- **THEN** validation fails

#### Scenario: Negative value rejected

- **WHEN** a `DocumentMeta` is constructed with `update_interval_sec = -1`
- **THEN** validation fails

#### Scenario: Positive value accepted

- **WHEN** a `DocumentMeta` is constructed with `update_interval_sec = 86400`
- **THEN** validation succeeds

### Requirement: Parsed documents inherit update_interval_sec

When a parser activity produces output documents, each output `DocumentMeta` SHALL inherit the `UpdateIntervalSec` value from its parent input document.

#### Scenario: ParseSearchPage propagates to children

- **WHEN** `ParseSearchPage` receives a `DocumentMeta` with `UpdateIntervalSec = 3600`
- **THEN** each output `DocumentMeta` in the returned slice has `UpdateIntervalSec = 3600`

#### Scenario: ParseAdvertContent propagates to output

- **WHEN** `ParseAdvertContent` receives a `DocumentMeta` with `UpdateIntervalSec = 7200`
- **THEN** the output `DocumentMeta` has `UpdateIntervalSec = 7200`
