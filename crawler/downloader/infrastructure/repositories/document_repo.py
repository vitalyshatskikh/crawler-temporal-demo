import logging

from downloader.domain import downloading
from surfer.domain import adverts

logger = logging.getLogger(__name__)


class PGDocumentRepository(downloading.IDocumentRepository):
    async def save(self, document: adverts.Document) -> None:
        # TODO implement me
        logger.info("saving document: %s", document.model_dump(exclude={'body'}))
