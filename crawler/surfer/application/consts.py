import enum


class QueueName(enum.StrEnum):
    SURFING_TASK = "surfing"
    ADVERT_PROCESSING = "advert-processing"
    DOWNLOADING = "downloading"
    PARSING = "parsing"


class ActivityName(enum.StrEnum):
    GET_SURF_CONFIG = "GetSurfParams"
    GET_DOCUMENTS_META = "GetDocumentsMeta"
    PARSE_SEARCH_PAGE = "ParseSearchPage"
    PARSE_ADVERT_CONTENT = "ParseAdvertContent"


class WorkflowName(enum.StrEnum):
    SEARCH_ADVERTS = "SearchAdverts"
    PROCESS_SEARCH_BRANCH = "ProcessSearchBranch"
    PROCESS_SEARCH_PAGE = "ProcessSearchPage"
    PROCESS_ADVERT = "ProcessAdvert"
    DOWNLOAD_SEARCH_PAGE = "DownloadSearchPage"
    DOWNLOAD_ADVERT_CONTENT = "DownloadAdvertContent"
