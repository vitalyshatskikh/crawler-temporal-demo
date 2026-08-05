import datetime as dt

import pydantic

from shared.py.settings import RetryConfig

_MIN_TIMEOUT = dt.timedelta(seconds=1)


class SurferConfig(pydantic.BaseModel):
    search_adverts_timeout: dt.timedelta = pydantic.Field(dt.timedelta(minutes=20), ge=_MIN_TIMEOUT)

    process_branch_wf_timeout: dt.timedelta = pydantic.Field(dt.timedelta(minutes=15), ge=_MIN_TIMEOUT)
    process_search_page_wf_timeout: dt.timedelta = pydantic.Field(dt.timedelta(minutes=5), ge=_MIN_TIMEOUT)
    process_advert_wf_timeout: dt.timedelta = pydantic.Field(dt.timedelta(minutes=5), ge=_MIN_TIMEOUT)
    # should be a little bit less then process_*_timeout
    download_search_page_wf_timeout: dt.timedelta = pydantic.Field(dt.timedelta(minutes=4), ge=_MIN_TIMEOUT)
    download_advert_content_wf_timeout: dt.timedelta = pydantic.Field(dt.timedelta(minutes=4), ge=_MIN_TIMEOUT)
    # should be a little bit less then download_*_wf_timeout
    download_search_page_timeout: dt.timedelta = pydantic.Field(dt.timedelta(minutes=4), ge=_MIN_TIMEOUT)
    download_advert_content_timeout: dt.timedelta = pydantic.Field(dt.timedelta(minutes=4), ge=_MIN_TIMEOUT)

    repo_request_timeout: dt.timedelta = pydantic.Field(dt.timedelta(seconds=15), ge=_MIN_TIMEOUT)
    repo_request_retry: RetryConfig = RetryConfig()

    parse_search_page_timeout: dt.timedelta = pydantic.Field(dt.timedelta(seconds=30), ge=_MIN_TIMEOUT)
    parse_search_page_retry: RetryConfig = RetryConfig()

    parse_advert_content_timeout: dt.timedelta = pydantic.Field(dt.timedelta(seconds=30), ge=_MIN_TIMEOUT)
    parse_advert_content_retry: RetryConfig = RetryConfig()