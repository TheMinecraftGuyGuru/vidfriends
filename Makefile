SHELL := /bin/bash

BACKEND_DIR := backend
FRONTEND_DIR := frontend
PNPM := pnpm

.PHONY: help test lint backend-test backend-lint backend-migrate-up backend-migrate-down backend-migrate-status frontend-install frontend-test frontend-lint frontend-build

help: ## Show available make targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / {printf "\033[36m%-24s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: backend-test frontend-test ## Run backend and frontend test suites

lint: backend-lint frontend-lint ## Run linting for both services

backend-test: ## Run Go unit tests for the backend service
	cd $(BACKEND_DIR) && go test ./...

backend-lint: ## Run Go linters for the backend service
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd $(BACKEND_DIR) && golangci-lint run; \
	else \
		echo "golangci-lint not installed; skipping backend lint."; \
	fi

backend-migrate-up: ## Apply database migrations (up) using the backend CLI
	cd $(BACKEND_DIR) && go run ./cmd/vidfriends migrate up

backend-migrate-down: ## Roll back the most recent migration (down) using the backend CLI
	cd $(BACKEND_DIR) && go run ./cmd/vidfriends migrate down

backend-migrate-status: ## Show the status of database migrations using the backend CLI
	cd $(BACKEND_DIR) && go run ./cmd/vidfriends migrate status

frontend-install: ## Install frontend dependencies with pnpm
	$(PNPM) --dir $(FRONTEND_DIR) install

frontend-test: ## Run frontend unit tests with Vitest
	$(PNPM) --dir $(FRONTEND_DIR) test

frontend-lint: ## Run frontend linting with ESLint
	$(PNPM) --dir $(FRONTEND_DIR) lint

frontend-build: ## Build the production frontend bundle with Vite
	$(PNPM) --dir $(FRONTEND_DIR) build
