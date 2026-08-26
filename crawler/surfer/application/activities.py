from temporalio import activity

from surfer.application import consts
from surfer.domain import adverts, surfing


class SurfConfigRepo:
    def __init__(self, repo: surfing.ISurfingRepository):
        self._repo = repo

    @activity.defn(name=consts.ActivityName.GET_SURF_CONFIG)
    async def get_surf_params(self, name: str) -> surfing.Params:
        return await self._repo.get_surf_config(name)


# TODO move into separate app
@activity.defn(name=consts.ActivityName.PARSE_SEARCH_PAGE)
async def dummy_parse_search_page(meta: adverts.DocumentMeta) -> list[adverts.DocumentMeta]:
    activity.logger.info("parse search page %s", meta.external_url)
    return []


# TODO move into separate app
@activity.defn(name=consts.ActivityName.PARSE_ADVERT_CONTENT)
async def dummy_parse_advert_content(sdocid: adverts.SdocID) -> None:
    activity.logger.info("parse advert content %s", sdocid)
