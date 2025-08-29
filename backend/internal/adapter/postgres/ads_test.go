package postgres_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/berezovskyivalerii/adsieve/internal/adapter/postgres"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
)

func newAdsRepo(t *testing.T) (*postgres.AdsRepo, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	return postgres.NewAdsRepo(db), mock, func() { _ = db.Close() }
}

func TestAdsRepo_ListByUser_HappyWithFilters(t *testing.T) {
	repo, mock, done := newAdsRepo(t)
	defer done()

	userID := int64(42)
	status := "active"
	platform := "facebook"
	query := "sale"
	f := entity.AdsFilter{
		Status:   &status,
		Platform: &platform,
		Query:    &query,
		AdIDs:    []int64{87, 112},
		Limit:    2,
		Offset:   0,
		Sort:     "-created_at",
	}

	// COUNT(*)
	mock.ExpectQuery(`SELECT\s+COUNT\(\*\)\s+FROM\s+ads a\s+JOIN\s+ad_accounts aa\s+ON aa\.account_id = a\.account_id\s+WHERE\s+aa\.user_id = \$1\s+AND a\.status = \$2\s+AND a\.platform = \$3\s+AND a\.name ILIKE \$4\s+AND a\.ad_id = ANY\(\$5\)`).
		WithArgs(userID, status, platform, "%"+query+"%", pq.Array([]int64{87, 112})).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// SELECT items
	mock.ExpectQuery(`SELECT\s+a\.ad_id, a\.account_id, a\.name, a\.status, a\.platform\s+FROM\s+ads a\s+JOIN\s+ad_accounts aa\s+ON aa\.account_id = a\.account_id\s+WHERE\s+aa\.user_id = \$1\s+AND a\.status = \$2\s+AND a\.platform = \$3\s+AND a\.name ILIKE \$4\s+AND a\.ad_id = ANY\(\$5\)\s+ORDER BY a\.created_at DESC\s+LIMIT \$6 OFFSET \$7`).
		WithArgs(userID, status, platform, "%"+query+"%", pq.Array([]int64{87, 112}), 2, 0).
		WillReturnRows(sqlmock.NewRows([]string{"ad_id", "account_id", "name", "status", "platform"}).
			AddRow(int64(87), int64(1001), "Summer Sale Shoes", "active", "facebook").
			AddRow(int64(112), int64(1001), "Sale – Leads", "active", "facebook"))

	items, total, err := repo.ListByUser(context.Background(), userID, f)
	require.NoError(t, err)
	require.Equal(t, 2, total)
	require.Len(t, items, 2)
	require.Equal(t, int64(87), items[0].AdID)
	require.Equal(t, "Summer Sale Shoes", items[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdsRepo_ListByUser_Empty(t *testing.T) {
	repo, mock, done := newAdsRepo(t)
	defer done()

	userID := int64(42)
	f := entity.AdsFilter{Limit: 10, Offset: 0}

	// COUNT(*) → 0
	mock.ExpectQuery(`SELECT\s+COUNT\(\*\)\s+FROM\s+ads a\s+JOIN\s+ad_accounts aa\s+ON aa\.account_id = a\.account_id\s+WHERE\s+aa\.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	items, total, err := repo.ListByUser(context.Background(), userID, f)
	require.NoError(t, err)
	require.Equal(t, 0, total)
	require.Len(t, items, 0)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdsRepo_ListByUser_CountError(t *testing.T) {
	repo, mock, done := newAdsRepo(t)
	defer done()

	userID := int64(42)
	f := entity.AdsFilter{Limit: 10, Offset: 0}

	mock.ExpectQuery(`SELECT\s+COUNT\(\*\)\s+FROM\s+ads a\s+JOIN\s+ad_accounts aa\s+ON aa\.account_id = a\.account_id\s+WHERE\s+aa\.user_id = \$1`).
		WithArgs(userID).
		WillReturnError(sqlmock.ErrCancelled)

	_, _, err := repo.ListByUser(context.Background(), userID, f)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdsRepo_ListByUser_SelectError(t *testing.T) {
	repo, mock, done := newAdsRepo(t)
	defer done()

	userID := int64(42)
	f := entity.AdsFilter{Limit: 10, Offset: 0, Sort: "name"}

	// COUNT(*) → 2
	mock.ExpectQuery(`SELECT\s+COUNT\(\*\)\s+FROM\s+ads a\s+JOIN\s+ad_accounts aa\s+ON aa\.account_id = a\.account_id\s+WHERE\s+aa\.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// SELECT → ошибка
	mock.ExpectQuery(`SELECT\s+a\.ad_id, a\.account_id, a\.name, a\.status, a\.platform`).
		WithArgs(userID, 10, 0).
		WillReturnError(sqlmock.ErrCancelled)

	_, _, err := repo.ListByUser(context.Background(), userID, f)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---- ДОПОЛНЕНИЯ: сортировки, Scan-ошибка и rows.Err() ----

func TestAdsRepo_ListByUser_SortByNameAsc(t *testing.T) {
	repo, mock, done := newAdsRepo(t)
	defer done()

	userID := int64(7)
	f := entity.AdsFilter{Limit: 2, Offset: 0, Sort: "name"}

	mock.ExpectQuery(`SELECT\s+COUNT\(\*\)\s+FROM\s+ads a\s+JOIN\s+ad_accounts aa\s+ON aa\.account_id = a\.account_id\s+WHERE\s+aa\.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT\s+a\.ad_id, a\.account_id, a\.name, a\.status, a\.platform\s+FROM\s+ads a\s+JOIN\s+ad_accounts aa\s+ON aa\.account_id = a\.account_id\s+WHERE\s+aa\.user_id = \$1\s+ORDER BY a\.name ASC\s+LIMIT \$2 OFFSET \$3`).
		WithArgs(userID, 2, 0).
		WillReturnRows(sqlmock.NewRows([]string{"ad_id", "account_id", "name", "status", "platform"}).
			AddRow(int64(1), int64(10), "A", "active", "facebook"))

	items, total, err := repo.ListByUser(context.Background(), userID, f)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, items, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdsRepo_ListByUser_SortByNameDesc(t *testing.T) {
	repo, mock, done := newAdsRepo(t)
	defer done()

	userID := int64(7)
	f := entity.AdsFilter{Limit: 1, Offset: 0, Sort: "-name"}

	mock.ExpectQuery(`SELECT\s+COUNT\(\*\)\s+FROM\s+ads a\s+JOIN\s+ad_accounts aa\s+ON aa\.account_id = a\.account_id\s+WHERE\s+aa\.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT\s+a\.ad_id, a\.account_id, a\.name, a\.status, a\.platform\s+FROM\s+ads a\s+JOIN\s+ad_accounts aa\s+ON aa\.account_id = a\.account_id\s+WHERE\s+aa\.user_id = \$1\s+ORDER BY a\.name DESC\s+LIMIT \$2 OFFSET \$3`).
		WithArgs(userID, 1, 0).
		WillReturnRows(sqlmock.NewRows([]string{"ad_id", "account_id", "name", "status", "platform"}).
			AddRow(int64(2), int64(10), "Z", "active", "facebook"))

	_, _, err := repo.ListByUser(context.Background(), userID, f)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdsRepo_ListByUser_ScanError(t *testing.T) {
	repo, mock, done := newAdsRepo(t)
	defer done()

	userID := int64(9)
	f := entity.AdsFilter{Limit: 10, Offset: 0}

	// COUNT(*) → 1
	mock.ExpectQuery(`SELECT\s+COUNT\(\*\)\s+FROM\s+ads a\s+JOIN\s+ad_accounts aa\s+ON aa\.account_id = a\.account_id\s+WHERE\s+aa\.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// SELECT → тип в первой колонке ломает Scan (строка вместо BIGINT)
	mock.ExpectQuery(`SELECT\s+a\.ad_id, a\.account_id, a\.name, a\.status, a\.platform`).
		WithArgs(userID, 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"ad_id", "account_id", "name", "status", "platform"}).
			AddRow("oops", int64(10), "Name", "active", "facebook"))

	_, _, err := repo.ListByUser(context.Background(), userID, f)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdsRepo_ListByUser_ScanError_SecondRow(t *testing.T) {
	repo, mock, done := newAdsRepo(t)
	defer done()

	userID := int64(9)
	f := entity.AdsFilter{Limit: 10, Offset: 0}

	// COUNT(*) → 2 (будем читать 2 строки)
	mock.ExpectQuery(`SELECT\s+COUNT\(\*\)\s+FROM\s+ads a\s+JOIN\s+ad_accounts aa\s+ON aa\.account_id = a\.account_id\s+WHERE\s+aa\.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Первая строка валидна, вторая ломает Scan (строка вместо BIGINT для ad_id)
	rows := sqlmock.NewRows([]string{"ad_id", "account_id", "name", "status", "platform"}).
		AddRow(int64(1), int64(10), "A", "active", "facebook").
		AddRow("oops", int64(10), "B", "active", "facebook")

	mock.ExpectQuery(`SELECT\s+a\.ad_id, a\.account_id, a\.name, a\.status, a\.platform`).
		WithArgs(userID, 10, 0).
		WillReturnRows(rows)

	_, _, err := repo.ListByUser(context.Background(), userID, f)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

