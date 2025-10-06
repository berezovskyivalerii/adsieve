package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/berezovskyivalerii/adsieve/internal/adapter/postgres"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
)

func newRepoTokens(t *testing.T) (*postgres.TokensRepo, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	return postgres.NewTokensRepo(db), mock, func() {
		require.NoError(t, mock.ExpectationsWereMet())
		_ = db.Close()
	}
}

// Create
func TestTokensRepo_Create(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo, mock, done := newRepoTokens(t)
		defer done()

		rs := entity.RefreshToken{
			UserID:       7,
			RefreshToken: "abc123",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}

		mock.ExpectExec(`INSERT INTO refresh_tokens`).
			WithArgs(rs.UserID, rs.RefreshToken, rs.ExpiresAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), rs)
		require.NoError(t, err)
	})

	t.Run("db error bubbles up", func(t *testing.T) {
		repo, mock, done := newRepoTokens(t)
		defer done()

		rs := entity.RefreshToken{
			UserID:       7,
			RefreshToken: "abc123",
			ExpiresAt:    time.Now(),
		}

		mock.ExpectExec(`INSERT INTO refresh_tokens`).
			WithArgs(rs.UserID, rs.RefreshToken, rs.ExpiresAt).
			WillReturnError(errors.New("boom"))

		err := repo.Create(context.Background(), rs)
		require.Error(t, err)
	})
}

// Get (single-use → удаляем все токены этого user)
func TestTokensRepo_Get(t *testing.T) {
	const raw = "toktoktok"
	const uid = int64(42)
	const rid = int64(1)
	expires := time.Now().Add(30 * time.Minute)

	happyRows := sqlmock.NewRows(
		[]string{"token_id", "user_id", "token", "expires_at"},
	).AddRow(rid, uid, raw, expires)

	t.Run("happy path", func(t *testing.T) {
		repo, mock, done := newRepoTokens(t)
		defer done()

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT token_id, user_id, refresh_token, expires_at FROM\s+refresh_tokens`).
			WithArgs(raw).
			WillReturnRows(happyRows)
		mock.ExpectExec(`DELETE FROM refresh_tokens WHERE user_id =`).
			WithArgs(uid).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		rs, err := repo.Get(context.Background(), raw)
		require.NoError(t, err)
		require.Equal(t, rid, rs.TokenID)
		require.Equal(t, uid, rs.UserID)
		require.Equal(t, raw, rs.RefreshToken)
		require.WithinDuration(t, expires, rs.ExpiresAt, time.Second)
	})

	t.Run("token not found → sql.ErrNoRows (rollback)", func(t *testing.T) {
		repo, mock, done := newRepoTokens(t)
		defer done()

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT token_id, user_id, refresh_token, expires_at FROM\s+refresh_tokens`).
			WithArgs(raw).
			WillReturnError(sql.ErrNoRows)
		mock.ExpectRollback()

		_, err := repo.Get(context.Background(), raw)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("delete fails → rollback", func(t *testing.T) {
		repo, mock, done := newRepoTokens(t)
		defer done()

		rows := sqlmock.NewRows(
			[]string{"token_id", "user_id", "refresh_token", "expires_at"},
		).AddRow(rid, uid, raw, expires)

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT token_id, user_id, refresh_token, expires_at FROM\s+refresh_tokens`).
			WithArgs(raw).
			WillReturnRows(rows)
		mock.ExpectExec(`DELETE FROM refresh_tokens WHERE user_id =`).
			WithArgs(uid).
			WillReturnError(errors.New("delete failed"))
		mock.ExpectRollback()

		_, err := repo.Get(context.Background(), raw)
		require.Error(t, err)
	})

	t.Run("begin tx fails → error", func(t *testing.T) {
		repo, mock, done := newRepoTokens(t)
		defer done()

		mock.ExpectBegin().WillReturnError(errors.New("begin failed"))

		_, err := repo.Get(context.Background(), "whatever")
		require.Error(t, err)
	})

	t.Run("select returns non-no-rows error → rollback", func(t *testing.T) {
		repo, mock, done := newRepoTokens(t)
		defer done()

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT token_id, user_id, refresh_token, expires_at FROM\s+refresh_tokens`).
			WithArgs("toktoktok").
			WillReturnError(errors.New("select failed"))
		mock.ExpectRollback()

		_, err := repo.Get(context.Background(), "toktoktok")
		require.Error(t, err)
	})

}
