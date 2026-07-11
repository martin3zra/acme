# Spec: `app/` File Organization

## Goal

Keep the flat `package app` navigable as it grows by splitting oversized files
along a consistent, predictable naming scheme — so any concern (an invoice's
create path, a payment's DTO, a status enum) lives where its name says it should.

This is **file organization, not an architecture change**. The package stays flat,
logic stays on the `*Server` receiver, and no behavior changes. Every split in this
spec is a same-package move: zero import churn, the compiler is the guardrail.

## Convention

`<domain>-<kind>.go`, hyphenated. Every domain repeats the same `<kind>` suffixes.
This already exists across the 76 files in `app/`:

| Kind | Purpose | Examples |
|---|---|---|
| `*-handlers.go` | HTTP handlers (`*Server` methods) | `invoice-handlers.go`, `item-handlers.go` |
| `*-repository.go` | DB reads/writes (`*Server` methods) | `invoice-repository.go`, `company-repository.go` |
| `*-types.go` | Form DTOs, enums, JSON value-objects | `invoice-types.go`, `payment-types.go` |
| `*-mapper.go` | DTO ↔ model mapping | `invoice-mapper.go`, `item-mapper.go` |
| `*-pdf.go` | PDF rendering | `invoice-pdf.go`, `payment-pdf.go` |
| `*-events.go` / `*-scheduler.go` | domain events, background jobs | `invoice-events.go`, `invoice-scheduler.go` |

New `<kind>` suffixes are introduced below to split the largest repositories by
concern, mirroring how `invoice-pdf.go` / `invoice-scheduler.go` / `invoice-events.go`
are already carved out of the invoice domain.

## Done

### `types.go` split (this round)

`types.go` was a 2367-line grab-bag of every domain's Form DTOs, enums, and JSON
value-objects. It was split into 14 `*-types.go` files by domain, leaving a
~350-line residual `types.go` holding only the shared identity/date/prereq kernel
(`User` + methods, `Role`/`RoleMap`, sequences, `Date`, `Missing`,
`PrerequisiteResult`, prereq cache).

Splitting rules applied (carry these forward to future moves):

1. An enum travels with its registry var (`InvoiceStatus` + `InvoiceStatuses`).
2. A Form travels with its private helpers (`StoreInvoiceForm` + `computeTax`).
3. A JSON value-object keeps its `Value`/`Scan` and travels with its owning domain.
4. A value-object shared by two domains lives with its primary owner and is
   referenced freely (same package). `TransactionKind`/`TransactionSource` and
   `Line`/`Discount` live in `invoice-types.go`; `purchase-types.go` uses them.

## Backlog (same convention, not yet done)

Split the oversized repositories by concern. Each becomes several
`<domain>-<concern>.go` files, all still `package app`, all still `*Server` methods.
Target order (biggest first):

| File | Lines | Split into |
|---|---:|---|
| `purchase-repository.go` | 1213 | `purchase-create.go`, `purchase-update.go`, `purchase-receive.go`, `purchase-queries.go` |
| `inventory-repository.go` | 1090 | `inventory-movements.go`, `inventory-adjustments.go`, `inventory-transfers.go`, `inventory-queries.go` |
| `company-repository.go` | 965 | `company-create.go`, `company-update.go`, `company-queries.go` |
| `invoice-repository.go` | 930 | `invoice-create.go`, `invoice-update.go`, `invoice-void.go`, `invoice-queries.go` |

Also pending:

- **Relocate the orphaned invoice email send.** It currently sits in the generic
  store flow at `server.go` (`mail.NewInvoiceMail(...)`), not in an invoice file.
  Move it into an invoice-owned file (e.g. `invoice-mail.go`).
- `playsql_models.go` (2088 lines) — the other giant. Out of scope until the
  repository splits land; revisit whether to split it per domain or leave it as the
  generated-style model registry.

## Non-goals

- No new packages or sub-packages. `app/` stays flat (`mail` remains the sole
  subpackage).
- No underscore renames of existing files — hyphens throughout.
- **No service layer, no repository interfaces.** Handlers keep calling `*Server`
  methods directly. Introducing indirection is a separate decision, explicitly not
  part of this reorg.
- No behavior, signature, or exported-name changes in any split.

## Verification (for each future split)

Same as the `types.go` split — a pure same-package move:

1. `gofmt -l app/` — clean.
2. `go build ./...` and `go vet ./app/...` — pass.
3. `go test ./app/...` — unchanged.
4. `git diff` shows only moves; no `*_test.go` touched; every new file is `package app`.
