import datetime as dt

import pydantic

from shared.py.settings import RetryConfig
from surfer.domain import adverts

_MIN_TIMEOUT = dt.timedelta(seconds=1)


class DownloadIn(pydantic.BaseModel):
    meta: adverts.DocumentMeta

    download_timeout: dt.timedelta = pydantic.Field(dt.timedelta(minutes=2), ge=_MIN_TIMEOUT)
    download_retry: RetryConfig = RetryConfig(max_attempts=1)

    config_request_timeout: dt.timedelta = pydantic.Field(dt.timedelta(seconds=15), ge=_MIN_TIMEOUT)
    config_request_retry: RetryConfig = RetryConfig()
