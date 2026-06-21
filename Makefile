.PHONY: help dev build run test tidy lint migrate-up migrate-down migrate-create sqlc-gen docker-up docker-down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

dev: ## Run with hot reload (Air)
	air

build: ## Build binary into ./tmp/api
	go build -o ./tmp/api ./cmd/api

run: build ## Build then run once
	./tmp/api

test: ## Run tests
	go test ./... -cover

tidy: ## go mod tidy
	go mod tidy

lint: ## Run golangci-lint
	golangci-lint run ./...

## ── Database migrations (golang-migrate) ─────────────
# Default matches .env; override with: make migrate-up DATABASE_URL=...
DATABASE_URL ?= postgres://transport:transport_secret@localhost:5434/transport_db?sslmode=disable

migrate-up: ## Apply migrations up
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down: ## Rollback last migration
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-create: ## Create a pair of migrations: make migrate-create name=add_vehicles
	migrate create -ext sql -dir migrations -seq $(name)

sqlc-gen: ## Generate sqlc code
	sqlc generate

## ── Docker ───────────────────────────────────────────
docker-up: ## Start full stack via root docker-compose
	cd .. && docker compose up -d

docker-down: ## Stop full stack
	cd .. && docker compose down
