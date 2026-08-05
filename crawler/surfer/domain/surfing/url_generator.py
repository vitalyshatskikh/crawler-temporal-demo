import typing as tp

import chevron  # type: ignore[import-untyped]

from surfer.domain import errors, surfing


class URLGenerator:
    def __init__(self, template: str) -> None:
        try:
            chevron.render(template, {})
        except chevron.ChevronError as exc:
            raise errors.ValidationError(f"invalid mustache template: {exc}") from exc
        self._template = template

    def branch(self, ctx: surfing.TemplateContext) -> "BranchURLGenerator":
        return BranchURLGenerator(self._template, ctx.values)


class BranchURLGenerator:
    def __init__(self, template: str, params: dict[str, str]) -> None:
        self._template = template
        self._params = params

    def page(self, page_num: int) -> str:
        self._params[surfing.URL_TEMPLATE_PAGE_PARAM] = str(page_num)
        try:
            return tp.cast("str", chevron.render(self._template, self._params))
        except chevron.ChevronError as exc:
            raise errors.ValidationError(f"cannot render url: {exc}") from exc
