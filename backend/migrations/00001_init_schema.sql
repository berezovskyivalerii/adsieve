-- +goose up
CREATE TABLE
    users (
        user_id BIGSERIAL PRIMARY KEY,
        email TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE TABLE
    ad_accounts (
        account_id BIGSERIAL PRIMARY KEY,
        user_id BIGINT NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
        platform TEXT NOT NULL,
        external_account_id TEXT NOT NULL,
        access_token TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        UNIQUE (platform, external_account_id)
    );

CREATE TABLE
    ads (
        ad_id BIGINT PRIMARY KEY,
        account_id BIGINT NOT NULL REFERENCES ad_accounts (account_id) ON DELETE CASCADE,
        name TEXT NOT NULL,
        status TEXT NOT NULL DEFAULT 'active',
        platform TEXT NOT NULL,
        UNIQUE (account_id, ad_id)
    );

CREATE TABLE
    clicks (
        id BIGSERIAL PRIMARY KEY,
        click_id VARCHAR(64) NOT NULL UNIQUE, 
        ad_id BIGINT NOT NULL REFERENCES ads (ad_id) ON DELETE CASCADE,
        clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        click_ref UUID UNIQUE
    );

CREATE INDEX idx_clicks_ad_time ON clicks (ad_id, clicked_at);

CREATE TABLE
    conversions (
        conversion_id BIGSERIAL PRIMARY KEY,
        ad_id BIGINT NOT NULL REFERENCES ads (ad_id) ON DELETE CASCADE,
        converted_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        revenue NUMERIC(15, 2) NOT NULL,
        order_id TEXT,
        click_ref UUID,
        UNIQUE (order_id, ad_id)
    );

CREATE INDEX idx_conv_ad_time ON conversions (ad_id, converted_at);

CREATE TABLE
    ads_insights (
        ad_id BIGINT NOT NULL REFERENCES ads (ad_id) ON DELETE CASCADE,
        insight_date DATE NOT NULL,
        spend NUMERIC(15, 2) NOT NULL,
        PRIMARY KEY (ad_id, insight_date)
    );

CREATE TABLE
    ad_daily_metrics (
        ad_id BIGINT NOT NULL REFERENCES ads (ad_id) ON DELETE CASCADE,
        metric_date DATE NOT NULL,
        clicks INT NOT NULL,
        conversions INT NOT NULL,
        revenue NUMERIC(15, 2) NOT NULL,
        spend NUMERIC(15, 2) NOT NULL,
        PRIMARY KEY (ad_id, metric_date)
    );

CREATE TABLE
    refresh_tokens (
        token_id BIGSERIAL PRIMARY KEY,
        user_id BIGINT NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
        refresh_token TEXT NOT NULL UNIQUE,
        expires_at TIMESTAMPTZ NOT NULL
    );

-- +goose down
DROP TABLE IF EXISTS refresh_tokens;

DROP TABLE IF EXISTS ad_daily_metrics;

DROP TABLE IF EXISTS ads_insights;

DROP INDEX IF EXISTS idx_conv_ad_time;

DROP TABLE IF EXISTS conversions;

DROP INDEX IF EXISTS idx_clicks_ad_time;

DROP TABLE IF EXISTS clicks;

DROP TABLE IF EXISTS ads;

DROP TABLE IF EXISTS ad_accounts;

DROP TABLE IF EXISTS users;