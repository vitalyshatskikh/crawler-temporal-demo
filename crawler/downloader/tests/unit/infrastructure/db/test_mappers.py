
from downloader.domain import downloading
from downloader.infrastructure.db import mappers
from downloader.infrastructure.db.orm import download_configs as orm
from surfer.domain import adverts


class TestDownloadConfigToParams:
    def test_download_config_to_params__when_valid_row__then_returns_params(self) -> None:
        row = orm.DownloadConfigORM(
            id=1,
            source_id="test.com",
            doc_type="search_page",
            name="test.com search page",
            headers={"User-Agent": "curl/7.0"},
        )
        result = mappers.download_config_to_params(row)
        assert isinstance(result, downloading.Params)
        assert result.id == 1
        assert result.source_id == adverts.SourceID("test.com")
        assert result.headers == {"User-Agent": "curl/7.0"}
