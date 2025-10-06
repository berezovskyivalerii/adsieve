-- +goose Up
ALTER TABLE google_user_tokens
  ADD COLUMN IF NOT EXISTS needs_consent BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_ad_accounts_platform_ext
  ON ad_accounts(platform, external_account_id);

CREATE INDEX IF NOT EXISTS idx_ads_insights_ad_date
  ON ads_insights(ad_id, insight_date);

-- +goose Down
ALTER TABLE google_user_tokens DROP COLUMN IF EXISTS needs_consent;
DROP INDEX IF EXISTS idx_ad_accounts_platform_ext;
DROP INDEX IF EXISTS idx_ads_insights_ad_date;
