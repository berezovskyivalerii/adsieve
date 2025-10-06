package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/berezovskyivalerii/adsieve/internal/adapter/postgres"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
)

func newRepoClicks(t *testing.T) (*postgres.ClicksRepo, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	return postgres.NewClicksRepo(db), mock, func() { _ = db.Close() }
}

func TestClick(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		repo, mock, done := newRepoClicks(t)
		defer done()

		clk := entity.Click{
			ClickID:   "abc_12345",
			AdID:      1,
			ClickedAt: time.Now(),
			ClickRef:  uuid.UUID{},
		}

		// INSERT ... RETURNING id — возвращаем одну колонку "id"
		mock.ExpectQuery(`INSERT\s+INTO\s+clicks`).
			WithArgs(clk.ClickID, clk.AdID, clk.ClickedAt, clk.ClickRef).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(12345)))

		gotID, err := repo.Click(context.Background(), clk)
		require.NoError(t, err)
		require.Equal(t, int64(12345), gotID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate click", func(t *testing.T) {
		repo, mock, done := newRepoClicks(t)
		defer done()

		clk := entity.Click{
			ClickID:   "click_12345",
			AdID:      1,
			ClickedAt: time.Now(),
			ClickRef:  uuid.New(),
		}

		// unique_violation
		pgErr := &pq.Error{Code: "23505"}

		mock.ExpectQuery(`INSERT\s+INTO\s+clicks`).
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
			rowID   int64 = 1
			clickID       = "click_12345"
			adID    int64 = 17
		)
		clickedAt := time.Now()
		clickRef := uuid.New() // используем тот же в моке и в проверке

		mock.ExpectQuery(`SELECT\s+id,\s+click_id,\s+ad_id,\s+clicked_at,\s+click_ref\s+FROM\s+clicks\s+WHERE\s+click_id\s*=\s*\$1`).
			WithArgs(clickID).
			// Важно: значения должны быть совместимы с database/sql.
			// Для UUID безопаснее отдавать строку (или []byte), Scanner из google/uuid это понимает.
			WillReturnRows(sqlmock.NewRows([]string{"id", "click_id", "ad_id", "clicked_at", "click_ref"}).
				AddRow(rowID, clickID, adID, clickedAt, clickRef.String()))

		clk, err := repo.ByClickID(context.Background(), clickID)
		require.NoError(t, err)
		require.Equal(t, rowID, clk.ID)
		require.Equal(t, clickID, clk.ClickID)
		require.Equal(t, adID, clk.AdID)
		require.WithinDuration(t, clickedAt, clk.ClickedAt, time.Second)
		require.Equal(t, clickRef, clk.ClickRef) // сравниваем с тем, что положили в мок

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock, done := newRepoClicks(t)
		defer done()

		mock.ExpectQuery(`SELECT\s+id,\s+click_id,\s+ad_id,\s+clicked_at,\s+click_ref\s+FROM\s+clicks`).
			WithArgs("absent_click").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.ByClickID(context.Background(), "absent_click")
		require.ErrorIs(t, err, errs.ErrClickNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
