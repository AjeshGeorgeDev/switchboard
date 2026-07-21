.PHONY: dev-up dev-down prod-up prod-down prod-logs prod-ps migrate-up migrate-down sqlc-generate build run backend-dev test frontend-install frontend-dev frontend-build db-backup postgres-upgrade-16-to-17 screenshot-data screenshots docker-build

DATABASE_URL ?= postgres://switchboard:switchboard@localhost:5432/switchboard?sslmode=disable
REDIS_URL ?= redis://localhost:6379/0
PORT ?= 8080
COMPOSE_PROD := docker compose -f docker-compose.prod.yml

-include .env
export
GOPATH_BIN := $(shell go env GOPATH)/bin

dev-up:
	docker compose up -d

dev-down:
	docker compose down

# Full production stack (Postgres + Redis + app). Requires .env with POSTGRES_PASSWORD, JWT_SECRET, APP_BASE_URL.
prod-up:
	$(COMPOSE_PROD) up -d --build

prod-down:
	$(COMPOSE_PROD) down

prod-logs:
	$(COMPOSE_PROD) logs -f app

prod-ps:
	$(COMPOSE_PROD) ps

docker-build:
	docker build -t switchboard:latest .

migrate-up:
	cd backend && "$(GOPATH_BIN)/migrate" -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	cd backend && "$(GOPATH_BIN)/migrate" -path migrations -database "$(DATABASE_URL)" down 1

sqlc-generate:
	cd backend && "$(GOPATH_BIN)/sqlc" generate

build: frontend-build
	cd backend && go build -o bin/server ./cmd/server

# .env is included + exported above; recipes inherit HARBOR_*, JWT_*, etc.
run:
	cd backend && go run ./cmd/server

backend-dev:
	@test -x "$(GOPATH_BIN)/air" || (echo 'Install air: go install github.com/air-verse/air@latest' && exit 1)
	cd backend && "$(GOPATH_BIN)/air"

test:
	cd backend && go test ./...

frontend-install:
	cd frontend && pnpm install

frontend-dev:
	cd frontend && PORT="$(PORT)" pnpm run dev

frontend-build:
	cd frontend && pnpm run build

db-backup:
	bash scripts/postgres-backup.sh

postgres-upgrade-16-to-17:
	bash scripts/postgres-upgrade-16-to-17.sh

screenshot-data:
	bash scripts/ensure-screenshot-user.sh
	APP_BASE_URL=http://localhost:$(PORT) bash scripts/seed-screenshot-data.sh

screenshots:
	@test -d node_modules/playwright || npm install --no-save playwright@1.50.0
	@npx playwright install chromium
	node scripts/capture-screenshots.mjs
