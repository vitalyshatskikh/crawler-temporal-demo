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
            update_interval_sec=86400,
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
            update_interval_sec=86400,
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
            content_url="",
            body="<html>test</html>",
            created_at=now,
            updated_at=now,
            update_interval_sec=86400,
        )
        result = shared_mappers.document_to_meta(row)
        assert isinstance(result, adverts.DocumentMeta)
        assert result.sdoc_id == adverts.SdocID("abc123")
        assert result.source_id == adverts.SourceID("test.com")
        assert result.type == adverts.DocumentType.SURFED_ADVERT
        assert result.content_url == ""

    def test_document_to_meta__when_content_url_set__then_returns_content_url(self) -> None:
        now = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)
        row = shared_orm.DocumentORM(
            sdoc_id="abc123",
            source_id="test.com",
            doc_type="surfed_advert",
            external_url="https://test.com/abc123",
            content_url="https://cdn.example.com/abc123",
            body="<html>test</html>",
            created_at=now,
            updated_at=now,
            update_interval_sec=86400,
        )
        result = shared_mappers.document_to_meta(row)
        assert isinstance(result, adverts.DocumentMeta)
        assert result.content_url == "https://cdn.example.com/abc123"


class TestDocumentToOrm:
    def test_document_to_orm__when_valid_document__then_returns_orm(self) -> None:
        now = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)
        doc = adverts.Document(
            sdoc_id=adverts.SdocID("xyz789"),
            source_id=adverts.SourceID("test.com"),
            type=adverts.DocumentType.DOWNLOADED_ADVERT,
            external_url="https://test.com/xyz789",
            content_url="",
            body="<html>downloaded</html>",
            created_at=now,
            updated_at=now,
            update_interval_sec=86400,
        )
        result = shared_mappers.document_to_orm(doc)
        assert result["sdoc_id"] == "xyz789"
        assert result["body"] == "<html>downloaded</html>"
        assert result["doc_type"] == "downloaded_advert"
        assert result["content_url"] == ""

    def test_document_to_orm__when_content_url_set__then_returns_content_url(self) -> None:
        now = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)
        doc = adverts.Document(
            sdoc_id=adverts.SdocID("xyz789"),
            source_id=adverts.SourceID("test.com"),
            type=adverts.DocumentType.DOWNLOADED_ADVERT,
            external_url="https://test.com/xyz789",
            content_url="https://cdn.example.com/xyz789",
            body="<html>downloaded</html>",
            created_at=now,
            updated_at=now,
            update_interval_sec=86400,
        )
        result = shared_mappers.document_to_orm(doc)
        assert result["content_url"] == "https://cdn.example.com/xyz789"

    def test_document_to_meta__roundtrip__then_preserves_content_url(self) -> None:
        now = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)
        original_row = shared_orm.DocumentORM(
            sdoc_id="roundtrip123",
            source_id="test.com",
            doc_type="surfed_advert",
            external_url="https://test.com/abc",
            content_url="https://cdn.example.com/abc",
            body="<html>test</html>",
            created_at=now,
            updated_at=now,
            update_interval_sec=86400,
        )
        meta = shared_mappers.document_to_meta(original_row)
        assert meta.content_url == "https://cdn.example.com/abc"
        orm_dict = shared_mappers.document_to_orm(adverts.Document(**meta.model_dump(), body="<html>test</html>"))
        assert orm_dict["content_url"] == "https://cdn.example.com/abc"
