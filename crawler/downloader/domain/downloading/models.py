import pydantic

from surfer.domain import adverts


class Params(pydantic.BaseModel):
    id: int
    name: str = pydantic.Field(..., min_length=1)
    source_id: adverts.SourceID = pydantic.Field(..., min_length=1)
    headers: dict[str, str] = pydantic.Field(default_factory=dict)
