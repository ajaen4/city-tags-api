.PHONY: build run test clean docs migrations setup

-include .env

setup:
	@$(eval DB_STRING=postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable)

build-local:
	@go build -o ./bin/main cmd/api/main.go

build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/main cmd/api/main.go

run:
	@make docs
	@air

docs:
	@swag init --dir ./cmd/api,./internal/server,./internal/api_errors,./internal/database

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
	@goose -dir $(MIGRATION_DIR) create $(MIGRATION_NAME) sql

# GOOSE_DBSTRING must have "" around it so the make file interprets correctly certain special characters
run-migrations: setup
	@goose -dir $(MIGRATION_DIR) $(DB_DRIVER) "$(DB_STRING)" up

reset-migrations: setup
	@goose -dir $(MIGRATION_DIR) $(DB_DRIVER) "$(DB_STRING)" reset

migrations-status: setup
	@goose -dir $(MIGRATION_DIR) $(DB_DRIVER) "$(DB_STRING)" status

single-image-run: setup
	@docker build -t city-tags-api .
	@docker run --name city-tags-api --env-file .env -p $(SERVER_PORT):$(SERVER_PORT) --entrypoint /bin/sh city-tags-api -c "./main"

single-image-down:
	@docker container rm city-tags-api

test-env-up:
	@echo "Running test env..."
	@docker compose up --build -d
	@echo "Test env deployed"

test-env-down:
	@echo "Destroying test env..."
	@docker compose down --volumes
	@echo "Test env destroyed"

unit-tests:
	@echo "Running unit tests..."
	@go test ./internal/... -v

integration-tests:
	make test-env-up
	@docker compose exec -T integration-tests make run-int-tests-in-container;result=$$?; make test-env-down; exit $$result

run-int-tests-in-container:
	@echo "Running migrations..."
	@make run-migrations
	@echo "Initializing db..."
	@make init-db DATA_FILE=./integration_tests/init_db/init_queries.sql
	@echo "Running integration tests..."
	@go test ./integration_tests/... -v; result=$$?; exit $$result

clean:
	@echo "Cleaning..."
	@rm -f bin/main

test-cicd:
	@act -P ubuntu-latest=catthehacker/ubuntu:act-22.04 -e .github/created_pr.json --job tests --secret-file .github/.env --container-architecture linux/amd64
