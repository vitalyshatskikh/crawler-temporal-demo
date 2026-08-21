import typing as tp

import pytest
import temporalio.client
import temporalio.testing
import temporalio.worker

from downloader.application import activities, consts, workflows
from downloader.tests._factories import (
    assert_workflow_failure_message,
)
from downloader.tests._interceptors import (
    WorkflowMockInterceptor,
    clear_mocks,
    get_mock_calls,
    set_mock,
)
from downloader.tests.conftest import WorkerSpec
from surfer.application import consts as surfer_consts
from surfer.tests._factories import DownloadInFactory


async def test_run__when_all_activities_succeed__then_completes(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
    downloading_repo: activities.DownloadingRepo,
    web_downloader: activities.WebDownloader,
) -> None:
    clear_mocks()
    set_mock(consts.ActivityName.GET_DOWNLOADING_CONFIG, result=None)
    set_mock(consts.ActivityName.DOWNLOAD_TO_REPO, result=None)
    in_ = DownloadInFactory()

    await make_workers([
        WorkerSpec(
            workflows=[workflows.DownloadSearchPage],
            activities=[
                downloading_repo.get_downloading_config,
                web_downloader.download_to_repo,
            ],
            task_queue=surfer_consts.QueueName.DOWNLOADING,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    result = await env.client.execute_workflow(  # type: ignore[misc]
        workflows.DownloadSearchPage.run,
        in_,
        id="test-dsp-success",
        task_queue=surfer_consts.QueueName.DOWNLOADING,
    )
    assert result is None

    assert len(get_mock_calls(consts.ActivityName.GET_DOWNLOADING_CONFIG)) == 1
    assert len(get_mock_calls(consts.ActivityName.DOWNLOAD_TO_REPO)) == 1


async def test_run__when_get_config_fails__then_propagates_error(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
    downloading_repo: activities.DownloadingRepo,
    web_downloader: activities.WebDownloader,
) -> None:
    clear_mocks()
    set_mock(
        consts.ActivityName.GET_DOWNLOADING_CONFIG,
        error=RuntimeError("config boom"),
    )
    in_ = DownloadInFactory()

    await make_workers([
        WorkerSpec(
            workflows=[workflows.DownloadSearchPage],
            activities=[
                downloading_repo.get_downloading_config,
                web_downloader.download_to_repo,
            ],
            task_queue=surfer_consts.QueueName.DOWNLOADING,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    with pytest.raises(temporalio.client.WorkflowFailureError) as excinfo:
        await env.client.execute_workflow(  # type: ignore[misc]
            workflows.DownloadSearchPage.run,
            in_,
            id="test-dsp-config-fail",
            task_queue=surfer_consts.QueueName.DOWNLOADING,
        )

    assert_workflow_failure_message(excinfo, "config boom")


async def test_run__when_download_to_repo_fails__then_propagates_error(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
    downloading_repo: activities.DownloadingRepo,
    web_downloader: activities.WebDownloader,
) -> None:
    clear_mocks()
    set_mock(consts.ActivityName.GET_DOWNLOADING_CONFIG, result=None)
    set_mock(
        consts.ActivityName.DOWNLOAD_TO_REPO,
        error=RuntimeError("save boom"),
    )
    in_ = DownloadInFactory()

    await make_workers([
        WorkerSpec(
            workflows=[workflows.DownloadSearchPage],
            activities=[
                downloading_repo.get_downloading_config,
                web_downloader.download_to_repo,
            ],
            task_queue=surfer_consts.QueueName.DOWNLOADING,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    with pytest.raises(temporalio.client.WorkflowFailureError) as excinfo:
        await env.client.execute_workflow(  # type: ignore[misc]
            workflows.DownloadSearchPage.run,
            in_,
            id="test-dsp-save-fail",
            task_queue=surfer_consts.QueueName.DOWNLOADING,
        )

    assert_workflow_failure_message(excinfo, "save boom")
