package postgres

import (
	"context"
	"database/sql"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
	"github.com/lib/pq"
)

type ConversionRepo struct {
	db *sql.DB
}

func NewConversionRepo(db *sql.DB) *ConversionRepo{
	return &ConversionRepo{db: db}
}

func (r *ConversionRepo) Create(ctx context.Context, conv entity.Conversion) (int64, error) {
	const q = `
		INSERT INTO conversions (
			ad_id,
			converted_at,
			revenue,
			order_id,
			click_ref
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING conversion_id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		q,
		conv.AdID,
		conv.ConvertedAt,
		conv.Revenue,
		conv.OrderID,  // допускает NULL
		conv.ClickRef, // допускает NULL
	).Scan(&id)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return 0, errs.ErrConversionExists
		}
		return 0, err
	}
	return id, nil
}