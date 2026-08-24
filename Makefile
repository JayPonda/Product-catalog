.PHONY: test test-backend test-frontend coverage coverage-backend coverage-frontend lint lint-backend lint-frontend format format-backend format-frontend

# Run both backend and frontend test suites.
test: test-backend test-frontend

test-backend:
	$(MAKE) -C server test

test-frontend:
	pnpm --dir app exec vitest run

# Generate coverage reports for both sides.
coverage: coverage-backend coverage-frontend

coverage-backend:
	$(MAKE) -C server coverage

coverage-frontend:
	pnpm --dir app run coverage

# Run linting for both frontend and backend.
lint: lint-backend lint-frontend

lint-backend:
	$(MAKE) -C server lint

lint-frontend:
	pnpm --dir app run lint

# Format codebase for both frontend and backend.
format: format-backend format-frontend

format-backend:
	$(MAKE) -C server format

format-frontend:
	pnpm --dir app run format

