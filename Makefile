.PHONY: run build tidy migrate

run:
	go run ./cmd/server/main.go

build:
	go build -o bin/briers-api ./cmd/server/main.go

tidy:
	go mod tidy

migrate:
	psql "$(DATABASE_URL)" -f migrations/001_init_schema.sql

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
