## 1. Edit 0003 Migration

- [x] 1.1 Add `external_url_jmespath TEXT NOT NULL DEFAULT ''` column to `0003_create_parsing_configs.py` `create_table` call
- [x] 1.2 Add `external_url_template TEXT NOT NULL DEFAULT ''` column to `0003_create_parsing_configs.py` `create_table` call
- [x] 1.3 Add `content_url_template TEXT NOT NULL DEFAULT ''` column to `0003_create_parsing_configs.py` `create_table` call
- [x] 1.4 Verify `downgrade()` in 0003 correctly drops the table (no changes needed — `op.drop_table` cascades)

## 2. Remove 0007 Migration

- [x] 2.1 Delete `crawler/migrations/versions/0007_add_parsing_config_url_fields.py`
- [x] 2.2 Delete `crawler/migrations/versions/__pycache__/0007_add_parsing_config_url_fields.cpython-314.pyc`

## 3. Verify

- [x] 3.1 Run `cd crawler && poetry run ruff check migrations/versions/` — lint clean
- [x] 3.2 Run `cd crawler && poetry run alembic history` — chain ends at 0006, no 0007
- [x] 3.3 Run `cd crawler && poetry run alembic upgrade head` against a fresh DB — succeeds, `parsing_configs` has all 6 columns
- [x] 3.4 Run `cd crawler && poetry run alembic downgrade base && poetry run alembic upgrade head` — both directions work
- [x] 3.5 Run `cd crawler && poetry run pytest -m "not linting and not integration"` — tests pass
