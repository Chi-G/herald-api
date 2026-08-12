.PHONY: run db seed build

default: run

# Run the Herald server (loads .env automatically)
run:

	go run cmd/server/main.go

# Start local PostgreSQL container
db:
	docker-compose up -d

# Seed database schema & test API key
seed:
	docker exec -i herald_postgres psql -U herald -d herald < schema.sql
	docker exec -i herald_postgres psql -U herald -d herald < seed.sql

# Build application binary
build:
	go build -o bin/herald cmd/server/main.go
