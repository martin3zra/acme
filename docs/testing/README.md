# Testing Documentation

How testing works in this repo. Start with the strategy, then dip into whichever
layer you need.

1. [Testing Strategy](00-testing-strategy.md) — the philosophy and taxonomy.
2. [Test Environment](01-test-environment.md) — prerequisites, `.env.test`, how
   the suite bootstraps itself, running locally and in CI.
3. [Fixtures & Harness](02-fixtures.md) — the per-test server, transaction
   isolation, assertions.
4. [Test Data](03-test-data.md) — the tenant fixture, builders, factories, faker.
5. [Unit Tests](04-unit-tests.md) — the DB-free slice.
6. [Integration (Flow) Tests](05-integration-tests.md) — the dominant style.
7. [API Tests](06-api-tests.md) — *gap*: what exists, recommended approach.
8. [Playwright / E2E](07-playwright-tests.md) — *gap*: not adopted, a path to it.

Module documentation (per business area) lives in [`../modules/`](../modules/)
— added in a later batch.

## TL;DR to run tests

```
cp .env.test.sample .env.test   # once; needs Postgres on :5433 and `camel`
make test                        # full suite
go test -race ./app/             # what CI runs
make test-unit                   # DB-free subset
```
