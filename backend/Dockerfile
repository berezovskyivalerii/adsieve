FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o adsieve ./cmd/api

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/adsieve .

ENV PORT=8080
ENV DATABASE_URL=postgres://adsieve:adsieve@postgres:5432/adsieve?sslmode=disable

EXPOSE 8080
ENTRYPOINT ["./adsieve"]
