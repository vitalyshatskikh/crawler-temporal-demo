import http
import logging

import aiohttp
from temporalio import activity

from downloader.application import consts
from downloader.domain import downloading, errors
from surfer.domain import adverts

logger = logging.getLogger(__name__)


class WebDownloader:
    def __init__(
        self,
        doc_repo: downloading.IDocumentRepository,
        http_client: aiohttp.ClientSession | None = None,
    ) -> None:
        self._doc_repo = doc_repo
        self._http_client = http_client or aiohttp.ClientSession()

    @activity.defn(name=consts.ActivityName.DOWNLOAD_TO_REPO)
    async def download_to_repo(self, conf: downloading.Params, doc_meta: adverts.DocumentMeta) -> None:
        actual_url = doc_meta.content_url or doc_meta.external_url
        async with self._http_client.get(
            url=actual_url,
            headers=conf.headers,
            raise_for_status=_raise_for_status,
        ) as response:
            body = await response.text()

        if response.status == http.HTTPStatus.NOT_FOUND:
            logger.info(
                "page not found: url=%s sdoc_id=%s",
                actual_url,
                doc_meta.sdoc_id,
            )

        doc = adverts.Document(
            **doc_meta.model_dump(),
            body=body,
        )
        await self._doc_repo.save(doc)

    async def aclose(self) -> None:
        await self._http_client.close()


async def _raise_for_status(response: aiohttp.ClientResponse) -> None:
    if response.status >= http.HTTPStatus.INTERNAL_SERVER_ERROR:
        raise errors.DownloaderError("server error", response.status, response.reason, str(response.url))
    if response.status >= http.HTTPStatus.BAD_REQUEST and response.status != http.HTTPStatus.NOT_FOUND:
        raise errors.ValidationError("client error", response.status, response.reason, str(response.url))
