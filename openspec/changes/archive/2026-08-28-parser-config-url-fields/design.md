## Context

`ParseSearchPage` currently identifies advert URLs to extract via a convention: a `ParsingParam` whose `Name` equals the constant `PropExternalURL` (`"external_url"`). `ParsingConfig.Validate()` enforces that at least one such param exists for `search_page` configs, and `parser.go:59` retrieves the URLs via `urls := parsed[PropExternalURL]`. This implicit contract is easy to misconfigure and does not support URL transformation (e.g., CDN override, query parameter injection) or `content_url` population.

The `add-content-url-column` change added the `content_url` column to the `documents` table and updated the Python downloader to prefer it over `external_url` when set. The deferred follow-up ("Parser `content_url` population") was noted in its design: the Go parser would eventually extract or derive a `content_url` from search page data. This change implements that deferred work.

`ParseAdvertContent` is intentionally unchanged by this change. `ContentURLTemplate` is allowed on `downloaded_advert` configs (stored, not yet consumed) — `ParseAdvertContent` does not render mustache templates.

## Goals / Non-Goals

**Goals:**

- Replace the `PropExternalURL`-as-magic-param convention with explicit, required `ExternalURLJMESPath` on `ParsingConfig`.
- Add `ExternalURLTemplate` and `ContentURLTemplate` (mustache) to support URL transformation and `content_url` population.
- `ParsingConfig.Validate()` validates mustache template syntax at config load time (fail-fast).
- All existing `PropExternalURL` references in test fixtures and production code updated to use `ExternalURLJMESPath`.
- Three new columns added to `parsing_configs` table: `external_url_jmespath`, `external_url_template`, `content_url_template`.
- Reject `Param.Name` starting with `_` in `Validate()` to reserve the prefix for system-injected mustache context keys.

**Non-Goals:**

- Changing the `Params`-based body zipping mechanism (remains for extracting title, price, etc.).
- Adding a separate save flow for `content_url` — the produced `Document` is already saved in one call with both `ExternalURL` and `ContentURL` fields.
- Supporting mustache lambdas or custom tags — standard mustache spec only.
- Data migration to remove stale `Params` entries with `Name == "external_url"` (no underscore) from existing rows — deferred to a follow-up change.
- Template variable integrity check (verifying all `{{var}}` references resolve to a `Param.Name` or the reserved `_external_url` key) — intentionally out of scope; coupling between templates and param names is documented.

## Decisions

### Decision: `github.com/cbroglie/mustache` for Go mustache rendering

The Python side uses `chevron` (a mustache compiler). For Go, `github.com/cbroglie/mustache` is the most widely used pure-Go implementation with no C dependencies, making it compatible with all build environments.

**Alternatives considered:**

- `github.com/hoisie/mustache`: older, less maintained, same API shape.
- `github.com/Joker/jface`: more complex, not pure mustache.
- Build a custom mini-template engine: unnecessary scope.

**Chosen:** `github.com/cbroglie/mustache`. Pure Go, no external deps, supports the full mustache spec including lambdas.

### Decision: Three TEXT NOT NULL DEFAULT '' columns on `parsing_configs`

```
external_url_jmespath TEXT NOT NULL DEFAULT ''
external_url_template TEXT NOT NULL DEFAULT ''
content_url_template  TEXT NOT NULL DEFAULT ''
```

Using `NOT NULL DEFAULT ''` (consistent with the existing `content_url TEXT NOT NULL DEFAULT ''` on `documents`) avoids nullable column issues in Go's sqlc-generated code and the Python SQLAlchemy ORM. Empty string means "not configured" — no template applied for URL fields, `ExternalURLJMESPath` validation handles the required field for `search_page`.

**Alternatives considered:**

- Nullable columns: adds complexity in Go/Python mapping; `COALESCE` needed in queries.
- JSONB single column: would nest the new fields inside the existing `config JSONB` blob — cleaner for schema but requires parsing the JSON to read individual URL fields; mixes top-level config fields with the body-extraction `Params` array.

**Chosen:** Three top-level TEXT columns. Aligned with the existing `content_url` convention on `documents`, easy to index or constrain later, and sqlc generates clean scalar fields.

### Decision: Template validation in `ParsingConfig.Validate()`

`Validate()` uses `mustache.ParseString` to check the syntax of both mustache templates. If either template fails to parse, `Validate()` returns `ErrValidation` with a descriptive message. The parsed template result is discarded; rendering at parse time calls `mustache.Render` directly (re-compiling on each call).

**Rationale:** Fail-fast on configuration errors at service startup (or first use), not lazily at parse time where a bad template would corrupt the document stream. Re-compiling on each render is a simplicity trade-off; pre-compilation and caching is deferred.

**Known limitation:** `Validate()` does not verify that every `{{var}}` reference in a template corresponds to either the reserved `_external_url` key or a `Param.Name`. If a template references a param name that does not exist, rendering at parse time will produce an empty string for that variable. This coupling is documented; a future iteration may add an integrity check.

### Decision: Mustache context shape

