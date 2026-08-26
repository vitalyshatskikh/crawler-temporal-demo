import typing as tp

import sqlalchemy as sa
import sqlalchemy.orm as sa_orm
from sqlalchemy import text

from shared.py.db.metadata import Base


class SurfConfigORM(Base):
    __tablename__ = "surf_configs"

    id: sa_orm.Mapped[int] = sa_orm.mapped_column(sa.BigInteger, primary_key=True)
    name: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text, unique=True)
    source_id: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text, index=True)
    url_template: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text)
    url_template_params: sa_orm.Mapped[list[dict[str, tp.Any]]] = sa_orm.mapped_column(sa.JSON)
    max_pages: sa_orm.Mapped[int] = sa_orm.mapped_column(sa.Integer)
    cron_schedule: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text)
    update_interval_sec: sa_orm.Mapped[int] = sa_orm.mapped_column(
        sa.Integer, server_default=text("86400")
    )
