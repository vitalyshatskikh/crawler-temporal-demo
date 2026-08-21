from temporalio import activity

from downloader.application import consts
from downloader.domain import downloading
from surfer.domain import adverts


class DownloadingRepo:
    def __init__(self, repo: downloading.IDownloadingRepository) -> None:
        self._repo = repo

    @activity.defn(name=consts.ActivityName.GET_DOWNLOADING_CONFIG)
    async def get_downloading_config(
        self,
        source_id: adverts.SourceID,
        doc_type: adverts.DocumentType,
    ) -> downloading.Params:
        return await self._repo.get_download_config(source_id, doc_type)
