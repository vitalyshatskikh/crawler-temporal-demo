# parser-config Specification

## Purpose

Defines the configuration structure for the parser's JMESPath-based document extraction. Controls URL extraction from search pages and allows mustache templates to transform URLs into canonical and content-fetch forms.

## ADDED Requirements

### Requirement: ParsingConfig structure

The `ParsingConfig` type SHALL contain the following fields:

- `SourceID` (string): identifies the document source.
- `DocumentType` (string): `search_page` or `downloaded_advert`.
- `ExternalURLJMESPath` (string): JMESPath expression evaluated against the search page body to produce a list of URL strings. Required when `DocumentType == "search_page"`. Empty for `downloaded_advert` configs.
- `ExternalURLTemplate` (string): Optional mustache template rendered for each extracted URL. When empty, the URL string itself is used as the external URL.
- `ContentURLTemplate` (string): Optional mustache template rendered for each extracted URL. When empty, the produced document's `ContentURL` field is set to the empty string.
- `Params` ([]ParsingParam): extraction rules for additional fields added to each produced document's body. Each `ParsingParam` has `Name`, `JMESPath`, and `Default`.

The `_` (underscore) prefix is reserved for system-injected mustache context keys. `Param.Name` MUST NOT start with `_`.

#### Scenario: search_page config with ExternalURLJMESPath

- **WHEN** `ParsingConfig{DocumentType: "search_page", ExternalURLJMESPath: "urls[*]", ExternalURLTemplate: "", ContentURLTemplate: "", Params: []}` is validated
- **THEN** validation succeeds

#### Scenario: search_page config without ExternalURLJMESPath

- **WHEN** `ParsingConfig{DocumentType: "search_page", ExternalURLJMESPath: "", Params: []}` is validated
- **THEN** validation fails with an error indicating `ExternalURLJMESPath` is required

#### Scenario: downloaded_advert config with empty ExternalURLJMESPath

- **WHEN** `ParsingConfig{DocumentType: "downloaded_advert", ExternalURLJMESPath: "", Params: []}` is validated
- **THEN** validation succeeds (URL is already known from the input document)

#### Scenario: downloaded_advert config with non-empty ContentURLTemplate

- **WHEN** `ParsingConfig{DocumentType: "downloaded_advert", ExternalURLJMESPath: "", ContentURLTemplate: "https://cdn.example.com/{{_external_url}}", Params: []}` is validated
- **THEN** validation succeeds (template is stored; `ParseAdvertContent` does not consume it in this change)

#### Scenario: Param.Name starting with underscore rejected

- **WHEN** `ParsingConfig{DocumentType: "search_page", ExternalURLJMESPath: "urls", Params: [{Name: "_external_url", JMESPath: "url", Default: ""}]}` is validated
- **THEN** validation fails with an error indicating `_`-prefixed param names are reserved

#### Scenario: Param.Name with underscore in middle accepted

- **WHEN** `ParsingConfig{DocumentType: "search_page", ExternalURLJMESPath: "urls", Params: [{Name: "my_external_url", JMESPath: "url", Default: ""}]}` is validated
- **THEN** validation succeeds

### Requirement: ExternalURLTemplate rendering

When `ExternalURLTemplate` is non-empty, it SHALL be rendered as a mustache template using the extracted snippet as context, and the result SHALL be used as the `ExternalURL` of each produced surfed advert document.

The mustache context map SHALL contain:

- The key `"_external_url"` mapped to the raw URL string extracted via `ExternalURLJMESPath`.
- For each entry in `Params`, the key matching `param.Name` mapped to the param's value at the snippet index (or the param's `Default` if the index is out of range).

#### Scenario: identity template

- **WHEN** `ExternalURLTemplate` is `"{{_external_url}}"` and the extracted URL is `"https://example.com/item1"`
- **THEN** the rendered URL is `"https://example.com/item1"`

#### Scenario: template with query parameter

- **WHEN** `ExternalURLTemplate` is `"{{_external_url}}?ref=search"` and the extracted URL is `"https://cdn.example.com/item1"`
- **THEN** the rendered URL is `"https://cdn.example.com/item1?ref=search"`

