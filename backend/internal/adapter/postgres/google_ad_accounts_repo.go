package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type GoogleAdAccountsRepo struct {
	db *sql.DB
}

func NewGoogleAdAccountsRepo(db *sql.DB) *GoogleAdAccountsRepo {
	return &GoogleAdAccountsRepo{db: db}
}

// LinkGoogleAccounts — массовый UPSERT выбранных customerIds для пользователя.
// platform='google', external_account_id='<customerId>', status='linked'.
func (r *GoogleAdAccountsRepo) LinkGoogleAccounts(
	ctx context.Context,
	userID int64,
	tokenOwnerGoogleUserID string,
	customerIDs []string,
) error {
	if len(customerIDs) == 0 {
		return nil
	}
	const q = `
	INSERT INTO ad_accounts (user_id, platform, external_account_id, token_owner, status, created_at, updated_at)
	VALUES ($1, 'google', $2, $3, 'linked', NOW(), NOW())
	ON CONFLICT (platform, external_account_id) DO UPDATE
	SET user_id    = EXCLUDED.user_id,
		token_owner= EXCLUDED.token_owner,
		status     = 'linked',
		updated_at = NOW()`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, q)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, cid := range customerIDs {
		if _, err := stmt.ExecContext(ctx, userID, cid, tokenOwnerGoogleUserID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// UpsertLinked — одиночный upsert с возвратом account_id и признака already.
func (r *GoogleAdAccountsRepo) UpsertLinked(
	ctx context.Context,
	userID int64,
	platform, externalID, tokenOwner string,
) (accountID int64, already bool, err error) {
	const q = `
INSERT INTO ad_accounts (user_id, platform, external_account_id, token_owner, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, 'linked', NOW(), NOW())
ON CONFLICT (platform, external_account_id) DO UPDATE
SET user_id    = EXCLUDED.user_id,
    token_owner= EXCLUDED.token_owner,
    status     = 'linked',
    updated_at = NOW()
RETURNING account_id, (xmax <> 0) AS updated`
	if err = r.db.QueryRowContext(ctx, q, userID, platform, externalID, tokenOwner).
		Scan(&accountID, &already); err != nil {
		return 0, false, fmt.Errorf("upsert linked account: %w", err)
	}
	return accountID, already, nil
}

func (r *GoogleAdAccountsRepo) GetAccountID(
	ctx context.Context,
	userID int64,
	platform, externalID string,
) (int64, error) {
	const q = `SELECT account_id FROM ad_accounts
	           WHERE user_id=$1 AND platform=$2 AND external_account_id=$3
	           LIMIT 1`
	var id int64
	if err := r.db.QueryRowContext(ctx, q, userID, platform, externalID).Scan(&id); err != nil {
		return 0, fmt.Errorf("get account id: %w", err)
	}
	return id, nil
}

func (r *GoogleAdAccountsRepo) UpsertAdIfMissing(
	ctx context.Context,
	accountID, adID int64,
) error {
	const qPlatform = `SELECT platform FROM ad_accounts WHERE account_id=$1`
	var platform string
	if err := r.db.QueryRowContext(ctx, qPlatform, accountID).Scan(&platform); err != nil {
		return fmt.Errorf("lookup platform by account_id: %w", err)
	}
	const q = `
INSERT INTO ads (ad_id, account_id, name, status, platform)
VALUES ($1, $2, CONCAT('ad-', $1::text), 'active', $3)
ON CONFLICT (ad_id) DO NOTHING`
	if _, err := r.db.ExecContext(ctx, q, adID, accountID, platform); err != nil {
		return fmt.Errorf("upsert ad if missing: %w", err)
	}
	return nil
}

func (r *GoogleAdAccountsRepo) UpsertSpend(
	ctx context.Context,
	adID int64,
	date string,
	costMicros int64,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { if err != nil { _ = tx.Rollback() } }()

	const upsertInsights = `
INSERT INTO ads_insights (ad_id, insight_date, spend)
VALUES ($1, $2::date, ($3::numeric / 1000000.0))
ON CONFLICT (ad_id, insight_date)
DO UPDATE SET spend = EXCLUDED.spend`
	if _, err = tx.ExecContext(ctx, upsertInsights, adID, date, costMicros); err != nil {
		return fmt.Errorf("upsert ads_insights: %w", err)
	}

	const upsertDaily = `
INSERT INTO ad_daily_metrics (ad_id, metric_date, clicks, conversions, revenue, spend)
VALUES ($1, $2::date, 0, 0, 0, ($3::numeric / 1000000.0))
ON CONFLICT (ad_id, metric_date)
DO UPDATE SET spend = EXCLUDED.spend`
	if _, err = tx.ExecContext(ctx, upsertDaily, adID, date, costMicros); err != nil {
		return fmt.Errorf("upsert ad_daily_metrics: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
