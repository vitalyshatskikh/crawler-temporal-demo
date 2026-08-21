import abc

import temporalio.client
import temporalio.worker

from surfer.application import activities, consts, workflows
from surfer.domain import adverts, surfing


class Worker(abc.ABC):
    @abc.abstractmethod
    async def run(self) -> None:
        ...


class SurfingWorker(Worker):
    def __init__(
        self,
        client: temporalio.client.Client,
        surf_config_repo: surfing.ISurfingRepository,
        adverts_repo: adverts.IAdvertsRepository,
    ) -> None:
        self._surf_config_repo = activities.SurfConfigRepo(surf_config_repo)
        self._adverts_repo = activities.AdvertsRepo(adverts_repo)
        self._w = temporalio.worker.Worker(
            client=client,
            task_queue=consts.QueueName.SURFING_TASK,
            workflows=[
                workflows.SearchAdverts,
                workflows.ProcessSearchBranch,
                workflows.ProcessSearchPage,
            ],
            activities=[
                self._surf_config_repo.get_surf_params,
                self._adverts_repo.get_documents_meta,
            ],
        )

    async def run(self) -> None:
        await self._w.run()


class AdvertsWorker(Worker):
    def __init__(
        self,
        client: temporalio.client.Client,
    ) -> None:
        self._w = temporalio.worker.Worker(
            client=client,
            task_queue=consts.QueueName.ADVERT_PROCESSING,
            workflows=[
                workflows.ProcessAdvert,
            ],
            activities=[],
        )

    async def run(self) -> None:
        await self._w.run()


# TODO move into separate app
class ParsingWorker(Worker):
    def __init__(
        self,
        client: temporalio.client.Client,
    ) -> None:
        self._w = temporalio.worker.Worker(
            client=client,
            task_queue=consts.QueueName.PARSING,
            workflows=[],
            activities=[
                activities.dummy_parse_search_page,
                activities.dummy_parse_advert_content,
            ],
        )

    async def run(self) -> None:
        await self._w.run()