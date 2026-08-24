import sqlalchemy.ext.asyncio as sa_asyncio
from sqlalchemy.dialects.postgresql import insert as pg_insert

from downloader.domain import downloading
from shared.py.db import mappers, orm
from surfer.domain import adverts


class PGDocumentRepository(downloading.IDocumentRepository):
    def __init__(self, sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession]) -> None:
        self._sessionmaker = sessionmaker

    async def save(self, document: adverts.Document) -> None:
        row = mappers.document_to_orm(document)
        async with self._sessionmaker() as session:
            stmt = pg_insert(orm.DocumentORM).values(**row)
            stmt = stmt.on_conflict_do_update(
                index_elements=[orm.DocumentORM.sdoc_id],
                set_={
                    "body": stmt.excluded.body,
                    "updated_at": stmt.excluded.updated_at,
                    "external_url": stmt.excluded.external_url,
                    "doc_type": stmt.excluded.doc_type,
                    "source_id": stmt.excluded.source_id,
                },
            )
            await session.execute(stmt)
            await session.commit()
