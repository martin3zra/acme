# Module Documentation

One doc per business area. Each has a **For users** section (what it does, real
workflows, states, gotchas) and a **For developers** section (data model, routes,
repository entry points, rules, permissions, tests).

## Available now

- [Customers](customers.md) — who you sell to; terms, credit limit, opening balance
- [Vendors](vendors.md) — who you buy from; the buy-side mirror of customers
- [Invoices](invoices.md) — the core sales document; **canonical doc** for the
  sales-transaction family
- [Estimates](estimates.md) — quotes; invoices in `estimate` mode
- [Sales Orders](sales-orders.md) — confirmed orders; invoices in `order` mode

## Planned (later batches)

- Receivables — credit invoices owed by customers
- Payments — customer payments and allocation
- Purchases — purchase orders, receipts, vendor bills
- Inventory — items, variants, warehouses, stock movement
- Taxes — tax rates and fiscal tax-receipt (NCF) sequences
- Reports — sales and other reporting
- Permissions — roles and the ACL (`viewAny/create/update/...:module`)

> Cross-references to planned modules are forward-links: they resolve once those
> batches land.

For how the code is tested, see [`../testing/`](../testing/README.md).
