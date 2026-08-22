import asyncio
import contextlib
import logging
import sys

import aiohttp
import temporalio.client
from temporalio.contrib.pydantic import pydantic_data_converter

from downloader.application import config as downloader_config
from downloader.infrastructure import repositories, workers
from shared.py import observability, settings

logger = logging.getLogger(__name__)


class AppConfig(settings.Config):
    downloader: downloader_config.DownloaderConfig = downloader_config.DownloaderConfig()


async def main() -> None:
    app_config = AppConfig()
    observability.setup_logger(app_config)

    logger.info("Starting downloader application...")

    client = await temporalio.client.Client.connect(
        app_config.temporal_host,
        namespace=app_config.temporal_namespace,
        data_converter=pydantic_data_converter,
    )

    http_timeout = aiohttp.ClientTimeout(
        total=app_config.downloader.http_total_timeout.total_seconds(),
        connect=app_config.downloader.http_connect_timeout.total_seconds(),
    )
    connector = aiohttp.TCPConnector(limit=app_config.downloader.http_connector_limit)
    http_client = aiohttp.ClientSession(
        timeout=http_timeout,
        connector=connector,
        connector_owner=True,
    )

    downloading_repo = repositories.PGDownloadingRepository()
    document_repo = repositories.PGDocumentRepository()

    w = workers.DownloadingWorker(
        client=client,
        downloading_repo=downloading_repo,
        document_repo=document_repo,
        http_client=http_client,
    )

    async with contextlib.aclosing(w):
        await w.run()


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        sys.exit(1)
