import abc

from surfer.domain.surfing import models


class ISurfingRepository(abc.ABC):
    @abc.abstractmethod
    async def get_surf_config(self, name: str) -> models.Params: ...


class DummyConfigRepository(ISurfingRepository):
    def __init__(
        self,
        *,
        result: models.Params | None = None,
        error: Exception | None = None,
    ) -> None:
        self._result = result
        self._error = error

    async def get_surf_config(self, name: str) -> models.Params:
        if self._error is not None:
            raise self._error
        assert self._result is not None
        return self._result
