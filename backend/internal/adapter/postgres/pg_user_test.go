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

func newRepoUser(t *testing.T) (*postgres.UserRepo, sqlmock.Sqlmock, func()){
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	return postgres.NewUserRepo(db), mock, func() { db.Close() }
}

func TestCreateUser(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo, mock, done := newRepoUser(t)
		defer done()

		u := entity.User{Email: "me@example.com", PassHash: "hash123hash456"}

		mock.ExpectQuery(`INSERT INTO users`).
					WithArgs(u.Email, u.PassHash).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

		id, err := repo.CreateUser(context.Background(), u)
		require.NoError(t, err)
		require.Equal(t, int64(1), id)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate email", func(t *testing.T){
		repo, mock, done := newRepoUser(t)
		defer done()

		u := entity.User{Email: "me@example.com", PassHash: "hash123hash456"}

		mock.ExpectQuery(`INSERT INTO users`).
					WithArgs(u.Email, u.PassHash).
					WillReturnError(&pq.Error{Code: "23505"})
		
		_, err := repo.CreateUser(context.Background(), u)
		require.ErrorIs(t, err, errs.ErrEmailTaken)
		require.NoError(t, mock.ExpectationsWereMet())

	})
}

func TestByEmail(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo, mock, done := newRepoUser(t)
		defer done()

		const (
			id        = int64(42)
			email     = "alice@example.com"
			passHash  = "hash"
		)
		created := time.Now()

		mock.ExpectQuery(`SELECT .* FROM\s+users`).
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "pass_hash", "created_at"}).
				AddRow(id, email, passHash, created))

		u, err := repo.ByEmail(context.Background(), email)
		require.NoError(t, err)
		require.Equal(t, id, u.UserID)
		require.Equal(t, email, u.Email)
		require.Equal(t, passHash, u.PassHash)
		require.WithinDuration(t, created, u.RegisteredAt, time.Second)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no rows → ErrUserNotFound", func(t *testing.T) {
		repo, mock, done := newRepoUser(t)
		defer done()

		mock.ExpectQuery(`SELECT .* FROM\s+users`).
			WithArgs("missing@example.com").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.ByEmail(context.Background(), "missing@example.com")
		require.ErrorIs(t, err, errs.ErrUserNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}