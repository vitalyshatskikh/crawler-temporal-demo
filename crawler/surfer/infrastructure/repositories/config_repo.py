import sqlalchemy as sa
import sqlalchemy.ext.asyncio as sa_asyncio

from surfer.domain import errors, surfing
from surfer.infrastructure.db import mappers, orm


class PGConfigRepository(surfing.ISurfingRepository):
    def __init__(self, sessionmaker: sa_asyncio.async_sessionmaker[sa_asyncio.AsyncSession]) -> None:
        self._sessionmaker = sessionmaker

    async def get_surf_config(self, name: str) -> surfing.Params:
        async with self._sessionmaker() as session:
            stmt = sa.select(orm.SurfConfigORM).where(orm.SurfConfigORM.name == name)
            row = (await session.execute(stmt)).scalar_one_or_none()
            if row is None:
                raise errors.NotFoundError(f"surf config not found: name={name}", name)
            return mappers.surf_config_to_params(row)

    async def get_surf_schedules(self) -> dict[str, str]:
        async with self._sessionmaker() as session:
            stmt = sa.select(orm.SurfConfigORM.name, orm.SurfConfigORM.cron_schedule)
            rows = (await session.execute(stmt)).all()
            return {name: cron for name, cron in rows}
