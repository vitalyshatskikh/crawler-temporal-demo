from temporalio import workflow

with workflow.unsafe.imports_passed_through():
    from downloader.application import activities
    from surfer.application import consts as surfer_consts
    from surfer.application.workflows.models import DownloadIn


@workflow.defn(name=surfer_consts.WorkflowName.DOWNLOAD_SEARCH_PAGE)
class DownloadSearchPage:
    @workflow.run
    async def run(self, in_: DownloadIn) -> None:
        workflow.logger.info(
            "starting %s", surfer_consts.WorkflowName.DOWNLOAD_SEARCH_PAGE,
            extra={"sdoc_id": str(in_.meta.sdoc_id)},
        )

        conf = await workflow.execute_local_activity_method(
            activities.DownloadingRepo.get_downloading_config,
            args=[in_.meta.source_id, in_.meta.type],
            schedule_to_close_timeout=in_.config_request_timeout,
            retry_policy=in_.config_request_retry.to_retry_policy(),
        )

        await workflow.execute_local_activity_method(
            activities.WebDownloader.download_to_repo,
            args=[conf, in_.meta],
            schedule_to_close_timeout=in_.download_timeout,
            retry_policy=in_.download_retry.to_retry_policy(
                non_retryable_error_types=["DownloaderError", "ValidationError"],
            ),
        )
