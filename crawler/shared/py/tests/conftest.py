import asyncio
import os
import typing as tp

import alembic.command
import alembic.config
import asyncpg
import pytest
import sqlalchemy.ext.asyncio as sa_asyncio

from shared.py import settings
from shared.py.db import engine as db_engine
from shared.py.db import settings as db_settings

__all__ = [
    "admin_pool",
    "apply_migrations",
    "ensure_test_db",
    "pg_config",
    "sessionmaker",
    "test_db_dsn",
    "test_db_name",
]


@pytest.fixture(scope="session")
def test_db_name() -> str:
    return f"crawler-test-{os.getpid()}"


@pytest.fixture(scope="session")
def pg_config(test_db_name: str) -> db_settings.PGConfig:
    cfg = settings.Config()
    cfg.postgres.database = test_db_name
    return cfg.postgres


@pytest.fixture(scope="session")
def test_db_dsn(pg_config: db_settings.PGConfig) -> str:
    return pg_config.to_dsn()


@pytest.fixture(scope="session")
async def admin_pool(pg_config: db_settings.PGConfig) -> asyncpg.Pool:
    pool = await asyncpg.create_pool(
        f"postgresql://{pg_config.user}:{pg_config.password}@{pg_config.hosts}",
        min_size=1,
        max_size=5,
    )
    yield pool
    await pool.close()


@pytest.fixture(scope="session", autouse=True)
async def ensure_test_db(
    admin_pool: asyncpg.Pool,
    test_db_name: str,
) -> None:
    async with admin_pool.acquire() as conn:
        await conn.execute(f"""DROP DATABASE IF EXISTS "{test_db_name}" """)
        await conn.execute(f"""CREATE DATABASE "{test_db_name}" """)

    yield

    async with admin_pool.acquire() as conn:
        await conn.execute(
            "SELECT pg_terminate_backend(pid) "
            "FROM pg_stat_activity "
            "WHERE datname = $1 AND pid <> pg_backend_pid()",
            test_db_name,
        )
        await conn.execute(f"""DROP DATABASE IF EXISTS "{test_db_name}" """)


@pytest.fixture(scope="session", autouse=True)
async def apply_migrations(ensure_test_db: tp.Any, test_db_dsn: str) -> None:
    cfg = alembic.config.Config(config_args={"script_location": "migrations",})
    cfg.set_main_option("sqlalchemy.url", test_db_dsn)
    await asyncio.to_thread(alembic.command.upgrade, cfg, "head")


@pytest.fixture
async def sessionmaker(test_db_dsn: str) -> sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession]:
    engine = sa_asyncio.create_async_engine(
        test_db_dsn,
        pool_size=10,
        max_overflow=10,
        pool_pre_ping=True,
        future=True,
    )
    yield db_engine.make_sessionmaker(engine)
    await engine.dispose()
