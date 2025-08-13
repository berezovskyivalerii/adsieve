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
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func newRepoClicks(t *testing.T) (*postgres.ClicksRepo, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	return postgres.NewClicksRepo(db), mock, func() { db.Close() }
}

func TestClick(t *testing.T) {
	t.Parallel()
	t.Run("happy path", func(t *testing.T) {
		repo, mock, done := newRepoClicks(t)
		defer done()

		clk := entity.Click{ClickID: "abc_12345", AdID: 1, ClickedAt: time.Now(), ClickRef: uuid.UUID{}}

		mock.ExpectQuery(`INSERT INTO clicks`).
			WithArgs(clk.ClickID, clk.AdID, clk.ClickedAt, clk.ClickRef).
			WillReturnRows(sqlmock.NewRows([]string{"click_id"}).AddRow(int64(12345)))

		click_id, err := repo.Click(context.Background(), clk)
		require.NoError(t, err)
		require.Equal(t, int64(12345), click_id)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate click", func(t *testing.T) {
		repo, mock, done := newRepoClicks(t)
		defer done()

		clk := entity.Click{
			ClickID:   "click_12345", // строка
			AdID:      1,
			ClickedAt: time.Now(),
			ClickRef:  uuid.New(),
		}

		// код ошибки unique_violation (PRIMARY KEY или UNIQUE)
		pgErr := &pq.Error{Code: "23505"}

		mock.ExpectQuery(`INSERT INTO clicks`).
			WithArgs(clk.ClickID, clk.AdID, clk.ClickedAt, clk.ClickRef).
			WillReturnError(pgErr)

		_, err := repo.Click(context.Background(), clk)
		require.ErrorIs(t, err, errs.ErrDuplicateClick)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestById(t *testing.T) {
	t.Parallel()
	t.Run("happy path", func(t *testing.T) {
		repo, mock, done := newRepoClicks(t)
		defer done()

		const (
			rowID     int64     = 1
			clickID             = "click_12345"
			adID      int64     = 17
		)
		click_ref := uuid.UUID{}
		clickedAt := time.Now()

		mock.ExpectQuery(`SELECT id, click_id, ad_id, clicked_at, click_ref\s+FROM\s+clicks\s+WHERE\s+click_id\s*=\s*\$1`).
			WithArgs(clickID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "click_id", "ad_id", "clicked_at", "click_ref"}).
				AddRow(rowID, clickID, adID, clickedAt, click_ref))

		clk, err := repo.ByClickID(context.Background(), clickID)
		require.NoError(t, err)
		require.Equal(t, rowID, clk.ID)
		require.Equal(t, clickID, clk.ClickID)
		require.Equal(t, adID, clk.AdID)
		require.Equal(t, click_ref, clk.ClickRef)
		require.WithinDuration(t, clickedAt, clk.ClickedAt, time.Second)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock, done := newRepoClicks(t)
		defer done()

		mock.ExpectQuery(`SELECT id, click_id, ad_id, clicked_at, click_ref\s+FROM\s+clicks`).
			WithArgs("absent_click").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.ByClickID(context.Background(), "absent_click")
		require.ErrorIs(t, err, errs.ErrClickNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
