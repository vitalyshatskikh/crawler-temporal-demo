

import typing as tp

import pytest
import temporalio.client
import temporalio.testing
import temporalio.worker

from surfer.application import consts
from surfer.application.workflows import models as wf_models
from surfer.application.workflows.process_advert import ProcessAdvert
from surfer.tests._factories import (
    ProcessAdvertInFactory,
    assert_workflow_failure_message,
)
from surfer.tests._interceptors import (
    WorkflowMockInterceptor,
    clear_mocks,
    get_mock_calls,
    set_mock,
)
from surfer.tests.conftest import WorkerSpec


async def test_run__when_download_and_parse_succeed__then_completes(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    set_mock(consts.WorkflowName.DOWNLOAD_ADVERT_CONTENT, result=None)
    set_mock(consts.ActivityName.PARSE_ADVERT_CONTENT, result=None)
    in_ = ProcessAdvertInFactory()

    await make_workers([
        WorkerSpec(
            workflows=[ProcessAdvert],
            activities=[],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    result = await env.client.execute_workflow(  # type: ignore[misc]
        ProcessAdvert.run,
        in_,
        id='test-pa-success',
        task_queue=consts.QueueName.SURFING_TASK,
    )
    assert result is None

    assert len(get_mock_calls(consts.WorkflowName.DOWNLOAD_ADVERT_CONTENT)) == 1
    assert len(get_mock_calls(consts.ActivityName.PARSE_ADVERT_CONTENT)) == 1

    download_call = get_mock_calls(consts.WorkflowName.DOWNLOAD_ADVERT_CONTENT)[0]
    assert download_call.args[0] == wf_models.DownloadIn(
        meta=in_.doc_meta,
        download_timeout=in_.surfer_config.download_advert_content_timeout,  # type: ignore[attr-defined]
        config_request_timeout=in_.surfer_config.repo_request_timeout,  # type: ignore[attr-defined]
    )


async def test_run__when_download_fails__then_propagates_error(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    set_mock(
        consts.WorkflowName.DOWNLOAD_ADVERT_CONTENT,
        error=RuntimeError('download boom'),
    )
    in_ = ProcessAdvertInFactory()

    await make_workers([
        WorkerSpec(
            workflows=[ProcessAdvert],
            activities=[],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    with pytest.raises(temporalio.client.WorkflowFailureError) as excinfo:
        await env.client.execute_workflow(  # type: ignore[misc]
            ProcessAdvert.run,
            in_,
            id='test-pa-download-fail',
            task_queue=consts.QueueName.SURFING_TASK,
        )

    assert_workflow_failure_message(excinfo, 'download boom')


async def test_run__when_parse_fails__then_propagates_error(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    set_mock(consts.WorkflowName.DOWNLOAD_ADVERT_CONTENT, result=None)
    set_mock(
        consts.ActivityName.PARSE_ADVERT_CONTENT,
        error=RuntimeError('parse boom'),
    )
    in_ = ProcessAdvertInFactory()

    await make_workers([
        WorkerSpec(
            workflows=[ProcessAdvert],
            activities=[],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    with pytest.raises(temporalio.client.WorkflowFailureError) as excinfo:
        await env.client.execute_workflow(  # type: ignore[misc]
            ProcessAdvert.run,
            in_,
            id='test-pa-parse-fail',
            task_queue=consts.QueueName.SURFING_TASK,
        )

    assert_workflow_failure_message(excinfo, 'parse boom')
