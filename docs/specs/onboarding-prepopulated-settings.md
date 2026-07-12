# Onboarding — Pre-populated Company Settings

**Mode:** Technical spec
**Status:** Draft
**Date:** 2026-07-12

## Goal

Cut the manual setup between "create company" and "create first invoice" by
auto-seeding the settings a new company always needs. Today `createCompany`
already provisions **units**, **default sequences**, and **redirect
preferences**; the user still has to manually add a **tax**, a **warehouse**,
and **expense categories** before the daily cycle works (see
`docs/getting-started.md` A3–A5). This spec extends provisioning to cover those,
and mark tax + warehouse as the default selection so the first product and first
invoice are effectively one-click.

## Current state (what already happens on company create)

`app/company-repository.go` → `createCompany` runs one `WithTransaction`:

1. Insert `companies` row.
2. Insert `company_users` (`role=owner`, `current=true`).
3. `copySharedData` → `SELECT copy_shared_data(company_id)` — copies
   `shared_units` → `units`.
4. `linkCompanyDefaultSequences` — upserts `company_settings` with default
   `sequences` + `redirect_preferences` JSON.

Existing shared-data plumbing to mirror:
- `shared_units` + `copy_shared_data(company_id)` — the pattern to extend.
- `shared_tax_receipts` + `upsert_tax_receipts(company_id, receipts jsonb)` —
  NCF catalog is DGII types; the user supplies real serie/start/end ranges. This
  stays user-driven.

## Scope — settings to seed

| Setting | Seed? | Mark default? | Notes |
|---|---|---|---|
| **Tax — ITBIS 18%** | ✅ | ✅ default on product form | National DR rate; safe to seed. |
| **Warehouse — "General"** | ✅ | ✅ default stock location | Singleton per new company. |
| **Expense categories** | ✅ (starter set) | n/a | Common DR categories. |
| **NCF / tax receipts** | ❌ no DB row | n/a | Onboarding **checklist item** only; deep-links to Settings → Comprobantes. Real serials come from DGII — never fabricate. |

## Approach

### 1. Mechanism: extend `copy_shared_data` + new `shared_*` tables

Add company-agnostic seed tables and copy them inside the same stored proc, so
all list-like defaults flow through one call in the existing transaction:

- `shared_taxes` (name, rate) → copy into `taxes`. Seed row: `ITBIS`, `18.00`.
- `shared_warehouses` (name, location) → copy into `warehouses`. Seed row:
  `General`.
- `shared_expense_categories` (name) → copy into expense categories. Seed the
  starter set.

Extend `copy_shared_data(p_company_id)` to `INSERT ... SELECT` from each new
`shared_*` table, matching the existing `shared_units` block. No change to the
Go call site — `copySharedData` already runs it inside `createCompany`'s tx.

### 2. Defaulting: store default ids in `company_settings` JSON

Neither `taxes` nor `warehouses` has an `is_default` column, and there is no
default-warehouse concept today. Rather than add boolean columns to those hot
tables, record defaults in the existing `company_settings` row (which already
holds `sequences` + `redirect_preferences` JSON):

```
defaults: { "tax_id": <seeded>, "warehouse_id": <seeded> }
```

Because ids aren't known until the shared rows are copied, resolve them
**after** `copySharedData` and write them in a new `linkCompanyDefaults` step
(sibling to `linkCompanyDefaultSequences`), same transaction. Product-create and
invoice/stock code read `company_settings.defaults` to pre-select.

### 3. NCF checklist item

No DB row. Surface NCF configuration as a required onboarding to-do that
deep-links to Settings → Comprobantes, where the existing
`upsert_tax_receipts` flow lets the user enter real DGII ranges. Invoicing stays
blocked until they do — unchanged from today.

## Constraints

- **Transaction:** all seeding stays inside the single `createCompany`
  `WithTransaction`; a failure rolls back the whole company create.
- **Multi-tenant:** every seeded row is scoped by `company_id`; seed tables are
  global/company-agnostic.
- **Schema is Camel-managed** (`db/migrations`, YAML). New `shared_*` tables,
  their seed rows, and the `copy_shared_data` redefinition are new migrations.
  (See `[[camel-migrations]]`.)
- **No invalid comprobantes:** do not fabricate NCF numbers under any option.
- **playsql timestamps:** seed rows stamp `created_at`/`updated_at` via `NOW()`
  in the proc; don't compare them across rows (`[[playsql-timestamp-skew]]`).

## Non-goals

- Not adding `is_default` columns to `taxes`/`warehouses` (use settings JSON).
- Not auto-creating NCF/tax-receipt rows (DGII-owned; user-driven).
- Not seeding products, customers, or vendors (Part B / real business data).
- Not changing the onboarding form fields (name, RNC, city, address).

## Open questions

1. **Default-read wiring:** confirm every consumer (product form default tax,
   inventory/invoice default warehouse) reads `company_settings.defaults` — or
   do some paths need updating to honor it?
2. **Expense category starter set:** which exact categories? (Needs a
   DR-accounting-sensible list.)
3. **Editability:** if a user deletes the seeded tax/warehouse, should the
   stored default id be cleared/re-pointed, or just tolerated as dangling?
4. **Idempotency:** `copy_shared_data` currently assumes a fresh company — keep
   it single-shot on create, or make the new inserts `ON CONFLICT DO NOTHING`
   for safety on retries?
