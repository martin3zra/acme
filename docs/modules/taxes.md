# Taxes

Two related things live here: the **tax rates** you charge (e.g. ITBIS 18%) and
the **fiscal tax-receipt sequences** (NCF) that fiscal invoices draw their
government receipt number from.

## For users

### Tax rates

A tax is a named rate (e.g. "ITBIS", 18%) applied to invoice and purchase lines.
Each line references a tax, and the document total includes the computed tax.

### Fiscal tax receipts (NCF)

A tax receipt is a **numbered sequence** authorized for fiscal documents. It has
a serie (e.g. `B01`), a type, and a numeric range (`sequence_start` …
`sequence_end`) with a `current` position. When you issue a **fiscal invoice**,
the next number in the sequence is assigned as its **NCF** and the sequence
advances.

### Managing them

- **Tax rates**: created/edited directly.
- **Tax receipts** and **document code sequences** (the `INV-…`, `CUST-…`
  prefixes) are configured under the company's **settings/profile**.

### Gotchas

- A tax-receipt sequence is **finite**. When `current` reaches `sequence_end`,
  fiscal invoicing is blocked until a new sequence is configured — plan capacity
  ahead of running out.
- Tax receipts are assigned per company; a customer can carry a default tax
  receipt ([customers.md](customers.md)).

## For developers

### Data model

- **`taxes`** — `company_id`, `name`, `rate`. Model: `taxRead`.
- **`tax_receipts`** — `company_id`, `name`, `serie`, `type`,
  `sequence_start`, `sequence_end`, `current`. Model: `taxReceiptRead`.
- **Document sequences** (customer/invoice/vendor codes) live in
  `companies_settings.sequences` (jsonb), not here — see the company settings
  routes.

The invoice header records which receipt slot it used via `tax_receipt_id`,
`tax_receipt_sequence`, and the assigned `tax_number` (the NCF itself) — see
[invoices.md](invoices.md).

### Routes (`app/route.go`)

- Tax rates: `POST /taxes`, `PUT /taxes/:id`.
- Tax receipts: `PUT /settings/companies/:id/tax-receipts`
  (`companyUpdateTaxReceipts`).
- Document sequences: `PUT /settings/companies/:id/sequences`
  (`companyUpdateSequences`).

### Key rules

- **NCF assignment** — a fiscal invoice pulls the next number from the tax-receipt
  sequence and advances `current`; the slot is recorded on the invoice.
- **Rate application** — line `rate` drives the per-line tax; totals roll up on
  the header.
- **Exhaustion** — running past `sequence_end` stops fiscal invoicing (pin the
  behaviour: `flow_tax_receipt_sequence_test.go`).

### Permissions

Tax receipts/sequences sit under `:setting` (owner/admin manage settings); tax
rates are managed by roles with the relevant setting/company permissions. See
[permissions.md](permissions.md).

### Entry points & tests

- Handlers: `app/taxes-handler.go`, company-settings handlers
- Repositories: `app/taxes-repository.go`, `app/tax-receipt-repository.go`
- Tests: `app/flow_tax_receipt_test.go`,
  `app/flow_tax_receipt_sequence_test.go`, `app/flow_sales_test.go`

## Related

[invoices.md](invoices.md) · [reports.md](reports.md) · [customers.md](customers.md)
