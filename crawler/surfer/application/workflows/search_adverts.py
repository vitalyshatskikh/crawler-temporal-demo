import pydantic
from temporalio import exceptions, workflow

# Pass the activities through the sandbox
# https://github.com/temporalio/sdk-python#workflow-sandbox
with workflow.unsafe.imports_passed_through():
    from surfer.application import activities, config, consts

from surfer.application.workflows import process_search_branch


class SearchAdvertsIn(pydantic.BaseModel):
    surfer_config: config.SurferConfig
    surf_config_name: str = pydantic.Field(..., min_length=1)


@workflow.defn(name=consts.WorkflowName.SEARCH_ADVERTS)
class SearchAdverts:
    @workflow.run
    async def run(self, in_: SearchAdvertsIn) -> None:
        workflow.logger.info(
            "starting %s", consts.WorkflowName.SEARCH_ADVERTS, extra={"surf_config_name": in_.surf_config_name},
        )

        surf_params = await workflow.execute_local_activity_method(
            activities.SurfConfigRepo.get_surf_params,
            in_.surf_config_name,
            schedule_to_close_timeout=in_.surfer_config.repo_request_timeout,
            retry_policy=in_.surfer_config.repo_request_retry.to_retry_policy(),
        )

        handles = []
        for i, _ in enumerate(surf_params.url_template_params):
            wf_id = f"{consts.WorkflowName.PROCESS_SEARCH_BRANCH}/{surf_params.name}/branch/{i}"
            handle = await workflow.start_child_workflow(
                process_search_branch.ProcessSearchBranch.run,
                process_search_branch.ProcessSearchBranchIn(
                    surfer_config=in_.surfer_config,
                    surf_params=surf_params,
                    branch_idx=i,
                ),
                id=wf_id,
                execution_timeout=in_.surfer_config.process_branch_wf_timeout,
            )
            handles.append(handle)

        branch_errors: list[BaseException] = []
        for h in handles:
            try:
                await h
            except Exception as exc:  # noqa: BLE001
                branch_errors.append(exc)
        if branch_errors:
            raise exceptions.ApplicationError("some branches failed", *map(repr, branch_errors))