# Purchases

The buy side. Purchases move through three document kinds — **purchase order →
purchase receipt → vendor bill** — mirroring the sales invoice family. This is
the canonical doc for the purchase-transaction family and for paying vendors
(payables).

## For users

### The three kinds

All three are the same underlying document in a different `transaction_kind`:

- **Purchase order** — what you intend to buy from a vendor. No stock, no money.
- **Purchase receipt** — goods actually arrived; **stock comes in**. A receipt
  can be **partial** (some lines/quantities received now, the rest later).
- **Vendor bill** — the vendor's invoice to you; creates an **accounts-payable**
  entry (money you owe), settled by vendor payments.

### Lifecycle

A purchase carries a status: **draft → partially_received → received →
partially_paid → closed** (plus **posted**). Receiving advances the received
side; paying advances the paid side. You **confirm** a purchase to move it out of
draft.

### Paying vendors (payables)

The **payables** screen records a vendor payment against one or more open
accounts-payable entries, lowering what you owe that vendor — the buy-side mirror
of customer [payments](payments.md).

### Gotchas

- Only **receipts** move inventory; orders and bills don't.
- A vendor with an opening balance already has an AP entry from day one
  ([vendors.md](vendors.md)).
- Converting an order to a receipt/bill carries the origin forward.

## For developers

### Data model

- **`purchases`** — header. `transaction_kind`
  (`purchase_order`/`purchase_receipt`/`vendor_bill`), `purchase_status`,
  `vendor_id`, `warehouse_id`, `code`, dates, `subtotal`/`tax_amount`/…,
  `source` (jsonb, conversion origin). Model: `purchaseRead`.
- **`purchase_items`** — the lines. Model: `PurchaseItem` / `purchaseLineRead`.
- **`accounts_payable`** — the payable created by a vendor bill (and by vendor
  opening balances). `amount_payable` is a **generated** column. Model:
  `accountsPayableModel`.
- **`payables`** — the register row cross-referencing an AP entry
  (`payableRegister` / `payableRegisterRead`).
- **`vendor_payments`** + **`vendor_payment_items`** — a vendor payment and its
  allocations. Models: `vendorPaymentModel`, `vendorPaymentItemRead`.

Kinds/statuses: `PurchaseTransactionKinds` and `PurchaseStatuses` in
`app/types.go`.

### Routes (`app/route.go`)

Purchases: `GET /purchases` and `/purchases/orders|receipts|vendor-bills`
(list), `GET /purchases/create` (+ per-kind create), `POST /purchases`,
`GET /purchases/:id`, `GET /purchases/:id/edit`, `PUT /purchases/:id`,
`PUT /purchases/:id/confirm`, `DELETE /purchases/:id`.
Payables: `GET /payables` (list), `GET /payables/create`, `POST /payables`,
`PUT /payables/:id/void`, `GET /payables/:id`.

### Key rules

- **Receiving moves stock in** — a receipt records inventory movements and bumps
  balances ([inventory.md](inventory.md)); partial receipts move only what
  arrived (`flow_partial_receipt_test.go`).
- **Vendor bill → AP** — creates an `accounts_payable` row (map insert;
  `amount_payable` left to the generated column) and a `payables` register row.
- **Vendor payment** — `storeVendorPayment` allocates across AP entries and
  calls `updateAPBalance` (a self-referencing `amount_paid += …`, kept raw) and
  `updateVendorAmountPayable`.
- **Conversion** PO → receipt/bill records the origin in `source`
  (`flow_purchase_conversion_test.go`).
- **Locking** — AP/balance updates take row locks to stay correct under
  concurrency (see the transfer/lock tests for the same pattern).

### Permissions

`viewAny/create/update/confirm/delete:purchase`; payables:
`viewAny/create/void:payable` ([permissions.md](permissions.md)).

### Entry points & tests

- Handlers: `app/purchase-handlers.go`, `app/payable-handlers.go`
- Repositories: `app/purchase-repository.go`, `app/payable-repository.go`
- Builder: `app/purchase_builder_test.go`
- Tests: `app/flow_purchase_test.go`, `flow_purchase_writes_test.go`,
  `flow_purchase_reads_test.go`, `flow_purchase_conversion_test.go`,
  `flow_purchase_lines_playsql_test.go`, `flow_purchase_variant_lines_test.go`,
  `flow_partial_receipt_test.go`, `flow_payable_reads_test.go`

## Related

[vendors.md](vendors.md) · [inventory.md](inventory.md) ·
[payments.md](payments.md) · [invoices.md](invoices.md)
