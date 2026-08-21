import abc

from downloader.domain.downloading import models
from surfer.domain import adverts


class IDownloadingRepository(abc.ABC):
    @abc.abstractmethod
    async def get_download_config(self, source_id: adverts.SourceID, doc_type: adverts.DocumentType) -> models.Params:
        ...


class IDocumentRepository(abc.ABC):
    @abc.abstractmethod
    async def save(self, document: adverts.Document) -> None: ...


class DummyConfigRepository(IDownloadingRepository):
    def __init__(
        self,
        *,
        result: models.Params | None = None,
        error: Exception | None = None,
    ) -> None:
        self._result = result
        self._error = error

    async def get_download_config(self, source_id: adverts.SourceID, doc_type: adverts.DocumentType) -> models.Params:
        if self._error is not None:
            raise self._error
        assert self._result is not None
        return self._result


class DummyDocumentRepository(IDocumentRepository):
    def __init__(
        self,
        *,
        error: Exception | None = None,
    ) -> None:
        self._error = error
        self._saved: list[adverts.Document] = []

    async def save(self, document: adverts.Document) -> None:
        if self._error is not None:
            raise self._error
        self._saved.append(document)

    @property
    def saved(self) -> list[adverts.Document]:
        return self._saved
