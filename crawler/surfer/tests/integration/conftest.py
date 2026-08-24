import datetime as dt
import typing as tp
from uuid import uuid4

import pytest
import sqlalchemy as sa
import sqlalchemy.ext.asyncio as sa_asyncio

from shared.py.db import orm as shared_orm
from shared.py.tests.conftest import *
from surfer.infrastructure.db import orm as surfer_orm
from surfer.infrastructure.repositories.adverts_repo import PGAdvertsRepo
from surfer.infrastructure.repositories.config_repo import PGConfigRepository


@pytest.fixture
def pg_adverts_repo(
    sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
) -> PGAdvertsRepo:
    return PGAdvertsRepo(sessionmaker)


@pytest.fixture
def pg_config_repo(
    sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
) -> PGConfigRepository:
    return PGConfigRepository(sessionmaker)


async def insert_document(
    session: sa_asyncio.AsyncSession,
    *,
    sdoc_id: str | None = None,
    source_id: str | None = None,
    doc_type: str = "surfed_advert",
    external_url: str | None = None,
    body: str = "<html></html>",
    created_at: dt.datetime | None = None,
    updated_at: dt.datetime | None = None,
) -> str:
    if sdoc_id is None:
        sdoc_id = uuid4().hex[:16]
    if source_id is None:
        source_id = uuid4().hex[:12]
    if external_url is None:
        external_url = f"https://{source_id}.com/{sdoc_id}"
    if created_at is None:
        created_at = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)
    if updated_at is None:
        updated_at = created_at

    await session.execute(
        sa.insert(shared_orm.DocumentORM).values(
            sdoc_id=sdoc_id,
            source_id=source_id,
            doc_type=doc_type,
            external_url=external_url,
            body=body,
            created_at=created_at,
            updated_at=updated_at,
        ),
    )
    await session.commit()
    return sdoc_id


async def insert_surf_config(
    session: sa_asyncio.AsyncSession,
    *,
    id: int | None = None,
    name: str | None = None,
    source_id: str | None = None,
    url_template: str = "https://example.com/{{page}}",
    url_template_params: list[dict[str, tp.Any]] | None = None,
    max_pages: int = 5,
    cron_schedule: str = "0 * * * *",
) -> int:
    if id is None:
        id = int(uuid4().hex[:8], 16) % 100000
    if name is None:
        name = f"test-{uuid4().hex[:12]}"
    if source_id is None:
        source_id = name
    if url_template_params is None:
        url_template_params = [{"values": {"page": "1"}, "comment": ""}]

    await session.execute(
        sa.insert(surfer_orm.SurfConfigORM).values(
            id=id,
            name=name,
            source_id=source_id,
            url_template=url_template,
            url_template_params=url_template_params,
            max_pages=max_pages,
            cron_schedule=cron_schedule,
        ),
    )
    await session.commit()
    return id
