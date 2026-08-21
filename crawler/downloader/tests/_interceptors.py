"""
Workflow interceptor for mocking child workflows and remote activities in tests.

Replaces ContextVar-based stub workflows with a single interceptor that
intercept calls from workflows under test.  Mock state lives in
module-level dicts — no ContextVar needed (interceptor methods run with this
module's __globals__, which are shared with the test code).
"""



import typing as tp
from collections.abc import Sequence
from dataclasses import dataclass

import temporalio.exceptions
import temporalio.worker

# ---------------------------------------------------------------------------
# Module-level mock state (shared between test code and interceptor)
# ---------------------------------------------------------------------------

_mock_results: dict[str, tp.Any] = {}
_mock_errors: dict[str, BaseException] = {}


@dataclass
class MockCall:
    """A recorded intercepted call (child workflow or activity)."""

    kind: str
    name: str
    args: Sequence[tp.Any]
    id: str


_mock_calls: list[MockCall] = []


def set_mock(
    name: str,
    *,
    result: tp.Any = None,
    error: BaseException | None = None,
) -> None:
    """Configure mock behaviour for a workflow or activity by name.

    If ``error`` is set, the mock handle raises it when awaited.
    Otherwise the mock handle returns ``result``.
    """
    if error is not None:
        _mock_errors[name] = error
        _mock_results.pop(name, None)
    else:
        _mock_results[name] = result
        _mock_errors.pop(name, None)


def clear_mocks() -> None:
    """Clear all mock state. Call at the start of each test."""
    _mock_results.clear()
    _mock_errors.clear()
    _mock_calls.clear()


def get_mock_calls(name: str | None = None) -> list[MockCall]:
    """Return recorded calls. If ``name`` is given, filter by it."""
    if name is None:
        return list(_mock_calls)
    return [c for c in _mock_calls if c.name == name]


# ---------------------------------------------------------------------------
# Mock handle — awaitable stand-in for ChildWorkflowHandle / ActivityHandle
# ---------------------------------------------------------------------------


class MockHandle:
    """Minimal awaitable handle returned by the interceptor.

    Mimics ``temporalio.workflow.ChildWorkflowHandle`` just enough:
    - ``await handle``  → returns ``result`` or raises ``error``
    - ``handle.id``     → workflow id (or empty string for activities)
    - ``handle.first_execution_run_id`` → stub value
    - ``await handle.signal(...)`` → no-op
    """

    def __init__(
        self,
        wf_id: str = "",
        result: tp.Any = None,
        error: BaseException | None = None,
    ) -> None:
        self._id = wf_id
        self._result = result
        self._error = error
        self.first_execution_run_id: str | None = "mock-run-id"

    @property
    def id(self) -> str:
        return self._id

    async def signal(
        self,
        signal: str | tp.Callable,
        arg: tp.Any = ...,
        *,
        args: Sequence[tp.Any] = [],
    ) -> None:
        pass

    def __await__(self) -> tp.Generator[None, None, tp.Any]:
        if self._error is not None:
            raise temporalio.exceptions.ApplicationError(
                str(self._error),
                type=self._error.__class__.__name__,
            )
        yield self._result


# ---------------------------------------------------------------------------
# Interceptor implementation
# ---------------------------------------------------------------------------


class WorkflowMockInterceptor(temporalio.worker.Interceptor):
    """Worker-level interceptor that injects mock outbound for every workflow.

    Usage::

        Worker(
            ...,
            interceptors=[WorkflowMockInterceptor()],
        )
    """

    def workflow_interceptor_class(
        self,
        input: temporalio.worker.WorkflowInterceptorClassInput,
    ) -> type[temporalio.worker.WorkflowInboundInterceptor] | None:
        return _MockInbound


class _MockInbound(temporalio.worker.WorkflowInboundInterceptor):
    def init(self, outbound: temporalio.worker.WorkflowOutboundInterceptor) -> None:
        super().init(_MockOutbound(outbound))


class _MockOutbound(temporalio.worker.WorkflowOutboundInterceptor):
    async def start_child_workflow(
        self,
        input: temporalio.worker.StartChildWorkflowInput,
    ) -> tp.Any:
        _mock_calls.append(
            MockCall(kind="workflow", name=input.workflow, args=input.args, id=input.id)
        )
        error = _mock_errors.get(input.workflow)
        result = _mock_results.get(input.workflow)
        return MockHandle(wf_id=input.id, result=result, error=error)

    def start_activity(
        self,
        input: temporalio.worker.StartActivityInput,
    ) -> tp.Any:
        _mock_calls.append(
            MockCall(
                kind="activity",
                name=input.activity,
                args=input.args,
                id=input.activity_id or "",
            )
        )
        error = _mock_errors.get(input.activity)
        result = _mock_results.get(input.activity)
        return MockHandle(result=result, error=error)

    def start_local_activity(
        self,
        input: temporalio.worker.StartLocalActivityInput,
    ) -> tp.Any:
        _mock_calls.append(
            MockCall(
                kind="local_activity",
                name=input.activity,
                args=input.args,
                id=input.activity_id or "",
            )
        )
        error = _mock_errors.get(input.activity)
        result = _mock_results.get(input.activity)
        return MockHandle(result=result, error=error)
