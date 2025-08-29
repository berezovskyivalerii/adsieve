-- +goose Up
-- Demo seed: user + ad_account + ads (101/102/103) + user_ads
-- ВНИМАНИЕ: password_hash сейчас 'stub'. Для реального логина подставь bcrypt-хеш.

-- 1) ensure demo user
INSERT INTO users (email, password_hash)
VALUES ('testuser@adsieve.local', 'superpass')
ON CONFLICT (email) DO NOTHING;

-- 2) ensure ad_account 'facebook/acc_demo'
INSERT INTO ad_accounts (user_id, platform, external_account_id, access_token)
SELECT u.user_id, 'facebook', 'acc_demo', 'stub'
FROM users u
WHERE u.email = 'testuser@adsieve.local'
ON CONFLICT (platform, external_account_id) DO NOTHING;

-- 3) ensure ads 101/102/103 привязаны к аккаунту
WITH acc AS (
  SELECT account_id
  FROM ad_accounts
  WHERE platform='facebook' AND external_account_id='acc_demo'
)
INSERT INTO ads (ad_id, account_id, name, status, platform)
SELECT x.ad_id, acc.account_id, x.name, 'active', 'facebook'
FROM acc
CROSS JOIN (VALUES
  (101, 'Test Ad 101'),
  (102, 'Test Ad 102'),
  (103, 'Test Ad 103')
) AS x(ad_id, name)
ON CONFLICT (ad_id) DO NOTHING;

-- 4) ensure связи user -> ads
WITH u AS (SELECT user_id FROM users WHERE email='testuser@adsieve.local')
INSERT INTO user_ads (user_id, ad_id)
SELECT u.user_id, x.ad_id
FROM u
JOIN (VALUES (101),(102),(103)) AS x(ad_id) ON TRUE
ON CONFLICT DO NOTHING;

-- +goose Down
-- Аккуратно удаляем только созданные здесь объекты
WITH u AS (SELECT user_id FROM users WHERE email='testuser@adsieve.local'),
     a AS (SELECT account_id FROM ad_accounts WHERE platform='facebook' AND external_account_id='acc_demo')
DELETE FROM user_ads WHERE (user_id) IN (SELECT user_id FROM u) AND ad_id IN (101,102,103);

DELETE FROM ads
WHERE ad_id IN (101,102,103)
  AND account_id IN (SELECT account_id FROM ad_accounts WHERE platform='facebook' AND external_account_id='acc_demo');

DELETE FROM ad_accounts
WHERE platform='facebook' AND external_account_id='acc_demo'
  AND user_id IN (SELECT user_id FROM users WHERE email='testuser@adsieve.local');

DELETE FROM users WHERE email='testuser@adsieve.local';
