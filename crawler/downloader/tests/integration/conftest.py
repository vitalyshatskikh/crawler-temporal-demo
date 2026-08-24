import datetime as dt
from uuid import uuid4

import pytest
import sqlalchemy as sa
import sqlalchemy.ext.asyncio as sa_asyncio

from downloader.infrastructure.db import orm as downloader_orm
from downloader.infrastructure.repositories.config_repo import PGDownloadingRepository
from downloader.infrastructure.repositories.document_repo import PGDocumentRepository
from shared.py.db import orm as shared_orm
from shared.py.tests.conftest import *


@pytest.fixture
def pg_document_repo(
    sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
) -> PGDocumentRepository:
    return PGDocumentRepository(sessionmaker)


@pytest.fixture
def pg_downloading_repo(
    sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
) -> PGDownloadingRepository:
    return PGDownloadingRepository(sessionmaker)


async def insert_document(
    session: sa_asyncio.AsyncSession,
    *,
    sdoc_id: str | None = None,
    source_id: str | None = None,
    doc_type: str = "downloaded_advert",
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


async def insert_download_config(
    session: sa_asyncio.AsyncSession,
    *,
    id: int | None = None,
    source_id: str | None = None,
    doc_type: str = "search_page",
    name: str | None = None,
    headers: dict[str, str] | None = None,
) -> int:
    if id is None:
        id = int(uuid4().hex[:8], 16) % 100000
    if source_id is None:
        source_id = uuid4().hex[:12]
    if name is None:
        name = f"config-{uuid4().hex[:12]}"
    if headers is None:
        headers = {"User-Agent": "test-agent"}

    await session.execute(
        sa.insert(downloader_orm.DownloadConfigORM).values(
            id=id,
            source_id=source_id,
            doc_type=doc_type,
            name=name,
            headers=headers,
        ),
    )
    await session.commit()
    return id
