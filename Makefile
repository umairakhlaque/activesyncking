.PHONY: all build test lint clean docker-up docker-down migrate-up migrate-down dev

BINARY_DIR   := bin
AUTHSVC      := $(BINARY_DIR)/authsvc
GATEWAY      := $(BINARY_DIR)/gateway
ADMINSVC     := $(BINARY_DIR)/adminsvc

GO           := go
GOFLAGS      :=
LDFLAGS      := -ldflags "-s -w -X main.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)"

all: build

## build: compile all service binaries
build: $(AUTHSVC) $(GATEWAY) $(ADMINSVC)

$(AUTHSVC):
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $@ ./cmd/authsvc

$(GATEWAY):
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $@ ./cmd/gateway

$(ADMINSVC):
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $@ ./cmd/adminsvc

## test: run all unit tests
test:
	$(GO) test -race -timeout 120s ./...

## test-cover: run tests with coverage report
test-cover:
	$(GO) test -race -timeout 120s -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## vet: run go vet
vet:
	$(GO) vet ./...

## tidy: tidy and verify modules
tidy:
	$(GO) mod tidy
	$(GO) mod verify

## clean: remove build artifacts
clean:
	rm -rf $(BINARY_DIR) coverage.out coverage.html

## docker-up: start local development stack
docker-up:
	docker compose -f docker-compose.yml up -d

## docker-down: stop local development stack
docker-down:
	docker compose -f docker-compose.yml down

## docker-logs: tail all container logs
docker-logs:
	docker compose -f docker-compose.yml logs -f

## migrate-up: apply all pending migrations
migrate-up:
	@./scripts/migrate.sh up

## migrate-down: roll back last migration
migrate-down:
	@./scripts/migrate.sh down

## dev-authsvc: run auth service with hot reload (requires air)
dev-authsvc:
	air -c .air.authsvc.toml

## seed: seed development database
seed:
	$(GO) run ./scripts/seed/main.go

## web-install: install frontend dependencies
web-install:
	cd web/admin && npm install

## web-dev: run frontend dev server
web-dev:
	cd web/admin && npm run dev

## web-build: build frontend for production
web-build:
	cd web/admin && npm run build

help:
	@echo ''
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
