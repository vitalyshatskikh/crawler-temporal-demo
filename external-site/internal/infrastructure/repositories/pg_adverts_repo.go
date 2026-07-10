package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vitalyshatskikh/crawler-temporal-demo/external-site/internal/domain"
)

var (
	_ domain.AdvertsRepository = (*PGAdvertsRepo)(nil)

	advertsTable = "adverts"
)

type PGAdvertsRepo struct {
	dbPool *pgxpool.Pool
}

func NewPGAdvertsRepo(pool *pgxpool.Pool) *PGAdvertsRepo {
	return &PGAdvertsRepo{dbPool: pool}
}

func (r PGAdvertsRepo) GetAdvert(ctx context.Context, region, id string) (domain.Advert, error) {
	advert := domain.Advert{}

	sql, args, err := goqu.From(advertsTable).
		Select("id", "region", "title", "description", "price", "pub_date").
		Where(goqu.Ex{"region": region, "id": id}).
		ToSQL()
	if err != nil {
		return advert, fmt.Errorf("cannot create sql: %w", err)
	}

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return advert, fmt.Errorf("cannot acquire connection: %w", err)
	}
	defer conn.Release()

	err = conn.QueryRow(ctx, sql, args...).
		Scan(&advert.ID, &advert.Region, &advert.Title, &advert.Description, &advert.Price, &advert.PubDate)
	if errors.Is(err, pgx.ErrNoRows) {
		return advert, domain.ErrNotFound
	}
	if err != nil {
		return advert, fmt.Errorf("cannot fetch row: %w", err)
	}

	return advert, nil
}

func (r PGAdvertsRepo) SearchAdverts(ctx context.Context, params domain.AdvertSearchParams) ([]domain.Advert, error) {
	var adverts []domain.Advert

	if params.PageSize <= 0 {
		return adverts, fmt.Errorf("%w: invalid page size %d", domain.ErrValidation, params.PageSize)
	}
	if params.PageNum <= 0 {
		return adverts, fmt.Errorf("%w: invalid page num %d", domain.ErrValidation, params.PageNum)
	}

	sql, args, err := goqu.From(advertsTable).
		Select("id", "region", "title", "description", "price", "pub_date").
		Where(goqu.Ex{"region": params.Region}).
		Limit(uint(params.PageSize)).
		Offset(uint((params.PageNum - 1) * params.PageSize)).
		ToSQL()
	if err != nil {
		return adverts, fmt.Errorf("cannot create sql: %w", err)
	}

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return adverts, fmt.Errorf("cannot acquire connection: %w", err)
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return adverts, fmt.Errorf("cannot query rows: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var adv domain.Advert
		err = rows.Scan(&adv)
		if err != nil {
			return adverts, fmt.Errorf("cannot scan row: %w", err)
		}
		adverts = append(adverts, adv)
	}
	if rows.Err() != nil {
		return adverts, fmt.Errorf("cannot fetch rows: %w", rows.Err())
	}

	return adverts, nil
}

func (r PGAdvertsRepo) UpsertAdvert(ctx context.Context, advert domain.Advert) (bool, error) {
	sql, args, err := goqu.Insert(advertsTable).
		Rows(advert).
		OnConflict(
			goqu.DoUpdate(
				"region, id",
				goqu.Record{
					"title":       advert.Title,
					"description": advert.Description,
					"price":       advert.Price,
					"pub_date":    advert.PubDate,
					"version":     goqu.L("version + 1"),
				},
			),
		).
		Returning("version").
		ToSQL()
	if err != nil {
		return false, fmt.Errorf("cannot create sql: %w", err)
	}

	conn, err := r.dbPool.Acquire(ctx)
	if err != nil {
		return false, fmt.Errorf("cannot acquire connection: %w", err)
	}
	defer conn.Release()

	var version int
	err = conn.QueryRow(ctx, sql, args...).Scan(&version)
	if err != nil {
		return false, fmt.Errorf("cannot insert row: %w", err)
	}

	created := version == 0
	return created, nil
}

func (r PGAdvertsRepo) DeleteAdvert(ctx context.Context, region, id string) error {
	sql, args, err := goqu.From(advertsTable).Delete().Where(goqu.Ex{"region": region, "id": id}).ToSQL()
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
		return fmt.Errorf("cannot delete row: %w", err)
	}
	return nil
}
