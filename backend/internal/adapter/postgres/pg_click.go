package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
	"github.com/lib/pq"
)

type ClicksRepo struct {
	db *sql.DB
}

func NewClicksRepo(db *sql.DB) *ClicksRepo { return &ClicksRepo{db: db}}

func (r *ClicksRepo) Click(ctx context.Context, clk entity.Click) (int64, error) {
	const q = `INSERT INTO clicks (click_id, ad_id, occurred_at)
			   VALUES ($1, $2, $3)
			   RETURNING id`
	
	var id int64
	if err := r.db.QueryRowContext(ctx, q, clk.ClickID, clk.AdID, clk.OccurredAt).Scan(&id); err != nil{
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return 0, errs.ErrDuplicateClick
		}
		return 0, err
	}
	return id, nil
}

// этот метод существует только в репозитории
func (r *ClicksRepo) ByClickID(ctx context.Context, id string) (entity.Click, error){
	const q = `SELECT id, click_id, ad_id, occurred_at
	           FROM   clicks
	           WHERE  click_id = $1`

	var c entity.Click
	err := r.db.QueryRowContext(ctx, q, id).
		Scan(&c.ID, &c.ClickID, &c.AdID, &c.OccurredAt)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.Click{}, errs.ErrClickNotFound
	}
	return c, err
}