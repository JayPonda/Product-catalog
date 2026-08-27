.PHONY: mfile mup mseed mdown dev prod swagger dedup dedup-dry-range test coverage cover-html format lint

# 1. Load your existing .env file
ifneq (,$(wildcard .env))
    include .env
    export
endif

# 2. Handle mfile arguments
ifeq ($(firstword $(MAKECMDGOALS)),mfile)
  ARGUMENT := $(word 2, $(MAKECMDGOALS))
  %:
	@:
endif

# Create a migration file (scaffolds Up/Down markers)
mfile:
	@if [ -z "$(ARGUMENT)" ]; then \
		echo "Error: Please provide a migration name. Example: make mfile add_users_table"; \
		exit 1; \
	fi
	go run . migrate create $(ARGUMENT)

# Apply ALL pending migrations
mup:
	go run . migrate up

# One-time: register already-applied migrations (golang-migrate transition) so
# `up` won't replay them. Run once after switching tools, on an existing DB.
mseed:
	go run . migrate seed

# Rollback the single LAST migration (use `make mdown s=3` for more)
mdown:
	go run . migrate down $(if $(s),--steps $(s),)

# Start development server using Air
dev:
	air

# Run production-ready binary
prod:
	swag init -g main.go -o docs --parseDependency --parseInternal
	go build -o bin/server . && ./bin/server server

# Generate swagger docs
swagger:
	swag init -g main.go -o docs --parseDependency --parseInternal
	@echo "Swagger docs generated at docs/"

# Makefile for the Product Catalog server CLI.
#
# Run the duplicate-order removal command. Pass extra flags through ARGS, e.g.:
#   make dedup ARGS="--dry-run"
#   make dedup ARGS="--starttime 2024-01-01T00:00:00Z --endtime 2024-01-01T03:00:00Z"
#   make dedup ARGS="--window 2h30m --nearby 5m --batch 200"
dedup:
	go run . dedup-orders-remove $(ARGS)

# Dry-run over an explicit time range. Pass the bounds via START/END (RFC3339):
#   make dedup-dry-range START="2024-01-01T00:00:00Z" END="2024-01-01T03:00:00Z"
dedup-dry-range:
	go run . dedup-orders-remove --dry-run --starttime $(START) --endtime $(END)

# Run all backend unit + e2e tests.
test:
	go test ./... -count=1

# Run backend tests with a coverage report (function-level).
# -coverpkg attributes statements executed from other packages (e.g. controllers
# and middleware exercised via the routes E2E tests) to their own package.
COVPKGS := ./src/repositories/...,./src/services/...,./src/controllers/...,./src/middleware/...,./src/routes/...,./utils/...

coverage:
	go test ./... -count=1 -coverpkg='$(COVPKGS)' -coverprofile=cover.out
	go tool cover -func=cover.out | tail -1
	go tool cover -func=cover.out

# Generate an HTML coverage report and open it.
cover-html: coverage
	go tool cover -html=cover.out -o cover.html

# Run formatting on backend files
format:
	go fmt ./...

# Run linter on backend files
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1 run ./...

