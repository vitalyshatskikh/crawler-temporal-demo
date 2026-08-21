
import datetime as dt

import factory
import pytest
import temporalio.client
from faker import Faker

from downloader.domain import downloading
from surfer.domain import adverts
from surfer.domain.adverts import models as adverts_models

_fake = Faker()


class DownloadParamsFactory(factory.Factory):
    class Meta:
        model = downloading.Params

    id_ = factory.Sequence(lambda n: n + 1)
    name = "default"
    source_id = factory.LazyAttribute(lambda _: adverts.SourceID(_fake.slug()))
    headers = factory.LazyAttribute(lambda _: {"User-Agent": "test-agent"})
    id = factory.LazyAttribute(lambda o: o.id_)

    class Params:
        exclude = ("id_",)


class DownloaderDocMetaFactory(factory.Factory):
    class Meta:
        model = adverts_models.DocumentMeta

    _TS = dt.datetime(2024, 1, 1, tzinfo=dt.UTC)
    sdoc_id = factory.LazyAttribute(lambda _: adverts_models.SdocID(_fake.uuid4()))
    created_at = _TS
    updated_at = _TS
    source_id = adverts.SourceID("test")
    type = adverts_models.DocumentType.SURFED_ADVERT
    external_url = factory.LazyAttribute(
        lambda o: f"https://example.com/advert/{o.sdoc_id}"
    )


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
