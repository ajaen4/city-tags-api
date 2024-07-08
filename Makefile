.PHONY: build run test clean docs migrations setup

-include .env

setup:
	@$(eval GOOSE_DBSTRING=postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable)

build-local:
	@go build -o ./bin/main cmd/api/main.go

build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/main cmd/api/main.go

run:
	@make docs
	@air

docs:
	if [ -d docs/ ]; then rm -r docs/; fi && swag init --dir ./cmd/api,./internal/server,./internal/api_errors,./internal/database

open-db-conn: setup
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USERNAME) -d $(DB_NAME)

init-db: setup
	@if [ -z "$(DATA_FILE)" ]; then \
		echo "Error: DATA_FILE is not set"; \
		exit 1; \
	fi
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USERNAME) -d $(DB_NAME) -f $(DATA_FILE)

create-migration: setup
	@if [ -z "$(MIGRATION_NAME)" ]; then \
		echo "Error: MIGRATION_NAME is not set"; \
		exit 1; \
	fi
	@echo "Creating migration: $(MIGRATION_NAME)"
	@goose -dir $(GOOSE_MIGRATION_DIR) create $(MIGRATION_NAME) sql

# GOOSE_DBSTRING must have "" around it so the make file interprets correctly certain special characters
run-migrations: setup
	@goose -dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" up

reset-migrations: setup
	@goose -dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" reset

migrations-status: setup
	@goose -dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" status

single-image-run: setup
	docker build -t city-tags-api .
	docker run --name city-tags-api --env-file .env -p $(SERVER_PORT):$(SERVER_PORT) --entrypoint /bin/sh city-tags-api -c "./main"

single-image-down:
	docker rm city-tags-api

test-env-up:
	@echo "Running test env..."
	@docker compose up -d
	@echo "Waiting for test env to be ready..."
	@echo "Running migrations..."
	@make run-migrations
	@echo "Initializing db..."
	@make init-db DATA_FILE=$(DATA_FILE)

test-env-down:
	@docker compose down --volumes
	@docker images | grep 'city-tags-api-test' | awk '{print $$3}' | xargs -r docker rmi || true

unit-tests:
	@echo "Running unit tests..."
	@go test ./internal/... -v

integration-tests:
	@make test-env-up DATA_FILE=./integration_tests/init_db/init_queries.sql
	@echo "Running integration tests..."
	@go test ./integration_tests/... -v; result=$$?; make test-env-down; exit $$result

clean:
	@echo "Cleaning..."
	@rm -f bin/main