When rendering a template for snippet `i` (0-based), the mustache context `map[string]any` contains:

```
{
  "_external_url": <raw URL string from ExternalURLJMESPath at index i>,
  <param.Name>: <parsed[param.Name][i] if i < len(parsed[param.Name]) else param.Default>
}
```

The `"_external_url"` key is the raw URL — it is always present (required for templates to reference). Other keys come from `Params` and are only present if the param's result array has a value at index `i`; otherwise the param's `Default` is used.

**Rationale:** The underscore prefix reserves the `_external_url` key as a system-injected mustache context key. This prevents any user-defined `Param.Name` from colliding with the canonical URL key. The `_` prefix convention is forward-compatible: future system keys (e.g., `_source_id`, `_doc_type`) will follow the same pattern. `Validate()` rejects any `Param.Name` starting with `_` to enforce this reservation.

### Decision: JMESParser extension

`NewJMESParser` is extended to also compile `ExternalURLJMESPath`. The compiled URL expression is stored on the `JMESParser` struct and its result is included in `Parse()`'s returned map under the reserved key `"_external_url"`. This keeps URL extraction within the same compilation/execution model as `Params` and avoids splitting responsibility.

**Alternatives considered:**

- Execute `ExternalURLJMESPath` separately in `parser.go` outside `JMESParser`: splits compilation and parsing, less consistent with how `Params` are handled.
- Compile URL expression lazily on first `ParseSearchPage` call: adds conditional logic; compile-once-at-construction is simpler.

**Chosen:** Extend `JMESParser`. Single compile, single `Parse()` call, result map shared for both mustache rendering and body zipping.

### Decision: `_external_url` is included in output `Document.Body`

`ParseSearchPage` builds the output `Document.Body` by iterating over all keys in the parsed result map (including `_external_url`) and zipping values at the snippet index. The `_external_url` key therefore appears in the serialized JSON body alongside other extracted fields. There is no filtering of `_`-prefixed keys from the output body.

### Decision: Drop `PropExternalURL` constant and implicit param requirement

The `PropExternalURL` constant and the `Validate()` loop checking `p.Name == PropExternalURL` are removed entirely. The required URL extraction signal moves to the `ExternalURLJMESPath` field.

**Migration:** Existing `parsing_configs` rows that used `Name: "external_url"` in `Params` for URL extraction must be migrated: remove that entry from `Params` and instead populate the new `external_url_jmespath` column with the same JMESPath expression. **Deferred to a follow-up change.** In the interim, the stale `Params` entry is tolerated — the new code ignores it for URL extraction (uses `_external_url` via `ExternalURLJMESPath` instead), but the stale entry will appear as a key in `Document.Body` for affected rows until cleaned up.

**Alternatives considered:**

- Keep `PropExternalURL` as a no-op constant: confuses future readers, does not clarify the new design.
- Deprecate `PropExternalURL` with a warning: unnecessary — this was an internal constant, not a user-facing API.

**Chosen:** Remove `PropExternalURL`. Update all test fixtures and factories that referenced it.

## Risks / Trade-offs

- **Stale `Params` entries with `Name == "external_url"` (no underscore)**: The data migration to remove these is deferred. Affected rows will produce an extra `external_url` key in `Document.Body` (from the stale `Params` entry) until a follow-up migration cleans them. URL extraction itself is correct because the canonical path is `_external_url` via `ExternalURLJMESPath`. `search_page` configs with empty `external_url_jmespath` will fail `Validate()` at runtime.
- **Mustache rendering per snippet adds latency**: For a search page with 1000 snippets, each snippet's mustache templates are rendered 1000 times. In practice, most `ExternalURLTemplate` values are likely empty (identity) or very simple (single variable expansion), so the cost is minimal. If performance becomes a concern, templates can be compiled once per URL-count bucket.
- **Template context coupling**: If a user changes a param name in `Params`, any mustache template referencing that param's name must also be updated. No automated check for this. Mitigation: document the coupling clearly; keep param names stable.
- **Go library addition**: Adding `github.com/cbroglie/mustache` to `go.mod` is a new transitive dependency. It is a mature, well-maintained library with no history of security issues.

## Migration Plan

1. **DB migration** (Alembic Python + Go schema.sql sync): Add three `TEXT NOT NULL DEFAULT ''` columns to `parsing_configs`. All existing rows get empty strings — valid for `downloaded_advert` configs (no `ExternalURLJMESPath` required), but `search_page` configs with empty `external_url_jmespath` will fail `Validate()` until the row is updated.
2. **Application deploy**: Code that reads the new columns. `search_page` configs with empty `external_url_jmespath` will fail validation at runtime (not silently broken). `downloaded_advert` configs are unaffected.
3. **Data update** (deferred to follow-up): Remove stale `Params` entries with `Name == "external_url"` (no underscore) from existing `parsing_configs` rows; populate `external_url_jmespath` from the extracted JMESPath expression.
4. **Rollback**: Revert columns (or set `external_url_jmespath` back to empty string for affected rows). Application code that references the new fields degrades gracefully when columns are absent or empty.
