# acme — common dev/test tasks.
#
# Prereqs for tests:
#   - Postgres reachable per .env.test (cp .env.test.sample .env.test)
#   - `camel` on PATH (or ~/go/bin/camel) to build the test schema
#   - GOPRIVATE set for the private forge module
export GOPRIVATE := github.com/martin3zra/*

.PHONY: test test-db test-unit lint

# Full Go test suite (TestMain creates + migrates acme_test automatically).
test:
	go test ./...

# Force-(re)create and migrate the acme_test database. Optional — TestMain also
# does this on demand; useful to rebuild the schema by hand.
test-db:
	@set -a; . ./.env.test; set +a; \
	export DB_SOURCE="postgres://$$DB_USERNAME:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSLMODE"; \
	psql "postgres://$$DB_USERNAME:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/postgres?sslmode=$$DB_SSLMODE" -c "CREATE DATABASE $$DB_NAME" 2>/dev/null || true; \
	camel migrate

# Only the pure unit tests (no DB needed).
test-unit:
	go test ./app/ -run 'Test(Map|Get|Parse|Normalize|Detect|NextOccurrence|MapHeaders)' ./...
