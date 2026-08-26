import datetime as dt

import sqlalchemy as sa
import sqlalchemy.orm as sa_orm
from sqlalchemy import text

from shared.py.db.metadata import Base


class DocumentORM(Base):
    __tablename__ = "documents"

    sdoc_id: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text, primary_key=True)
    source_id: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text, primary_key=True)
    doc_type: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text, primary_key=True)
    external_url: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text)
    body: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text)
    created_at: sa_orm.Mapped[dt.datetime] = sa_orm.mapped_column(sa.DateTime(timezone=True))
    updated_at: sa_orm.Mapped[dt.datetime] = sa_orm.mapped_column(sa.DateTime(timezone=True))
    update_interval_sec: sa_orm.Mapped[int] = sa_orm.mapped_column(
        sa.Integer, server_default=text("86400")
    )
