# Getting Started — Full Business Cycle

A step-by-step walkthrough from accepting your invite to running the whole cycle:
company setup → catalog → sales (estimate → order → invoice) → collections →
purchasing → vendor payments → inventory → reports.

Sidebar labels below are shown as they appear in the app (Spanish) with the
route in parentheses. The sidebar is grouped: **Resumen**, **Ventas**,
**Compras**, **Inventario**, **Clientes y Catálogo**, **Finanzas**, **Análisis**.

---

## 1. Accept your invite & sign in

1. Open the invitation email and click the activation link
   (`/verify-account/:uuid/:hash`). The link is signed and time-limited.
2. **Set your password** on the activation screen, then submit.
3. If prompted, **verify your email** (`/verify-email`) — open the verification
   email and click through, or use *Resend* on the prompt screen.
4. Sign in at **/login** with your email + new password.

> If your account isn't linked to a company yet you'll land on an
> *Awaiting association* screen — an owner/admin must add you, or you create the
> company yourself in the next step.

---

## 2. Create your company

First login takes you to **Onboarding** (`/onboarding`).

1. Fill in **company name**, **RNC** (tax id), **city**, **address**.
2. Submit. This provisions the company and, automatically:
   - copies the shared **units** of measure,
   - seeds document **sequences** (invoices, estimates, orders, payments,
     customers, vendors, templates),
   - sets default redirect preferences,
   - makes you the **owner**.

You now land on the **Panel** (`/home`) dashboard.

---

## 3. One-time setup (Settings)

Open **Ajustes / Settings** (gear icon → `/settings/:account/profile`). Do these
once before invoicing — some are required.

1. **Taxes / Impuestos** — create at least one tax rate (e.g. *ITBIS 18%*).
   **Required:** every product needs a tax.
2. **Tax receipts (NCF) / Comprobantes** — configure your fiscal sequences
   (serie, start, end). **Required to issue invoices** of kind *Factura* — each
   invoice consumes one NCF number. Set the end high enough; an exhausted
   sequence blocks invoicing.
3. **Sequences / Secuencias** — review prefixes/padding for invoice, estimate,
   order, payment, customer, vendor numbering. (Purchasing sequences may need to
   be added if you use purchases.)
4. **Units / Unidades** — add or adjust units of measure if needed.
5. **Expense categories / Categorías de gastos** — for recording expenses.
6. **Users / Usuarios** — invite teammates (roles: owner, admin, supervisor,
   standard). Invited users start at step 1 above.

### Create a warehouse
Go to **Inventario → Almacenes** (`/inventories/warehouses`) and create at least
one (e.g. *General*). **Required** to receive/transfer/adjust stock and as the
default stock location on sales lines.

---

## 4. Build your catalog

### Products / Productos (`/items`)
1. **Clientes y Catálogo → Productos → New**.
2. Enter **name**, **price**, **description**, **tax**, **unit**, type
   (product/service).
3. Save. A **default variant** is created automatically (this is what inventory
   is tracked against — SKU is generated, stock starts at 0).

> Variants: today each product has one default variant. (A full
> attribute/variant matrix — size/color etc. — is on the roadmap.)

### Customers / Clientes (`/customers`)
1. **Clientes y Catálogo → Clientes → New**.
2. Enter name, contact, email (unique), phone, **payment method**, **payment
   terms** (`pia` = cash/pay-in-advance, or `netN` = credit), and optionally a
   **credit limit** and an **opening balance** (creates an opening receivable).
3. Save — the customer gets a code (e.g. `CUST-…`).

### Vendors / Proveedores (`/vendors`)
Same shape as customers (lead time, payment terms, optional opening balance).
Needed before you can purchase.

---

## 5. Sales cycle — estimate → order → invoice

All three are the same document with a different *kind*; the create screen is
identical (header + line table + totals).

### A. Estimate / Cotización (`/estimates/create`)
1. **Ventas → Cotizaciones → New**.
2. Pick the **customer**, set the **date**, add **line items** (type a reference
   and press Enter, or ⌘K to search; set qty, price, tax).
3. Save. Estimates don't touch stock or receivables.

### B. Convert estimate → order, or order → invoice
1. Open the document, use the **convert** action (e.g. *Convert to invoice*).
2. The new document opens pre-filled; review and save.
3. The source document is marked **closed** and linked to the new one
   (bidirectional link), so you can trace estimate → order → invoice.

### C. Invoice / Factura (`/invoices/create`)
1. **Ventas → Facturas → New** (or arrive here via conversion).
2. Choose **terms**:
   - **Cash (`pia`)** — you must record the **payment** (cash/card/check/bank)
     and the tendered amount **must equal the total**, or the save is rejected.
     Result: invoice **closed/paid**, stock shipped immediately.
   - **Credit (`netN`, e.g. net30)** — no payment now. Result: invoice
     **sent/unpaid**, a **receivable** is registered, the customer's balance
     increases, a due date is set, and stock is shipped.
