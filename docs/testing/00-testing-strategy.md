# Testing Strategy

## The short version

This codebase is tested almost entirely through **Go integration tests that
drive real production code against a real Postgres schema**, inside a
transaction that rolls back after each test. There are ~76 test files in `app/`.
They favour end-to-end flows (store a customer → invoice them → take a payment →
assert balances) over isolated unit tests, because the value of this system is in
the business rules that span repositories, sequences, tenancy, and the ledger —
not in individual functions.

## Why this shape

- **The bugs live in the seams.** NCF sequencing, stock movement, soft-delete
  scoping, opening balances, tenant isolation — these only fail when several
  pieces interact. A flow test that runs the actual `storeInvoice` /
  `storePayment` path catches them; a mock-heavy unit test would not.
- **Real schema, real SQL.** Tests run against the same Camel migrations that
  build production, so a column rename or constraint change surfaces in tests.
- **Transaction-per-test isolation** (`go-txdb`) makes flow tests cheap: no
  truncation, no fixtures to tear down, full parallel-safe isolation. See
  [02-fixtures.md](02-fixtures.md).
- **Production paths, not raw INSERTs.** Test data is built through the same
  handlers/repositories users hit, via builders and factories
  ([03-test-data.md](03-test-data.md)), so setup itself exercises validation and
  invariants.

## What is covered

- Backend business logic: repositories, handlers, ACL, mappers, the scheduler,
  imports, inventory, the full sales and purchase ledgers.
- The playsql data layer (models, reads, writes) — see the merge history in
  `app/playsql_models.go`.

## What is NOT covered (today)

- **Frontend / React** — no component or unit tests. CI gates it with
  `eslint` + `prettier` + `tsc` only (`.github/workflows/ci.yml`).
- **End-to-end / browser** — no Playwright. See
  [07-playwright-tests.md](07-playwright-tests.md) (gap + recommendation).
- **A dedicated API-test layer** — the JSON fetch endpoints are exercised
  indirectly through handler/repo tests, not a contract suite. See
  [06-api-tests.md](06-api-tests.md).

## The test taxonomy in this repo

| Kind | Where | Needs DB | Doc |
|---|---|---|---|
| Pure unit | `acl_test.go`, `config_test.go`, `import-headers_test.go`, `mapper-helpers_test.go`, recurrence math | No | [04-unit-tests.md](04-unit-tests.md) |
| Integration / flow | `flow_*_test.go` (the majority) | Yes (txdb) | [05-integration-tests.md](05-integration-tests.md) |
| Infrastructure | `harness_test.go`, `main_test.go`, builders, factories, faker | — | [02](02-fixtures.md), [03](03-test-data.md) |

## Running

```
make test        # full suite (TestMain builds/migrates acme_test on demand)
make test-unit   # DB-free subset only
go test -race ./app/   # race detector; CI runs this
```

See [01-test-environment.md](01-test-environment.md) for prerequisites.

## Related files

- `app/main_test.go` — suite bootstrap
- `app/harness_test.go` — per-test server + assertions
- `Makefile` — `test`, `test-db`, `test-unit`
- `.github/workflows/ci.yml` — the gate
