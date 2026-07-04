# Getting Started

Set up Acme and run the full business cycle: company → catalog → sales →
collections → purchasing → vendor payments → inventory → reports.

Sidebar labels are shown as they appear in the app (Spanish) with the route in
parentheses. The guide is split into **Part A — Admin setup** (one-time) and
**Part B — Daily use** (the cycle you repeat).

> 📸 _Screenshot: app dashboard (Panel) with the sidebar groups visible._

---

## Quick start (happy path)

1. Accept the email invite → set password → verify email → sign in.
2. Create your **company** (Onboarding).
3. **Settings:** add a **tax**, configure **NCF / tax receipts**, review
   sequences; then create a **warehouse**.
4. Create a **product**, a **customer**, and a **vendor**.
5. **Sales:** Estimate → convert to Order → convert to **Invoice** (cash or credit).
6. **Collect** a credit invoice in *Cuentas por cobrar*.
7. **Purchase:** Purchase order → receipt (**confirm = stock in**) or vendor bill.
8. **Pay** the vendor bill in *Cuentas por pagar*.
9. Review **Existencias / Movimientos** and **Reportes**.

Everything below expands these steps.

---

# Part A — Admin setup (one-time)

## A1. Accept your invite & sign in

1. Open the invitation email and click the activation link
   (`/verify-account/:uuid/:hash`) — it's signed and time-limited.
2. **Set your password**, then submit.
3. If prompted, **verify your email** (`/verify-email`) via the verification
   email, or use *Resend*.
4. Sign in at **/login**.

> 📸 _Screenshot: activation / set-password screen._

> If your account isn't linked to a company yet you'll see *Awaiting
> association* — an owner/admin adds you, or you create the company next.

## A2. Create your company

First login opens **Onboarding** (`/onboarding`).

1. Enter **company name**, **RNC**, **city**, **address**. Submit.
2. This auto-provisions: units of measure, document **sequences**, redirect
   preferences, and makes you the **owner**.

> 📸 _Screenshot: onboarding company form._

## A3. Configure Settings

Open **Ajustes / Settings** (`/settings/:account/profile`). Do these once.

| Setting | Where | Required? |
|---|---|---|
| **Tax / Impuesto** (e.g. ITBIS 18%) | Settings → Taxes | ✅ products need a tax |
| **NCF / Tax receipts** (serie, start, end) | Settings → Comprobantes | ✅ to issue invoices |
| **Sequences** (prefixes/padding) | Settings → Secuencias | review |
| **Units / Unidades** | Settings → Units | optional |
| **Expense categories** | Settings → Categorías | for expenses |
| **Users / Usuarios** (roles) | Settings → Users | to invite teammates |

> 📸 _Screenshot: Settings — NCF / tax-receipts configuration._

> **NCF matters:** each invoice consumes one number; set the sequence end high
> enough — an exhausted NCF blocks invoicing.

## A4. Create a warehouse

**Inventario → Almacenes** (`/inventories/warehouses`) → create at least one
(e.g. *General*). Required to receive/transfer/adjust stock and as the default
sales stock location.

> 📸 _Screenshot: Almacenes list with a "General" warehouse._

## A5. Build your catalog

**Products / Productos** (`/items`) — name, price, tax, unit, type; save. A
**default variant** is created automatically (inventory is tracked against it;
SKU generated, stock 0).

**Customers / Clientes** (`/customers`) — name, email (unique), payment terms
(`pia` = cash, `netN` = credit), optional credit limit / opening balance.

**Vendors / Proveedores** (`/vendors`) — same shape; needed before purchasing.

> 📸 _Screenshot: product create form with the generated default variant._

---

# Part B — Daily use (the cycle)

## B1. Sales — estimate → order → invoice

The three are one document with a different *kind*; the create screen (header +
line table + totals) is identical.

**Estimate / Cotización** (`/estimates/create`) — pick customer, add lines
(type a reference + Enter, or ⌘K to search; qty, price, tax). Save. No stock or
receivable impact.

> 📸 _Screenshot: document create screen — customer, date, line table, totals._

**Convert** — open a doc → *Convert to order/invoice*. The new doc opens
pre-filled; the source is marked **closed** and linked both ways (trace
estimate → order → invoice).

**Invoice / Factura** (`/invoices/create`) — choose terms:
- **Cash (`pia`)** — record the **payment**; tendered amount **must equal the
  total** or the save is rejected. Invoice → *closed/paid*, stock ships now.
- **Credit (`netN`)** — no payment now. Invoice → *sent/unpaid*, a **receivable**
  is registered, customer balance rises, due date set, stock ships.

Saving consumes one **NCF** and stamps it on the invoice. Over-credit-limit
credit invoices are blocked.

## B2. Collect — receivables (`/payments`)

**Finanzas → Cuentas por cobrar → New** — pick customer, apply amounts to their
open invoices (full or partial), choose method, save. Effects: invoice balances
drop (→ paid/partial), customer **amount due** decreases, payment record stored.
Voiding a payment re-opens the invoice and restores the balance.

> 📸 _Screenshot: receivables — applying a payment to open invoices._

## B3. Purchasing — PO → receipt → vendor bill

**Purchase order** (`/purchases/orders/create`) — vendor + lines (item, qty,
unit cost). A commitment; **no stock movement** yet.

**Receipt** (`/purchases/receipts/create`) — optionally **converted from a PO**
(links back, tracks remaining qty; receive partially across multiple receipts).
**Confirm** → stock **comes in**, status *received*.

**Vendor bill** (`/purchases/vendor-bills/create`) — enter the vendor invoice
number; registers an **account payable** and raises the vendor balance.
**Confirm** → brings stock in (unless already received), status *posted*.

> 📸 _Screenshot: purchase document + the Confirm action._

## B4. Pay vendors — payables (`/payables`)

**Finanzas → Cuentas por pagar → New** — pick vendor, apply amounts to open
payables, save. Payable balance drops, vendor **amount payable** decreases,
vendor-payment recorded. Can't overpay a settled payable; voiding re-opens it.

## B5. Inventory (Inventario)

- **Existencias / Stock** (`/inventories/stocks`) — on-hand per variant/warehouse.
- **Transferencias** (`/inventories/transfers`) — move stock between warehouses:
  create → **dispatch** (leaves source) → **receive** (arrives). Cancel before
  dispatch.
- **Ajustes** (`/inventories/adjustments`) — manual +/- corrections with a reason.
- **Movimientos** (`/inventories/movements`) — full ledger of every stock in/out.

> 📸 _Screenshot: Movimientos ledger._

## B6. Expenses & Reports

- **Gastos** (`/expenses`) — record an expense against a category (optional
  vendor/attachment).
- **Reportes** (`/reports/sales`) — Sales / Profit & Loss / Expenses / Taxes;
  pick a date range → generate PDF.

---

## Gotchas

- **NCF required & finite** — invoices need a configured tax receipt; exhausted
  sequences block invoicing.
- **Cash invoices must balance** — tendered payment must equal the total.
- **Credit limit** — over-limit credit invoices are blocked.
- **Stock needs a warehouse** and is tracked per **variant** (the default
  variant ships with the product).
- **Confirm purchases** to move stock — creating a receipt/bill alone doesn't; a
  PO never moves stock.
- **Roles gate the menu** — owner sees everything; admin/supervisor/standard see
  progressively less.