#### Scenario: template with param reference

- **WHEN** `ExternalURLTemplate` is `"{{_external_url}}?item={{title}}"` and the extracted URL is `"https://cdn.example.com/item1"` and `Params` contains `{Name: "title", JMESPath: "titles"}` with value `"Widget"` at snippet index
- **THEN** the rendered URL is `"https://cdn.example.com/item1?item=Widget"`

#### Scenario: empty template uses raw URL

- **WHEN** `ExternalURLTemplate` is `""` and the extracted URL is `"https://example.com/item1"`
- **THEN** the `ExternalURL` is `"https://example.com/item1"`

### Requirement: ContentURLTemplate rendering

When `ContentURLTemplate` is non-empty, it SHALL be rendered as a mustache template using the same snippet context as `ExternalURLTemplate`, and the result SHALL be used as the `ContentURL` of each produced surfed advert document.

When `ContentURLTemplate` is empty, the `ContentURL` field SHALL be set to the empty string, preserving the existing downloader fallback behavior (download URL falls back to `external_url`).

#### Scenario: CDN override template

- **WHEN** `ContentURLTemplate` is `"https://cdn.example.com{{_external_url}}"` and the extracted URL is `"https://search.example.com/item1"`
- **THEN** the produced document has `ContentURL = "https://cdn.example.com/https://search.example.com/item1"`

#### Scenario: empty template produces empty ContentURL

- **WHEN** `ContentURLTemplate` is `""` and the extracted URL is `"https://example.com/item1"`
- **THEN** the produced document has `ContentURL = ""`

### Requirement: Malformed template rejected at validation

A mustache template that fails to parse SHALL cause `ParsingConfig.Validate()` to return an error.

#### Scenario: unclosed mustache tag

- **WHEN** `ExternalURLTemplate` is `"{{_external_url"` (unclosed tag)
- **THEN** `Validate()` returns an error

#### Scenario: invalid mustache syntax

- **WHEN** `ExternalURLTemplate` is `"{{#if}}"` (incomplete section)
- **THEN** `Validate()` returns an error

### Requirement: ParseSearchPage produces documents using templates

For each URL string extracted via `ExternalURLJMESPath`, `ParseSearchPage` SHALL produce one `Document` with:

- `SdocID = md5_hex(normalized_url)` (stable, deterministic).
- `ExternalURL` set to the rendered `ExternalURLTemplate` result (or raw URL if template is empty).
- `ContentURL` set to the rendered `ContentURLTemplate` result (or `""` if template is empty).
- `Body` constructed by zipping parallel arrays from the parsed result (including `_external_url`) at the corresponding index.
- `UpdateIntervalSec` inherited from the parent search page document.

#### Scenario: single snippet with templates

- **WHEN** `ParseSearchPage` processes a search page with body `{"urls": ["https://a.com"], "titles": ["Item A"]}` and config has `ExternalURLJMESPath: "urls"`, `ExternalURLTemplate: "{{_external_url}}?src=parser"`, `ContentURLTemplate: ""`, `Params: [{Name: "title", JMESPath: "titles"}]`
- **THEN** exactly one document is produced with `ExternalURL = "https://a.com?src=parser"`, `ContentURL = ""`, `Body["title"] = "Item A"`

#### Scenario: multiple snippets with ContentURLTemplate

- **WHEN** `ParseSearchPage` processes a search page with body `{"urls": ["https://a.com", "https://b.com"]}` and config has `ExternalURLJMESPath: "urls"`, `ExternalURLTemplate: ""`, `ContentURLTemplate: "https://cdn.example.com{{_external_url}}"`
- **THEN** two documents are produced: first with `ExternalURL = "https://a.com"`, `ContentURL = "https://cdn.example.com/https://a.com"`; second with `ExternalURL = "https://b.com"`, `ContentURL = "https://cdn.example.com/https://b.com"`

#### Scenario: empty ExternalURLJMESPath result produces zero documents

- **WHEN** `ParseSearchPage` processes a search page with body `{"urls": []}` and config has `ExternalURLJMESPath: "urls"`
- **THEN** zero documents are produced and no save calls occur


