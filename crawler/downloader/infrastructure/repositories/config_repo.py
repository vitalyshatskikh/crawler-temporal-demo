import sqlalchemy as sa
import sqlalchemy.ext.asyncio as sa_asyncio

from downloader.domain import downloading, errors
from downloader.infrastructure.db import mappers
from downloader.infrastructure.db.orm import download_configs as orm
from surfer.domain import adverts


class PGDownloadingRepository(downloading.IDownloadingRepository):
    def __init__(self, sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession]) -> None:
        self._sessionmaker = sessionmaker

    async def get_download_config(
        self, source_id: adverts.SourceID, doc_type: adverts.DocumentType,
    ) -> downloading.Params:
        async with self._sessionmaker() as session:
            stmt = sa.select(orm.DownloadConfigORM).where(
                orm.DownloadConfigORM.source_id == source_id,
                orm.DownloadConfigORM.doc_type == doc_type.value,
            )
            row = (await session.execute(stmt)).scalar_one_or_none()
            if row is None:
                raise errors.NotFoundError(
                    f"download config not found: source={source_id} doc_type={doc_type}",
                )
            return mappers.download_config_to_params(row)
