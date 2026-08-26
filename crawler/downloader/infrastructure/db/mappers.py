from downloader.domain import downloading
from downloader.infrastructure.db.orm import download_configs as orm
from surfer.domain import adverts


def download_config_to_params(row: orm.DownloadConfigORM) -> downloading.Params:
    return downloading.Params(
        id=row.id,
        name=row.name,
        source_id=adverts.SourceID(row.source_id),
        headers=row.headers,
    )
