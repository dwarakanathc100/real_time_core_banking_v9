.PHONY: build run docker-build docker-up migrate test

build:
    go build -o bin/rtcb cmd/api/main.go

run: build
    ./bin/rtcb

docker-build:
    docker build -t rtcb:latest .

docker-up: docker-build
    docker-compose up --build -d

migrate:
    psql postgresql://postgres:postgres@localhost:5432/rtcb?sslmode=disable -f internal/db/migrations/001_init.sql || true

test:
    go test ./... -v
