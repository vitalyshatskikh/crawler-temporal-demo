import datetime as dt

from surfer.domain import adverts


class PGAdvertsRepo(adverts.IAdvertsRepository):
    def __init__(self) -> None:
        pass

    async def get_documents_meta_by_sdoc_id(
        self,
        sdoc_ids: list[adverts.SdocID],
    ) -> dict[adverts.SdocID, adverts.DocumentMeta]:
        return {
            id_: adverts.DocumentMeta(
                sdoc_id=id_,
                created_at=dt.datetime.now(tz=dt.UTC),
                updated_at=dt.datetime.now(tz=dt.UTC),
                source_id=adverts.SourceID("web"),
                type=adverts.DocumentType.SURFED_ADVERT,
                external_url="https://example.com/100500",
            ) for id_ in sdoc_ids
        }

