
from surfer.domain import adverts, surfing
from surfer.infrastructure.db import orm


def surf_config_to_params(row: orm.SurfConfigORM) -> surfing.Params:
    return surfing.Params(
        id=row.id,
        name=row.name,
        source_id=adverts.SourceID(row.source_id),
        url_template=row.url_template,
        url_template_params=[
            surfing.TemplateContext(values=p.get("values", {}), comment=p.get("comment", ""))
            for p in row.url_template_params
        ],
        max_pages=row.max_pages,
        update_interval_sec=row.update_interval_sec,
    )


def params_to_surf_config(params: surfing.Params, cron_schedule: str) -> orm.SurfConfigORM:
    return orm.SurfConfigORM(
        id=params.id,
        name=params.name,
        source_id=str(params.source_id),
        url_template=params.url_template,
        url_template_params=[t.model_dump() for t in params.url_template_params],
        max_pages=params.max_pages,
        cron_schedule=cron_schedule,
        update_interval_sec=params.update_interval_sec,
    )
