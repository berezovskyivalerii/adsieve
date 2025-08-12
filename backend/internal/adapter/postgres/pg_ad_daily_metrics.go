package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	"github.com/lib/pq"
)

type MetricsRepo struct {
	db *sql.DB
}

func NewMetricsRepo(db *sql.DB) *MetricsRepo { return &MetricsRepo{db: db} }

const upsertSQL = `
	INSERT INTO ad_daily_metrics (
		ad_id, day, clicks, conversations, revenue, spend
	) VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (ad_id, metric_date) DO UPDATE
	SET    clicks        = EXCLUDED.clicks,
		   conversations = EXCLUDED.conversations,
		   revenue       = EXCLUDED.revenue,
		   spend         = EXCLUDED.spend;
`

func (r *MetricsRepo) Upsert(ctx context.Context, m entity.AdDailyMetric) error {
	_, err := r.db.ExecContext(ctx, upsertSQL, m.AdID, m.MetricDate, m.Clicks, m.Conversions, m.Revenue, m.Spend)
	return err
}

const listMetricsSQL = `
	SELECT  m.ad_id,
			m.metric_date,
			m.clicks,
			m.conversions,
			m.revenue,
			m.spend,
			a.name,
			a.status
	FROM    ad_daily_metrics m
	JOIN    ads a  ON a.ad_id = m.ad_id
	WHERE   a.user_id = $currentUser
	AND   m.metric_date BETWEEN $1 AND $2
	AND   a.ad_id = ANY($3);
`

func (r *MetricsRepo) List(
	ctx context.Context,
	adIDs []int64,
	from, to time.Time,
) ([]entity.AdDailyMetric, error) {
	rows, err := r.db.QueryContext(ctx, listMetricsSQL, from, to, pq.Array(adIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []entity.AdDailyMetric
	for rows.Next() {
		var m entity.AdDailyMetric
		if err := rows.Scan(
			&m.AdID,
			&m.MetricDate,
			&m.Clicks,
			&m.Conversions,
			&m.Revenue,
			&m.Spend,
			&m.Name,
			&m.Status,
		); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	return res, rows.Err()
}
