# Receivables

What customers owe you. A receivable is the link between a **credit invoice** and
the customer responsible for paying it — the accounts-receivable ledger.

## For users

### What it is

Whenever you issue a **credit invoice** (terms "netN"), the system records a
receivable: an outstanding amount that customer owes. Cash invoices don't create
one (they're paid on the spot). An **opening balance** on a new customer also
creates a receivable, so their starting debt is on the books.

### Where you see it

You don't manage receivables on a screen of their own — they surface as:

- the customer's **amount due**;
- the list of open invoices you allocate against when
  [recording a payment](payments.md);
- aging/owed figures in [reports](reports.md).

### Lifecycle

- **Open** — the invoice is unpaid or partly paid.
- **Settled** — payments have covered it.
- **Reversed** — if the invoice is voided, the receivable unwinds.

## For developers

### Data model

**`receivables`** — the cross-reference row: `uuid`, `company_id`, `invoice_id`,
`customer_id`, `deleted_at` (soft-delete). It ties an `invoices` row (the credit
invoice) to its `customers` row. Model: `receivableRead` (`app/playsql_models.go`),
with a `belongsTo` to the invoice.

### How it's created & settled

- **Created** by the invoice store path for credit terms (`registerReceivable`),
  and by the customer/vendor opening-balance flow
  (`storeCustomerInternal` → opening invoice → `registerReceivable`).
- **Settled** by payment allocation — see [payments.md](payments.md)
  (`updateInvoiceBalance` / `updateCustomerAmountDue`).
- **Reversed** when the invoice is voided ([invoices.md](invoices.md)).

Receivables have no dedicated routes; they're a ledger construct read through the
customer balance, the payment-create screen, and reports.

### Permissions

Governed by the surrounding modules (`:invoice`, `:payment`, `:customer`); there
is no standalone `receivable` permission. See [permissions.md](permissions.md).

### Entry points & tests

- Created/settled in `app/invoice-repository.go`, `app/customer-repository.go`,
  `app/payment-repository.go`
- Model: `app/playsql_models.go` (`receivableRead`)
- Tests: `app/flow_receivables_test.go`, `flow_opening_balance_test.go`,
  `flow_void_test.go`

## Related

[invoices.md](invoices.md) · [payments.md](payments.md) ·
[customers.md](customers.md)
