import sqlalchemy as sa
import sqlalchemy.orm as sa_orm

from shared.py.db.metadata import Base


class DownloadConfigORM(Base):
    __tablename__ = "download_configs"

    id: sa_orm.Mapped[int] = sa_orm.mapped_column(sa.BigInteger, primary_key=True)
    source_id: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text, index=True)
    doc_type: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text)
    name: sa_orm.Mapped[str] = sa_orm.mapped_column(sa.Text)
    headers: sa_orm.Mapped[dict[str, str]] = sa_orm.mapped_column(sa.JSON)
