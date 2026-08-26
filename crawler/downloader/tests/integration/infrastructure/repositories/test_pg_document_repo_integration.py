import datetime as dt
from uuid import uuid4

import pytest
import sqlalchemy as sa
import sqlalchemy.ext.asyncio as sa_asyncio

from downloader.infrastructure.repositories.document_repo import PGDocumentRepository
from shared.py.db import orm as shared_orm
from surfer.domain import adverts
from surfer.domain.adverts.models import DocumentType, SdocID, SourceID

pytestmark = pytest.mark.integration


class TestSave:
    async def test_when_new_document_then_inserts_row(
        self,
        sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
        pg_document_repo: PGDocumentRepository,
    ) -> None:
        now = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)
        sdoc_id = "ndoc-" + uuid4().hex[:12]

        doc = adverts.Document(
            sdoc_id=SdocID(sdoc_id),
            source_id=SourceID("test-source"),
            type=DocumentType.DOWNLOADED_ADVERT,
            external_url=f"https://example.com/{sdoc_id}",
            body="<html>new document body</html>",
            created_at=now,
            updated_at=now,
            update_interval_sec=86400,
        )

        await pg_document_repo.save(doc)

        async with sessionmaker() as session:
            row = await session.execute(
                sa.select(shared_orm.DocumentORM).where(
                    shared_orm.DocumentORM.sdoc_id == sdoc_id
                ),
            )
            result_row = row.scalars().one_or_none()

        assert result_row is not None
        assert result_row.sdoc_id == sdoc_id
        assert result_row.body == "<html>new document body</html>"
        assert result_row.source_id == "test-source"
        assert result_row.doc_type == "downloaded_advert"

    async def test_when_existing_sdoc_id_then_updates_body_and_preserves_created_at(
        self,
        sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
        pg_document_repo: PGDocumentRepository,
    ) -> None:
        now = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)
        sdoc_id = "udoc-" + uuid4().hex[:12]

        async with sessionmaker() as session:
            await session.execute(
                sa.insert(shared_orm.DocumentORM).values(
                    sdoc_id=sdoc_id,
                    source_id="update-source",
                    doc_type="downloaded_advert",
                    external_url=f"https://example.com/{sdoc_id}",
                    body="<html>original body</html>",
                    created_at=now,
                    updated_at=now,
                ),
            )
            await session.commit()

        updated_at = dt.datetime(2025, 6, 1, tzinfo=dt.UTC)
        updated_doc = adverts.Document(
            sdoc_id=SdocID(sdoc_id),
            source_id=SourceID("update-source"),
            type=DocumentType.DOWNLOADED_ADVERT,
            external_url=f"https://example.com/{sdoc_id}",
            body="<html>updated body with new content</html>",
            created_at=now,
            updated_at=updated_at,
            update_interval_sec=86400,
        )

        await pg_document_repo.save(updated_doc)

        async with sessionmaker() as session:
            row = await session.execute(
                sa.select(shared_orm.DocumentORM).where(
                    shared_orm.DocumentORM.sdoc_id == sdoc_id
                ),
            )
            result_row = row.scalars().one()

        assert result_row.body == "<html>updated body with new content</html>"
        assert result_row.created_at == now
        assert result_row.updated_at == updated_at

    async def test_when_existing_sdoc_id_then_does_not_create_duplicate(
        self,
        sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
        pg_document_repo: PGDocumentRepository,
    ) -> None:
        now = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)
        sdoc_id = "ddoc-" + uuid4().hex[:12]

        async with sessionmaker() as session:
            await session.execute(
                sa.insert(shared_orm.DocumentORM).values(
                    sdoc_id=sdoc_id,
                    source_id="dup-source",
                    doc_type="downloaded_advert",
                    external_url=f"https://example.com/{sdoc_id}",
                    body="<html>first</html>",
                    created_at=now,
                    updated_at=now,
                ),
            )
            await session.commit()

        doc = adverts.Document(
            sdoc_id=SdocID(sdoc_id),
            source_id=SourceID("dup-source"),
            type=DocumentType.DOWNLOADED_ADVERT,
            external_url=f"https://example.com/{sdoc_id}",
            body="<html>second</html>",
            created_at=now,
            updated_at=now,
            update_interval_sec=86400,
        )

        await pg_document_repo.save(doc)

        async with sessionmaker() as session:
            count = await session.execute(
                sa.select(sa.func.count()).select_from(shared_orm.DocumentORM).where(
                    shared_orm.DocumentORM.sdoc_id == sdoc_id
                ),
            )
            total = count.scalar()

        assert total == 1


