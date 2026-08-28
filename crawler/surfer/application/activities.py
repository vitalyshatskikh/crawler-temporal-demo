from temporalio import activity

from surfer.application import consts
from surfer.domain import surfing


class SurfConfigRepo:
    def __init__(self, repo: surfing.ISurfingRepository):
        self._repo = repo

    @activity.defn(name=consts.ActivityName.GET_SURF_CONFIG)
    async def get_surf_params(self, name: str) -> surfing.Params:
        return await self._repo.get_surf_config(name)
