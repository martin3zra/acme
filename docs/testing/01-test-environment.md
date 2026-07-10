# Test Environment

Everything the suite needs to run, and how it bootstraps itself.

## Prerequisites

1. **Postgres** reachable per `.env.test`. The sample points at `localhost:5433`
   (a dedicated instance, separate from the dev `acme` database).
2. **`camel`** on `PATH` (or at `~/go/bin/camel`) — the migration tool that
   builds the test schema. Install: `go install github.com/martin3zra/camel/cmd/camel@latest`.
3. **Go** matching `go.mod` (1.25.x).

## Configure

```
cp .env.test.sample .env.test
```

`.env.test` (never the dev `.env`) drives the test database connection:

| Var | Sample | Notes |
|---|---|---|
| `DB_NAME` | `acme_test` | Must not be `acme` — `TestMain` refuses to run against the dev DB |
| `DB_HOST` / `DB_PORT` | `localhost` / `5433` | The test Postgres |
| `DB_USERNAME` / `DB_PASSWORD` | `postgres` / `secret` | |
| `APP_KEY` | `base64:…` | A fixed non-secret test key |
| `APP_ENV` | `test` | |

## What `TestMain` does (`app/main_test.go`)

Before any test runs, `TestMain`:

1. **Loads `.env.test`** from the repo root or `../` (tests run from `app/`).
2. **Guards the dev DB** — aborts if `DB_NAME=acme`.
3. **Ensures `acme_test` exists** (`ensureTestDatabase`) — creates it if missing,
   connecting to the admin `postgres` database.
4. **Runs `camel migrate`** (`runMigrations`) against the test DSN — idempotent,
   so the schema is always current. Camel runs from the repo root where
   `camel.yaml` and `db/migrations` live; `DB_SOURCE` is passed explicitly.
5. **Registers the `txdb` driver** — `txdb.Register("txdb", "postgres", testDSN)`.
   Every `sql.Open("txdb", …)` after this is a transaction against `acme_test`
   that rolls back on `Close`. This is the isolation primitive
   ([02-fixtures.md](02-fixtures.md)).

You never create or migrate the test DB by hand for a normal run — `TestMain`
handles it on demand.

## Running

```
make test                     # go test ./...
go test ./app/                # just the app package
go test ./app/ -run TestFoo   # one test
go test -race ./app/          # race detector (CI runs this; ~3x slower)
make test-unit                # DB-free subset (see 04-unit-tests.md)
make test-db                  # force-(re)create + migrate acme_test by hand
```

## In CI (`.github/workflows/ci.yml`)

The `go` job spins up a `postgres:17` service mapped to host port **5433** (to
match `.env.test.sample`), builds the frontend (the binary embeds `public/build`),
installs a pinned `camel`, copies `.env.test.sample` to `.env.test`, then runs
`go test -race ./...`. `TestMain` creates and migrates `acme_test` inside that
service.

## Related files

- `app/main_test.go` — `TestMain`, `loadTestEnv`, `ensureTestDatabase`, `runMigrations`
- `.env.test.sample`
- `camel.yaml`, `db/migrations/`
- `Makefile`
- `.github/workflows/ci.yml`
