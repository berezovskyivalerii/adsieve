package postgres

import (
	"context"
	"database/sql"
)

type UserAdsRepo struct{ db *sql.DB }

func NewUserAdsRepo(db *sql.DB) *UserAdsRepo {
	return &UserAdsRepo{db: db}
}

const ensureSQL = `
	INSERT INTO user_ads (user_id, ad_id)
	VALUES ($1, $2)
	ON CONFLICT DO NOTHING;
`

func (r *UserAdsRepo) Ensure(ctx context.Context, userID, adID int64) error {
	_, err := r.db.ExecContext(ctx, ensureSQL, userID, adID)
	return err
}

const listSQL = `
	SELECT ad_id
	FROM user_ads
	WHERE user_id = $1
	ORDER BY ad_id
`

func (r *UserAdsRepo) IDsByUser(ctx context.Context, userID int64) ([]int64, error){
	rows, err := r.db.QueryContext(ctx, listSQL, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    ids := make([]int64, 0, 8)
    for rows.Next() {
        var id int64
        if err := rows.Scan(&id); err != nil {
            return nil, err
        }
        ids = append(ids, id)
    }
    return ids, rows.Err()
}
