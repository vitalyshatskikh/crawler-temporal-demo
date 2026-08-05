import abc

from surfer.domain.adverts import models


class IAdvertsRepository(abc.ABC):
    @abc.abstractmethod
    async def get_documents_meta_by_sdoc_id(
        self, sdoc_ids: list[models.SdocID]
    ) -> dict[models.SdocID, models.DocumentMeta]: ...


class DummyAdvertsRepository(IAdvertsRepository):
    def __init__(
        self,
        *,
        result: dict[models.SdocID, models.DocumentMeta] | None = None,
        error: Exception | None = None,
    ) -> None:
        self._result = result
        self._error = error

    async def get_documents_meta_by_sdoc_id(
        self, sdoc_ids: list[models.SdocID]
    ) -> dict[models.SdocID, models.DocumentMeta]:
        if self._error is not None:
            raise self._error
        assert self._result is not None
        return self._result
