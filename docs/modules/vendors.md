# Vendors

The suppliers you buy from. A vendor is the counterparty on purchases, vendor
bills, and vendor payments (payables). It mirrors [customers.md](customers.md) on
the buy side.

## For users

### What it is

A vendor record holds contact details, purchasing preferences (terms, payment
method, lead time), a purchase note, and a running **amount payable** (what you
currently owe them).

### Before you can create one

Like a customer, a vendor is a base record — nothing else needs to exist first.
The **email** must be unique across your vendors, and recording an **opening
balance** starts an accounts-payable entry from day one. Assigning a fiscal
**tax-receipt** is optional but, if set, must be a real sequence
([taxes.md](taxes.md)).

### Creating one

1. Enter name/contact/email/phone and pick **individual** or **business**.
2. Set payment terms/method and, optionally, a **lead time** (days) and a
   purchase note.
3. Optionally record an **opening balance**: money you already owed the vendor
   before adopting the system. This creates an opening accounts-payable entry so
   your payables start correct.

The vendor gets a sequential **code** automatically.

### States & lifecycle

- **Enabled / Disabled** — change-status; disabled vendors drop out of the vendor
  picker.
- **Deleted** — soft delete; hidden from lists/lookups, kept for audit; editing
  or re-deleting a deleted vendor is not-found.

### Gotchas

- The opening balance is a one-time starting liability; it shows up in payables
  immediately ([receivables.md](receivables.md) covers the customer side; the
  vendor side lives under payables/purchases).

## For developers

### Data model

**`vendors`** — `uuid` (DB-generated, guarded), `code`, `name`, `contact_name`,
`email`, `phone`, `status`, `amount_payable`, `purchase_note`, `lead_time_days`,
`address`, `vendor_type`, `payment_method`, `payment_terms`, `deleted_at`
(soft-delete). Model: `vendorModel` in `app/playsql_models.go`.

### Required fields & dependencies

Enforced by `StoreVendorForm.Rules` (`app/vendor-types.go:30`).

| Field | Rule | Note |
|---|---|---|
| `name` | required, 3–120 | — |
| `email` | required, `email`, lowercase, unique(vendors) | — |
| `credit_limited` | required | boolean |
| `tax_receipt` | `sometimes`, `tenantExists` tax_receipts | an NCF sequence, only if assigned |

### Routes (`app/route.go`)

`GET /vendors` (list/search), `POST /vendors`, `PUT /vendors/:id`,
`PUT /vendors/:id/change-status`, `DELETE /vendors/:id` (soft delete).

### Key rules

- **Code** from the per-company `vendor` sequence.
- **Opening balance** — `storeVendorOpenBalance` inserts an opening entry into
  `accounts_payable` (status `draft`/`pending`) and registers it in `payables`,
  in the same transaction as the vendor insert.
- **Soft-delete scoping** via the model's `softdelete` tag; mutations on a
  trashed vendor are not-found.
- **amount_payable** tracks what you owe, adjusted by purchases/bills and vendor
  payments.
- Tenant-scoped on `company_id`.

### Permissions

`viewAny/create/update/delete:vendor` ([permissions.md](permissions.md)).

### Entry points & tests

- Handlers: `app/vendor-handlers.go`
- Repository: `app/vendor-repository.go`
- Builder: `app/vendor_builder_test.go`
- Tests: `app/flow_vendor_reads_test.go`,
  `flow_customer_vendor_playsql_test.go`, `flow_opening_balance_test.go`

## Related

[purchases.md](purchases.md) · [payments.md](payments.md) ·
[customers.md](customers.md)
