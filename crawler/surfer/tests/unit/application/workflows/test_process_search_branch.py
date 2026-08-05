

import typing as tp

import pydantic
import pytest
import temporalio.client
import temporalio.testing
import temporalio.worker

from surfer.application import consts
from surfer.application.workflows.process_search_branch import (
    ProcessSearchBranch,
)
from surfer.tests._factories import (
    ProcessSearchBranchInFactory,
    SurfParamsFactory,
    TemplateContextFactory,
    assert_workflow_failure_message,
)
from surfer.tests._interceptors import (
    WorkflowMockInterceptor,
    clear_mocks,
    get_mock_calls,
    set_mock,
)
from surfer.tests.conftest import WorkerSpec


async def test_run__when_max_pages_2__then_starts_two_page_children(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    surf_params = SurfParamsFactory(
        url_template_params=[TemplateContextFactory.build(values={'category': 'x'})],
        max_pages=2,
    )
    set_mock(consts.WorkflowName.PROCESS_SEARCH_PAGE, result=None)
    in_ = ProcessSearchBranchInFactory(branch_idx=0, surf_params=surf_params)

    await make_workers([
        WorkerSpec(
            workflows=[ProcessSearchBranch],
            activities=[],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    await env.client.execute_workflow(  # type: ignore[misc]
        ProcessSearchBranch.run,
        in_,
        id='test-psb-2-pages',
        task_queue=consts.QueueName.SURFING_TASK,
    )

    calls = get_mock_calls(consts.WorkflowName.PROCESS_SEARCH_PAGE)
    assert len(calls) == 2
    assert [c.args[0].page_num for c in calls] == [1, 2]


async def test_run__when_some_pages_fail__then_raises(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    surf_params = SurfParamsFactory(
        url_template_params=[TemplateContextFactory.build(values={'category': 'x'})],
        max_pages=3,
    )
    set_mock(
        consts.WorkflowName.PROCESS_SEARCH_PAGE,
        error=RuntimeError('page boom'),
    )
    in_ = ProcessSearchBranchInFactory(branch_idx=0, surf_params=surf_params)

    await make_workers([
        WorkerSpec(
            workflows=[ProcessSearchBranch],
            activities=[],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    with pytest.raises(temporalio.client.WorkflowFailureError) as excinfo:
        await env.client.execute_workflow(  # type: ignore[misc]
            ProcessSearchBranch.run,
            in_,
            id='test-psb-fail',
            task_queue=consts.QueueName.SURFING_TASK,
        )

    assert_workflow_failure_message(excinfo, 'page boom')


def test_run__when_invalid_branch_idx__then_validation_error() -> None:
    surf_params = SurfParamsFactory(
        url_template_params=[TemplateContextFactory.build(values={'category': 'x'})],
    )
    with pytest.raises(pydantic.ValidationError) as excinfo:
        ProcessSearchBranchInFactory(branch_idx=5, surf_params=surf_params)
    assert 'invalid branch index: 5' in str(excinfo.value)
