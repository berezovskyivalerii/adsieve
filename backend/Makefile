DC          = docker compose
COMPOSE_CFG = -f docker-compose.yml

## Собрать образы
build:
	$(DC) $(COMPOSE_CFG) build

## Поднять всё в фоне (с миграциями)
up: build
	$(DC) $(COMPOSE_CFG) up -d

## Запустить миграции отдельно
migrate:
	$(DC) $(COMPOSE_CFG) run --rm goose -dir=/migrations up

## Остановить и убрать контейнеры
down:
	$(DC) $(COMPOSE_CFG) down

## Полная перезагрузка (удобно при смене схемы)
restart: down up

## Логи API-сервиса
logs:
	$(DC) $(COMPOSE_CFG) logs -f api

## Быстро зайти psql внутрь контeйнера Postgres
psql:
	$(DC) $(COMPOSE_CFG) exec postgres psql -U postgres -d adsieve

## Шелл в контейнере БД
db-sh:
	$(DC) $(COMPOSE_CFG) exec postgres bash

## Удалить всё подчистую: контейнеры, образы, тома, orphans
clean:
	$(DC) $(COMPOSE_CFG) down -v --rmi all --remove-orphans

.PHONY: build up migrate down restart logs psql db-sh clean
