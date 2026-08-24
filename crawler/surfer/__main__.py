import asyncio
import logging

import temporalio.client
import temporalio.worker
from temporalio.contrib.pydantic import pydantic_data_converter

from shared.py import observability, settings
from shared.py.db import engine as db_engine
from surfer.application import config
from surfer.infrastructure import repositories, schedules, workers

logger = logging.getLogger(__name__)


class AppConfig(settings.Config):
    surfer: config.SurferConfig = config.SurferConfig()


async def main() -> None:
    app_config = AppConfig()
    observability.setup_logger(app_config)

    logger.info("Starting surfer application...")

    client = await temporalio.client.Client.connect(
        app_config.temporal_host,
        namespace=app_config.temporal_namespace,
        data_converter=pydantic_data_converter,
    )

    engine = db_engine.make_engine(app_config.postgres)
    sessionmaker = db_engine.make_sessionmaker(engine)

    surf_config_repo = repositories.PGConfigRepository(sessionmaker)
    adverts_repo = repositories.PGAdvertsRepo(sessionmaker)

    schedules_conf = await surf_config_repo.get_surf_schedules()
    await schedules.setup_surfing(client, app_config.surfer, schedules_conf)

    to_run: list[workers.Worker] = [
        workers.SurfingWorker(
            client=client,
            surf_config_repo=surf_config_repo,
            adverts_repo=adverts_repo,
        ),
        workers.AdvertsWorker(client=client),
        workers.ParsingWorker(client=client),
    ]

    async with asyncio.TaskGroup() as tg:
        for w in to_run:
            tg.create_task(w.run())


if __name__ == "__main__":
    import sys
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        sys.exit(1)
