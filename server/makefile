.PHONY: mfile mup mdown dev prod

# 1. Load your existing .env file 
ifneq (,$(wildcard .env))
    include .env
    export
endif

# 2. Handle empty DB_SSLMODE fallback (default to disable if empty in .env)
ifeq ($(DB_SSLMODE),)
    SSLMODE := disable
else
    SSLMODE := $(DB_SSLMODE)
endif

# 3. Construct the PostgreSQL connection string from your variables
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(SSLMODE)

# 4. ONLY intercept arguments if the first command typed is 'mfile'
ifeq ($(firstword $(MAKECMDGOALS)),mfile)
  ARGUMENT := $(word 2, $(MAKECMDGOALS))
  %:
	@:
endif

# Create a migration file for this project
mfile: 
	@if [ -z "$(ARGUMENT)" ]; then \
		echo "Error: Please provide a migration name. Example: make mfile add_users_table"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations $(ARGUMENT)

# Apply ALL pending migrations
mup:
	migrate -path migrations -database "$(DB_URL)" -verbose up

# Rollback the single LAST migration
mdown:
	migrate -path migrations -database "$(DB_URL)" -verbose down 1

# Start development server using Air
dev:
	air

# Run production-ready binary
prod:
	go build -o /tmp/final . && /tmp/final
