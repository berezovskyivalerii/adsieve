-- +goose Up

-- 1) Одноразовые состояния OAuth (state + PKCE)
CREATE TABLE IF NOT EXISTS oauth_states (
  state         TEXT PRIMARY KEY,
  code_verifier TEXT        NOT NULL,
  user_id       BIGINT      NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
  expires_at    TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_oauth_states_expires ON oauth_states (expires_at);

-- 2) Хранилище refresh-токенов Google (в шифре)
CREATE TABLE IF NOT EXISTS google_user_tokens (
  user_id             BIGINT NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
  google_user_id      TEXT   NOT NULL,
  refresh_token_enc   TEXT   NOT NULL,
  refresh_token_scope TEXT   NOT NULL,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, google_user_id)
);

-- 3) Расширение ad_accounts под Google
--    - access_token делаем NULLABLE (для Google он не нужен)
--    - добавляем владельца токена (google_user_id/id_token-sub), статус и служебные таймстемпы
ALTER TABLE ad_accounts
  ALTER COLUMN access_token DROP NOT NULL;

ALTER TABLE ad_accounts
  ADD COLUMN IF NOT EXISTS token_owner TEXT,
  ADD COLUMN IF NOT EXISTS status      TEXT        NOT NULL DEFAULT 'linked',
  ADD COLUMN IF NOT EXISTS last_sync_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- +goose Down

-- 1) Вернуть строгий NOT NULL для access_token:
--    Перед этим заменим NULLы на 'stub' чтобы откат не упал.
UPDATE ad_accounts SET access_token = 'stub' WHERE access_token IS NULL;
ALTER TABLE ad_accounts
  ALTER COLUMN access_token SET NOT NULL;

-- 2) Удалить добавленные поля в ad_accounts
ALTER TABLE ad_accounts
  DROP COLUMN IF EXISTS token_owner,
  DROP COLUMN IF EXISTS status,
  DROP COLUMN IF EXISTS last_sync_at,
  DROP COLUMN IF EXISTS updated_at;

-- 3) Удалить вспомогательные таблицы
DROP TABLE IF EXISTS google_user_tokens;
DROP TABLE IF EXISTS oauth_states;
