"""seed download_configs

Revision ID: 0006
Revises: 0005
Create Date: 2026-08-24

"""
from collections.abc import Sequence

import sqlalchemy as sa
from alembic import op
from sqlalchemy.dialects.postgresql import insert as pg_insert

revision: str = "0006"
down_revision: str | None = "0005"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    download_configs_table = sa.table(
        "download_configs",
        sa.column("source_id", sa.Text()),
        sa.column("doc_type", sa.Text()),
        sa.column("name", sa.Text()),
        sa.column("headers", sa.JSON()),
    )

    rows = [
        {
            "source_id": "siteapi",
            "doc_type": "search_page",
            "name": "siteapi-search",
            "headers": {"X-Demo": "demo"},
        },
        {
            "source_id": "siteapi",
            "doc_type": "surfed_advert",
            "name": "siteapi-advert",
            "headers": {"X-Demo": "demo"},
        },
    ]

    stmt = pg_insert(download_configs_table).values(rows).on_conflict_do_nothing(
        index_elements=["source_id", "doc_type"]
    )
    op.get_bind().execute(stmt)


def downgrade() -> None:
    op.execute(
        sa.text(
            "DELETE FROM download_configs WHERE source_id = 'siteapi' "
            "AND doc_type IN ('search_page', 'surfed_advert')"
        )
    )
