# Invoices

Invoices are the core sales document. This is also the **canonical doc for the
sales-transaction family**: estimates ([estimates.md](estimates.md)) and sales
orders ([sales-orders.md](sales-orders.md)) are the same document in a different
mode and link back here for the shared mechanics.

## For users

### What it is

An invoice bills a customer for items. It can be **cash** (paid on the spot) or
**credit** (due later, which creates a receivable the customer owes). A fiscal
invoice also carries an **NCF** (a government tax-receipt number) drawn from a
tax-receipt sequence.

### Creating one

1. Pick a customer (must exist and be enabled — see [customers.md](customers.md)).
2. Add line items: item, quantity, unit price, tax rate. Totals compute as you go.
3. Choose terms:
   - **Cash / "pia"** — paid immediately; no receivable.
   - **Credit / "netN"** — a due date is set N days out and a **receivable** is
     recorded ([receivables.md](receivables.md)).
4. Save. The invoice gets a document code, and a fiscal invoice gets its NCF.

Selling an inventory-tracked item **moves stock out** of the warehouse.

### States

- **Draft / Sent / Paid** — the lifecycle of a normal invoice; a cash invoice is
  paid at creation, a credit invoice becomes paid once payments cover it.
- **Void** — reversed. Voiding gives back stock and unwinds the receivable; the
  row stays for the audit trail. A voided invoice can't be edited.

### Gotchas

- A credit invoice to a customer over their **credit limit** is refused
  ([customers.md](customers.md)).
- The NCF sequence is finite (per tax receipt); running it out blocks fiscal
  invoicing until a new sequence is configured ([taxes.md](taxes.md)).
- **Recurrence**: an invoice can be marked recurring; the scheduler then emits
  copies on a cadence (estimates and orders can't recur).
- **Print** links are signed URLs — they carry a hash and can be shared without a
  login.

## For developers

### Data model

- **`invoices`** — the header. Key columns: `transaction_kind`
  (`invoice`/`estimate`/`order`/`template`), `customer_id`, `code`, `date`,
  `due_on`, `amount`/`tax`/`total`/`amount_due`, `status`, `paid_status`,
  `tax_receipt_id` + `tax_receipt_sequence` + `tax_number` (NCF), `payment` and
  `discount` and `source` and `recurrence` (all jsonb), `movement_recorded`.
- **`invoices_items`** — the lines (item/variant, qty, unit, price, rate).
- **`receivables`** — links a credit invoice to its customer (created for credit
  terms; see [receivables.md](receivables.md)).

Model: `invoiceModel` in `app/playsql_models.go` (the merged read/write model;
jsonb columns are `[]byte`).

### The transaction-kind family

`/invoices`, `/estimates`, and `/orders` route to the **same handlers**
(`invoicesHandler`, `createInvoiceHandler`, `storeInvoiceHandler`,
`updateInvoiceHandler`, `voidInvoiceHandler`) and differ by `transaction_kind`
and permission. Kind-specific behaviour branches in `invoice-handlers.go` (e.g.
the post-store redirect at `invoice-handlers.go:304`). `TransactionKinds` is
defined in `app/types.go`. Estimates and orders do **not** touch inventory or
receivables; conversion to an invoice does (`flow_conversion_test.go`).

### Routes (`app/route.go`)

`GET /invoices` (list), `GET /invoices/create`, `POST /invoices`,
`GET /invoices/:id`, `GET /invoices/:id/edit`, `PUT /invoices/:id`,
`PUT /invoices/:id/void`, `POST /invoices/:id/recurrence`,
`GET /invoices/:id/print/:hash` (signed). Mirrored under `/estimates` and
`/orders`.

### Key rules

- **NCF / tax-receipt sequencing** — a fiscal invoice draws `tax_number` from a
  tax-receipt sequence; `tax_receipt_sequence` records which slot was used.
- **Stock movement** — an invoice for an inventory-tracked variant records a
  movement (`movement_recorded`) and decrements the balance
  ([inventory.md](inventory.md)).
- **Credit → receivable** — credit terms compute `due_on` and register a
  receivable; cash invoices settle immediately.
- **Void reverses** both the stock movement and the receivable.
- **uuid** is DB-generated and read back after insert (the store path uses a map
  insert; see the playsql merge history).

### Permissions

`viewAny/create/update/void` scoped per kind: `:invoice`, `:estimate`, `:order`.
See [permissions.md](permissions.md).

### Entry points & tests

- Handlers: `app/invoice-handlers.go`
- Repository: `app/invoice-repository.go`
- Tests: `app/flow_invoice_create_test.go`, `flow_invoice_lines_test.go`,
  `flow_invoice_update_stock_test.go`, `flow_invoice_guards_test.go`,
  `flow_sales_test.go`, `flow_conversion_test.go`, `flow_recurring_test.go`,
  `flow_credit_limit_test.go`, `flow_void_test.go`,
  `flow_invoice_variant_lines_test.go`

## Related

[customers.md](customers.md) · [receivables.md](receivables.md) ·
[payments.md](payments.md) · [inventory.md](inventory.md) · [taxes.md](taxes.md)
· [estimates.md](estimates.md) · [sales-orders.md](sales-orders.md)
