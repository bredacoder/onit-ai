# DATABASE_URL must be set in the environment before running migrate or test.
# Copy .env.example to .env.local (KEY=VALUE format) and run:
#   export $(grep -v '^#' .env.local | xargs)
# or simply export DATABASE_URL directly.

.PHONY: db-up db-down migrate lint test

db-up:
	docker compose up -d --wait db

db-down:
	docker compose down

migrate:
	go tool goose -dir db/migrations postgres "$(DATABASE_URL)" up

lint:
	go tool golangci-lint run

test: db-up migrate
	go test ./...
