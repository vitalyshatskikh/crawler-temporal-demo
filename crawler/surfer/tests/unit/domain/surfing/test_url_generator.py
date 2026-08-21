import threading

import pytest

from surfer.domain import errors, surfing

TEMPLATE = "https://example.com/adverts/{{category}}?page={{page}}"


@pytest.mark.parametrize(
    "category,page,expected",
    [
        ("x", 1, "https://example.com/adverts/x?page=1"),
        ("x", 2, "https://example.com/adverts/x?page=2"),
        ("y", 1, "https://example.com/adverts/y?page=1"),
        ("z", 10, "https://example.com/adverts/z?page=10"),
    ],
)
def test_page__when_valid_template__then_renders_url(
    category: str, page: int, expected: str
) -> None:
    # Given
    gen = surfing.URLGenerator(TEMPLATE)
    branch = gen.branch(surfing.TemplateContext(values={"category": category}))

    # When
    result = branch.page(page)

    # Then
    assert result == expected


def test_new_url_generator__when_bad_template__then_raises_validation_error() -> None:
    # When / Then
    with pytest.raises(errors.ValidationError):
        surfing.URLGenerator("{{unclosed")


def test_page__when_concurrent_branches_share_template__then_no_cross_talk() -> None:
    # Given
    gen = surfing.URLGenerator(TEMPLATE)
    branch_x = gen.branch(surfing.TemplateContext(values={"category": "x"}))
    branch_y = gen.branch(surfing.TemplateContext(values={"category": "y"}))

    results: list[str] = []
    lock = threading.Lock()

    def render(branch: surfing.BranchURLGenerator, page: int) -> None:
        url = branch.page(page)
        with lock:
            results.append(url)

    # When
    t1 = threading.Thread(target=render, args=(branch_x, 1))
    t2 = threading.Thread(target=render, args=(branch_y, 1))
    t1.start()
    t2.start()
    t1.join()
    t2.join()

    # Then
    assert "https://example.com/adverts/x?page=1" in results
    assert "https://example.com/adverts/y?page=1" in results
