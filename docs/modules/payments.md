# Payments

Recording money received from customers and applying it to what they owe.

## For users

### What it is

A payment is cash (or card/check/bank transfer) a customer gives you, allocated
across one or more of their outstanding **credit invoices**
([receivables.md](receivables.md)). Recording a payment reduces the invoices'
balances and the customer's overall **amount due**.

### Recording one

1. Pick the customer. Their open credit invoices (receivables) are listed.
2. Enter the total received and how it breaks down by method (cash, card, check,
   bank transfer).
3. Allocate the amount across invoices. Each allocation lowers that invoice's
   remaining balance; fully-covered invoices flip to paid.
4. Save. The payment gets a document code and a signed print link.

### States & gotchas

- **Void** — a payment can be voided, which reverses every allocation (invoices
  go back to owing) and restores the customer's balance. A voided payment can't
  be edited.
- Allocations can't exceed what an invoice still owes.
- This module is the **sell side**. The buy-side equivalent — paying vendors — is
  "payables" / vendor payments, covered in [purchases.md](purchases.md).

## For developers

### Data model

- **`receivables_income`** — the payment header (`uuid`, `code`, `customer_id`,
  `date`, `amount`, `payment` jsonb = method breakdown, `status`). Model:
  `paymentModel` (`app/playsql_models.go`).
- **`receivables_income_items`** — one allocation row per invoice
  (`receivable_income_id`, `invoice_id`, `amount_due`, `payment_amount`). Model:
  `paymentItemRead` / `paymentItem`.

### Routes (`app/route.go`)

`GET /payments` (list), `GET /payments/create`, `POST /payments`,
`GET /payments/:id/edit`, `PUT /payments/:id`, `PUT /payments/:id/void`,
`GET /payments/:id/print/:hash` (signed).

### Key rules

- **Allocation** — `attachPaymentLines` inserts the allocation rows and calls
  `updateInvoiceBalance` per invoice, then `updateCustomerAmountDue` for the
  customer total. All inside one transaction (`storePayment`).
- **Void** — `voidPayment` walks the allocation rows and gives each invoice's
  balance back, then restores the customer balance.
- `payment` is jsonb (the per-method split); decoded in the mapper.
- Tenant-scoped on `company_id`; `uuid` DB-generated.

### Permissions

`viewAny/create/update:payment`, plus the signed print route
([permissions.md](permissions.md)).

### Entry points & tests

- Handlers: `app/payment-handlers.go`
- Repository: `app/payment-repository.go`
- Tests: `app/flow_payment_test.go`, `flow_payment_reads_test.go`,
  `flow_void_test.go`, `flow_receivables_test.go`

## Related

[receivables.md](receivables.md) · [invoices.md](invoices.md) ·
[customers.md](customers.md) · [purchases.md](purchases.md)