3. Saving consumes one **NCF** (tax receipt) number and stamps it on the invoice.

> Credit limit: if the customer is credit-limited and this invoice would push
> their balance over the limit, the save is blocked.

---

## 6. Collect — receivables (`/payments`)

**Finanzas → Cuentas por cobrar → New** (record a customer payment).

1. Pick the **customer**; their open (credit) invoices appear.
2. Enter the **amount** applied to each invoice (full or partial) and the
   payment method.
3. Save. Effects:
   - each invoice's balance drops (paid → *closed/paid*, partial → *partial*),
   - the customer's **amount due** decreases,
   - a payment record (with code `PAY-…`) and its line items are stored.

Voiding a payment re-opens the invoice and restores the customer balance.

---

## 7. Purchasing cycle — PO → receipt → vendor bill

Documents under **Compras**, all vendor-facing. Like sales, they convert and link.

### A. Purchase order / Orden de compra (`/purchases/orders/create`)
1. **Compras → Órdenes de compra → New**.
2. Pick the **vendor**, add lines (item, qty, unit cost). Save. A PO is a
   commitment; it does **not** move stock yet.

### B. Receipt / Recibo de compra (`/purchases/receipts/create`)
1. Create a receipt, optionally **converted from a PO** (it links back and
   tracks remaining quantity — you can receive partially, in multiple receipts).
2. **Confirm** the receipt → stock **comes in** to the warehouse (inventory
   movement recorded), status → *received*.

### C. Vendor bill / Factura de proveedor (`/purchases/vendor-bills/create`)
1. Create a vendor bill (enter the vendor's invoice number). This registers an
   **account payable** and increases the vendor's balance.
2. **Confirm** it → if it isn't linked to a receipt that already moved stock,
   confirming brings stock in; status → *posted*.

---

## 8. Pay vendors — payables (`/payables`)

**Finanzas → Cuentas por pagar → New**.

1. Pick the **vendor**; their open payables appear.
2. Enter the amount applied to each payable (full or partial). Save. Effects:
   - the payable's balance drops (paid → *paid*, partial → *partial*),
   - the vendor's **amount payable** decreases,
   - a vendor-payment record is stored.

> You can't overpay a settled payable, and voiding a vendor payment re-opens it
> and restores the vendor balance.

---

## 9. Inventory (Inventario)

- **Almacenes** (`/inventories/warehouses`) — your stock locations.
- **Existencias / Stock** (`/inventories/stocks`) — on-hand quantity per
  variant/warehouse.
- **Transferencias** (`/inventories/transfers`) — move stock between warehouses
  as a request: create → **dispatch** (stock leaves source) → **receive** (stock
  arrives at destination). Cancel allowed before dispatch.
- **Ajustes** (`/inventories/adjustments`) — manual corrections (positive or
  negative) with a reason.
- **Movimientos** (`/inventories/movements`) — the full ledger of every stock
  in/out (sales, purchases, transfers, adjustments).

---

## 10. Expenses (`/expenses`)

**Finanzas → Gastos → New** — record an expense against a category, optionally
with a vendor and attachment.

---

## 11. Reports (`/reports/sales`)

**Análisis → Reportes** — tabs for **Sales**, **Profit & Loss**, **Expenses**,
**Taxes**. Pick a date range and generate the PDF.

---

## Quick reference — the happy path

1. Accept invite → set password → verify email → sign in.
2. Create company (Onboarding).
3. Settings: tax, **NCF**, sequences, units; create a **warehouse**.
4. Create a **product**, a **customer**, a **vendor**.
5. Estimate → convert to Order → convert to **Invoice** (cash or credit).
6. Credit invoice → **collect** in *Cuentas por cobrar*.
7. **Purchase order → receipt (confirm = stock in)** or **vendor bill**.
8. Vendor bill → **pay** in *Cuentas por pagar*.
9. Check **Existencias / Movimientos** and **Reportes**.

## Gotchas

- **NCF required & finite:** invoices need a configured tax receipt; an exhausted
  sequence blocks invoicing.
- **Cash invoices must balance:** tendered payment must equal the total.
- **Credit limit:** over-limit credit invoices are blocked.
- **Stock needs a warehouse** and is tracked per **variant** (the default variant
  is created with the product).
- **Confirm purchases** to actually move stock — creating a receipt/bill alone
  doesn't (a PO never moves stock).
- **Roles** gate the menu: owner sees everything; admin/supervisor/standard see
  progressively less.
