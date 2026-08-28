## MODIFIED Requirements

### Requirement: content_url column

The `documents` table SHALL include a `content_url TEXT NOT NULL DEFAULT ''` column. When empty, the `external_url` column SHALL be used as the fetch URL. When non-empty, the `content_url` value SHALL be used instead of `external_url` for content retrieval.

#### Scenario: Default empty value

- **WHEN** a document row is inserted without specifying `content_url` (using raw SQL, bypassing the ORM)
- **THEN** the column value is `''`

Note: The ORM always emits a full column list, so this default is not exercised by application code paths.

#### Scenario: Non-empty override value

- **WHEN** a document row is inserted with `content_url = 'https://cdn.example.com/page1'`
- **THEN** the stored value is `'https://cdn.example.com/page1'`

### Requirement: content_url field on DocumentMeta

The `DocumentMeta` type in both Python (surfer) and Go (parser) SHALL include a `content_url` string field with no minimum length constraint. An empty string is a valid value.

#### Scenario: Empty content_url is valid

- **WHEN** a `DocumentMeta` is constructed with `content_url = ''`
- **THEN** validation succeeds

#### Scenario: Non-empty content_url is valid

- **WHEN** a `DocumentMeta` is constructed with `content_url = 'https://cdn.example.com/content'`
- **THEN** validation succeeds

### Requirement: Downloader uses content_url override

The `WebDownloader` activity SHALL fetch from `doc_meta.content_url` when it is non-empty, otherwise from `doc_meta.external_url`. The actual URL used for fetching SHALL be reported in log messages.

#### Scenario: content_url set — fetches from content_url

- **WHEN** `download_to_repo` is called with `doc_meta.content_url = 'https://cdn.example.com/x'` and `doc_meta.external_url = 'https://src.example.com/x'`
- **THEN** the HTTP GET targets `https://cdn.example.com/x`

#### Scenario: content_url empty — falls back to external_url

- **WHEN** `download_to_repo` is called with `doc_meta.content_url = ''` and `doc_meta.external_url = 'https://src.example.com/x'`
- **THEN** the HTTP GET targets `https://src.example.com/x`

#### Scenario: Logged URL reflects actual fetch target

- **WHEN** `download_to_repo` is called with `doc_meta.content_url = 'https://cdn.example.com/x'`
- **THEN** log messages referencing the fetched URL report `https://cdn.example.com/x`

#### Scenario: Logged URL reflects fallback when content_url is empty

- **WHEN** `download_to_repo` is called with `doc_meta.content_url = ''` and `doc_meta.external_url = 'https://src.example.com/x'`
- **THEN** log messages referencing the fetched URL report `https://src.example.com/x`
