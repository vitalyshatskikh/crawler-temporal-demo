import typing as tp

from shared.py.db import orm
from surfer.domain import adverts


def document_to_meta(row: orm.DocumentORM) -> adverts.DocumentMeta:
    return adverts.DocumentMeta(
        sdoc_id=adverts.SdocID(row.sdoc_id),
        source_id=adverts.SourceID(row.source_id),
        type=adverts.DocumentType(row.doc_type),
        external_url=row.external_url,
        created_at=row.created_at,
        updated_at=row.updated_at,
    )


def document_to_orm(doc: adverts.Document) -> dict[str, tp.Any]:
    return {
        "sdoc_id": str(doc.sdoc_id),
        "source_id": str(doc.source_id),
        "doc_type": doc.type.value,
        "external_url": doc.external_url,
        "body": doc.body,
        "created_at": doc.created_at,
        "updated_at": doc.updated_at,
    }
