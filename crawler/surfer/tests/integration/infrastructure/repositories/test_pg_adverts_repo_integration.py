import datetime as dt
from uuid import uuid4

import pytest
import sqlalchemy as sa
import sqlalchemy.ext.asyncio as sa_asyncio

from shared.py.db import orm as shared_orm
from surfer.domain import adverts
from surfer.infrastructure.repositories.adverts_repo import PGAdvertsRepo

pytestmark = pytest.mark.integration


class TestGetDocumentsMetaBySdocId:
    async def test_when_rows_exist_then_returns_dict(
        self,
        sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
        pg_adverts_repo: PGAdvertsRepo,
    ) -> None:
        now = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)

        sdoc_id_1 = f"adoc-{uuid4().hex[:12]}"
        sdoc_id_2 = f"adoc-{uuid4().hex[:12]}"
        sdoc_id_3 = f"adoc-{uuid4().hex[:12]}"

        async with sessionmaker() as session:
            for sdoc_id in (sdoc_id_1, sdoc_id_2, sdoc_id_3):
                await session.execute(
                    sa.insert(shared_orm.DocumentORM).values(
                        sdoc_id=sdoc_id,
                        source_id="test-source",
                        doc_type="surfed_advert",
                        external_url=f"https://example.com/{sdoc_id}",
                        body="<html>test</html>",
                        created_at=now,
                        updated_at=now,
                    ),
                )
            await session.commit()

        result = await pg_adverts_repo.get_documents_meta_by_sdoc_id([
            adverts.SdocID(sdoc_id_1),
            adverts.SdocID(sdoc_id_2),
            adverts.SdocID(sdoc_id_3),
        ])

        assert len(result) == 3
        assert adverts.SdocID(sdoc_id_1) in result
        assert result[adverts.SdocID(sdoc_id_1)].source_id == adverts.SourceID("test-source")

    async def test_when_partial_match_then_returns_only_matches(
        self,
        sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
        pg_adverts_repo: PGAdvertsRepo,
    ) -> None:
        now = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)

        sdoc_id_1 = f"pdoc-{uuid4().hex[:12]}"
        sdoc_id_2 = f"pdoc-{uuid4().hex[:12]}"

        async with sessionmaker() as session:
            for sdoc_id in (sdoc_id_1, sdoc_id_2):
                await session.execute(
                    sa.insert(shared_orm.DocumentORM).values(
                        sdoc_id=sdoc_id,
                        source_id="partial-source",
                        doc_type="surfed_advert",
                        external_url=f"https://example.com/{sdoc_id}",
                        body="<html>partial</html>",
                        created_at=now,
                        updated_at=now,
                    ),
                )
            await session.commit()

        missing_id = adverts.SdocID(f"missing-{uuid4().hex[:8]}")

        result = await pg_adverts_repo.get_documents_meta_by_sdoc_id([
            adverts.SdocID(sdoc_id_1),
            adverts.SdocID(sdoc_id_2),
            missing_id,
        ])

        assert len(result) == 2
        assert adverts.SdocID(sdoc_id_1) in result
        assert adverts.SdocID(sdoc_id_2) in result
        assert missing_id not in result

    async def test_when_no_matches_then_returns_empty_dict(
        self,
        pg_adverts_repo: PGAdvertsRepo,
    ) -> None:
        result = await pg_adverts_repo.get_documents_meta_by_sdoc_id([
            adverts.SdocID(f"none1-{uuid4().hex[:8]}"),
            adverts.SdocID(f"none2-{uuid4().hex[:8]}"),
        ])

        assert result == {}

    async def test_when_empty_input_then_returns_empty_dict(
        self,
        pg_adverts_repo: PGAdvertsRepo,
    ) -> None:
        result = await pg_adverts_repo.get_documents_meta_by_sdoc_id([])

        assert result == {}
