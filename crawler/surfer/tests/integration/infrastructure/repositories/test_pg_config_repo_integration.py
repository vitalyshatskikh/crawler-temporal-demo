from uuid import uuid4

import pytest
import sqlalchemy as sa
import sqlalchemy.ext.asyncio as sa_asyncio

from surfer.domain import errors, surfing
from surfer.infrastructure.db import orm as surfer_orm
from surfer.infrastructure.repositories.config_repo import PGConfigRepository

pytestmark = pytest.mark.integration


class TestGetSurfConfig:
    async def test_when_row_exists_then_returns_params(
        self,
        sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
        pg_config_repo: PGConfigRepository,
    ) -> None:
        config_name = f"cfg-{uuid4().hex[:12]}"

        async with sessionmaker() as session:
            await session.execute(
                sa.insert(surfer_orm.SurfConfigORM).values(
                    id=12345,
                    name=config_name,
                    source_id="source-for-" + config_name,
                    url_template="https://{source}/search?page={{page}}",
                    url_template_params=[
                        {"values": {"source": "example.com"}, "comment": "main"},
                    ],
                    max_pages=10,
                    cron_schedule="0/15 * * * *",
                    update_interval_sec=86400,
                ),
            )
            await session.commit()

        result = await pg_config_repo.get_surf_config(config_name)

        assert isinstance(result, surfing.Params)
        assert result.id == 12345
        assert result.name == config_name
        assert result.source_id == "source-for-" + config_name
        assert result.url_template == "https://{source}/search?page={{page}}"
        assert len(result.url_template_params) == 1
        assert result.url_template_params[0].values == {"source": "example.com"}
        assert result.max_pages == 10
        assert result.update_interval_sec == 86400

    async def test_when_row_missing_then_raises_not_found(
        self,
        pg_config_repo: PGConfigRepository,
    ) -> None:
        with pytest.raises(errors.NotFoundError):
            await pg_config_repo.get_surf_config("nonexistent-config-" + uuid4().hex[:8])


class TestGetSurfSchedules:
    async def test_when_rows_exist_then_returns_dict(
        self,
        sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
        pg_config_repo: PGConfigRepository,
    ) -> None:
        cfg_name_1 = f"sched1-{uuid4().hex[:12]}"
        cfg_name_2 = f"sched2-{uuid4().hex[:12]}"
        cfg_name_3 = f"sched3-{uuid4().hex[:12]}"

        async with sessionmaker() as session:
            for name, schedule in (
                (cfg_name_1, "0 * * * *"),
                (cfg_name_2, "0/15 * * * *"),
                (cfg_name_3, "0/30 * * * *"),
            ):
                await session.execute(
                    sa.insert(surfer_orm.SurfConfigORM).values(
                        id=int(uuid4().hex[:8], 16) % 100000,
                        name=name,
                        source_id="src-" + name,
                        url_template="https://" + name + ".com/{{page}}",
                        url_template_params=[{"values": {}, "comment": ""}],
                        max_pages=5,
                        cron_schedule=schedule,
                        update_interval_sec=86400,
                    ),
                )
            await session.commit()

        result = await pg_config_repo.get_surf_schedules()

        assert result[cfg_name_1] == "0 * * * *"
        assert result[cfg_name_2] == "0/15 * * * *"
        assert result[cfg_name_3] == "0/30 * * * *"

    async def test_when_non_default_update_interval_sec_then_round_trips(
        self,
        sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
        pg_config_repo: PGConfigRepository,
    ) -> None:
        config_name = f"cfg-nd-{uuid4().hex[:12]}"

        async with sessionmaker() as session:
            await session.execute(
                sa.insert(surfer_orm.SurfConfigORM).values(
                    id=int(uuid4().hex[:8], 16) % 100000,
                    name=config_name,
                    source_id="src-" + config_name,
                    url_template="https://" + config_name + ".com/{{page}}",
                    url_template_params=[{"values": {}, "comment": ""}],
                    max_pages=3,
                    cron_schedule="0 * * * *",
                    update_interval_sec=300,
                ),
            )
            await session.commit()

        result = await pg_config_repo.get_surf_config(config_name)

        assert isinstance(result, surfing.Params)
        assert result.update_interval_sec == 300

    async def test_when_no_rows_then_returns_empty_dict(
        self,
        sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession],
        pg_config_repo: PGConfigRepository,
    ) -> None:
        async with sessionmaker() as session:
            await session.execute(sa.delete(surfer_orm.SurfConfigORM))
            await session.commit()

        result = await pg_config_repo.get_surf_schedules()

        assert result == {}
