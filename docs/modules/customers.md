# Customers

The people and businesses you sell to. A customer is the counterparty on
invoices, estimates, orders, and payments.

## For users

### What it is

A customer record holds contact details, billing preferences (cash vs credit
terms, credit limit), a fiscal tax-receipt assignment, and a running
**amount due** (what they currently owe you).

### Before you can create one

A customer is a starting point — nothing else needs to exist first. Two
constraints to know:

- The **email** must be unique across your customers.
- If you record an **opening balance**, you must also give the **as-of date** it
  applied on.
- Assigning a fiscal **tax-receipt** is optional, but if set it must be a real
  sequence ([taxes.md](taxes.md)).

### Creating one

1. Enter name/contact/email/phone and pick **individual** or **business**.
2. Set payment terms — **cash ("pia")** or **credit ("netN")** — and, for credit
   customers, an optional **credit limit**.
3. Optionally record an **opening balance**: money the customer already owed you
   before you started using the system. This creates an opening invoice and a
   receivable so their balance starts correct.

The customer gets a sequential **code** automatically.

### States & lifecycle

- **Enabled / Disabled** — toggle a customer's status (change-status). A disabled
  customer is filtered from the "pick a customer" search.
- **Deleted** — a soft delete. The record is hidden from lists and lookups but
  kept for the audit trail; editing or re-deleting a deleted customer is treated
  as not-found.

### Gotchas

- A **credit limit** is enforced when you invoice on credit — an invoice that
  would push the customer over their limit is refused
  ([invoices.md](invoices.md)).
- Search matches by name and only returns **enabled** customers.

## For developers

### Data model

**`customers`** — `uuid` (DB-generated, guarded), `code`, `name`,
`contact_name`, `email`, `phone`, `status`, `amount_due`, `address`,
`customer_type`, `payment_method`, `payment_terms`, `credit_limited`,
`credit_limit`, `tax_receipt_id`, `deleted_at` (soft-delete). Model:
`customerModel` in `app/playsql_models.go`.

### Required fields & dependencies

Enforced by `StoreCustomerForm.Rules` (`app/customer-types.go:28`).

| Field | Rule | Note |
|---|---|---|
| `name` | required, 3–120 | — |
| `email` | required, `email`, lowercase, unique(customers) | — |
| `credit_limited` | required | boolean |
| `tax_receipt` | `sometimes`, `tenantExists` tax_receipts | an NCF sequence, only if assigned |
| `open_balance_as_of` | required when `open_balance` > 0 | runtime guard, `app/customer-handlers.go:69` |

### Routes (`app/route.go`)

`GET /customers` (list/search), `POST /customers`, `PUT /customers/:id`,
`PUT /customers/:id/change-status`, `DELETE /customers/:id` (soft delete).

### Key rules

- **Code** comes from the per-company `customer` sequence
  (`GetNextSequence(..., "customer")`).
- **Opening balance** — on create, a non-zero opening balance stores an opening
  invoice (`type = opening`, credit-invoice sequence) and registers a receivable
  in the same transaction (`storeCustomerInternal`).
- **Soft-delete scoping** — reads pick up `deleted_at IS NULL` from the model's
  `softdelete` tag; update/delete/toggle on a trashed customer are not-found.
- **amount_due** tracks the customer's outstanding balance, adjusted by invoices
  and payments.
- Tenant-scoped: every query filters `company_id`.

### Permissions

`viewAny/create/update/delete:customer` ([permissions.md](permissions.md)).

### Entry points & tests

- Handlers: `app/customer-handlers.go`
- Repository: `app/customer-repository.go`
- Builder: `app/customer_builder_test.go`
- Tests: `app/flow_customer_reads_test.go`,
  `flow_customer_vendor_playsql_test.go`, `flow_credit_limit_test.go`,
  `flow_opening_balance_test.go`

## Related

[invoices.md](invoices.md) · [receivables.md](receivables.md) ·
[payments.md](payments.md) · [vendors.md](vendors.md)
