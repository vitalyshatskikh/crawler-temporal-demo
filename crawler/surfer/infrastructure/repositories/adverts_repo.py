import sqlalchemy as sa
import sqlalchemy.ext.asyncio as sa_asyncio

from shared.py.db import mappers, orm
from surfer.domain import adverts
from surfer.domain.adverts.models import SdocID


class PGAdvertsRepo(adverts.IAdvertsRepository):
    def __init__(self, sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession]) -> None:
        self._sessionmaker = sessionmaker

    async def get_documents_meta_by_sdoc_id(
        self, sdoc_ids: list[adverts.SdocID],
    ) -> dict[adverts.SdocID, adverts.DocumentMeta]:
        if not sdoc_ids:
            return {}
        async with self._sessionmaker() as session:
            stmt = sa.select(orm.DocumentORM).where(
                orm.DocumentORM.sdoc_id.in_(sdoc_ids)
            )
            rows = (await session.execute(stmt)).scalars().all()
            return {SdocID(row.sdoc_id): mappers.document_to_meta(row) for row in rows}
