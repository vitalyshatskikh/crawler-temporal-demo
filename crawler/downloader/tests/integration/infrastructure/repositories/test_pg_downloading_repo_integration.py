from uuid import uuid4

import pytest
import sqlalchemy as sa
import sqlalchemy.ext.asyncio as sa_asyncio

from downloader.domain import errors
from downloader.infrastructure.db import orm as downloader_orm
from downloader.infrastructure.repositories.config_repo import PGDownloadingRepository
from surfer.domain import adverts

pytestmark = pytest.mark.integration


class TestGetDownloadConfig:
    async def test_when_row_exists_then_returns_params(
        self,
        sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
        pg_downloading_repo: PGDownloadingRepository,
    ) -> None:
        source_id = f"dlsrc-{uuid4().hex[:12]}"

        async with sessionmaker() as session:
            await session.execute(
                sa.insert(downloader_orm.DownloadConfigORM).values(
                    source_id=source_id,
                    doc_type="search_page",
                    name="test download config",
                    headers={"User-Agent": "crawler-test/1.0", "Accept": "text/html"},
                ),
            )
            await session.commit()

        result = await pg_downloading_repo.get_download_config(
            adverts.SourceID(source_id),
            adverts.DocumentType.SEARCH_PAGE,
        )

        assert result.id is not None
        assert result.name == "test download config"
        assert result.source_id == source_id
        assert result.headers == {"User-Agent": "crawler-test/1.0", "Accept": "text/html"}

    async def test_when_row_missing_then_raises_not_found(
        self,
        pg_downloading_repo: PGDownloadingRepository,
    ) -> None:
        with pytest.raises(errors.NotFoundError):
            await pg_downloading_repo.get_download_config(
                adverts.SourceID(f"missing-{uuid4().hex[:8]}"),
                adverts.DocumentType.SEARCH_PAGE,
            )

    async def test_when_source_and_type_match_then_filters_correctly(
        self,
        sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
        pg_downloading_repo: PGDownloadingRepository,
    ) -> None:
        source_1 = f"src1-{uuid4().hex[:8]}"
        source_2 = f"src2-{uuid4().hex[:8]}"

        async with sessionmaker() as session:
            await session.execute(
                sa.insert(downloader_orm.DownloadConfigORM).values(
                    source_id=source_1,
                    doc_type="search_page",
                    name="config for src1 search",
                    headers={},
                ),
            )
            await session.execute(
                sa.insert(downloader_orm.DownloadConfigORM).values(
                    source_id=source_2,
                    doc_type="search_page",
                    name="config for src2 search",
                    headers={},
                ),
            )
            await session.execute(
                sa.insert(downloader_orm.DownloadConfigORM).values(
                    source_id=source_1,
                    doc_type="downloaded_advert",
                    name="config for src1 advert",
                    headers={},
                ),
            )
            await session.commit()

        result = await pg_downloading_repo.get_download_config(
            adverts.SourceID(source_1),
            adverts.DocumentType.SEARCH_PAGE,
        )

        assert result.id is not None
        assert result.name == "config for src1 search"
        assert result.source_id == source_1
