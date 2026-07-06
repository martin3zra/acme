package app

import (
	"testing"
	"time"
)

// TestFlowUpdateInvoiceLineActions drives updateInvoice through all three
// playsql-backed processInvoiceLines branches at once: an existing line is
// UPDATED (new qty), another is DELETED, and a new one is ADDED. A cash invoice
// with the same customer keeps the receivable/balance logic out of the way so
// the assertions isolate the line mutations.
func TestFlowUpdateInvoiceLineActions(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	item1, _ := mkItem(t, f, 100, 60)
	item2, _ := mkItem(t, f, 50, 30)
	item3, _ := mkItem(t, f, 30, 20)
	custID, _ := newCustomer(t, f, g).Build()

	// Original cash invoice: item1 x2, item2 x1.
	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithLine(item1, 2, 100, 18).
		WithLine(item2, 1, 50, 18).
		Build()

	var invID int
	is.NoErr(s.db.QueryRow(`SELECT id FROM invoices WHERE uuid = $1`, uuid).Scan(&invID))
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM invoices_items WHERE invoice_id = $1`, invID), 2)

	// Update: bump item1 qty, drop item2, add item3.
	form := &UpdateInvoiceForm{StoreInvoiceForm: StoreInvoiceForm{
		CustomerID: custID,
		Date:       time.Now(),
		Terms:      "pia",
		TaxReceipt: f.taxReceiptID,
		Discount:   Discount{Type: "percentage"},
		Kind:       TransactionKinds.Invoice,
		Lines: []*Line{
			{ID: item1, Unit: f.unitID, WarehouseID: f.warehouseID, Qty: 5, Price: 100, Rate: 18, Action: UPDATED},
			{ID: item2, Unit: f.unitID, WarehouseID: f.warehouseID, Qty: 1, Price: 50, Rate: 18, Action: DELETED},
			{ID: item3, Unit: f.unitID, WarehouseID: f.warehouseID, Qty: 3, Price: 30, Rate: 18, Action: ADDED},
		},
	}}
	form.Compute()
	is.NoErr(f.s.updateInvoice(f.ctx, uuid, form))

	// item1 updated (qty 5), item2 gone, item3 added (qty 3) — two rows.
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM invoices_items WHERE invoice_id = $1`, invID), 2)
	assertRow(t, s.db, "invoices_items", map[string]any{"invoice_id": invID, "item_id": item1, "qty": 5})
	assertRow(t, s.db, "invoices_items", map[string]any{"invoice_id": invID, "item_id": item3, "qty": 3})
	assertNoRow(t, s.db, "invoices_items", map[string]any{"invoice_id": invID, "item_id": item2})

	// The invoice-level UPDATE wrote the recomputed total: (5*100 + 3*30)*1.18.
	is.EqualFloat(scalarFloat(t, s.db, `SELECT total FROM invoices WHERE id = $1`, invID), 696.2)
}