class TestCompositeKeyIndependence:
    async def test_when_same_sdoc_id_different_source_id_then_both_rows_persist(
        self,
        sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
        pg_document_repo: PGDocumentRepository,
    ) -> None:
        now = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)
        sdoc_id = "composite-" + uuid4().hex[:12]

        async with sessionmaker() as session:
            await session.execute(
                sa.insert(shared_orm.DocumentORM).values(
                    sdoc_id=sdoc_id,
                    source_id="src-a",
                    doc_type="downloaded_advert",
                    external_url=f"https://example.com/{sdoc_id}",
                    body="<html>first</html>",
                    created_at=now,
                    updated_at=now,
                    update_interval_sec=86400,
                ),
            )
            await session.execute(
                sa.insert(shared_orm.DocumentORM).values(
                    sdoc_id=sdoc_id,
                    source_id="src-b",
                    doc_type="downloaded_advert",
                    external_url=f"https://example.com/{sdoc_id}",
                    body="<html>second</html>",
                    created_at=now,
                    updated_at=now,
                    update_interval_sec=86400,
                ),
            )
            await session.commit()

        async with sessionmaker() as session:
            row_a = await session.execute(
                sa.select(shared_orm.DocumentORM).where(
                    sa.and_(
                        shared_orm.DocumentORM.sdoc_id == sdoc_id,
                        shared_orm.DocumentORM.source_id == "src-a",
                        shared_orm.DocumentORM.doc_type == "downloaded_advert",
                    )
                ),
            )
            row_b = await session.execute(
                sa.select(shared_orm.DocumentORM).where(
                    sa.and_(
                        shared_orm.DocumentORM.sdoc_id == sdoc_id,
                        shared_orm.DocumentORM.source_id == "src-b",
                        shared_orm.DocumentORM.doc_type == "downloaded_advert",
                    )
                ),
            )

        result_a = row_a.scalars().one_or_none()
        result_b = row_b.scalars().one_or_none()

        assert result_a is not None
        assert result_a.body == "<html>first</html>"
        assert result_b is not None
        assert result_b.body == "<html>second</html>"

    async def test_when_save_with_existing_composite_key_then_only_matching_row_updated(
        self,
        sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
        pg_document_repo: PGDocumentRepository,
    ) -> None:
        now = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)
        sdoc_id = "composite-update-" + uuid4().hex[:12]

        async with sessionmaker() as session:
            await session.execute(
                sa.insert(shared_orm.DocumentORM).values(
                    sdoc_id=sdoc_id,
                    source_id="src-x",
                    doc_type="downloaded_advert",
                    external_url=f"https://example.com/{sdoc_id}",
                    body="<html>original x</html>",
                    created_at=now,
                    updated_at=now,
                    update_interval_sec=86400,
                ),
            )
            await session.execute(
                sa.insert(shared_orm.DocumentORM).values(
                    sdoc_id=sdoc_id,
                    source_id="src-y",
                    doc_type="downloaded_advert",
                    external_url=f"https://example.com/{sdoc_id}",
                    body="<html>original y</html>",
                    created_at=now,
                    updated_at=now,
                    update_interval_sec=86400,
                ),
            )
            await session.commit()

        later = dt.datetime(2025, 6, 1, tzinfo=dt.UTC)
        doc_x_update = adverts.Document(
            sdoc_id=SdocID(sdoc_id),
            source_id=SourceID("src-x"),
            type=DocumentType.DOWNLOADED_ADVERT,
            external_url=f"https://example.com/{sdoc_id}-updated",
            body="<html>updated x</html>",
            created_at=now,
            updated_at=later,
            update_interval_sec=300,
        )

        await pg_document_repo.save(doc_x_update)

        async with sessionmaker() as session:
            row_x = await session.execute(
                sa.select(shared_orm.DocumentORM).where(
                    sa.and_(
                        shared_orm.DocumentORM.sdoc_id == sdoc_id,
                        shared_orm.DocumentORM.source_id == "src-x",
                    )
                ),
            )
            row_y = await session.execute(
                sa.select(shared_orm.DocumentORM).where(
                    sa.and_(
                        shared_orm.DocumentORM.sdoc_id == sdoc_id,
                        shared_orm.DocumentORM.source_id == "src-y",
                    )
                ),
            )

        result_x = row_x.scalars().one()
        result_y = row_y.scalars().one()

        assert result_x.body == "<html>updated x</html>"
        assert result_x.updated_at == later
        assert result_x.update_interval_sec == 300
        assert result_y.body == "<html>original y</html>"
        assert result_y.update_interval_sec == 86400
