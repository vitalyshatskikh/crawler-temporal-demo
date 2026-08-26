"""create download_configs table

Revision ID: 0002
Revises: 0001
Create Date: 2026-08-23

"""
from collections.abc import Sequence

import sqlalchemy as sa
from alembic import op

revision: str = "0002"
down_revision: str | None = "0001"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.create_table(
        "download_configs",
        sa.Column("id", sa.BigInteger(), primary_key=True, autoincrement=True),
        sa.Column("source_id", sa.Text(), nullable=False),
        sa.Column("doc_type", sa.Text(), nullable=False),
        sa.Column("name", sa.Text(), nullable=False),
        sa.Column("headers", sa.JSON(), nullable=False, server_default="{}"),
    )
    op.create_index("idx_download_configs_source_id", "download_configs", ["source_id"])
    op.create_unique_constraint(
        "uq_download_configs_source_doc_type",
        "download_configs",
        ["source_id", "doc_type"],
    )


def downgrade() -> None:
    op.drop_constraint(
        "uq_download_configs_source_doc_type", "download_configs", type_="unique"
    )
    op.drop_index("idx_download_configs_source_id", table_name="download_configs")
    op.drop_table("download_configs")
