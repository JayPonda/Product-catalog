.PHONY: mfile mup mseed mdown dev prod swagger

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
