package app

import "testing"

// balanceOf reads the on-hand quantity for a variant in a warehouse (0 when no
// balance row exists yet).
func balanceOf(t *testing.T, f *fixture, variantID int) float64 {
	t.Helper()
	var q float64
	_ = f.s.db.QueryRow(
		`SELECT quantity FROM inventory_balances WHERE company_id = $1 AND variant_id = $2 AND warehouse_id = $3`,
		f.company.ID, variantID, f.warehouseID,
	).Scan(&q)
	return q
}

// TestFlowInvoiceUpdateReconcilesQtyChange: bumping a live invoice line's qty
// moves the balance to match the new qty — the update path re-records stock, it
// does not leave the original quantity debited.
func TestFlowInvoiceUpdateReconcilesQtyChange(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, variantID := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Build()

	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithLine(itemID, 2, 100, 18).Build()
	is.EqualFloat(balanceOf(t, f, variantID), -2) // create debited 2

	// Bump the line to qty 5.
	is.NoErr(updateInvoiceVariantLines(t, f, uuid, custID, []*Line{
		{ID: itemID, VariantID: variantID, Unit: f.unitID, WarehouseID: f.warehouseID, Qty: 5, Price: 100, Rate: 18, Action: UPDATED},
	}))

	is.EqualFloat(balanceOf(t, f, variantID), -5) // reconciled to the new qty
}

// TestFlowInvoiceUpdateReconcilesVariantSwap: swapping a line's variant (delete
// old + add new) restocks the old variant and debits the new one, so the balance
// follows the sold variant rather than stranding stock on the original.
func TestFlowInvoiceUpdateReconcilesVariantSwap(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, variantIDs := mkVariantItem(t, f, 100)
	red, blue := variantIDs[0], variantIDs[1]
	custID, _ := newCustomer(t, f, g).Build()

	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithVariantLine(itemID, blue, 2, 100, 18).Build()
	is.EqualFloat(balanceOf(t, f, blue), -2)
	is.EqualFloat(balanceOf(t, f, red), 0)

	// Swap blue → red at a new qty.
	is.NoErr(updateInvoiceVariantLines(t, f, uuid, custID, []*Line{
		{ID: itemID, VariantID: blue, Unit: f.unitID, WarehouseID: f.warehouseID, Qty: 2, Price: 100, Rate: 18, Action: DELETED},
		{ID: itemID, VariantID: red, Unit: f.unitID, WarehouseID: f.warehouseID, Qty: 3, Price: 100, Rate: 18, Action: ADDED},
	}))

	is.EqualFloat(balanceOf(t, f, blue), 0) // old variant restocked
	is.EqualFloat(balanceOf(t, f, red), -3) // new variant debited
}

// TestFlowInvoiceUpdateReconcileIsIdempotent: re-saving an invoice with no line
// changes leaves the balance where it was — reconcile nets to zero delta and does
// not drift on repeated edits.
func TestFlowInvoiceUpdateReconcileIsIdempotent(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, variantID := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Build()

	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithLine(itemID, 4, 100, 18).Build()
	is.EqualFloat(balanceOf(t, f, variantID), -4)

	for range 3 {
		is.NoErr(updateInvoiceVariantLines(t, f, uuid, custID, []*Line{
			{ID: itemID, VariantID: variantID, Unit: f.unitID, WarehouseID: f.warehouseID, Qty: 4, Price: 100, Rate: 18, Action: UPDATED},
		}))
		is.EqualFloat(balanceOf(t, f, variantID), -4) // stable across edits
	}
}
