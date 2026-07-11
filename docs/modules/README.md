# Module Documentation

One doc per business area. Each has a **For users** section (what it does, real
workflows, states, gotchas) and a **For developers** section (data model, routes,
repository entry points, rules, permissions, tests).

## Sales

- [Invoices](invoices.md) — the core sales document; **canonical doc** for the
  sales-transaction family
- [Estimates](estimates.md) — quotes; invoices in `estimate` mode
- [Sales Orders](sales-orders.md) — confirmed orders; invoices in `order` mode
- [Receivables](receivables.md) — the AR ledger: credit invoices customers owe
- [Payments](payments.md) — customer payments and allocation

## Buying

- [Purchases](purchases.md) — purchase orders, receipts, vendor bills, payables
- [Vendors](vendors.md) — who you buy from; the buy-side mirror of customers

## Catalog & stock

- [Inventory](inventory.md) — items, variants, warehouses, stock movement,
  transfers, adjustments

## Parties & fiscal

- [Customers](customers.md) — who you sell to; terms, credit limit, opening balance
- [Taxes](taxes.md) — tax rates and fiscal tax-receipt (NCF) sequences
- [Reports](reports.md) — sales, profit & loss, expenses, taxes (PDF)
- [Permissions](permissions.md) — roles and the ACL (`action:module`)

For how the code is tested, see [`../testing/`](../testing/README.md).
