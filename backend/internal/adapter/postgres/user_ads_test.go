package postgres_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/berezovskyivalerii/adsieve/internal/adapter/postgres"
)

func newUserAdsRepo(t *testing.T) (*postgres.UserAdsRepo, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	return postgres.NewUserAdsRepo(db), mock, func() { _ = db.Close() }
}

func TestUserAdsRepo_Ensure_HappyPath(t *testing.T) {
	repo, mock, done := newUserAdsRepo(t)
	defer done()

	const userID, adID int64 = 42, 101

	mock.ExpectExec(`INSERT\s+INTO\s+user_ads`).
		WithArgs(userID, adID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Ensure(context.Background(), userID, adID)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserAdsRepo_Ensure_Idempotent_NoRowsAffected(t *testing.T) {
	// ON CONFLICT DO NOTHING -> RowsAffected = 0, и это ОК
	repo, mock, done := newUserAdsRepo(t)
	defer done()

	const userID, adID int64 = 42, 101

	mock.ExpectExec(`INSERT\s+INTO\s+user_ads`).
		WithArgs(userID, adID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Ensure(context.Background(), userID, adID)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserAdsRepo_Ensure_DBError(t *testing.T) {
	repo, mock, done := newUserAdsRepo(t)
	defer done()

	const userID, adID int64 = 7, 200

	mock.ExpectExec(`INSERT\s+INTO\s+user_ads`).
		WithArgs(userID, adID).
		WillReturnError(sqlmock.ErrCancelled)

	err := repo.Ensure(context.Background(), userID, adID)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserAdsRepo_IDsByUser_QueryError(t *testing.T) {
	repo, mock, done := newUserAdsRepo(t)
	defer done()

	const userID int64 = 99

	mock.ExpectQuery(`SELECT\s+ad_id\s+FROM\s+user_ads\s+WHERE\s+user_id\s*=\s*\$1`).
		WithArgs(userID).
		WillReturnError(sqlmock.ErrCancelled)

	ids, err := repo.IDsByUser(context.Background(), userID)
	require.Error(t, err)
	require.Nil(t, ids)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserAdsRepo_IDsByUser_Empty(t *testing.T) {
	repo, mock, done := newUserAdsRepo(t)
	defer done()

	const userID int64 = 1

	rows := sqlmock.NewRows([]string{"ad_id"}) // пусто
	mock.ExpectQuery(`SELECT\s+ad_id\s+FROM\s+user_ads\s+WHERE\s+user_id\s*=\s*\$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	ids, err := repo.IDsByUser(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, ids, 0)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserAdsRepo_IDsByUser_MultipleRows(t *testing.T) {
	repo, mock, done := newUserAdsRepo(t)
	defer done()

	const userID int64 = 5

	rows := sqlmock.NewRows([]string{"ad_id"}).
		AddRow(int64(10)).
		AddRow(int64(11)).
		AddRow(int64(12))

	mock.ExpectQuery(`SELECT\s+ad_id\s+FROM\s+user_ads\s+WHERE\s+user_id\s*=\s*\$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	ids, err := repo.IDsByUser(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, []int64{10, 11, 12}, ids)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserAdsRepo_IDsByUser_ScanError(t *testing.T) {
	repo, mock, done := newUserAdsRepo(t)
	defer done()

	const userID int64 = 77

	// Подсовываем строку вместо int64 -> Scan упадёт
	rows := sqlmock.NewRows([]string{"ad_id"}).
		AddRow("oops")

	mock.ExpectQuery(`SELECT\s+ad_id\s+FROM\s+user_ads\s+WHERE\s+user_id\s*=\s*\$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	ids, err := repo.IDsByUser(context.Background(), userID)
	require.Error(t, err)
	require.Nil(t, ids)
	require.NoError(t, mock.ExpectationsWereMet())
}
