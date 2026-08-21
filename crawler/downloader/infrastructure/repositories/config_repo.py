from downloader.domain import downloading, errors
from downloader.domain.downloading import models
from surfer.domain import adverts


class PGDownloadingRepository(downloading.IDownloadingRepository):
    async def get_download_config(
        self,
        source_id: adverts.SourceID,
        doc_type: adverts.DocumentType,
    ) -> models.Params:
        raise errors.NotFoundError(
            f"download config not found: source={source_id} doc_type={doc_type}",
        )
