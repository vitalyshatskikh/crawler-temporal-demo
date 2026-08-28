import typing as tp

import sqlalchemy as sa
import sqlalchemy.orm as sa_orm

from shared.py.db.metadata import Base


class ParsingConfigORM(Base):
    __tablename__ = "parsing_configs"

    id: sa_orm.Mapped[int] = sa_orm.mapped_column(sa.BigInteger, primary_key=True)
    source_id: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text, index=True)
    doc_type: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text)
    name: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text)
    config: sa_orm.Mapped[dict[str, tp.Any]] = sa_orm.mapped_column(sa.JSON)
    external_url_jmespath: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text, default="")
    external_url_template: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text, default="")
    content_url_template: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text, default="")
