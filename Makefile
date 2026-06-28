.PHONY: test test-integration

# Fast unit tests — no database required.
test:
	go test ./...

# Integration tests — build a throwaway Postgres from the Camel baseline and
# exercise repository code against the real schema. Requires the acme Postgres
# (see .env DB_*) to be reachable; loads .env for connection settings.
test-integration:
	@set -a; . ./.env; set +a; go test -tags integration ./app/ -run Integration -count=1 -v
