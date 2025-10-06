# --- build stage ---
FROM golang:1.24.0-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g ./cmd/api/main.go -o ./docs -d .

# api
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api
# cron: sync_fb_insights
RUN CGO_ENABLED=0 go build -o /out/sync_fb_insights ./cmd/cron/sync_fb_insights
# cron: aggregate_daily  <-- ДОБАВЬ ЭТУ СТРОКУ
RUN CGO_ENABLED=0 go build -o /out/aggregate_daily ./cmd/cron/aggregate_daily
# goose (если пользуешься)
RUN GOBIN=/out go install github.com/pressly/goose/v3/cmd/goose@latest

# --- runtime stage ---
FROM alpine:3.20
WORKDIR /app
RUN adduser -D -h /app app && chown -R app:app /app
USER app

COPY --from=build /out/api /app/api
COPY --from=build /out/sync_fb_insights /app/sync_fb_insights
COPY --from=build /out/aggregate_daily /app/aggregate_daily
COPY --from=build /out/goose /app/goose
COPY sql /sql
COPY migrations /app/migrations

ENTRYPOINT ["/app/api"]
