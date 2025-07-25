package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/berezovskyivalerii/adsieve/internal/adapter/postgres"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
)

func newMetricsRepo(t *testing.T) (*postgres.MetricsRepo, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	return postgres.NewMetricsRepo(db), mock, func() { _ = db.Close() }
}

func TestMetricsRepo_Upsert(t *testing.T) {
	repo, mock, done := newMetricsRepo(t)
	defer done()

	m := entity.AdDailyMetric{
		AdID:        87,
		MetricDate:  time.Date(2025, 7, 24, 0, 0, 0, 0, time.UTC),
		Clicks:      10,
		Conversions: 2,
		Revenue:     decimal.NewFromFloat(199.90),
		Spend:       decimal.NewFromFloat(40.00),
	}

	mock.ExpectExec(`INSERT INTO ad_daily_metrics`).
		WithArgs(
			m.AdID,
			m.MetricDate,
			m.Clicks,
			m.Conversions,
			m.Revenue,
			m.Spend,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Upsert(context.Background(), m)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepo_List(t *testing.T) {
	repo, mock, done := newMetricsRepo(t)
	defer done()

	from := time.Date(2025, 7, 20, 0, 0, 0, 0, time.UTC)
	to   := time.Date(2025, 7, 24, 0, 0, 0, 0, time.UTC)
	adIDs := []int64{87}

	row := entity.AdDailyMetric{
		AdID:        87,
		MetricDate:  from,
		Clicks:      5,
		Conversions: 1,
		Revenue:     decimal.NewFromFloat(50),
		Spend:       decimal.NewFromFloat(25),
	}

	rows := sqlmock.NewRows([]string{
		"ad_id",
		"metric_date",
		"clicks",
		"conversions",
		"revenue",
		"spend",
	}).
		AddRow(row.AdID, row.MetricDate, row.Clicks, row.Conversions, row.Revenue, row.Spend)

	mock.ExpectQuery(`SELECT\s+ad_id`).
		WithArgs(from, to, sqlmock.AnyArg()).
		WillReturnRows(rows)

	got, err := repo.List(context.Background(), adIDs, from, to)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, row.AdID, got[0].AdID)
	require.Equal(t, row.MetricDate, got[0].MetricDate)
	require.Equal(t, row.Clicks, got[0].Clicks)
	require.NoError(t, mock.ExpectationsWereMet())
}
