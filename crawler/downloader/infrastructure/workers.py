import logging

import aiohttp
import temporalio.client
import temporalio.worker

from downloader.application import activities, workflows
from downloader.domain import downloading
from surfer.application import consts as surfer_consts

logger = logging.getLogger(__name__)


class DownloadingWorker:
    def __init__(
        self,
        client: temporalio.client.Client,
        downloading_repo: downloading.IDownloadingRepository,
        document_repo: downloading.IDocumentRepository,
        http_client: aiohttp.ClientSession,
    ) -> None:
        self._download_repo = activities.DownloadingRepo(downloading_repo)
        self._web_download = activities.WebDownloader(document_repo, http_client)
        self._w = temporalio.worker.Worker(
            client=client,
            task_queue=surfer_consts.QueueName.DOWNLOADING,
            workflows=[
                workflows.DownloadSearchPage,
                workflows.DownloadAdvertContent,
            ],
            activities=[
                self._download_repo.get_downloading_config,
                self._web_download.download_to_repo,
            ],
        )

    async def run(self) -> None:
        await self._w.run()

    async def aclose(self) -> None:
        await self._web_download.aclose()
