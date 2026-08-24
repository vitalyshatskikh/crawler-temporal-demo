import datetime as dt

from shared.py.db import mappers as shared_mappers
from shared.py.db import orm as shared_orm
from surfer.domain import adverts, surfing
from surfer.infrastructure.db import mappers, orm


class TestSurfConfigToParams:
    def test_surf_config_to_params__when_valid_row__then_returns_params(self) -> None:
        row = orm.SurfConfigORM(
            id=1,
            name="test.com",
            source_id="test.com",
            url_template="https://test.com/{{page}}",
            url_template_params=[{"values": {"page": "1"}, "comment": "home"}],
            max_pages=5,
            cron_schedule="0 * * * *",
        )
        result = mappers.surf_config_to_params(row)
        assert isinstance(result, surfing.Params)
        assert result.id == 1
        assert result.name == "test.com"
        assert result.source_id == adverts.SourceID("test.com")
        assert result.url_template == "https://test.com/{{page}}"
        assert len(result.url_template_params) == 1
        assert result.url_template_params[0].values == {"page": "1"}
        assert result.max_pages == 5


class TestParamsToSurfConfig:
    def test_params_to_surf_config__when_valid_params__then_returns_orm(self) -> None:
        params = surfing.Params(
            id=2,
            name="example.com",
            source_id=adverts.SourceID("example.com"),
            url_template="https://example.com/{{page}}",
            url_template_params=[surfing.TemplateContext(values={"page": "1"})],
            max_pages=10,
        )
        result = mappers.params_to_surf_config(params, cron_schedule="0/1 * * * *")
        assert isinstance(result, orm.SurfConfigORM)
        assert result.id == 2
        assert result.name == "example.com"
        assert result.cron_schedule == "0/1 * * * *"


class TestDocumentToMeta:
    def test_document_to_meta__when_valid_row__then_returns_document_meta(self) -> None:
        now = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)
        row = shared_orm.DocumentORM(
            sdoc_id="abc123",
            source_id="test.com",
            doc_type="surfed_advert",
            external_url="https://test.com/abc123",
            body="<html>test</html>",
            created_at=now,
            updated_at=now,
        )
        result = shared_mappers.document_to_meta(row)
        assert isinstance(result, adverts.DocumentMeta)
        assert result.sdoc_id == adverts.SdocID("abc123")
        assert result.source_id == adverts.SourceID("test.com")
        assert result.type == adverts.DocumentType.SURFED_ADVERT


class TestDocumentToOrm:
    def test_document_to_orm__when_valid_document__then_returns_orm(self) -> None:
        now = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)
        doc = adverts.Document(
            sdoc_id=adverts.SdocID("xyz789"),
            source_id=adverts.SourceID("test.com"),
            type=adverts.DocumentType.DOWNLOADED_ADVERT,
            external_url="https://test.com/xyz789",
            body="<html>downloaded</html>",
            created_at=now,
            updated_at=now,
        )
        result = shared_mappers.document_to_orm(doc)
        assert result["sdoc_id"] == "xyz789"
        assert result["body"] == "<html>downloaded</html>"
        assert result["doc_type"] == "downloaded_advert"
