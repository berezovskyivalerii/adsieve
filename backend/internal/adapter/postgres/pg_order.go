package postgres

import (
	"context"
	"database/sql"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
	"github.com/lib/pq"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo{
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Create(ctx context.Context, order entity.Order) (int64, error) {
	const q = `INSERT INTO orders (click_id, order_value, occurred_at) VALUES ($1,$2,$3) RETURNING id`
	var id int64

	if err := r.db.QueryRowContext(ctx, q, order.ClickID, order.OrderValue, order.OccurredAt).Scan(&id); err != nil{
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505"{
			return 0, errs.ErrOrderExists
		}
		return 0, err
	}
	return id, nil
}