package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/berezovskyivalerii/adsieve/internal/adapter/postgres"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
)

func newConversionRepo(t *testing.T) (*postgres.ConversionRepo, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	return postgres.NewConversionRepo(db), mock, func() { _ = db.Close() }
}

func TestConversionRepo_Create(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo, mock, done := newConversionRepo(t)
		defer done()

		now := time.Now().UTC()
		orderID := "A100"
		clickRef := uuid.New()

		conv := entity.Conversion{
			AdID:        12345,
			ConvertedAt: now,
			Revenue:     decimal.NewFromFloat(49.99),
			OrderID:     &orderID,
			ClickRef:    &clickRef,
		}

		mock.ExpectQuery(`INSERT\s+INTO\s+conversions`).
			// Для устойчивости к типам времени/decimal/uuid используем AnyArg
			WithArgs(conv.AdID, sqlmock.AnyArg(), sqlmock.AnyArg(), conv.OrderID, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"conversion_id"}).AddRow(int64(1)))

		id, err := repo.Create(context.Background(), conv)
		require.NoError(t, err)
		require.Equal(t, int64(1), id)
	})

	t.Run("duplicate (pq 23505) -> ErrConversionExists", func(t *testing.T) {
		repo, mock, done := newConversionRepo(t)
		defer done()

		orderID := "A100"
		conv := entity.Conversion{
			AdID:        12345,
			ConvertedAt: time.Now().UTC(),
			Revenue:     decimal.NewFromFloat(20),
			OrderID:     &orderID,
		}

		pgErr := &pq.Error{Code: "23505"} // unique_violation

		mock.ExpectQuery(`INSERT\s+INTO\s+conversions`).
			WithArgs(conv.AdID, sqlmock.AnyArg(), sqlmock.AnyArg(), conv.OrderID, sqlmock.AnyArg()).
			WillReturnError(pgErr)

		id, err := repo.Create(context.Background(), conv)
		require.ErrorIs(t, err, errs.ErrConversionExists)
		require.Zero(t, id)
	})

	t.Run("pq error (not 23505) -> returns original error", func(t *testing.T) {
		repo, mock, done := newConversionRepo(t)
		defer done()

		conv := entity.Conversion{
			AdID:        777,
			ConvertedAt: time.Now().UTC(),
			Revenue:     decimal.NewFromFloat(10),
		}

		// 23503 = foreign_key_violation (пример «другая» ошибка)
		pgErr := &pq.Error{Code: "23503"}

		mock.ExpectQuery(`INSERT\s+INTO\s+conversions`).
			WithArgs(conv.AdID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(pgErr)

		id, err := repo.Create(context.Background(), conv)
		require.Error(t, err)
		require.Zero(t, id)

		var got *pq.Error
		require.ErrorAs(t, err, &got)
		require.Equal(t, "23503", string(got.Code))
	})

	t.Run("generic error (non-pq) -> returns original error", func(t *testing.T) {
		repo, mock, done := newConversionRepo(t)
		defer done()

		conv := entity.Conversion{
			AdID:        888,
			ConvertedAt: time.Now().UTC(),
			Revenue:     decimal.NewFromFloat(5),
		}

		mock.ExpectQuery(`INSERT\s+INTO\s+conversions`).
			WithArgs(conv.AdID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("boom"))

		id, err := repo.Create(context.Background(), conv)
		require.Error(t, err)
		require.Contains(t, err.Error(), "boom")
		require.Zero(t, id)
	})
}

func TestConversionRepo_GetByOrderID_happy_path(t *testing.T) {
	repo, mock, done := newConversionRepo(t)
	defer done()

	const (
		orderID      int64 = 555
		conversionID int64 = 1
		adID         int64 = 101
	)
	convertedAt := time.Now().Add(-time.Minute)
	revenueStr := "123.45"
	clickRef := uuid.New().String()

	rows := sqlmock.NewRows([]string{
		"conversion_id", "ad_id", "converted_at", "revenue", "order_id", "click_ref",
	}).AddRow(
		conversionID, adID, convertedAt, revenueStr, "ORD-555", clickRef,
	)

	mock.ExpectQuery(`SELECT\s+conversion_id,\s+ad_id,\s+converted_at,\s+revenue,\s+order_id,\s+click_ref\s+FROM\s+conversions\s+WHERE\s+order_id\s*=\s*\$1`).
		WithArgs(orderID).
		WillReturnRows(rows)

	got, err := repo.GetByOrderID(context.Background(), orderID)
	require.NoError(t, err)

	require.Equal(t, conversionID, got.ConversionID)
	require.Equal(t, adID, got.AdID)
	require.WithinDuration(t, convertedAt, got.ConvertedAt, time.Second)

	wantRev, _ := decimal.NewFromString(revenueStr)
	require.True(t, got.Revenue.Equal(wantRev))

	parsedRef, _ := uuid.Parse(clickRef)
	require.NotNil(t, got.ClickRef)
	require.Equal(t, parsedRef, *got.ClickRef)
}

func TestConversionRepo_GetByOrderID_not_found(t *testing.T) {
	repo, mock, done := newConversionRepo(t)
	defer done()

	const orderID int64 = 404

	mock.ExpectQuery(`SELECT\s+conversion_id,\s+ad_id,\s+converted_at,\s+revenue,\s+order_id,\s+click_ref\s+FROM\s+conversions\s+WHERE\s+order_id\s*=\s*\$1`).
		WithArgs(orderID).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByOrderID(context.Background(), orderID)
	require.ErrorIs(t, err, errs.ErrConversionNotFound)
}

func TestConversionRepo_GetByOrderID_db_error(t *testing.T) {
	repo, mock, done := newConversionRepo(t)
	defer done()

	const orderID int64 = 777

	mock.ExpectQuery(`SELECT\s+conversion_id,\s+ad_id,\s+converted_at,\s+revenue,\s+order_id,\s+click_ref\s+FROM\s+conversions\s+WHERE\s+order_id\s*=\s*\$1`).
		WithArgs(orderID).
		WillReturnError(sqlmock.ErrCancelled)

	_, err := repo.GetByOrderID(context.Background(), orderID)
	require.Error(t, err)
}
