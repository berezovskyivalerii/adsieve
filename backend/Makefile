# -------- Settings --------
SHELL := /bin/sh
COMPOSE := docker compose

# Demo auth (для sign-in цели)
EMAIL ?= testuser@adsieve.local
PASS  ?= superpass

# -------- Help --------
.PHONY: help
help: ## показать список команд
	@echo "AdSieve – Make команды:"
	@grep -E '^[a-zA-Z0-9_.-]+:.*?## ' $(MAKEFILE_LIST) | awk -F':.*?## ' '{printf "  %-24s %s\n", $$1, $$2}'

# -------- Bootstrap / Build --------
.PHONY: env
env: ## создать .env (если отсутствует)
	@[ -f .env ] || { \
	  echo 'POSTGRES_HOST=postgres'                 > .env; \
	  echo 'POSTGRES_USER=postgres'                >> .env; \
	  echo 'POSTGRES_PASSWORD=secret'              >> .env; \
	  echo 'POSTGRES_DB=adsieve'                   >> .env; \
	  echo 'POSTGRES_SSL=disable'                  >> .env; \
	  echo ''                                      >> .env; \
	  echo 'DB_DSN=postgres://postgres:secret@postgres:5432/adsieve?sslmode=disable' >> .env; \
	  echo ''                                      >> .env; \
	  echo 'GOOSE_DRIVER=postgres'                 >> .env; \
	  echo 'GOOSE_DBSTRING=host=postgres port=5432 user=postgres password=secret dbname=adsieve sslmode=disable' >> .env; \
	  echo ''                                      >> .env; \
	  echo 'PORT=8080'                             >> .env; \
	  echo 'JWT_SECRET=super-secret-key'           >> .env; \
	  echo 'BCRYPT_COST=12'                        >> .env; \
	  echo "[ok] .env создан"; \
	} || echo "[skip] .env уже существует"

.PHONY: build
build: ## собрать образы (api, cron, goose)
	$(COMPOSE) build --no-cache postgres api cron-sync-fb cron-aggregate goose

# -------- Up / Down --------
.PHONY: up-db
up-db: ## запустить только Postgres
	$(COMPOSE) up -d postgres

.PHONY: wait-db
wait-db: ## дождаться health Postgres
	@printf "[wait] postgres health"
	@until $(COMPOSE) ps | grep -E "adsieve-db\s+.*healthy" >/dev/null; do printf "."; sleep 1; done; echo " ok"

.PHONY: migrate
migrate: ## прогнать миграции (goose)
	$(COMPOSE) run --rm goose

.PHONY: up
up: env up-db wait-db migrate ## запустить Postgres + миграции
	@echo "[ok] DB и миграции готовы"

.PHONY: up-all
up-all: up ## запустить API и кроны
	$(COMPOSE) up -d api cron-sync-fb cron-aggregate
	@echo "[ok] api + cron подняты"

.PHONY: down
down: ## остановить и удалить контейнеры/тома
	$(COMPOSE) down -v --remove-orphans

.PHONY: ps logs
ps:   ## статус контейнеров
	$(COMPOSE) ps
logs: ## хвост логов (всех)
	$(COMPOSE) logs -f

# -------- Handy exec --------
.PHONY: psql
psql: ## открыть psql в контейнере Postgres
	$(COMPOSE) exec postgres sh -lc 'PGPASSWORD="$$POSTGRES_PASSWORD" psql -h 127.0.0.1 -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"'

.PHONY: sync-once
sync-once: ## разово выполнить sync_fb_insights
	$(COMPOSE) exec -T cron-sync-fb /app/sync_fb_insights || true

.PHONY: aggregate-once
aggregate-once: ## разово выполнить aggregate_daily
	$(COMPOSE) exec -T cron-aggregate /app/aggregate_daily || true

# -------- API quick checks --------
.PHONY: api-alive
api-alive: ## проверить, что API слушает /api/click (ожидаем 400/404)
	@code=$$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/api/click || true); \
	if echo $$code | grep -qE '^(400|404)$$'; then echo "[ok] api up (code=$$code)"; else echo "[err] api not responding (code=$$code)"; exit 1; fi

.PHONY: sign-up sign-in token
sign-up: ## создать пользователя (используй EMAIL=... PASS=...)
	curl -s -X POST http://localhost:8080/api/auth/sign-up \
	  -H 'Content-Type: application/json' \
	  -d '{"email":"$(EMAIL)","password":"$(PASS)"}'
	@echo

sign-in: ## залогиниться и получить access_token (EMAIL/PASS)
	curl -s -X POST http://localhost:8080/api/auth/sign-in \
	  -H 'Content-Type: application/json' \
	  -d '{"email":"$(EMAIL)","password":"$(PASS)"}'
	@echo

token: ## распарсить access_token (требует jq)
	@curl -s -X POST http://localhost:8080/api/auth/sign-in \
	  -H 'Content-Type: application/json' \
	  -d '{"email":"$(EMAIL)","password":"$(PASS)"}' | jq -r '.access_token'

.PHONY: metrics-today
metrics-today: ## запросить метрики за сегодня (UTC) – ACCESS=... обязателен
	@[ -n "$$ACCESS" ] || (echo "[err] Укажи ACCESS: make metrics-today ACCESS=<token>"; exit 1)
	@TODAY=$$(date -I -u); \
	curl -s -H "Authorization: Bearer $$ACCESS" \
	  "http://localhost:8080/api/metrics?ad_ids=101,102,103&from=$$TODAY&to=$$TODAY" | sed -e 's/^/> /'

# -------- Dev all-in-one --------
.PHONY: dev
dev: env build up-all api-alive ## полный цикл: .env → build → db+migrate → api+cron
	@echo "[done] проект запущен: http://localhost:8080"
