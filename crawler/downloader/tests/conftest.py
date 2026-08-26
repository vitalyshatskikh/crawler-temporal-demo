import asyncio
import typing as tp
from collections.abc import AsyncIterator, Awaitable, Callable
from dataclasses import dataclass, field

import pytest_asyncio
import temporalio.converter
import temporalio.testing
import temporalio.worker
from temporalio.contrib.pydantic import PydanticJSONPlainPayloadConverter


@dataclass(frozen=True)
class WorkerSpec:
    """Specification for a single Worker to be created by make_workers."""

    workflows: list[type]
    activities: list[tp.Any]
    task_queue: str
    interceptors: list[temporalio.worker.Interceptor] = field(default_factory=list)


_WorkerFactory = Callable[[list[WorkerSpec]], Awaitable[list[temporalio.worker.Worker]]]


@pytest_asyncio.fixture(scope="function")
async def env() -> AsyncIterator[temporalio.testing.WorkflowEnvironment]:
    """Start a time-skipping WorkflowEnvironment with pydantic v2 support."""

    class PydanticPayloadConverter(temporalio.converter.DefaultPayloadConverter):
        """Payload converter that uses pydantic v2 for JSON encoding/decoding."""

        def __init__(self) -> None:
            converters = list(
                temporalio.converter.DefaultPayloadConverter.default_encoding_payload_converters
            )
            for i, c in enumerate(converters):
                if hasattr(c, "encoding") and c.encoding == "json/plain":
                    converters[i] = PydanticJSONPlainPayloadConverter()
                    break
            temporalio.converter.CompositePayloadConverter.__init__(self, *converters)

    env_ = await temporalio.testing.WorkflowEnvironment.start_time_skipping(
        data_converter=temporalio.converter.DataConverter(
            payload_converter_class=PydanticPayloadConverter,
        ),
    )
    yield env_
    await env_.shutdown()


@pytest_asyncio.fixture
async def make_workers(
    env: temporalio.testing.WorkflowEnvironment,
) -> AsyncIterator[_WorkerFactory]:
    """Build and start N workers for a test; stop them on fixture teardown.

    Usage::

        async def test_something(env, make_workers):
            workers = await make_workers([
                WorkerSpec(
                    workflows=[DownloadAdvertContent],
                    activities=[get_downloading_config, download_to_repo],
                    task_queue="downloading",
                ),
            ])
            result = await env.client.execute_workflow(...)

    Workers are stopped automatically when the test ends.
    """
    worker_tasks: list[asyncio.Task[None]] = []
    workers: list[temporalio.worker.Worker] = []

    async def _make(specs: list[WorkerSpec]) -> list[temporalio.worker.Worker]:
        for spec in specs:
            worker = temporalio.worker.Worker(
                env.client,
                task_queue=spec.task_queue,
                workflows=spec.workflows,
                activities=spec.activities,
                interceptors=spec.interceptors,
            )
            workers.append(worker)
            task = asyncio.create_task(worker.run())
            worker_tasks.append(task)
        return workers

    yield _make

    for task in worker_tasks:
        task.cancel()
    for worker in workers:
        await worker.shutdown()
