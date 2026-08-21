import enum


class ActivityName(enum.StrEnum):
    GET_DOWNLOADING_CONFIG = "GetDownloadingConfig"
    DOWNLOAD_TO_REPO = "DownloadToRepo"
