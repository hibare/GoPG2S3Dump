SHELL=/bin/bash

UI := $(shell id -u)
GID := $(shell id -g)
MAKEFLAGS += -s
DOCKER_COMPOSE_PREFIX = HOST_UID=${UID} HOST_GID=${GID} docker-compose -f docker-compose.dev.yml

all: pg-up

pg-up:
	${DOCKER_COMPOSE_PREFIX} up -d postgres
	@echo "Waiting for POstgres to become healthy..."
	@until docker exec postgres pg_isready > /dev/null 2>&1; do \
		sleep 1; \
		echo "POstgres is not healthy yet, retrying..."; \
	done
	@printf "POstgres is now healthy!\n\n"

pg-down:
	${DOCKER_COMPOSE_PREFIX} rm -fsv postgres
	
clean: 
	${DOCKER_COMPOSE_PREFIX} down
	go mod tidy

test: 
ifndef GITHUB_ACTIONS
	$(MAKE) pg-up
endif
	go test ./... -cover
ifndef GITHUB_ACTIONS
	$(MAKE) pg-down
endif

.PHONY = all clean test pg-up pg-down