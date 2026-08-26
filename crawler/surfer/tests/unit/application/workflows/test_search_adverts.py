
import typing as tp

import pytest
import temporalio.client
import temporalio.exceptions
import temporalio.testing
import temporalio.worker

from surfer.application import activities, consts, workflows
from surfer.domain.surfing import repositories as surfing_repo
from surfer.tests._factories import (
    SearchAdvertsInFactory,
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


async def test_run__when_one_branch__then_starts_one_child_and_completes(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    surf_params = SurfParamsFactory(
        url_template_params=[TemplateContextFactory.build(values={'category': 'x'})],
    )
    set_mock(consts.WorkflowName.PROCESS_SEARCH_BRANCH, result=None)
    in_ = SearchAdvertsInFactory(surf_config_name='demo', surf_params=surf_params)

    await make_workers([
        WorkerSpec(
            workflows=[workflows.SearchAdverts],
            activities=[
                activities.SurfConfigRepo(
                    surfing_repo.DummyConfigRepository(result=surf_params),  # type: ignore[arg-type]
                ).get_surf_params,
            ],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    result = await env.client.execute_workflow(  # type: ignore[misc]
        workflows.SearchAdverts.run,
        in_,
        id='test-sa-one-branch',
        task_queue=consts.QueueName.SURFING_TASK,
    )
    assert result is None

    calls = get_mock_calls(consts.WorkflowName.PROCESS_SEARCH_BRANCH)
    assert len(calls) == 1
    assert calls[0].args[0].branch_idx == 0


async def test_run__when_three_branches__then_starts_three_children_with_sequential_ids(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    surf_params = SurfParamsFactory(
        url_template_params=[
            TemplateContextFactory.build(values={'category': c})
            for c in ('x', 'y', 'z')
        ],
    )
    set_mock(consts.WorkflowName.PROCESS_SEARCH_BRANCH, result=None)
    in_ = SearchAdvertsInFactory(surf_config_name='demo', surf_params=surf_params)

    await make_workers([
        WorkerSpec(
            workflows=[workflows.SearchAdverts],
            activities=[
                activities.SurfConfigRepo(
                    surfing_repo.DummyConfigRepository(result=surf_params),  # type: ignore[arg-type]
                ).get_surf_params,
            ],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    await env.client.execute_workflow(  # type: ignore[misc]
        workflows.SearchAdverts.run,
        in_,
        id='test-sa-three-branches',
        task_queue=consts.QueueName.SURFING_TASK,
    )

    calls = get_mock_calls(consts.WorkflowName.PROCESS_SEARCH_BRANCH)
    assert len(calls) == 3
    assert [c.args[0].branch_idx for c in calls] == [0, 1, 2]


async def test_run__when_some_branches_fail__then_raises_error(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    surf_params = SurfParamsFactory(
        url_template_params=[
            TemplateContextFactory.build(values={'category': c})
            for c in ('x', 'y', 'z')
        ],
    )
    set_mock(
        consts.WorkflowName.PROCESS_SEARCH_BRANCH,
        error=RuntimeError('branch boom'),
    )
    in_ = SearchAdvertsInFactory(surf_config_name='demo', surf_params=surf_params)

    await make_workers([
        WorkerSpec(
            workflows=[workflows.SearchAdverts],
            activities=[
                activities.SurfConfigRepo(
                    surfing_repo.DummyConfigRepository(result=surf_params),  # type: ignore[arg-type]
                ).get_surf_params,
            ],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    with pytest.raises(temporalio.client.WorkflowFailureError) as excinfo:
        await env.client.execute_workflow(  # type: ignore[misc]
            workflows.SearchAdverts.run,
            in_,
            id='test-sa-some-fail',
            task_queue=consts.QueueName.SURFING_TASK,
        )

    cause = excinfo.value.cause
    assert isinstance(cause, temporalio.exceptions.ApplicationError)
    assert len(cause.details) == 3


async def test_run__when_repo_raises__then_workflow_fails(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    await make_workers([
        WorkerSpec(
            workflows=[workflows.SearchAdverts],
            activities=[
                activities.SurfConfigRepo(
                    surfing_repo.DummyConfigRepository(
                        error=RuntimeError('repo boom'),
                    ),
                ).get_surf_params,
            ],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    with pytest.raises(temporalio.client.WorkflowFailureError) as excinfo:
        await env.client.execute_workflow(  # type: ignore[misc]
            workflows.SearchAdverts.run,
            SearchAdvertsInFactory(),
            id='test-sa-repo-fail',
            task_queue=consts.QueueName.SURFING_TASK,
        )

    assert_workflow_failure_message(excinfo, 'repo boom')


async def test_run__when_all_branches_succeed__then_completes(
    env: temporalio.testing.WorkflowEnvironment,
    make_workers: tp.Callable[[list[WorkerSpec]], tp.Awaitable[list[temporalio.worker.Worker]]],
) -> None:
    clear_mocks()
    surf_params = SurfParamsFactory(
        url_template_params=[
            TemplateContextFactory.build(values={'category': c})
            for c in ('x', 'y')
        ],
    )
    set_mock(consts.WorkflowName.PROCESS_SEARCH_BRANCH, result=None)
    in_ = SearchAdvertsInFactory(surf_config_name='demo', surf_params=surf_params)

    await make_workers([
        WorkerSpec(
            workflows=[workflows.SearchAdverts],
            activities=[
                activities.SurfConfigRepo(
                    surfing_repo.DummyConfigRepository(result=surf_params),  # type: ignore[arg-type]
                ).get_surf_params,
            ],
            task_queue=consts.QueueName.SURFING_TASK,
            interceptors=[WorkflowMockInterceptor()],
        ),
    ])

    result = await env.client.execute_workflow(  # type: ignore[misc]
        workflows.SearchAdverts.run,
        in_,
        id='test-sa-all-ok',
        task_queue=consts.QueueName.SURFING_TASK,
    )
    assert result is None
