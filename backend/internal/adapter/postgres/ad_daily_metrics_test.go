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

	mock.ExpectExec(`INSERT\s+INTO\s+ad_daily_metrics`).
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

func TestMetricsRepo_Upsert_DBError(t *testing.T) {
	repo, mock, done := newMetricsRepo(t)
	defer done()

	m := entity.AdDailyMetric{
		AdID:        99,
		MetricDate:  time.Date(2025, 7, 23, 0, 0, 0, 0, time.UTC),
		Clicks:      1,
		Conversions: 0,
		Revenue:     decimal.NewFromFloat(0),
		Spend:       decimal.NewFromFloat(0),
	}

	mock.ExpectExec(`INSERT\s+INTO\s+ad_daily_metrics`).
		WithArgs(m.AdID, m.MetricDate, m.Clicks, m.Conversions, m.Revenue, m.Spend).
		WillReturnError(sqlmock.ErrCancelled)

	err := repo.Upsert(context.Background(), m)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepo_List(t *testing.T) {
	repo, mock, done := newMetricsRepo(t)
	defer done()

	from := time.Date(2025, 7, 20, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 7, 24, 0, 0, 0, 0, time.UTC)
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
		"name",
		"status",
	}).AddRow(
		row.AdID,
		row.MetricDate,
		row.Clicks,
		row.Conversions,
		row.Revenue,
		row.Spend,
		"Test Ad",
		"active",
	)

	mock.ExpectQuery(`SELECT\s+m\.ad_id,\s+m\.metric_date,`).
		WithArgs(from, to, sqlmock.AnyArg()).
		WillReturnRows(rows)

	got, err := repo.List(context.Background(), adIDs, from, to)
	require.NoError(t, err)
	require.Len(t, got, 1)

	require.Equal(t, row.AdID, got[0].AdID)
	require.Equal(t, row.MetricDate, got[0].MetricDate)
	require.Equal(t, row.Clicks, got[0].Clicks)
	require.Equal(t, row.Conversions, got[0].Conversions)
	require.True(t, got[0].Revenue.Equal(row.Revenue), "revenue not equal")
	require.True(t, got[0].Spend.Equal(row.Spend), "spend not equal")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepo_List_DBError(t *testing.T) {
	repo, mock, done := newMetricsRepo(t)
	defer done()

	from := time.Date(2025, 7, 20, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 7, 24, 0, 0, 0, 0, time.UTC)
	adIDs := []int64{87}

	mock.ExpectQuery(`SELECT\s+m\.ad_id,\s+m\.metric_date,`).
		WithArgs(from, to, sqlmock.AnyArg()).
		WillReturnError(sqlmock.ErrCancelled)

	_, err := repo.List(context.Background(), adIDs, from, to)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// Пустой результат (ни одной строки)
func TestMetricsRepo_List_EmptyResult(t *testing.T) {
	repo, mock, done := newMetricsRepo(t)
	defer done()

	from := time.Date(2025, 7, 21, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 7, 22, 0, 0, 0, 0, time.UTC)
	adIDs := []int64{999}

	rows := sqlmock.NewRows([]string{
		"ad_id", "metric_date", "clicks", "conversions", "revenue", "spend", "name", "status",
	})

	mock.ExpectQuery(`SELECT\s+m\.ad_id,\s+m\.metric_date,`).
		WithArgs(from, to, sqlmock.AnyArg()).
		WillReturnRows(rows)

	got, err := repo.List(context.Background(), adIDs, from, to)
	require.NoError(t, err)
	require.Len(t, got, 0)
	require.NoError(t, mock.ExpectationsWereMet())
}

// Несколько строк и проверка порядка (ORDER BY m.ad_id, m.metric_date)
func TestMetricsRepo_List_MultipleRows(t *testing.T) {
	repo, mock, done := newMetricsRepo(t)
	defer done()

	from := time.Date(2025, 7, 20, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 7, 23, 0, 0, 0, 0, time.UTC)
	adIDs := []int64{10, 11}

	rows := sqlmock.NewRows([]string{
		"ad_id", "metric_date", "clicks", "conversions", "revenue", "spend", "name", "status",
	}).
		AddRow(int64(10), from, 3, 1, decimal.NewFromFloat(30), decimal.NewFromFloat(10), "Ad A", "active").
		AddRow(int64(10), from.Add(24*time.Hour), 4, 2, decimal.NewFromFloat(40), decimal.NewFromFloat(12), "Ad A", "active").
		AddRow(int64(11), from, 5, 1, decimal.NewFromFloat(50), decimal.NewFromFloat(20), "Ad B", "paused")

	mock.ExpectQuery(`SELECT\s+m\.ad_id,\s+m\.metric_date,`).
		WithArgs(from, to, sqlmock.AnyArg()).
		WillReturnRows(rows)

	got, err := repo.List(context.Background(), adIDs, from, to)
	require.NoError(t, err)
	require.Len(t, got, 3)

	require.Equal(t, int64(10), got[0].AdID)
	require.Equal(t, from, got[0].MetricDate)

	require.Equal(t, int64(10), got[1].AdID)
	require.Equal(t, from.Add(24*time.Hour), got[1].MetricDate)

	require.Equal(t, int64(11), got[2].AdID)
	require.Equal(t, from, got[2].MetricDate)

	require.NoError(t, mock.ExpectationsWereMet())
}

// Ошибка Scan на второй строке (подсовываем строку в clicks)
func TestMetricsRepo_List_ScanError_OnSecondRow(t *testing.T) {
	repo, mock, done := newMetricsRepo(t)
	defer done()

	from := time.Date(2025, 7, 20, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 7, 24, 0, 0, 0, 0, time.UTC)
	adIDs := []int64{87}

	rows := sqlmock.NewRows([]string{
		"ad_id", "metric_date", "clicks", "conversions", "revenue", "spend", "name", "status",
	}).
		AddRow(int64(87), from, 1, 0, decimal.NewFromFloat(10), decimal.NewFromFloat(5), "Ad X", "active").
		// ВОТ ТУТ: "clicks" как строка → Scan в int упадёт
		AddRow(int64(87), from.Add(24*time.Hour), "oops", 1, decimal.NewFromFloat(20), decimal.NewFromFloat(6), "Ad X", "active")

	mock.ExpectQuery(`SELECT\s+m\.ad_id,\s+m\.metric_date,`).
		WithArgs(from, to, sqlmock.AnyArg()).
		WillReturnRows(rows)

	_, err := repo.List(context.Background(), adIDs, from, to)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// revenue/spend приходят строками — проверяем Scanner у decimal.Decimal
func TestMetricsRepo_List_RevenueSpendAsString(t *testing.T) {
	repo, mock, done := newMetricsRepo(t)
	defer done()

	from := time.Date(2025, 7, 22, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 7, 23, 0, 0, 0, 0, time.UTC)
	adIDs := []int64{101}

	rows := sqlmock.NewRows([]string{
		"ad_id", "metric_date", "clicks", "conversions", "revenue", "spend", "name", "status",
	}).
		AddRow(int64(101), from, 7, 3, "123.45", "67.89", "Ad S", "active")

	mock.ExpectQuery(`SELECT\s+m\.ad_id,\s+m\.metric_date,`).
		WithArgs(from, to, sqlmock.AnyArg()).
		WillReturnRows(rows)

	got, err := repo.List(context.Background(), adIDs, from, to)
	require.NoError(t, err)
	require.Len(t, got, 1)

	require.Equal(t, int64(101), got[0].AdID)
	require.Equal(t, from, got[0].MetricDate)
	require.Equal(t, 7, got[0].Clicks)          // Clicks — int в сущности
	require.Equal(t, 3, got[0].Conversions)     // Conversions — int в сущности

	wantRev, _ := decimal.NewFromString("123.45")
	wantSpend, _ := decimal.NewFromString("67.89")
	require.True(t, got[0].Revenue.Equal(wantRev))
	require.True(t, got[0].Spend.Equal(wantSpend))

	require.NoError(t, mock.ExpectationsWereMet())
}
