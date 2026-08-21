import datetime as dt

import pydantic

_MIN_TIMEOUT = dt.timedelta(seconds=1)


class DownloaderConfig(pydantic.BaseModel):
    http_total_timeout: dt.timedelta = pydantic.Field(dt.timedelta(seconds=30), ge=_MIN_TIMEOUT)
    http_connect_timeout: dt.timedelta = pydantic.Field(dt.timedelta(seconds=10), ge=_MIN_TIMEOUT)
    http_connector_limit: int = pydantic.Field(100, ge=1)
    http_proxy: str | None = None
