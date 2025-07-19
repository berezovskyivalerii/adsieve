package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/berezovskyivalerii/adsieve/internal/adapter/postgres"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
)

func newRepoOrder(t *testing.T) (*postgres.OrderRepo, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	return postgres.NewOrderRepo(db), mock, func() { db.Close() }
}

func TestCreateOrder(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo, mock, done := newRepoOrder(t)
		defer done()

		order := entity.Order{ClickID: "123", OrderValue: decimal.NewFromInt(10), OccurredAt: time.Now()}

		mock.ExpectQuery("INSERT INTO orders").WithArgs(order.ClickID, order.OrderValue, order.OccurredAt).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

        id, err := repo.Create(context.Background(), order)
		require.NoError(t, err)
		require.Equal(t, int64(1), id)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate order error", func(t *testing.T) {
		repo, mock, done := newRepoOrder(t)
		defer done()

		order := entity.Order{ClickID: "123", OrderValue: decimal.NewFromInt(10), OccurredAt: time.Now()}

		dupErr := &pq.Error{Code: "23505"}

		mock.ExpectQuery("INSERT INTO orders").
			WithArgs(order.ClickID, order.OrderValue, sqlmock.AnyArg()).
			WillReturnError(dupErr)

		id, err := repo.Create(context.Background(), order)
		assert.ErrorIs(t, err, errs.ErrOrderExists)
		assert.Zero(t, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
