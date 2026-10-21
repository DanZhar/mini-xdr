COMPOSE := docker compose
TI_DB_URL ?= postgres://xdr:xdr@localhost:5432/ti?sslmode=disable

.PHONY: up down reset logs migrate-ti build test lint run-ti

## Infrastructure
up:            ## Start Postgres, Redis and apply migrations
	$(COMPOSE) up -d

down:          ## Stop containers (data is kept)
	$(COMPOSE) down

reset:         ## Stop containers and wipe volumes (re-runs init.sql)
	$(COMPOSE) down -v

logs:
	$(COMPOSE) logs -f

migrate-ti:    ## Re-run TI migrations
	$(COMPOSE) run --rm migrate-ti

## Go
build:
	go build -o bin/ti-service ./cmd/ti-service

run-ti:
	go run ./cmd/ti-service

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run ./...