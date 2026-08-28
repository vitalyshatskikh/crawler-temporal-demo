"""create parsing_configs table

Revision ID: 0003
Revises: 0002
Create Date: 2026-08-23

"""
from collections.abc import Sequence

import sqlalchemy as sa
from alembic import op

revision: str = "0003"
down_revision: str | None = "0002"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.create_table(
        "parsing_configs",
        sa.Column("id", sa.BigInteger(), primary_key=True, autoincrement=True),
        sa.Column("source_id", sa.Text(), nullable=False),
        sa.Column("doc_type", sa.Text(), nullable=False),
        sa.Column("name", sa.Text(), nullable=False),
        sa.Column("config", sa.JSON(), nullable=False, server_default="{}"),
        sa.Column("external_url_jmespath", sa.Text(), nullable=False, server_default=""),
        sa.Column("external_url_template", sa.Text(), nullable=False, server_default=""),
        sa.Column("content_url_template", sa.Text(), nullable=False, server_default=""),
    )
    op.create_index("idx_parsing_configs_source_id", "parsing_configs", ["source_id"])
    op.create_unique_constraint(
        "uq_parsing_configs_source_doc_type",
        "parsing_configs",
        ["source_id", "doc_type"],
    )


def downgrade() -> None:
    op.drop_constraint(
        "uq_parsing_configs_source_doc_type", "parsing_configs", type_="unique"
    )
    op.drop_index("idx_parsing_configs_source_id", table_name="parsing_configs")
    op.drop_table("parsing_configs")
