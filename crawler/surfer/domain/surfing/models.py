
import pydantic

from surfer.domain import adverts

URL_TEMPLATE_PAGE_PARAM = "page"


class TemplateContext(pydantic.BaseModel):
    values: dict[str, str] = pydantic.Field(default_factory=dict)
    comment: str = ""


class Params(pydantic.BaseModel):
    id: int
    name: str = pydantic.Field(..., min_length=1)
    source_id: adverts.SourceID = pydantic.Field(..., min_length=1)
    url_template: str = pydantic.Field(..., min_length=1)
    url_template_params: list[TemplateContext] = pydantic.Field(default_factory=list)
    max_pages: int = pydantic.Field(..., gt=0)
    update_interval_sec: int = pydantic.Field(..., gt=0)

    @pydantic.field_validator('url_template_params', mode='after')
    @classmethod
    def validate_url_template_params(cls, v: list[TemplateContext]) -> list[TemplateContext]:
        if not v:
            return [TemplateContext(values={})]
        return v
