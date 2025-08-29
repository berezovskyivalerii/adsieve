package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
)

type ConversionRepo struct {
	db *sql.DB
}

func NewConversionRepo(db *sql.DB) *ConversionRepo {
	return &ConversionRepo{db: db}
}

// Делает INSERT в таблицу conversions
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

// Получение заказа по order_id, метод существует только в репозитории
func (r *ConversionRepo) GetByOrderID(ctx context.Context, orderID int64) (entity.Conversion, error) {
	const q = `
		SELECT conversion_id, ad_id, converted_at, revenue, order_id, click_ref
		FROM conversions
		WHERE order_id = $1`
	var conv entity.Conversion
	err := r.db.QueryRowContext(ctx, q, orderID).Scan(
		&conv.ConversionID,
		&conv.AdID,
		&conv.ConvertedAt,
		&conv.Revenue,
		&conv.OrderID,
		&conv.ClickRef,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Conversion{}, errs.ErrConversionNotFound
	}
	return conv, err
}
