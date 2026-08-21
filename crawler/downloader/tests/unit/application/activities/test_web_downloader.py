import contextlib
import datetime as dt

import pytest
from aioresponses import aioresponses

from downloader.application.activities import WebDownloader
from downloader.domain import downloading, errors
from downloader.tests._factories import DownloaderDocMetaFactory, DownloadParamsFactory
from surfer.domain import adverts
from surfer.domain.adverts import models as adverts_models


async def test_download_to_repo__when_response_2xx__then_saves_document_with_body() -> None:
    doc_repo = downloading.DummyDocumentRepository()
    activity = WebDownloader(doc_repo)
    doc_meta = DownloaderDocMetaFactory(external_url="https://example.com/page1")
    params = DownloadParamsFactory(headers={"User-Agent": "test-agent"})

    with aioresponses() as m:
        m.get("https://example.com/page1", status=200, body="<html>hello world</html>")
        async with contextlib.aclosing(activity):
            await activity.download_to_repo(params, doc_meta)  # type: ignore[arg-type]

    assert len(doc_repo.saved) == 1
    saved_doc = doc_repo.saved[0]
    assert saved_doc == adverts.Document(**doc_meta.model_dump(), body="<html>hello world</html>")  # type: ignore[attr-defined]


async def test_download_to_repo__when_response_5xx__then_propagates_downloader_error() -> None:
    doc_repo = downloading.DummyDocumentRepository()
    activity = WebDownloader(doc_repo)
    doc_meta = DownloaderDocMetaFactory(external_url="https://example.com/notfound")
    params = DownloadParamsFactory()

    with (
        aioresponses() as m,
        pytest.raises(errors.DownloaderError),
    ):
        m.get("https://example.com/notfound", status=500, body="server error")
        async with contextlib.aclosing(activity):
            await activity.download_to_repo(params, doc_meta)  # type: ignore[arg-type]

    assert len(doc_repo.saved) == 0


async def test_download_to_repo__when_save_raises__then_propagates() -> None:
    doc_repo = downloading.DummyDocumentRepository(error=RuntimeError("save boom"))
    activity = WebDownloader(doc_repo)
    doc_meta = DownloaderDocMetaFactory(external_url="https://example.com/page2")
    params = DownloadParamsFactory()

    with (
        aioresponses() as m,
        pytest.raises(RuntimeError, match="save boom"),
    ):
        m.get("https://example.com/page2", status=200, body="<html>content</html>")
        async with contextlib.aclosing(activity):
            await activity.download_to_repo(params, doc_meta)  # type: ignore[arg-type]


@pytest.mark.parametrize("status,expected_error", [
    (400, errors.ValidationError),
    (401, errors.ValidationError),
    (403, errors.ValidationError),
    (429, errors.ValidationError),
])
async def test_download_to_repo__when_response_4xx_excluding_404__then_raises_validation_error(
    status: int,
    expected_error: type[Exception],
) -> None:
    doc_repo = downloading.DummyDocumentRepository()
    activity = WebDownloader(doc_repo)
    doc_meta = DownloaderDocMetaFactory(external_url="https://example.com/page")
    params = DownloadParamsFactory()

    with (
        aioresponses() as m,
        pytest.raises(expected_error),
    ):
        m.get("https://example.com/page", status=status, body="error page")
        async with contextlib.aclosing(activity):
            await activity.download_to_repo(params, doc_meta)  # type: ignore[arg-type]

    assert len(doc_repo.saved) == 0


async def test_download_to_repo__when_response_404__then_saves_document_with_body() -> None:
    doc_repo = downloading.DummyDocumentRepository()
    activity = WebDownloader(doc_repo)
    doc_meta = DownloaderDocMetaFactory(external_url="https://example.com/missing")
    params = DownloadParamsFactory()

    with aioresponses() as m:
        m.get("https://example.com/missing", status=404, body="Not Found")
        async with contextlib.aclosing(activity):
            await activity.download_to_repo(params, doc_meta)  # type: ignore[arg-type]

    assert len(doc_repo.saved) == 1
    saved_doc = doc_repo.saved[0]
    assert saved_doc == adverts.Document(**doc_meta.model_dump(), body="Not Found")  # type: ignore[attr-defined]


async def test_download_to_repo__when_response_2xx__then_all_doc_meta_fields_preserved() -> None:
    doc_repo = downloading.DummyDocumentRepository()
    activity = WebDownloader(doc_repo)

    doc_meta = adverts_models.DocumentMeta(
        sdoc_id=adverts_models.SdocID("sdoc-x"),
        created_at=dt.datetime(2024, 6, 1, tzinfo=dt.UTC),
        updated_at=dt.datetime(2024, 6, 2, tzinfo=dt.UTC),
        source_id=adverts.SourceID("source-y"),
        type=adverts_models.DocumentType.DOWNLOADED_ADVERT,
        external_url="https://example.com/advert/99",
    )
    params = DownloadParamsFactory()

    with aioresponses() as m:
        m.get("https://example.com/advert/99", status=200, body="<html>advert body</html>")
        async with contextlib.aclosing(activity):
            await activity.download_to_repo(params, doc_meta)  # type: ignore[arg-type]

    assert len(doc_repo.saved) == 1
    saved = doc_repo.saved[0]
    assert saved == adverts.Document(**doc_meta.model_dump(), body="<html>advert body</html>")
