package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vitalyshatskikh/crawler-temporal-demo/example-site/internal/domain"
)

var (
	_ domain.AdvertsRepository = (*PGAdvertsRepo)(nil)

	advertsTable = "adverts"
)

type AdvertRecord struct {
	domain.AdvertIdentity
	Title       string
	Description string
	Price       int
	PubDate     time.Time `db:"pub_date"`
}

type searchRow struct {
	AdvertRecord
	TotalCount int `db:"total_count"`
}

type PGAdvertsRepo struct {
	dbPool *pgxpool.Pool
}

func NewPGAdvertsRepo(pool *pgxpool.Pool) *PGAdvertsRepo {
	return &PGAdvertsRepo{dbPool: pool}
}

func (r PGAdvertsRepo) GetAdvert(ctx context.Context, id domain.AdvertIdentity) (domain.Advert, error) {
	sql, args, err := psql.
		Select(&AdvertRecord{}).
		From(advertsTable).
		Where(goqu.Ex{"region": id.Region, "id": id.ID, "deleted_at": nil}).
		ToSQL()
	if err != nil {
		return domain.Advert{}, fmt.Errorf("cannot create sql: %w", err)
	}

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return domain.Advert{}, fmt.Errorf("cannot acquire connection: %w", err)
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return domain.Advert{}, fmt.Errorf("cannot query row: %w", err)
	}

	advertRecord, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[AdvertRecord])
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Advert{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Advert{}, fmt.Errorf("cannot fetch row: %w", err)
	}

	return domain.Advert{
		AdvertIdentity: advertRecord.AdvertIdentity,
		Title:          advertRecord.Title,
		Description:    advertRecord.Description,
		Price:          advertRecord.Price,
		PubDate:        advertRecord.PubDate,
	}, nil
}

func (r PGAdvertsRepo) SearchAdverts(ctx context.Context, params domain.AdvertSearchParams) (domain.AdvertSearchResult, error) {
	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return domain.AdvertSearchResult{}, fmt.Errorf("cannot acquire connection: %w", err)
	}
	defer conn.Release()

	sql, args, err := psql.
		Select(
			"id", "region", "title", "description", "price", "pub_date",
			goqu.L("COUNT(*) OVER()").As("total_count"),
		).
		From(advertsTable).
		Where(goqu.Ex{"region": params.Region, "deleted_at": nil}).
		Order(goqu.C("pub_date").Desc()).
		Limit(uint(params.PageSize)).
		Offset(uint((params.PageNum - 1) * params.PageSize)).
		ToSQL()
	if err != nil {
		return domain.AdvertSearchResult{}, fmt.Errorf("cannot create sql: %w", err)
	}

	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return domain.AdvertSearchResult{}, fmt.Errorf("cannot query rows: %w", err)
	}
	defer rows.Close()

	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[searchRow])
	if err != nil {
		return domain.AdvertSearchResult{}, fmt.Errorf("cannot collect rows: %w", err)
	}

	adverts := make([]domain.Advert, len(records))
	var total int
	for i, rec := range records {
		adverts[i] = domain.Advert{
			AdvertIdentity: rec.AdvertIdentity,
			Title:          rec.Title,
			Description:    rec.Description,
			Price:          rec.Price,
			PubDate:        rec.PubDate,
		}
		total = rec.TotalCount
	}

	return domain.AdvertSearchResult{
		AdvertSearchParams: params,
		Adverts:            adverts,
		AdvertsTotal:       total,
	}, nil
}

func (r PGAdvertsRepo) UpsertAdvert(ctx context.Context, advert domain.Advert) (bool, error) {
	record := AdvertRecord{
		AdvertIdentity: advert.AdvertIdentity,
		Title:          advert.Title,
		Description:    advert.Description,
		Price:          advert.Price,
		PubDate:        advert.PubDate,
	}

	sql, args, err := psql.
		Insert(advertsTable).
		Rows(&record).
		OnConflict(
			goqu.DoUpdate(
				"region, id",
				goqu.Record{
					"title":       advert.Title,
					"description": advert.Description,
					"price":       advert.Price,
					"pub_date":    advert.PubDate,
					"version":     goqu.L(advertsTable + ".version + 1"),
				},
			),
		).
		Returning(goqu.L("(xmax = 0)").As("inserted")).
		ToSQL()
	if err != nil {
		return false, fmt.Errorf("cannot create sql: %w", err)
	}

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return false, fmt.Errorf("cannot acquire connection: %w", err)
	}
	defer conn.Release()

	var inserted bool
	err = conn.QueryRow(ctx, sql, args...).Scan(&inserted)
	if err != nil {
		return false, fmt.Errorf("cannot upsert row: %w", err)
	}

	return inserted, nil
}

func (r PGAdvertsRepo) DeleteAdvert(ctx context.Context, id domain.AdvertIdentity) error {
	sql, args, err := psql.
		Update(advertsTable).
		Set(goqu.Record{"deleted_at": time.Now()}).
		Where(goqu.Ex{"region": id.Region, "id": id.ID}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("cannot create sql: %w", err)
	}

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("cannot acquire connection: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("cannot update row: %w", err)
	}
	return nil
}

func (r PGAdvertsRepo) CleanupDeletedAdverts(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	sql, args, err := psql.
		Delete(advertsTable).
		Where(goqu.C("deleted_at").Lt(cutoff)).
		ToSQL()
	if err != nil {
		return fmt.Errorf("cannot create sql: %w", err)
	}

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("cannot acquire connection: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("cannot delete rows: %w", err)
	}
	return nil
}
