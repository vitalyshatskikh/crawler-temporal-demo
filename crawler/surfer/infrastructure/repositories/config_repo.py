from surfer.domain import adverts, errors, surfing


class PGConfigRepository(surfing.ISurfingRepository):
    async def get_surf_config(self, name: str) -> surfing.Params:
        # TODO implement me
        match name:
            case "example.com/fresh":
                return surfing.Params(
                    id=1,
                    name=name,
                    source_id=adverts.SourceID(name),
                    url_template="https://example.com/adverts/{{category}}?page={{page}}",
                    url_template_params=[
                        surfing.TemplateContext(values={"category": "x"}, comment="x"),
                        surfing.TemplateContext(values={"category": "y"}, comment="y"),
                        surfing.TemplateContext(values={"category": "z"}, comment="z"),
                    ],
                    max_pages=5,
                )
            case "example.com/all":
                return surfing.Params(
                    id=1,
                    name=name,
                    source_id=adverts.SourceID(name),
                    url_template="https://example.com/adverts/{{category}}?page={{page}}",
                    url_template_params=[
                        surfing.TemplateContext(values={"category": "x"}, comment="x"),
                        surfing.TemplateContext(values={"category": "y"}, comment="y"),
                        surfing.TemplateContext(values={"category": "z"}, comment="z"),
                    ],
                    max_pages=100,
                )
            case _:
                raise errors.NotFoundError("surf config not found", name)

    async def get_surf_schedules(self) -> dict[str, str]:
        return {
            "example.com/fresh": "0/1 * * * *",
            "example.com/all": "0 * * * *",
        }
