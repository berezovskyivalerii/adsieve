package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/berezovskyivalerii/adsieve/internal/adapter/postgres"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func newRepoClicks(t *testing.T) (*postgres.ClicksRepo, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	return postgres.NewClicksRepo(db), mock, func() { db.Close() }
}

func TestClick(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo, mock, done := newRepoClicks(t)
		defer done()

		clk := entity.Click{ClickID: "clickID", AdID: 1, OccurredAt: time.Now()}

		mock.ExpectQuery(`INSERT INTO clicks`).
			WithArgs(clk.ClickID, clk.AdID, clk.OccurredAt).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

		id, err := repo.Click(context.Background(), clk)
		require.NoError(t, err)
		require.Equal(t, int64(1), id)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate click", func(t *testing.T) {
		repo, mock, done := newRepoClicks(t)
		defer done()

		clk := entity.Click{ClickID: "clickID", AdID: 1, OccurredAt: time.Now()}

		mock.ExpectQuery(`INSERT INTO clicks`).
			WithArgs(clk.ClickID, clk.AdID, clk.OccurredAt).
			WillReturnError(&pq.Error{Code: "23505"})

		_, err := repo.Click(context.Background(), clk)
		require.ErrorIs(t, err, errs.ErrDuplicateClick)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestById(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo, mock, done := newRepoClicks(t)
		defer done()

		const (
			id      string = "abc"
			clickID       = "123id123"
			adID    int64 = 3
		)
		occurredAt := time.Now()

		mock.ExpectQuery(`SELECT id, click_id, ad_id, occurred_at FROM\s+clicks\s+WHERE`).
			WithArgs(id).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "click_id", "ad_id", "occurred_at"}).
					AddRow(id, clickID, adID, occurredAt),
			)

		clk, err := repo.ByID(context.Background(), id)
		require.NoError(t, err)
		require.Equal(t, id, clk.ID)
		require.Equal(t, clickID, clk.ClickID)
		require.Equal(t, adID, clk.AdID)
		require.WithinDuration(t, occurredAt, clk.OccurredAt, time.Second)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found → ErrClickNotFound", func(t *testing.T) {
		repo, mock, done := newRepoClicks(t)
		defer done()

		mock.ExpectQuery(`SELECT id, click_id, ad_id, occurred_at FROM\s+clicks\s+WHERE`).
			WithArgs("abc").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.ByID(context.Background(), "abc")
		require.ErrorIs(t, err, errs.ErrClickNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
