package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
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

		mock.ExpectQuery(`INSERT INTO conversions`).
			WithArgs(conv.AdID, conv.ConvertedAt, conv.Revenue, conv.OrderID, conv.ClickRef).
			WillReturnRows(sqlmock.NewRows([]string{"conversion_id"}).AddRow(int64(1)))

		id, err := repo.Create(context.Background(), conv)
		require.NoError(t, err)
		require.Equal(t, int64(1), id)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate order (23505)", func(t *testing.T) {
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

		mock.ExpectQuery(`INSERT INTO conversions`).
			WithArgs(conv.AdID, sqlmock.AnyArg(), conv.Revenue, conv.OrderID, sqlmock.AnyArg()).
			WillReturnError(pgErr)

		id, err := repo.Create(context.Background(), conv)
		assert.ErrorIs(t, err, errs.ErrConversionExists)
		assert.Zero(t, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
