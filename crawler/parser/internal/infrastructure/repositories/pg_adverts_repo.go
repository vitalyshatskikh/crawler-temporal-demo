package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	queries "github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/db/gen"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/application"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
)

var _ application.AdvertsRepository = (*PGAdvertsRepo)(nil)

type PGAdvertsRepo struct {
	q *queries.Queries
}

func NewPGAdvertsRepo(pool *pgxpool.Pool) *PGAdvertsRepo {
	return &PGAdvertsRepo{q: queries.New(pool)}
}

func (r *PGAdvertsRepo) GetDocument(ctx context.Context, sdocID domain.SdocID, sourceID domain.SourceID, docType domain.DocumentType) (domain.Document, error) {
	row, err := r.q.GetDocument(ctx, queries.GetDocumentParams{
		SdocID:   string(sdocID),
		SourceID: string(sourceID),
		DocType:  string(docType),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Document{}, domain.ErrNotFound
		}
		return domain.Document{}, err
	}

	return domain.Document{
		DocumentMeta: domain.DocumentMeta{
			SdocID:            domain.SdocID(row.SdocID),
			CreatedAt:         row.CreatedAt.Time,
			UpdatedAt:         row.UpdatedAt.Time,
			SourceID:          domain.SourceID(row.SourceID),
			Type:              domain.DocumentType(row.DocType),
			ExternalURL:       row.ExternalUrl,
			ContentURL:        row.ContentUrl,
			UpdateIntervalSec: int(row.UpdateIntervalSec),
		},
		Body: []byte(row.Body),
	}, nil
}

func (r *PGAdvertsRepo) SaveDocument(ctx context.Context, doc domain.Document) error {
	return r.q.UpsertDocument(ctx, queries.UpsertDocumentParams{
		SdocID:            string(doc.SdocID),
		SourceID:          string(doc.SourceID),
		DocType:           string(doc.Type),
		ExternalUrl:       doc.ExternalURL,
		ContentUrl:        doc.ContentURL,
		Body:              string(doc.Body),
		CreatedAt:         pgtype.Timestamptz{Time: doc.CreatedAt, Valid: true},
		UpdatedAt:         pgtype.Timestamptz{Time: doc.UpdatedAt, Valid: true},
		UpdateIntervalSec: int32(doc.UpdateIntervalSec),
	})
}
