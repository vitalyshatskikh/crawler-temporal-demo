"""create surf_configs table

Revision ID: 0001
Revises:
Create Date: 2026-08-23

"""
from collections.abc import Sequence

import sqlalchemy as sa
from alembic import op

revision: str = "0001"
down_revision: str | None = None
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.create_table(
        "surf_configs",
        sa.Column("id", sa.BigInteger(), primary_key=True, autoincrement=True),
        sa.Column("name", sa.Text(), nullable=False, unique=True),
        sa.Column("source_id", sa.Text(), nullable=False),
        sa.Column("url_template", sa.Text(), nullable=False),
        sa.Column("url_template_params", sa.JSON(), nullable=False),
        sa.Column("max_pages", sa.Integer(), nullable=False),
        sa.Column("cron_schedule", sa.Text(), nullable=False),
    )
    op.create_index("idx_surf_configs_source_id", "surf_configs", ["source_id"])


def downgrade() -> None:
    op.drop_index("idx_surf_configs_source_id", table_name="surf_configs")
    op.drop_table("surf_configs")
