

import datetime as dt
import hashlib
import typing as tp

import pytest
import temporalio.client
import temporalio.testing
import temporalio.worker

from surfer.application import consts
from surfer.application.workflows.process_search_page import (
    ProcessSearchPage,
)
from surfer.domain.adverts import models as adverts_models
from surfer.tests._factories import (
    DocMetaFactory,
    ProcessSearchPageInFactory,
    assert_workflow_failure_message,
)
from surfer.tests._interceptors import (
    WorkflowMockInterceptor,
    clear_mocks,
    get_mock_calls,
    set_mock,
)
from surfer.tests.conftest import WorkerSpec


async def test_run__when_parse_returns_empty__then_no_advert_children(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    set_mock(consts.WorkflowName.DOWNLOAD_SEARCH_PAGE, result={})
    set_mock(consts.ActivityName.PARSE_SEARCH_PAGE, result=[])
    set_mock(consts.WorkflowName.PROCESS_ADVERT, result=None)
    in_ = ProcessSearchPageInFactory()

    await make_workers([
        WorkerSpec(
            workflows=[ProcessSearchPage],
            activities=[],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    await env.client.execute_workflow(  # type: ignore[misc]
        ProcessSearchPage.run,
        in_,
        id='test-psp-empty',
        task_queue=consts.QueueName.SURFING_TASK,
    )

    assert len(get_mock_calls(consts.WorkflowName.PROCESS_ADVERT)) == 0

    download_call = get_mock_calls(consts.WorkflowName.DOWNLOAD_SEARCH_PAGE)[0]
    assert download_call.args[0].meta.model_dump(
        exclude={"created_at", "updated_at"},
    ) == {
        "sdoc_id": hashlib.md5(in_.page_url.encode()).hexdigest(),
        "source_id": in_.surf_params.source_id,  # type: ignore[attr-defined]
        "type": adverts_models.DocumentType.SEARCH_PAGE.value,
        "external_url": in_.page_url,
        "content_url": "",
        "update_interval_sec": in_.surf_params.update_interval_sec,  # type: ignore[attr-defined]
    }


async def test_run__when_parse_returns_sdocs__then_starts_advert_children(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    doc_meta_1 = DocMetaFactory(sdoc_id='1')
    doc_meta_2 = DocMetaFactory(sdoc_id='2')
    set_mock(consts.WorkflowName.DOWNLOAD_SEARCH_PAGE, result={})
    set_mock(consts.ActivityName.PARSE_SEARCH_PAGE, result=[doc_meta_1, doc_meta_2])
    set_mock(consts.WorkflowName.PROCESS_ADVERT, result=None)
    in_ = ProcessSearchPageInFactory()

    await make_workers([
        WorkerSpec(
            workflows=[ProcessSearchPage],
            activities=[],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    await env.client.execute_workflow(  # type: ignore[misc]
        ProcessSearchPage.run,
        in_,
        id='test-psp-two-sdocs',
        task_queue=consts.QueueName.SURFING_TASK,
    )

    advert_calls = get_mock_calls(consts.WorkflowName.PROCESS_ADVERT)
    assert len(advert_calls) == 2
    assert str(advert_calls[0].args[0].doc_meta.sdoc_id) == '1'
    assert str(advert_calls[1].args[0].doc_meta.sdoc_id) == '2'

    download_call = get_mock_calls(consts.WorkflowName.DOWNLOAD_SEARCH_PAGE)[0]
    assert download_call.args[0].meta.model_dump(
        exclude={"created_at", "updated_at"},
    ) == {
        "sdoc_id": hashlib.md5(in_.page_url.encode()).hexdigest(),
        "source_id": in_.surf_params.source_id,  # type: ignore[attr-defined]
        "type": adverts_models.DocumentType.SEARCH_PAGE.value,
        "external_url": in_.page_url,
        "content_url": "",
        "update_interval_sec": in_.surf_params.update_interval_sec,  # type: ignore[attr-defined]
    }


async def test_run__when_download_fails__then_propagates_error(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    set_mock(
        consts.WorkflowName.DOWNLOAD_SEARCH_PAGE,
        error=RuntimeError('download boom'),
    )
    in_ = ProcessSearchPageInFactory()

    await make_workers([
        WorkerSpec(
            workflows=[ProcessSearchPage],
            activities=[],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    with pytest.raises(temporalio.client.WorkflowFailureError) as excinfo:
        await env.client.execute_workflow(  # type: ignore[misc]
            ProcessSearchPage.run,
            in_,
            id='test-psp-download-fail',
            task_queue=consts.QueueName.SURFING_TASK,
        )

    assert_workflow_failure_message(excinfo, 'download boom')


async def test_run__when_parse_fails__then_propagates_error(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    set_mock(consts.WorkflowName.DOWNLOAD_SEARCH_PAGE, result={})
    set_mock(
        consts.ActivityName.PARSE_SEARCH_PAGE,
        error=RuntimeError('parse boom'),
    )
    in_ = ProcessSearchPageInFactory()

    await make_workers([
        WorkerSpec(
            workflows=[ProcessSearchPage],
            activities=[],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    with pytest.raises(temporalio.client.WorkflowFailureError) as excinfo:
        await env.client.execute_workflow(  # type: ignore[misc]
            ProcessSearchPage.run,
            in_,
            id='test-psp-parse-fail',
            task_queue=consts.QueueName.SURFING_TASK,
        )

    assert_workflow_failure_message(excinfo, 'parse boom')


async def test_run__when_doc_meta_recently_updated__then_skips_advert_children(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    recent_ts = dt.datetime.now(tz=dt.UTC)
    recent_1 = DocMetaFactory(sdoc_id='1', created_at=recent_ts, updated_at=recent_ts)
    recent_2 = DocMetaFactory(sdoc_id='2', created_at=recent_ts, updated_at=recent_ts)
    set_mock(consts.WorkflowName.DOWNLOAD_SEARCH_PAGE, result={})
    set_mock(consts.ActivityName.PARSE_SEARCH_PAGE, result=[recent_1, recent_2])
    set_mock(consts.WorkflowName.PROCESS_ADVERT, result=None)
    in_ = ProcessSearchPageInFactory()

    await make_workers([
        WorkerSpec(
            workflows=[ProcessSearchPage],
            activities=[],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    await env.client.execute_workflow(  # type: ignore[misc]
        ProcessSearchPage.run,
        in_,
        id='test-psp-recent-skip',
        task_queue=consts.QueueName.SURFING_TASK,
    )

    assert len(get_mock_calls(consts.WorkflowName.PROCESS_ADVERT)) == 0
