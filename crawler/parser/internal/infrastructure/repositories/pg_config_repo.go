package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	queries "github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/db/gen"
	"github.com/vitalyshatskikh/crawler-temporal-demo/crawler/parser/internal/domain"
)

var _ domain.ConfigRepository = (*PGConfigRepo)(nil)

var ErrUnmarshalConfig = errors.New("unmarshal parsing_configs.config")

type PGConfigRepo struct {
	q *queries.Queries
}

func NewPGConfigRepo(pool *pgxpool.Pool) *PGConfigRepo {
	return &PGConfigRepo{q: queries.New(pool)}
}

func (r *PGConfigRepo) GetConfig(ctx context.Context, sourceID domain.SourceID, docType domain.DocumentType) (domain.ParsingConfig, error) {
	row, err := r.q.GetParsingConfig(ctx, queries.GetParsingConfigParams{
		SourceID: string(sourceID),
		DocType:  string(docType),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ParsingConfig{}, domain.ErrNotFound
		}
		return domain.ParsingConfig{}, err
	}

	var params []domain.ParsingParam
	if err := json.Unmarshal(row.Config, &params); err != nil {
		return domain.ParsingConfig{}, fmt.Errorf("%w: %w", ErrUnmarshalConfig, err)
	}

	return domain.ParsingConfig{
		ID:           int(row.ID),
		Name:         row.Name,
		SourceID:     domain.SourceID(row.SourceID),
		DocumentType: domain.DocumentType(row.DocType),
		Params:       params,
	}, nil
}
