include .env
export

DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: run build test bdd migrate-up migrate-down migrate-create docker-up docker-down docker-logs

run:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

test:
	go test ./... -v

bdd:
	cd $(UTILITIES_DIR) && docker-compose up -d db
	go test -tags bdd ./tests/bdd -v
	cd $(UTILITIES_DIR) && docker-compose down

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# Docker (runs docker-compose from Utilities repo)
UTILITIES_DIR := ../Utilities

docker-up:
	cd $(UTILITIES_DIR) && docker-compose up --build -d

docker-down:
	cd $(UTILITIES_DIR) && docker-compose down

docker-logs:
	cd $(UTILITIES_DIR) && docker-compose logs -f backend
