package postgres

import (
	"context"
	"database/sql"
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

