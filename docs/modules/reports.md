# Reports

Read-only summaries over the ledger, each rendered to a downloadable PDF.

## For users

### What's available

Four reports, each driven by a date range (and any report-specific filters):

- **Sales** — invoiced sales over the period.
- **Profit & Loss** — revenue vs. cost/expenses.
- **Expenses** — expenses by category.
- **Taxes** — tax collected/owed, for fiscal filing.

### How it works

Open a report, choose the period, and generate — the result is a **PDF** you can
download or print. Reports never change data; they only read it.

### Gotchas

- Figures reflect the current ledger state, including voids and soft-deletes as
  each report defines them — a voided invoice is excluded from sales, for example.
- Reports are tenant-scoped: you only ever see your own company's data.

## For developers

### Shape

Each report is a **GET** (render the form/page) plus a **POST** (generate).
Generation reads through a report repository and renders a PDF via a dedicated
builder. No report writes to the database.

### Routes (`app/route.go`)

| Report | View | Generate |
|---|---|---|
| Sales | `GET /reports/sales` | `POST /reports/sales` |
| Profit & Loss | `GET /reports/profit-lost` | `POST /reports/profit-lost` |
| Expenses | `GET /reports/expenses` | `POST /reports/expenses` |
| Taxes | `GET /reports/taxes` | `POST /reports/taxes` |

All view routes are gated by `viewAny:reports`.

### Entry points

- Handlers: `app/sales-report-handler.go`, `app/taxes-report-handler.go`,
  and the profit-lost / expenses report handlers
- Repositories: `app/sales-report-repository.go`,
  `app/taxes-report-repository.go`
- PDF builders: `app/sales-report-pdf.go`, `app/taxes-report-pdf.go`,
  `app/profit-lost-report-pdf.go`, `app/expenses-report-pdf.go`

### Key rules

- **Read-only** — reports never mutate; they aggregate existing rows.
- **Date range** scoping is the primary filter; the taxes report feeds fiscal
  reconciliation ([taxes.md](taxes.md)).
- Tenant-scoped on `company_id`.

### Permissions

`viewAny:reports` for the report pages; generation follows the same gate. See
[permissions.md](permissions.md).

### Tests

Reports are exercised indirectly through the ledger flow tests (e.g. the sales
figures are pinned by the invoice/sales flows, `app/flow_sales_test.go`); there
is no separate report-rendering test suite.

## Related

[invoices.md](invoices.md) · [taxes.md](taxes.md) · [payments.md](payments.md)
