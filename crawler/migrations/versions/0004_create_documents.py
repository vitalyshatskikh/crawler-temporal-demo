"""create documents table

Revision ID: 0004
Revises: 0003
Create Date: 2026-08-23

"""
from collections.abc import Sequence

import sqlalchemy as sa
from alembic import op

revision: str = "0004"
down_revision: str | None = "0003"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.create_table(
        "documents",
        sa.Column("sdoc_id", sa.Text(), primary_key=False),
        sa.Column("source_id", sa.Text(), nullable=False),
        sa.Column("doc_type", sa.Text(), nullable=False),
        sa.Column("external_url", sa.Text(), nullable=False),
        sa.Column("body", sa.Text(), nullable=False),
        sa.Column(
            "created_at",
            sa.DateTime(timezone=True),
            nullable=False,
        ),
        sa.Column(
            "updated_at",
            sa.DateTime(timezone=True),
            nullable=False,
        ),
        sa.Column("update_interval_sec", sa.Integer(), nullable=False, server_default="86400"),
        sa.CheckConstraint("updated_at >= created_at", name="ck_documents_updated_at"),
        sa.PrimaryKeyConstraint("sdoc_id", "source_id", "doc_type", name="pk_documents"),
    )
    op.create_index("idx_documents_source_id", "documents", ["source_id"])
    op.create_index("idx_documents_doc_type", "documents", ["doc_type"])


def downgrade() -> None:
    op.drop_index("idx_documents_doc_type", table_name="documents")
    op.drop_index("idx_documents_source_id", table_name="documents")
    op.drop_constraint("pk_documents", table_name="documents", type_="primary")
    op.drop_table("documents")
