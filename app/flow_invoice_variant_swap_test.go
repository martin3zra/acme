package app

import (
	"testing"
	"time"
)

// updateInvoiceVariantLines drives updateInvoice with an explicit line set,
// mirroring TestFlowUpdateInvoiceLineActions but letting the caller set each
// line's VariantID so variant-swap scenarios can be exercised.
func updateInvoiceVariantLines(t *testing.T, f *fixture, uuid string, custID int, lines []*Line) error {
	t.Helper()
	form := &UpdateInvoiceForm{StoreInvoiceForm: StoreInvoiceForm{
		CustomerID: custID,
		Date:       time.Now(),
		Terms:      "pia",
		TaxReceipt: f.taxReceiptID,
		Discount:   Discount{Type: "percentage"},
		Kind:       TransactionKinds.Invoice,
		Lines:      lines,
	}}
	form.Compute()
	return f.s.updateInvoice(f.ctx, uuid, form)
}

// TestFlowInvoiceVariantSwapViaDeleteAdd: the supported way to change a line's
// variant is delete-the-old + add-the-new (the frontend keys lines by
// (item_id, variant_id), so a variant change is a new line, not an edit). The
// invoices_items row must move from the old variant to the new one.
//
// This test asserts the line rows only; the matching stock reconciliation is
// covered by TestFlowInvoiceUpdateReconcilesVariantSwap.
func TestFlowInvoiceVariantSwapViaDeleteAdd(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, variantIDs := mkVariantItem(t, f, 100)
	red, blue := variantIDs[0], variantIDs[1]
	custID, _ := newCustomer(t, f, g).Build()

	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithVariantLine(itemID, blue, 2, 100, 18).Build()

	var invID int
	is.NoErr(s.db.QueryRow(`SELECT id FROM invoices WHERE uuid = $1`, uuid).Scan(&invID))
	assertRow(t, s.db, "invoices_items", map[string]any{"invoice_id": invID, "item_id": itemID, "variant_id": blue})

	// Swap blue → red: DELETE the blue line, ADD a red one.
	is.NoErr(updateInvoiceVariantLines(t, f, uuid, custID, []*Line{
		{ID: itemID, VariantID: blue, Unit: f.unitID, WarehouseID: f.warehouseID, Qty: 2, Price: 100, Rate: 18, Action: DELETED},
		{ID: itemID, VariantID: red, Unit: f.unitID, WarehouseID: f.warehouseID, Qty: 3, Price: 100, Rate: 18, Action: ADDED},
	}))

	// Exactly one line, now on red at the new qty; blue is gone.
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM invoices_items WHERE invoice_id = $1`, invID), 1)
	assertRow(t, s.db, "invoices_items", map[string]any{"invoice_id": invID, "item_id": itemID, "variant_id": red, "qty": 3})
	assertNoRow(t, s.db, "invoices_items", map[string]any{"invoice_id": invID, "variant_id": blue})
}

// TestFlowInvoiceVariantSwapAsUpdateIsInert: an UPDATED line whose variant_id no
// longer matches the stored row must not silently corrupt data. processInvoiceLines
// keys the UPDATE on (item_id, variant_id); a changed variant matches zero rows,
// so the update is a no-op — the original line is left intact and no phantom row
// for the new variant appears. (This is why a real variant change goes through
// delete+add, not UPDATE.)
func TestFlowInvoiceVariantSwapAsUpdateIsInert(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, variantIDs := mkVariantItem(t, f, 100)
	red, blue := variantIDs[0], variantIDs[1]
	custID, _ := newCustomer(t, f, g).Build()

	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithVariantLine(itemID, blue, 2, 100, 18).Build()

	var invID int
	is.NoErr(s.db.QueryRow(`SELECT id FROM invoices WHERE uuid = $1`, uuid).Scan(&invID))

	// UPDATE the line but point it at red instead of the stored blue.
	is.NoErr(updateInvoiceVariantLines(t, f, uuid, custID, []*Line{
		{ID: itemID, VariantID: red, Unit: f.unitID, WarehouseID: f.warehouseID, Qty: 9, Price: 100, Rate: 18, Action: UPDATED},
	}))

	// Inert: still one row, still blue, still the original qty — no red row.
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM invoices_items WHERE invoice_id = $1`, invID), 1)
	assertRow(t, s.db, "invoices_items", map[string]any{"invoice_id": invID, "item_id": itemID, "variant_id": blue, "qty": 2})
	assertNoRow(t, s.db, "invoices_items", map[string]any{"invoice_id": invID, "variant_id": red})
}
