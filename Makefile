.PHONY: up restart build rebuild down logs ps db api front migrate-up migrate-down test test-cover test-db-setup fmt lint frontend-lint frontend-test frontend-typecheck audit check

up:
	docker compose up -d

restart:
	docker compose restart

build:
	docker compose build

rebuild:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

db:
	docker compose exec db psql -U app_user -d app_db

api:
	docker compose exec api sh

front:
	docker compose exec frontend sh

migrate-up:
	docker compose exec api sh -lc 'migrate -path /db/migrations -database "$$DATABASE_URL" up'

migrate-down:
	docker compose exec api sh -lc 'migrate -path /db/migrations -database "$$DATABASE_URL" down 1'

test:
	docker compose exec api sh -lc 'PATH="/usr/local/go/bin:$$PATH" DATABASE_URL="postgres://app_user:password@db:5432/app_db_test?sslmode=disable" REDIS_URL="redis://redis:6379/1" go test ./... --cover'

test-cover:
	docker compose exec api sh -lc 'PATH="/usr/local/go/bin:$$PATH"; DATABASE_URL="postgres://app_user:password@db:5432/app_db_test?sslmode=disable"; REDIS_URL="redis://redis:6379/1"; COVERPKG=$$(go list ./... | grep -v "/cmd/" | grep -v "smtp_tls" | tr "\n" "," | sed "s/,$$//"); go test ./... -coverprofile=/tmp/cover.out -coverpkg="$$COVERPKG" && go tool cover -func=/tmp/cover.out > /tmp/cover.txt && cat /tmp/cover.txt && grep "^total:" /tmp/cover.txt | grep -q "100.0%" || { echo "ERROR: coverage < 100.0%"; exit 1; }'

test-db-setup:
	docker compose exec db psql -U postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'app_db_test' AND pid <> pg_backend_pid();"
	docker compose exec db psql -U postgres -c "DROP DATABASE IF EXISTS app_db_test;"
	docker compose exec db psql -U postgres -c "CREATE DATABASE app_db_test OWNER app_user ENCODING 'UTF8' LC_COLLATE 'ja_JP.UTF-8' LC_CTYPE 'ja_JP.UTF-8' TEMPLATE template0;"
	docker compose exec db psql -U postgres -d app_db_test -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
	docker compose exec db psql -U postgres -d app_db_test -c "CREATE EXTENSION IF NOT EXISTS pgroonga;"
	docker compose exec api sh -lc 'migrate -path /db/migrations -database "postgres://app_user:password@db:5432/app_db_test?sslmode=disable" up'

fmt:
	docker compose exec api sh -lc 'gofmt -w $$(find . -name "*.go" -not -path "./tmp/*")'

lint:
	docker compose exec api golangci-lint run

frontend-lint:
	docker compose exec frontend npm run lint

frontend-test:
	docker compose exec frontend npm run test

frontend-typecheck:
	docker compose exec frontend npm run typecheck

audit:
	docker compose exec frontend npm audit

check:
	make lint
	make test-cover
	make frontend-lint
	make frontend-test
	make frontend-typecheck