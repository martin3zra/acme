package app

import "testing"

// TestFlowPartialReceipt: receiving part of a purchase order tracks the remaining
// quantity and brings in only the received stock; a second receipt completes it.
func TestFlowPartialReceipt(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	g := newFaker(t)
	vendorID, _ := newVendor(t, f, g).Build()
	itemID, variantID := mkItem(t, f, 100, 60)

	poUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseOrder).
		WithLine(itemID, 10, 100, 18).Build()

	var poID int
	is.NoErr(s.db.QueryRow(`SELECT id FROM purchases WHERE uuid = $1`, poUUID).Scan(&poID))

	// First receipt: 4 of 10.
	rc1 := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseReceipt).
		FromSource(&PurchaseSource{Type: PurchaseTransactionKinds.PurchaseOrder, ID: poUUID}).
		WithLine(itemID, 4, 100, 18).Build()
	is.NoErr(f.s.confirmPurchase(f.ctx, rc1))

	remaining := func() int64 {
		lines, err := f.s.findPurchaseLines(f.ctx, f.company.ID, poID)
		is.NoErr(err)
		is.NoErr(f.s.enrichLinesWithRemainingQty(f.ctx, f.company.ID, poUUID, lines))
		is.True(len(lines) == 1, "purchase order should have one line")
		is.True(lines[0].RemainingQty != nil, "remaining qty must be set")
		return *lines[0].RemainingQty
	}

	is.Equal(remaining(), int64(6))
	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT quantity FROM inventory_balances WHERE company_id = $1 AND variant_id = $2 AND warehouse_id = $3`,
		f.company.ID, variantID, f.warehouseID), 4)

	links, err := f.s.findLinkedReceiptsForOrder(f.ctx, f.company.ID, poUUID)
	is.NoErr(err)
	is.Equal(len(links), 1)

	// Second receipt: remaining 6 -> fully received.
	rc2 := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseReceipt).
		FromSource(&PurchaseSource{Type: PurchaseTransactionKinds.PurchaseOrder, ID: poUUID}).
		WithLine(itemID, 6, 100, 18).Build()
	is.NoErr(f.s.confirmPurchase(f.ctx, rc2))

	is.Equal(remaining(), int64(0))
	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT quantity FROM inventory_balances WHERE company_id = $1 AND variant_id = $2 AND warehouse_id = $3`,
		f.company.ID, variantID, f.warehouseID), 10)
}
