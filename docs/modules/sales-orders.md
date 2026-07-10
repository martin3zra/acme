# Sales Orders

A sales order is a **sales invoice in `order` mode** — the same document with a
different `transaction_kind`. It records a customer's confirmed order before it
is billed: no NCF, no receivable, no stock movement until it becomes an invoice.

## For users

### What it is

A confirmed order to fulfil. Like an estimate it is built with a customer and
line items, but it represents a commitment to deliver rather than a quote. It
doesn't bill the customer or move inventory on its own.

### Typical flow

1. Create the order for a customer with the agreed lines.
2. Fulfil / print it (signed print link, same as invoices).
3. **Convert to an invoice** to bill the customer — that applies the receivable
   (credit terms) and the stock movement.

### Gotchas

- Sales orders **cannot recur**.
- Inventory and the customer balance are only affected at conversion, not when
  the order is created.

## For developers

Sales orders are the invoice module in `order` mode; tables, handlers,
repository, and model are shared with [invoices.md](invoices.md).

- **Routes** (`app/route.go`): `GET /orders`, `GET /orders/create`,
  `POST /orders`, `GET/PUT /orders/:id`, `PUT /orders/:id/void`,
  `GET /orders/:id/print/:hash` — all on the shared invoice handlers.
- **Kind**: `TransactionKinds.Order` (`app/types.go`); branching in
  `app/invoice-handlers.go:308`.
- **Conversion** to an invoice carries the origin in the `source` jsonb; see
  `app/flow_conversion_test.go`.
- **Permissions**: `viewAny/create/update/void:order`
  ([permissions.md](permissions.md)).

For the data model, lines, and shared rules, see **[invoices.md](invoices.md)**.

## Related

[invoices.md](invoices.md) · [estimates.md](estimates.md) ·
[customers.md](customers.md) · [inventory.md](inventory.md)
