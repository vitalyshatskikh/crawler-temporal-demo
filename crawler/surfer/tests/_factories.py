
import datetime as dt

import factory
import pytest
import temporalio.client
from faker import Faker

from surfer.application.config import SurferConfig
from surfer.application.workflows import models as wf_models
from surfer.application.workflows.process_advert import ProcessAdvertIn
from surfer.application.workflows.process_search_branch import ProcessSearchBranchIn
from surfer.application.workflows.process_search_page import ProcessSearchPageIn
from surfer.application.workflows.search_adverts import SearchAdvertsIn
from surfer.domain import adverts
from surfer.domain.adverts import models as adverts_models
from surfer.domain.surfing import models as surfing_models

_fake = Faker()


class TemplateContextFactory(factory.Factory):
    class Meta:
        model = surfing_models.TemplateContext

    values = factory.LazyAttribute(lambda _: {"category": _fake.word()})
    comment = factory.LazyAttribute(lambda _: _fake.sentence(nb_words=3))


class SurfParamsFactory(factory.Factory):
    class Meta:
        model = surfing_models.Params

    id_ = factory.Sequence(lambda n: n + 1)
    name = "demo"
    source_id = factory.LazyAttribute(lambda o: adverts.SourceID(o.name))
    url_template = "https://example.com/adverts/{{category}}?page={{page}}"
    url_template_params = factory.LazyAttribute(
        lambda _: [
            surfing_models.TemplateContext(values={"category": c})
            for c in ("x", "y", "z")
        ]
    )
    max_pages = 5
    update_interval_sec = 86400
    id = factory.LazyAttribute(lambda o: o.id_)

    class Params:
        exclude = ("id_",)


class DocMetaFactory(factory.Factory):
    class Meta:
        model = adverts_models.DocumentMeta

    _TS = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)
    sdoc_id = factory.LazyAttribute(
        lambda _: adverts_models.SdocID(_fake.uuid4())
    )
    created_at = _TS
    updated_at = _TS
    source_id = adverts.SourceID("test")
    type = adverts_models.DocumentType.SURFED_ADVERT
    external_url = factory.LazyAttribute(lambda o: f"https://example.com/{o.sdoc_id}")
    update_interval_sec = 86400


class SurferConfigFactory(factory.Factory):
    class Meta:
        model = SurferConfig

    process_branch_timeout = dt.timedelta(minutes=15)
    process_search_page_timeout = dt.timedelta(minutes=5)
    process_advert_timeout = dt.timedelta(minutes=5)
    download_search_page_timeout = dt.timedelta(minutes=4)
    download_advert_content_timeout = dt.timedelta(minutes=4)
    repo_request_timeout = dt.timedelta(seconds=15)
    parse_search_page_timeout = dt.timedelta(seconds=30)
    parse_advert_content_timeout = dt.timedelta(seconds=30)


class SearchAdvertsInFactory(factory.Factory):
    class Meta:
        model = SearchAdvertsIn

    surfer_config = factory.SubFactory(SurferConfigFactory)
    surf_config_name = "demo"


class ProcessSearchBranchInFactory(factory.Factory):
    class Meta:
        model = ProcessSearchBranchIn

    surfer_config = factory.SubFactory(SurferConfigFactory)
    surf_params = factory.SubFactory(SurfParamsFactory)
    branch_idx = 0


class ProcessSearchPageInFactory(factory.Factory):
    class Meta:
        model = ProcessSearchPageIn

    surfer_config = factory.SubFactory(SurferConfigFactory)
    surf_params = factory.SubFactory(SurfParamsFactory)
    branch_idx = 0
    page_num = 1
    page_url = "https://example.com/adverts/x?page=1"


class ProcessAdvertInFactory(factory.Factory):
    class Meta:
        model = ProcessAdvertIn

    surfer_config = factory.SubFactory(SurferConfigFactory)
    surf_params = factory.SubFactory(SurfParamsFactory)
    doc_meta = factory.SubFactory(DocMetaFactory)


def assert_workflow_failure_message(
    excinfo: pytest.ExceptionInfo[temporalio.client.WorkflowFailureError],
    needle: str,
    *,
    anywhere: bool = True,
) -> None:
    """Assert that a WorkflowFailureError's cause chain contains needle.

    Descends the .cause chain. For ExceptionGroup, recurses into
    sub-exceptions and succeeds if any leaf contains needle.
    """
    exc_val: BaseException | None = excinfo.value
    while isinstance(exc_val, temporalio.client.WorkflowFailureError):
        exc_val = exc_val.cause

    def search(exc: BaseException | None) -> bool:
        if exc is None:
            return False
        if isinstance(exc, ExceptionGroup):
            return any(search(e) for e in exc.exceptions)
        return needle in str(exc) if anywhere else str(exc).startswith(needle)

    assert search(exc_val), (
        f"expected {needle!r} not found in workflow failure "
        f"(root cause: {exc_val!r})"
    )


class DownloadInFactory(factory.Factory):
    class Meta:
        model = wf_models.DownloadIn

    meta = factory.SubFactory(DocMetaFactory)
