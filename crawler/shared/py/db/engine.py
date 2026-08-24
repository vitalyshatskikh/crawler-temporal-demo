
import sqlalchemy.ext.asyncio as sa_asyncio

from shared.py.db.settings import PGConfig


def make_engine(cfg: PGConfig) -> sa_asyncio.AsyncEngine:
    return sa_asyncio.create_async_engine(
        cfg.to_dsn(),
        pool_size=10,
        max_overflow=10,
        pool_pre_ping=True,
        future=True,
    )


def make_sessionmaker(
    engine: sa_asyncio.AsyncEngine,
) -> sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession]:
    return sa_asyncio.async_sessionmaker(
        engine,
        expire_on_commit=False,
        class_=sa_asyncio.AsyncSession,
    )
