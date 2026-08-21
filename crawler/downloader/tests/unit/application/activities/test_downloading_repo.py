

import pytest

from downloader.application.activities import DownloadingRepo
from downloader.domain import downloading
from downloader.tests._factories import DownloadParamsFactory
from surfer.domain import adverts
from surfer.domain.adverts import models as adverts_models


async def test_get_downloading_config__when_repo_returns_params__then_returns_them() -> None:
    params = DownloadParamsFactory(headers={"X-Test": "value"})
    repo = downloading.DummyConfigRepository(result=params)  # type: ignore[arg-type]
    activity = DownloadingRepo(repo)

    result = await activity.get_downloading_config(
        adverts.SourceID("test"),
        adverts_models.DocumentType.DOWNLOADED_ADVERT,
    )

    assert result == params



async def test_get_downloading_config__when_repo_raises__then_propagates() -> None:
    repo = downloading.DummyConfigRepository(error=RuntimeError("boom"))
    activity = DownloadingRepo(repo)

    with pytest.raises(RuntimeError, match="boom"):
        await activity.get_downloading_config(
            adverts.SourceID("test"),
            adverts_models.DocumentType.DOWNLOADED_ADVERT,
        )
