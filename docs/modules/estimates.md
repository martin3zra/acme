# Estimates

An estimate (quote) is a **sales invoice in `estimate` mode** — same document,
same code, a different `transaction_kind`. It proposes prices to a customer
without billing them: no NCF, no receivable, no stock movement.

## For users

### What it is

A non-binding quote. You build it exactly like an invoice — customer, line
items, totals — but nothing financial happens: the customer isn't charged,
inventory isn't touched, and no tax receipt is consumed.

### Before you can create one

Same prerequisites as an invoice — an enabled **customer** and at least one
**item** with a **unit** per line — but an estimate consumes **no NCF**, so no
tax-receipt sequence is needed (and unlike an order, no `terms`; a `date` is
required). See [invoices.md](invoices.md#before-you-can-create-one) for the full
list.

### Typical flow

1. Create an estimate for a customer with the lines you're proposing.
2. Send / print it (signed print link, same as invoices).
3. If the customer accepts, **convert it to an invoice** — that's when billing,
   the receivable (for credit terms), and stock movement happen.

### Gotchas

- Estimates **cannot recur** (that's invoice-only).
- An estimate on its own never affects the customer's balance or inventory —
  conversion does.

## For developers

Estimates are the invoice module in a different mode. Everything — tables,
handlers, repository, model — is shared with [invoices.md](invoices.md); only the
`transaction_kind` (`estimate`) and the permission scope differ.

- **Routes** (`app/route.go`): `GET /estimates`, `GET /estimates/create`,
  `POST /estimates`, `GET/PUT /estimates/:id`, `PUT /estimates/:id/void`,
  `GET /estimates/:id/print/:hash`. All bound to the shared invoice handlers.
- **Kind**: `TransactionKinds.Estimate` (`app/types.go`); branching in
  `app/invoice-handlers.go`.
- **Create rules**: the shared `StoreInvoiceForm.Rules`
  (`app/invoice-types.go:227`) — for `estimate`, `date` is required and
  `tax_receipt` is **not**. See the table in
  [invoices.md](invoices.md#required-fields--dependencies).
- **Conversion**: estimate → invoice is where financial effects apply; the
  `source` jsonb on the new invoice records the origin. See
  `app/flow_conversion_test.go`.
- **Permissions**: `viewAny/create/update/void:estimate`
  ([permissions.md](permissions.md)).

For the data model, line handling, and all shared rules, see
**[invoices.md](invoices.md)** — this doc only records what makes the `estimate`
mode different.

## Related

[invoices.md](invoices.md) · [sales-orders.md](sales-orders.md) ·
[customers.md](customers.md)
