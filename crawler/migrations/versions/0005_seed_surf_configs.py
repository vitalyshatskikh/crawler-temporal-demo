"""seed surf_configs

Revision ID: 0005
Revises: 0004
Create Date: 2026-08-24

"""
from collections.abc import Sequence

import sqlalchemy as sa
from alembic import op
from sqlalchemy.dialects.postgresql import insert as pg_insert

revision: str = "0005"
down_revision: str | None = "0004"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    surf_configs_table = sa.table(
        "surf_configs",
        sa.column("name", sa.Text()),
        sa.column("source_id", sa.Text()),
        sa.column("url_template", sa.Text()),
        sa.column("url_template_params", sa.JSON()),
        sa.column("max_pages", sa.Integer()),
        sa.column("cron_schedule", sa.Text()),
        sa.column("update_interval_sec", sa.Integer()),
    )

    rows = [
        {
            "name": "siteapi-local-debug",
            "source_id": "siteapi",
            "url_template": "http://localhost:8090/adverts/{{region}}/search?page={{page}}",
            "url_template_params": [
                {"values": {"region": "moscow"}, "comment": "moscow"},
                {"values": {"region": "pekin"}, "comment": "pekin"},
                {"values": {"region": "new-york"}, "comment": "new-york"},
            ],
            "max_pages": 5,
            "cron_schedule": "0 1 * * *",
            "update_interval_sec": 600,
        },
        {
            "name": "siteapi-demo-fresh",
            "source_id": "siteapi",
            "url_template": "http://siteapi:8080/adverts/{{region}}/search?page={{page}}",
            "url_template_params": [
                {"values": {"region": "moscow"}, "comment": "moscow"},
                {"values": {"region": "pekin"}, "comment": "pekin"},
                {"values": {"region": "new-york"}, "comment": "new-york"},
            ],
            "max_pages": 5,
            "cron_schedule": "0/2 * * * *",
            "update_interval_sec": 7200,
        },
        {
            "name": "siteapi-demo-all",
            "source_id": "siteapi",
            "url_template": "http://siteapi:8080/adverts/{{region}}/search?page={{page}}",
            "url_template_params": [
                {"values": {"region": "moscow"}, "comment": "moscow"},
                {"values": {"region": "pekin"}, "comment": "pekin"},
                {"values": {"region": "new-york"}, "comment": "new-york"},
            ],
            "max_pages": 100,
            "cron_schedule": "0 * * * *",
            "update_interval_sec": 7200,
        },
    ]

    stmt = pg_insert(surf_configs_table).values(rows).on_conflict_do_nothing(
        index_elements=["name"]
    )
    op.get_bind().execute(stmt)


def downgrade() -> None:
    op.execute(
        sa.text(
            "DELETE FROM surf_configs WHERE name IN "
            "('siteapi-local-debug', 'siteapi-demo-fresh', 'siteapi-demo-all')"
        )
    )
