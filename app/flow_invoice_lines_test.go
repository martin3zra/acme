package app

import "testing"

// TestFlowMultiLineInvoiceBulkInsert exercises the true multi-row path of the
// playsql-backed attachInvoiceLines (InsertMany): a two-line invoice must persist
// both line rows and sum to the correct total.
func TestFlowMultiLineInvoiceBulkInsert(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	item1, _ := mkItem(t, f, 100, 60)
	item2, _ := mkItem(t, f, 50, 30)
	custID, _ := newCustomer(t, f, g).Build()

	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithLine(item1, 2, 100, 18).
		WithLine(item2, 3, 50, 18).
		Build()

	var invID int
	is.NoErr(s.db.QueryRow(`SELECT id FROM invoices WHERE uuid = $1`, uuid).Scan(&invID))

	// Both line rows landed in one bulk insert.
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM invoices_items WHERE invoice_id = $1`, invID), 2)
	assertRow(t, s.db, "invoices_items", map[string]any{"invoice_id": invID, "item_id": item1, "qty": 2})
	assertRow(t, s.db, "invoices_items", map[string]any{"invoice_id": invID, "item_id": item2, "qty": 3})

	// Subtotal 2*100 + 3*50 = 350; 18% tax = 63; total 413.
	is.EqualFloat(scalarFloat(t, s.db, `SELECT total FROM invoices WHERE id = $1`, invID), 413)
}
