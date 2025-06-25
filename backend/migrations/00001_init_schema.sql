-- +goose up
CREATE TABLE
    users (
        id BIGSERIAL PRIMARY KEY,
        email TEXT NOT NULL UNIQUE,
        pass_hash TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE TABLE
    clicks (
        id BIGSERIAL PRIMARY KEY,
        click_id VARCHAR(64) NOT NULL UNIQUE,
        ad_id BIGINT NOT NULL,
        occured_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE TABLE orders(
    id BIGSERIAL PRIMARY KEY,
    click_id VARCHAR(64) NOT NULL REFERENCES clicks(click_id),
    order_value NUMERIC(12,2) NOT NULL,
    occured_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
);

CREATE TABLE ads_insights(
    ad_id BIGINT NOT NULL,
    insight_date DATE NOT NULL,
    clicks INT NOT NULL DEFAULT 0,
    spend NUMERIC(12,2) NOT NULL DEFAULT 0,
    impressions INT NOT NULL DEFAULT 0,
    PRIMARY KEY (ad_id, insight_date)
);

-- +goose down
DROP TABLE IF EXISTS ads_insights;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS clicks;
DROP TABLE IF EXISTS users;