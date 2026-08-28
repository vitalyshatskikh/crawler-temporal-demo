import pydantic
from temporalio import workflow

# Pass the activities through the sandbox
# https://github.com/temporalio/sdk-python#workflow-sandbox
with workflow.unsafe.imports_passed_through():
    from surfer.application import config, consts
    from surfer.application.workflows.models import DownloadIn
    from surfer.domain import adverts, surfing


class ProcessAdvertIn(pydantic.BaseModel):
    surfer_config: config.SurferConfig
    surf_params: surfing.Params
    doc_meta: adverts.DocumentMeta


@workflow.defn(name=consts.WorkflowName.PROCESS_ADVERT)
class ProcessAdvert:
    @workflow.run
    async def run(self, in_: ProcessAdvertIn) -> None:
        workflow.logger.info(
            "starting %s", consts.WorkflowName.PROCESS_ADVERT,
            extra={"surf_config_name": in_.surf_params.name, "sdoc_id": str(in_.doc_meta.sdoc_id)},
        )

        download_wf_id = (
            f"{consts.WorkflowName.DOWNLOAD_ADVERT_CONTENT}/{in_.surf_params.name}"
            f"/sdocid/{in_.doc_meta.sdoc_id}"
        )
        await workflow.execute_child_workflow(
            consts.WorkflowName.DOWNLOAD_ADVERT_CONTENT,
            DownloadIn(
                meta=in_.doc_meta,
                download_timeout=in_.surfer_config.download_advert_content_timeout,
                config_request_timeout=in_.surfer_config.repo_request_timeout,
            ),
            id=download_wf_id,
            task_queue=consts.QueueName.DOWNLOADING,
            execution_timeout=in_.surfer_config.download_advert_content_wf_timeout,
        )

        await workflow.execute_activity(
            consts.ActivityName.PARSE_ADVERT_CONTENT,
            in_.doc_meta.model_copy(update={'type': adverts.DocumentType.DOWNLOADED_ADVERT}),
            task_queue=consts.QueueName.PARSING,
            start_to_close_timeout=in_.surfer_config.parse_advert_content_timeout,
            retry_policy=in_.surfer_config.parse_advert_content_retry.to_retry_policy(),
        )

        # TODO add indexing stage
