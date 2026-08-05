import typing as tp

import pydantic
from temporalio import workflow

# Pass the activities through the sandbox
# https://github.com/temporalio/sdk-python#workflow-sandbox
with workflow.unsafe.imports_passed_through():
    from surfer.application import config, consts
    from surfer.domain import surfing

from surfer.application.workflows import process_search_page


class ProcessSearchBranchIn(pydantic.BaseModel):
    surfer_config: config.SurferConfig
    surf_params: surfing.Params
    branch_idx: int

    @pydantic.model_validator(mode='after')
    def validate_branch_idx(self) -> tp.Self:
        if self.branch_idx < 0 or self.branch_idx >= len(self.surf_params.url_template_params):
            raise ValueError(f"invalid branch index: {self.branch_idx}")
        return self


@workflow.defn(name=consts.WorkflowName.PROCESS_SEARCH_BRANCH)
class ProcessSearchBranch:
    @workflow.run
    async def run(self, in_: ProcessSearchBranchIn) -> None:
        workflow.logger.info(
            "starting %s", consts.WorkflowName.PROCESS_SEARCH_BRANCH,
            extra={"surf_config_name": in_.surf_params.name, "branch_index": in_.branch_idx},
        )

        url_gen = surfing.URLGenerator(in_.surf_params.url_template)
        branch_gen = url_gen.branch(in_.surf_params.url_template_params[in_.branch_idx])

        for i in range(in_.surf_params.max_pages):
            page_num = i + 1
            wf_id = (
                f"{consts.WorkflowName.PROCESS_SEARCH_PAGE}/{in_.surf_params.name}"
                f"/branch{in_.branch_idx}/page{page_num}"
            )
            actual_url = branch_gen.page(page_num)
            await workflow.execute_child_workflow(
                process_search_page.ProcessSearchPage.run,
                process_search_page.ProcessSearchPageIn(
                    surfer_config=in_.surfer_config,
                    surf_params=in_.surf_params,
                    branch_idx=in_.branch_idx,
                    page_num=page_num,
                    page_url=actual_url,
                ),
                id=wf_id,
                execution_timeout=in_.surfer_config.process_search_page_wf_timeout,
            )

