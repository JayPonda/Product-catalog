.PHONY: mfile mup mdown mforce dev prod

# 1. Load your existing .env file
ifneq (,$(wildcard .env))
    include .env
    export
endif

# 2. Handle empty DB_SSLMODE fallback
ifeq ($(DB_SSLMODE),)
    SSLMODE := disable
else
    SSLMODE := $(DB_SSLMODE)
endif

# 3. Construct PostgreSQL connection string
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(SSLMODE)

# 4. Handle mfile arguments
ifeq ($(firstword $(MAKECMDGOALS)),mfile)
  ARGUMENT := $(word 2, $(MAKECMDGOALS))
  %:
	@:
endif

# 5. Handle mforce arguments
ifeq ($(firstword $(MAKECMDGOALS)),mforce)
  FORCE_VERSION := $(word 2, $(MAKECMDGOALS))
  %:
	@:
endif

# Create a migration file
mfile:
	@if [ -z "$(ARGUMENT)" ]; then \
		echo "Error: Please provide a migration name. Example: make mfile add_users_table"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations $(ARGUMENT)

# Apply ALL pending migrations
mup:
	migrate -path migrations -database "$(DB_URL)" -verbose up

# Force migration version - DEVELOPMENT ONLY
mforce:
	@if [ -z "$(FORCE_VERSION)" ]; then \
		echo "Error: Please provide a migration version."; \
		echo "Example: make mforce 20260818170000"; \
		exit 1; \
	fi
	@if [ "$(APP_ENV)" != "local" ]; then \
		echo "ERROR: mforce is only allowed in local."; \
		echo "Current APP_ENV=$(APP_ENV)"; \
		exit 1; \
	fi
	@echo "WARNING: Forcing migration version to $(FORCE_VERSION)..."
	migrate -path migrations -database "$(DB_URL)" -verbose force $(FORCE_VERSION)

# Rollback the single LAST migration
mdown:
	migrate -path migrations -database "$(DB_URL)" -verbose down 1

# Start development server using Air
dev:
	air

# Run production-ready binary
prod:
	go build -o /tmp/final . && /tmp/final
