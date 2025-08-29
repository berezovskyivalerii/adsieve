package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
)

type ClicksRepo struct {
	db *sql.DB
}

func NewClicksRepo(db *sql.DB) *ClicksRepo { return &ClicksRepo{db: db} }

// Делает INSERT в таблицу clicks
func (r *ClicksRepo) Click(ctx context.Context, clk entity.Click) (int64, error) {
	const q = `INSERT INTO clicks (click_id, ad_id, clicked_at, click_ref)
			   VALUES ($1, $2, $3, $4)
			   RETURNING id`

	var ID int64
	if err := r.db.QueryRowContext(ctx, q, clk.ClickID, clk.AdID, clk.ClickedAt, clk.ClickRef).Scan(&ID); err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return 0, errs.ErrDuplicateClick
		}
		return 0, err
	}
	return ID, nil
}

// этот метод существует только в репозитории
func (r *ClicksRepo) ByClickID(ctx context.Context, id string) (entity.Click, error) {
	const q = `SELECT id, click_id, ad_id, clicked_at, click_ref
	           FROM   clicks
	           WHERE  click_id = $1`

	var clk entity.Click
	err := r.db.QueryRowContext(ctx, q, id).
		Scan(&clk.ID, &clk.ClickID, &clk.AdID, &clk.ClickedAt, &clk.ClickRef)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.Click{}, errs.ErrClickNotFound
	}
	return clk, err
}
