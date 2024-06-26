.PHONY: build run test clean docs migrations setup

-include .env

setup:
	@$(eval GOOSE_DBSTRING=postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=require)

build-local:	
	@go build -o ./bin/main cmd/api/main.go

build:	
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/main cmd/api/main.go

run:
	make docs
	air

docs:
	if [ -d docs/ ]; then rm -r docs/; fi && swag init --dir ./cmd/api,./internal/server,./internal/api_errors

open-db-conn: setup
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USERNAME) -d $(DB_NAME)

init-db: setup
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USERNAME) -d $(DB_NAME) -f ./init_db/init_queries.sql

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

docker-run: setup
	docker build -t city-tags-api .
	docker run --name city-tags-api --env-file .env -p $(SERVER_PORT):$(SERVER_PORT) --entrypoint /bin/sh city-tags-api -c "./main"

docker-down:
	docker rm city-tags-api

unit-tests:
	@echo "Testing..."
	@go test ./internal -v

integration-tests:
	@echo "Testing..."
	@go test ./integration_tests -v

clean:
	@echo "Cleaning..."
	@rm -f bin/main
