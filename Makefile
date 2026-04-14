APP_NAME := server
BIN_DIR := bin
CMD_DIR := ./cmd/server
ENV ?= dev
PROFILE ?=

.PHONY: help tidy build run run-worker migrate-up migrate-down test test-unit swagger clean

help:
	@echo "Available targets:"
	@echo "  make tidy         - tidy go modules"
	@echo "  make build        - build server binary"
	@echo "  make run          - run HTTP server (ENV=dev/test/prod PROFILE=order)"
	@echo "  make run-worker   - run Asynq worker (ENV/PROFILE)"
	@echo "  make migrate-up   - run DB migration up (set DSN)"
	@echo "  make migrate-down - run DB migration down one step (set DSN)"
	@echo "  make test-unit    - run tests under tests/unit"
	@echo "  make test         - run all tests"
	@echo "  make swagger      - generate swagger docs"
	@echo "  make clean        - remove built binaries"

tidy:
	go mod tidy

build:
	go build -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)

run:
	go run $(CMD_DIR) server --env $(ENV) --profile $(PROFILE)

run-worker:
	go run $(CMD_DIR) worker --env $(ENV) --profile $(PROFILE)

migrate-up:
	go run ./cmd/migrate up --driver "$(DRIVER)" --dsn "$(DSN)"

migrate-down:
	go run ./cmd/migrate down --driver "$(DRIVER)" --dsn "$(DSN)"

test-unit:
	go test ./tests/unit/...

test:
	go test ./...

swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o docs -d ./cmd/server,./api

clean:
	rm -rf $(BIN_DIR)
