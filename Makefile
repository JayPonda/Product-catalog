.PHONY: test test-backend test-frontend coverage coverage-backend coverage-frontend

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
