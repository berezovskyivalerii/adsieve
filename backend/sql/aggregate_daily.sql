-- Агрегация дневных метрик за последние 7 дней (UTC).
-- Таблицы:
--   clicks(id, ad_id, clicked_at timestamptz, click_ref uuid)
--   conversions(conversion_id, ad_id, converted_at timestamptz, revenue numeric(15,2), order_id text, click_ref uuid)
--   ads_insights(ad_id, insight_date date, spend numeric(15,2))
--   ad_daily_metrics(ad_id, metric_date date, clicks int, conversions int, revenue numeric(15,2), spend numeric(15,2))

-- 1) клики → clicks
WITH bounds AS (
  SELECT
    (now() AT TIME ZONE 'UTC')::date AS today,
    ((now() AT TIME ZONE 'UTC')::date - INTERVAL '7 days')::date AS since
)
INSERT INTO ad_daily_metrics (ad_id, metric_date, clicks, conversions, revenue, spend)
SELECT
  c.ad_id,
  (c.clicked_at AT TIME ZONE 'UTC')::date AS metric_date,
  COUNT(*) AS clicks,
  0, 0, 0
FROM clicks c
JOIN bounds b ON TRUE
WHERE (c.clicked_at AT TIME ZONE 'UTC')::date BETWEEN b.since AND b.today
GROUP BY c.ad_id, (c.clicked_at AT TIME ZONE 'UTC')::date
ON CONFLICT (ad_id, metric_date) DO UPDATE
SET clicks = EXCLUDED.clicks;

-- 2) конверсии + revenue → conversions
WITH bounds AS (
  SELECT
    (now() AT TIME ZONE 'UTC')::date AS today,
    ((now() AT TIME ZONE 'UTC')::date - INTERVAL '7 days')::date AS since
),
agg AS (
  SELECT
    o.ad_id,
    (o.converted_at AT TIME ZONE 'UTC')::date AS metric_date,
    COUNT(*) AS conversions,
    COALESCE(SUM(o.revenue), 0) AS revenue
  FROM conversions o
  JOIN bounds b ON TRUE
  WHERE (o.converted_at AT TIME ZONE 'UTC')::date BETWEEN b.since AND b.today
  GROUP BY o.ad_id, (o.converted_at AT TIME ZONE 'UTC')::date
)
INSERT INTO ad_daily_metrics (ad_id, metric_date, clicks, conversions, revenue, spend)
SELECT ad_id, metric_date, 0, conversions, revenue, 0
FROM agg
ON CONFLICT (ad_id, metric_date) DO UPDATE
SET conversions = EXCLUDED.conversions,
    revenue     = EXCLUDED.revenue;

-- 3) расходы → ads_insights
WITH bounds AS (
  SELECT
    (now() AT TIME ZONE 'UTC')::date AS today,
    ((now() AT TIME ZONE 'UTC')::date - INTERVAL '7 days')::date AS since
)
INSERT INTO ad_daily_metrics (ad_id, metric_date, clicks, conversions, revenue, spend)
SELECT
  ai.ad_id,
  ai.insight_date AS metric_date,
  0, 0, 0,
  COALESCE(SUM(ai.spend), 0) AS spend
FROM ads_insights ai
JOIN bounds b ON TRUE
WHERE ai.insight_date BETWEEN b.since AND b.today
GROUP BY ai.ad_id, ai.insight_date
ON CONFLICT (ad_id, metric_date) DO UPDATE
SET spend = EXCLUDED.spend;
