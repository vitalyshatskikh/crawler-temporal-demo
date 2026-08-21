from temporalio import activity

from surfer.application import consts
from surfer.domain import adverts, surfing


class SurfConfigRepo:
    def __init__(self, repo: surfing.ISurfingRepository):
        self._repo = repo

    @activity.defn(name=consts.ActivityName.GET_SURF_CONFIG)
    async def get_surf_params(self, name: str) -> surfing.Params:
        return await self._repo.get_surf_config(name)


class AdvertsRepo:
    def __init__(self, repo: adverts.IAdvertsRepository):
        self._repo = repo

    @activity.defn(name=consts.ActivityName.GET_DOCUMENTS_META)
    async def get_documents_meta(self, sdoc_ids: list[adverts.SdocID]) -> dict[adverts.SdocID, adverts.DocumentMeta]:
        return await self._repo.get_documents_meta_by_sdoc_id(sdoc_ids)


# TODO move into separate app
@activity.defn(name=consts.ActivityName.PARSE_SEARCH_PAGE)
async def dummy_parse_search_page(url: str) -> list[adverts.SdocID]:
    activity.logger.info("parse search page %s", url)
    return list(map(adverts.SdocID, ("1", "2", "3")))


# TODO move into separate app
@activity.defn(name=consts.ActivityName.PARSE_ADVERT_CONTENT)
async def dummy_parse_advert_content(sdocid: adverts.SdocID) -> None:
    activity.logger.info("parse advert content %s", sdocid)
