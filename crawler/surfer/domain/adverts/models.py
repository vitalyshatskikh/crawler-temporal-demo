import datetime as dt
import enum
import typing as tp

import pydantic

SdocID = tp.NewType("SdocID", str)

SourceID = tp.NewType("SourceID", str)


class DocumentType(enum.StrEnum):
    SEARCH_PAGE = "search_page"
    SURFED_ADVERT = "surfed_advert"
    DOWNLOADED_ADVERT = "downloaded_advert"
    PARSED_ADVERT = "parsed_advert"


class DocumentMeta(pydantic.BaseModel):
    sdoc_id: SdocID = pydantic.Field(..., min_length=1)
    created_at: dt.datetime = pydantic.Field()
    updated_at: dt.datetime
    source_id: SourceID = pydantic.Field(..., min_length=1)
    type: DocumentType
    external_url: str = pydantic.Field(..., min_length=1)
    content_url: str = pydantic.Field(default="")
    update_interval_sec: int = pydantic.Field(..., gt=0)

    @pydantic.model_validator(mode='after')
    def validate_dates(self) -> tp.Self:
        if self.created_at is None:
            raise ValueError("created_at required")
        if not self.updated_at >= self.created_at:
            raise ValueError("updated_at must be not before created_at")
        return self


class Document(DocumentMeta):
    body: str = pydantic.Field(..., min_length=1)
