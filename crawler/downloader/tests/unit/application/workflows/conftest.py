import typing as tp

import pytest

from downloader.application import activities
from downloader.domain import downloading
from downloader.tests._factories import DownloadParamsFactory


@pytest.fixture
async def downloading_repo() -> activities.DownloadingRepo:
    repo = activities.DownloadingRepo(
        downloading.DummyConfigRepository(result=DownloadParamsFactory()),  # type: ignore[arg-type]
    )
    return repo

@pytest.fixture
async def web_downloader() -> tp.AsyncGenerator[activities.WebDownloader]:
    downloader = activities.WebDownloader(
        downloading.DummyDocumentRepository(),
    )
    try:
        yield downloader
    finally:
        await downloader.aclose()
