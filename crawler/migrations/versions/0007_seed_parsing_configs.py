"""seed parsing_configs

Revision ID: 0007
Revises: 0006
Create Date: 2026-08-29

"""
from collections.abc import Sequence

import sqlalchemy as sa
from alembic import op
from sqlalchemy.dialects.postgresql import insert as pg_insert

revision: str = "0007"
down_revision: str | None = "0006"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    parsing_configs_table = sa.table(
        "parsing_configs",
        sa.column("source_id", sa.Text()),
        sa.column("doc_type", sa.Text()),
        sa.column("name", sa.Text()),
        sa.column("config", sa.JSON()),
        sa.column("external_url_jmespath", sa.Text()),
        sa.column("external_url_template", sa.Text()),
        sa.column("content_url_template", sa.Text()),
    )

    rows = [
        {
            "source_id": "siteapi-local",
            "doc_type": "search_page",
            "name": "siteapi-search-page",
            "config": [],
            "external_url_jmespath": "adverts[*].url",
            "external_url_template": "http://siteapi:8080{{_external_url}}",
            "content_url_template": "http://localhost:8090{{_external_url}}",
        },
        {
            "source_id": "siteapi-local",
            "doc_type": "downloaded_advert",
            "name": "siteapi-downloaded-advert",
            "config": [
                {"name": "title", "jmespath": "title", "default": ""},
                {"name": "description", "jmespath": "description", "default": ""},
                {"name": "price", "jmespath": "price", "default": ""},
                {"name": "pub_date", "jmespath": "pubDate", "default": ""},
            ],
            "external_url_jmespath": "",
            "external_url_template": "",
            "content_url_template": "",
        },
        {
            "source_id": "siteapi",
            "doc_type": "search_page",
            "name": "siteapi-search-page",
            "config": [],
            "external_url_jmespath": "adverts[*].url",
            "external_url_template": "http://siteapi:8080{{_external_url}}",
            "content_url_template": "http://siteapi:8080{{_external_url}}",
        },
        {
            "source_id": "siteapi",
            "doc_type": "downloaded_advert",
            "name": "siteapi-downloaded-advert",
            "config": [
                {"name": "title", "jmespath": "title", "default": ""},
                {"name": "description", "jmespath": "description", "default": ""},
                {"name": "price", "jmespath": "price", "default": ""},
                {"name": "pub_date", "jmespath": "pubDate", "default": ""},
            ],
            "external_url_jmespath": "",
            "external_url_template": "",
            "content_url_template": "",
        },
    ]

    stmt = pg_insert(parsing_configs_table).values(rows).on_conflict_do_nothing(
        index_elements=["source_id", "doc_type"]
    )
    op.get_bind().execute(stmt)


def downgrade() -> None:
    op.execute(
        sa.text(
            "DELETE FROM parsing_configs "
            "WHERE source_id = 'siteapi' "
            "AND doc_type IN ('search_page', 'downloaded_advert')"
        )
    )
