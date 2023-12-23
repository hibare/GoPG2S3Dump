SHELL=/bin/bash

UI := $(shell id -u)
GID := $(shell id -g)
MAKEFLAGS += -s
DOCKER_COMPOSE_PREFIX = HOST_UID=${UID} HOST_GID=${GID} docker-compose -f docker-compose.dev.yml

# Bold
BCYAN=\033[1;36m
BBLUE=\033[1;34m

# No color (Reset)
NC=\033[0m

.DEFAULT_GOAL := help

.PHONY: pg-up
pg-up: ## Start Postgres
	${DOCKER_COMPOSE_PREFIX} up -d postgres
	@echo "Waiting for POstgres to become healthy..."
	@until docker exec postgres pg_isready > /dev/null 2>&1; do \
		sleep 1; \
		echo "Postgres is not healthy yet, retrying..."; \
	done
	@printf "Postgres is now healthy!\n\n"

.PHONY: pg-down
pg-down: ## Stop Postgres
	${DOCKER_COMPOSE_PREFIX} rm -fsv postgres
	
.PHONY: clean
clean: ## Clean up
	${DOCKER_COMPOSE_PREFIX} down
	go mod tidy

.PHONY: test
test: ## Run tests
ifndef GITHUB_ACTIONS
	$(MAKE) pg-up
endif
	go test ./... -cover
ifndef GITHUB_ACTIONS
	$(MAKE) pg-down
endif

.PHONY: help
help: ## Disply this help
		@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BCYAN)%-18s$(NC)%s\n", $$1, $$2}'