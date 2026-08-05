import pydantic
import pytest

from surfer.domain import surfing


@pytest.mark.parametrize(
    "payload,expect_error",
    [
        (
            {"id": 1, "name": "", "source_id": "x", "url_template": "https://x/{{page}}", "max_pages": 5},
            True,
        ),
        ({"id": 1, "name": "n", "source_id": "x", "url_template": "", "max_pages": 5}, True),
        (
            {
                "id": 1,
                "name": "n",
                "source_id": "x",
                "url_template": "https://x/{{page}}",
                "max_pages": 0,
            },
            True,
        ),
        (
            {
                "id": 1,
                "name": "n",
                "source_id": "x",
                "url_template": "https://x/{{page}}",
                "max_pages": -1,
            },
            True,
        ),
        (
            {
                "id": 1,
                "name": "n",
                "source_id": "x",
                "url_template": "https://x/{{page}}",
                "max_pages": 1,
            },
            False,
        ),
    ],
)
def test_params__when_invalid_fields__then_raises_validation_error(
    payload: dict, expect_error: bool
) -> None:
    # When
    if expect_error:
        with pytest.raises(pydantic.ValidationError):
            surfing.Params(**payload)
    else:
        result = surfing.Params(**payload)

        # Then
        assert result.max_pages == payload["max_pages"]


def test_template_context_defaults() -> None:
    # When
    ctx = surfing.TemplateContext()

    # Then
    assert ctx.values == {}
    assert ctx.comment == ""
