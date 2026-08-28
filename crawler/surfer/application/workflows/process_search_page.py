import datetime as dt
import hashlib

import pydantic
from temporalio import workflow

# Pass the activities through the sandbox
# https://github.com/temporalio/sdk-python#workflow-sandbox
with workflow.unsafe.imports_passed_through():
    from surfer.application import config, consts
    from surfer.application.workflows.models import DownloadIn
    from surfer.domain import adverts, surfing

from surfer.application.workflows import process_advert


class ProcessSearchPageIn(pydantic.BaseModel):
    surfer_config: config.SurferConfig
    surf_params: surfing.Params
    branch_idx: int
    page_num: int
    page_url: str


@workflow.defn(name=consts.WorkflowName.PROCESS_SEARCH_PAGE)
class ProcessSearchPage:
    @workflow.run
    async def run(self, in_: ProcessSearchPageIn) -> None:
        workflow.logger.info(
            "starting %s", consts.WorkflowName.PROCESS_SEARCH_PAGE,
            extra={"surf_config_name": in_.surf_params.name, "page_url": in_.page_url},
        )

        download_wf_id = (
            f"{consts.WorkflowName.DOWNLOAD_SEARCH_PAGE}/{in_.surf_params.name}"
            f"/branch{in_.branch_idx}/page{in_.page_num}"
        )
        now = workflow.now()
        page_doc_meta = adverts.DocumentMeta(
            sdoc_id=adverts.SdocID(hashlib.md5(in_.page_url.encode()).hexdigest()),
            created_at=now,
            updated_at=now,
            source_id=in_.surf_params.source_id,
            type=adverts.DocumentType.SEARCH_PAGE,
            external_url=in_.page_url,
            update_interval_sec=in_.surf_params.update_interval_sec,
        )
        await workflow.execute_child_workflow(
            consts.WorkflowName.DOWNLOAD_SEARCH_PAGE,
            DownloadIn(
                meta=page_doc_meta,
                download_timeout=in_.surfer_config.download_search_page_timeout,
                config_request_timeout=in_.surfer_config.repo_request_timeout,
            ),
            id=download_wf_id,
            task_queue=consts.QueueName.DOWNLOADING,
            execution_timeout=in_.surfer_config.download_search_page_wf_timeout,
        )

        documents_meta: list[adverts.DocumentMeta] = await workflow.execute_activity(
            consts.ActivityName.PARSE_SEARCH_PAGE,
            page_doc_meta,
            task_queue=consts.QueueName.PARSING,
            start_to_close_timeout=in_.surfer_config.parse_search_page_timeout,
            retry_policy=in_.surfer_config.parse_search_page_retry.to_retry_policy(),
        )

        processed = 0
        skipped = 0
        for doc_meta in documents_meta:
            if doc_meta.updated_at + dt.timedelta(seconds=doc_meta.update_interval_sec) > workflow.now():
                skipped += 1
                continue
            wf_id = f"{consts.WorkflowName.PROCESS_ADVERT}/{in_.surf_params.name}/sdocid/{doc_meta.sdoc_id}"
            await workflow.start_child_workflow(
                process_advert.ProcessAdvert.run,
                process_advert.ProcessAdvertIn(
                    surfer_config=in_.surfer_config,
                    surf_params=in_.surf_params,
                    doc_meta=doc_meta,
                ),
                id=wf_id,
                task_queue=consts.QueueName.ADVERT_PROCESSING,
                parent_close_policy=workflow.ParentClosePolicy.ABANDON,  # fire and forget
                execution_timeout=in_.surfer_config.process_advert_wf_timeout,
            )
            processed += 1

        workflow.logger.info(
            "documents processed",
            extra={
                "surf_config_name": in_.surf_params.name,
                "page_url": in_.page_url,
                "processed": processed,
                "skipped": skipped,
                "total": len(documents_meta),
            },
        )

