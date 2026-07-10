# Spec: Project Documentation Set (testing + modules)

## Goal

Author two hand-written documentation trees under `docs/`, drafted from the
actual codebase:

- **`docs/testing/`** — how testing works in this repo, for developers.
- **`docs/modules/`** — one doc per business module, each serving end users
  ("For users") and developers ("For developers").

Content is **descriptive**: it documents what exists today, grounded in real
files, routes, tables, and the 76 Go test files. Where a listed doc has no
implementation yet (API tests, Playwright), it ships as a short honest stub that
names the gap and a recommended approach — not invented content.

## Audience & scope

- Testing tree: contributors and QA. Technical.
- Module tree: split per document. "For users" = an operator of the ERP
  (workflows, states, gotchas). "For developers" = tables, routes, repository
  entry points, business rules, permissions, related tests.

## What exists today (grounding)

- Go backend (`app/`), React/Inertia frontend (`resources/js/`), Postgres via
  Camel migrations (`db/migrations`, `camel.yaml`).
- **76 `_test.go` files**, all Go, flow-style and table-driven. Isolation via
  `go-txdb` (one rolled-back transaction per test) — `harness_test.go`
  (`newTestServer`/`newHandlerServer`), `main_test.go` (`TestMain` creates and
  migrates `acme_test`).
- Test data via **builders and factories**: `customer_builder_test.go`,
  `invoice_builder_test.go`, `vendor_builder_test.go`, `purchase_builder_test.go`,
  `factories_test.go`, `faker_test.go`.
- Data access via **playsql** models (`app/playsql_models.go`), one model per
  table; reads through the Boot-cached plan, writes through `playTx`.
- `.env.test` / `.env.test.sample`, `Makefile` targets (`test`, `test-db`,
  `test-unit`), CI in `.github/workflows/ci.yml` (Go `test -race` + frontend gates).
- **No** Playwright, **no** REST API test layer, **no** frontend unit tests.
  ~41 JSON fetch endpoints exist alongside Inertia; 1 SSE stream.
- Existing docs: `docs/getting-started.md` only.

## Structure

```
docs/
  testing/
    00-testing-strategy.md    # philosophy: flow tests over units, txdb isolation, what is/not covered
    01-test-environment.md    # .env.test, Postgres:5433, camel, Makefile, TestMain lifecycle, running locally + CI
    02-fixtures.md            # harness_test.go: newTestServer/newHandlerServer, txdb, authCtx/handlerCtx, is-assertions
    03-test-data.md           # builders + factories + faker; how to seed a company/customer/invoice
    04-unit-tests.md          # the pure-unit slice (acl, mappers, config, import-headers); make test-unit
    05-integration-tests.md   # flow_* tests: the dominant style; repo+handler through real schema; examples
    06-api-tests.md           # GAP STUB: no API-test layer; the JSON endpoints that exist; recommended approach
    07-playwright-tests.md    # GAP STUB: no e2e today; recommended Playwright-on-Inertia setup + first-test sketch
  modules/
    customers.md  estimates.md  sales-orders.md  invoices.md  receivables.md
    payments.md   inventory.md  purchases.md     vendors.md   taxes.md
    reports.md    permissions.md
```

## Document conventions

**Testing docs** — descriptive, each ends with a "Related files" list of real
paths. `06`/`07` are one-screen gap stubs: what's missing, why, and a concrete
recommended path (framework, CI wiring), clearly marked as not-yet-adopted.

**Module docs** — every file has exactly two top sections:

```
# <Module>
## For users
  - What it is / when used
  - Primary workflow(s) and the states involved
  - Rules and gotchas a user would hit
## For developers
  - Data model (tables), key routes/handlers, repository file(s)
  - Load-bearing business rules (e.g. NCF sequencing, stock movement, soft-delete)
  - Permissions (subjects/actions)
  - Related tests
```

**Overlaps** — keep all 13 files (matches the request). Shared mechanics live in
one canonical doc and the others cross-link, no duplication:
- `estimates.md`, `sales-orders.md` are short: explain the `transaction_kind`
  mode, then link `invoices.md` for shared create/line/stock mechanics.
- `receivables.md` (credit invoice ↔ customer) and `payments.md` cross-reference.

**Accuracy** — docs cite `path:symbol`, not code copies, to limit drift. Each
module doc's "Related tests" points at the `flow_*_test.go` that pins its rules,
so a reader can verify behavior against green tests.

## Delivery plan

Batched PRs off `main` (repo squash-merges; no stacking):
1. `docs/testing/` (8 files) — one PR.
2. Modules batch A: customers, vendors, invoices, estimates, sales-orders.
3. Modules batch B: payments, receivables, purchases, inventory.
4. Modules batch C: taxes, reports, permissions.

Each PR is self-contained and reviewable; later batches branch off updated `main`.

## Non-goals

- No generated/auto-extracted docs tooling — hand-written now.
- No invented content for API/Playwright; those are honest gap stubs.
- No restructuring of code to match docs.
- No end-user screenshots/video; text + workflow descriptions only (revisit later).
- Not changing `docs/getting-started.md` (may cross-link it).

## Open questions

- **Index/nav**: add a top-level `docs/README.md` linking both trees? (Assume yes
  unless told otherwise.)
- **Module depth**: target length per module doc — tight one-pager vs deep
  reference? (Assume ~1–2 screens; deepen invoices/inventory/purchases where the
  rules are heaviest.)
- **Permissions doc source of truth**: derive from `acl_test.go` + the ACL
  definitions; confirm that is the authoritative list of subjects/actions.
