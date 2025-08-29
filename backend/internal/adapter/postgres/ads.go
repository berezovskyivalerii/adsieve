package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/lib/pq"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
)

type AdsRepo struct {
	db *sql.DB
}

func NewAdsRepo(db *sql.DB) *AdsRepo { return &AdsRepo{db: db} }

// ListByUser возвращает объявления пользователя с учётом фильтров/пагинации.
// Возвращает: items []entity.Ad и total (общее кол-во с учётом фильтров, без пагинации).
func (r *AdsRepo) ListByUser(ctx context.Context, userID int64, f entity.AdsFilter) ([]entity.Ad, int, error) {
	orderBy := orderClause(f.Sort)

	const baseFrom = `
		FROM ads a
		JOIN ad_accounts aa ON aa.account_id = a.account_id
	`
	conds := []string{"aa.user_id = $1"}
	args := []any{userID}
	next := 2

	if f.Status != nil && *f.Status != "" {
		conds = append(conds, "a.status = $"+itoa(next))
		args = append(args, *f.Status)
		next++
	}
	if f.Platform != nil && *f.Platform != "" {
		conds = append(conds, "a.platform = $"+itoa(next))
		args = append(args, *f.Platform)
		next++
	}
	if f.Query != nil && *f.Query != "" {
		conds = append(conds, "a.name ILIKE $"+itoa(next))
		args = append(args, "%"+*f.Query+"%")
		next++
	}
	if len(f.AdIDs) > 0 {
		conds = append(conds, "a.ad_id = ANY($"+itoa(next)+")")
		args = append(args, pq.Array(f.AdIDs))
		next++
	}

	where := "WHERE " + strings.Join(conds, " AND ")

	// ----- total -----
	countSQL := "SELECT COUNT(*) " + baseFrom + " " + where
	var total int
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []entity.Ad{}, 0, nil
	}

	// ----- items -----
	itemsSQL := `
		SELECT a.ad_id, a.account_id, a.name, a.status, a.platform
	` + baseFrom + " " + where + `
		` + orderBy + `
		LIMIT $` + itoa(next) + ` OFFSET $` + itoa(next+1)

	argsWithPage := append(append([]any{}, args...), f.Limit, f.Offset)

	rows, err := r.db.QueryContext(ctx, itemsSQL, argsWithPage...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]entity.Ad, 0, min(total, f.Limit))
	for rows.Next() {
		var a entity.Ad
		if err := rows.Scan(&a.AdID, &a.AccountID, &a.Name, &a.Status, &a.Platform); err != nil {
			return nil, 0, err
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func itoa(v int) string { return strconv.FormatInt(int64(v), 10) }

func orderClause(sort string) string {
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "-name":
		return "ORDER BY a.name DESC"
	case "created_at":
		return "ORDER BY a.created_at ASC"
	case "-created_at":
		return "ORDER BY a.created_at DESC"
	case "name":
		fallthrough
	default:	
		return "ORDER BY a.name ASC"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
