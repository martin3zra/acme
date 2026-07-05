package app

import (
	"database/sql"
	"testing"

	"github.com/martin3zra/forge/database"
)

// TestFlowVoidCreditInvoiceRemovesReceivable: voiding a credit invoice deletes
// its receivable row and reverses stock. Covers the playsql-backed
// deleteInvoiceFromReceivables on a row that actually exists (the cash-void flow
// only exercises the zero-match path).
func TestFlowVoidCreditInvoiceRemovesReceivable(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, variantID := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Credit("net30").Build()
	uuid := newInvoice(t, f, g).ForCustomer(custID).Credit("net30").
		WithLine(itemID, 1, 100, 18).Build()

	// Receivable exists before the void.
	assertRow(t, s.db, "receivables", map[string]any{"customer_id": custID, "company_id": f.company.ID})

	is.NoErr(f.s.voidInvoice(f.ctx, TransactionKinds.Invoice, uuid))

	is.Equal(scalarString(t, s.db, `SELECT status FROM invoices WHERE uuid = $1`, uuid), "void")
	assertNoRow(t, s.db, "receivables", map[string]any{"customer_id": custID})
	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT quantity FROM inventory_balances WHERE company_id = $1 AND variant_id = $2 AND warehouse_id = $3`,
		f.company.ID, variantID, f.warehouseID), 0)
}

// TestChangeCustomerFromReceivables: moving a receivable to a new customer
// rewrites receivables.customer_id. Covers the playsql-backed
// changeCustomerFromReceivables directly (its updateInvoice caller path is
// awkward to drive end-to-end).
func TestChangeCustomerFromReceivables(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	oldCust, _ := newCustomer(t, f, g).Credit("net30").Build()
	newCust, _ := newCustomer(t, f, g).Credit("net30").Build()
	uuid := newInvoice(t, f, g).ForCustomer(oldCust).Credit("net30").
		WithLine(itemID, 1, 100, 18).Build()

	var invID int
	is.NoErr(s.db.QueryRow(`SELECT id FROM invoices WHERE uuid = $1`, uuid).Scan(&invID))

	is.NoErr(database.WithTransaction(s.db, func(tx *sql.Tx) error {
		return f.s.changeCustomerFromReceivables(tx, f.company.ID, invID, oldCust, newCust)
	}))

	assertRow(t, s.db, "receivables", map[string]any{"invoice_id": invID, "customer_id": newCust})
	assertNoRow(t, s.db, "receivables", map[string]any{"invoice_id": invID, "customer_id": oldCust})
}
