## 1. Domain layer — struct and validation

- [x] 1.1 Add `github.com/cbroglie/mustache` to `crawler/go.mod` via `go get github.com/cbroglie/mustache` and `go mod tidy`
- [x] 1.2 Create `crawler/parser/internal/domain/mustache.go` — `RenderTemplate(tpl, ctx) string` that calls `mustache.Render` directly. Mustache context uses `_external_url` as the reserved key for the raw URL.
- [x] 1.3 Edit `crawler/parser/internal/domain/config.go` — remove `PropExternalURL` constant; add `ExternalURLJMESPath`, `ExternalURLTemplate`, `ContentURLTemplate` fields to `ParsingConfig`; rewrite `Validate()` to require `ExternalURLJMESPath != ""` for `search_page`, compile both templates (fail-fast on syntax errors), reject any `Param.Name` starting with `_` (reserved prefix for system keys), allow non-empty templates on `downloaded_advert` configs (stored but unused)
- [x] 1.3a Edit `crawler/parser/internal/domain/jmes.go` — extend `NewJMESParser` to also compile `ExternalURLJMESPath`; store it on the `JMESParser` struct; include the result under key `"_external_url"` in `Parse()`'s returned map

## 2. Domain layer — parser logic

- [x] 2.1 Edit `crawler/parser/internal/domain/parser.go` — in `ParseSearchPage`: replace `urls := parsed[PropExternalURL]` with `urls := parsed[ExternalURLKey]`; build snippet context map per iteration (including `_external_url`); render `ExternalURLTemplate` and `ContentURLTemplate` via `mustache.go` using the context map; set `ContentURL` on each output `Document`; wrap `SdocIDForURL` errors with template/snippet context

## 3. Database schema

- [x] 3.1 Create `crawler/migrations/versions/0007_add_parsing_config_url_fields.py` — Alembic migration adding `external_url_jmespath TEXT NOT NULL DEFAULT ''`, `external_url_template TEXT NOT NULL DEFAULT ''`, `content_url_template TEXT NOT NULL DEFAULT ''` to `parsing_configs`
- [x] 3.2 Edit `crawler/parser/db/schema.sql` — add same three columns to `parsing_configs` table
- [x] 3.3 Edit `crawler/parser/db/queries/parsing_configs.sql` — extend `GetParsingConfig` SELECT to include the three new columns
- [x] 3.4 Run `sqlc generate` in `crawler/parser/db/` to regenerate `gen/parsing_configs.sql.go` and `gen/models.go`

## 4. Infrastructure layer — repository

- [x] 4.1 Edit `crawler/parser/internal/infrastructure/repositories/pg_config_repo.go` — `GetConfig` maps the three new columns into `ParsingConfig`

## 5. Python ORM

- [x] 5.1 Edit `crawler/downloader/infrastructure/db/orm/parsing_configs.py` — add three `sa_orm.Mapped[str]` columns mapped to the new table columns

## 6. Tests — domain

- [x] 6.1 Edit `crawler/parser/internal/domain/testutil/factories.go` — update `ValidSearchPageConfigFactory` to set `ExternalURLJMESPath: "urls[*]"` (remove the `PropExternalURL` param); update `ValidAdvertConfigFactory` to remove the `PropExternalURL` param
- [x] 6.2 Create `crawler/parser/internal/domain/config_test.go` — tests for `ParsingConfig.Validate()`: `ExternalURLJMESPath` required for `search_page`, optional for `downloaded_advert`, malformed mustache template → error, empty template is valid, `Param.Name` starting with `_` → error, non-empty `ContentURLTemplate` on `downloaded_advert` → succeeds
- [x] 6.3 Edit `crawler/parser/internal/domain/parser_test.go` — replace all 6 occurrences of `{Name: domain.PropExternalURL, JMESPath: ...}` with `ExternalURLJMESPath: "urls"` (tests use `"urls"`, factory uses `"urls[*]"` — keep that distinction); add test cases: mustache-rendered `ExternalURL`, `ContentURL` populated from template, empty templates preserve identity/empty-ContentURL behavior, template using a `Params`-derived key

## 7. Tests — application layer

- [x] 7.1 Edit `crawler/parser/internal/application/activities/parser_test.go` — replace all 4 occurrences of `{Name: domain.PropExternalURL, ...}` with `ExternalURLJMESPath: "urls"`

## 8. Tests — infrastructure layer

- [x] 8.1 Edit `crawler/parser/internal/infrastructure/repositories/pg_config_repo_integration_test.go` — insert a config row with the three new columns; assert round-trip via `GetConfig` returns values matching the inserted data

## 9. Verification

- [x] 9.1 Run `go vet ./...` in `crawler/`
- [x] 9.2 Run `go test ./...` in `crawler/` (non-integration)
- [x] 9.3 Run `poetry run ruff check --fix .` in `crawler/` (Python ORM lint)
- [x] 9.4 Run `poetry run mypy surfer downloader` to verify Python types
- [x] 9.5 Run `poetry run pytest -m "not linting and not integration"` in `crawler/` to verify Python tests pass
