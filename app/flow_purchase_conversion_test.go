package app

import "testing"

// Converting a purchase order directly into a vendor bill used to abort on
// purchases_source_check.
//
// The vendor-bill path reads the source document's `source`, adds a `target` key,
// and writes it back. A receipt has a source — the order it was received against —
// so the stamp lands. A purchase order has none: its `source` is NULL, so the write
// was a bare {"target": ...}, and the constraint requires both `type` and `id`.

// TestVendorBillFromPurchaseOrder: the conversion succeeds, and the order's source
// is left alone.
func TestVendorBillFromPurchaseOrder(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	vendorID, _ := newVendor(t, f, g).Build()
	itemID, _ := mkItem(t, f, 100, 60)

	poUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseOrder).
		WithLine(itemID, 1, 100, 18).Build()
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM purchases WHERE uuid = $1 AND source IS NULL`, poUUID), 1)

	billUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		FromSource(&PurchaseSource{Type: PurchaseTransactionKinds.PurchaseOrder, ID: poUUID}).
		WithLine(itemID, 1, 100, 18).Build()

	// The bill points back at the order — that link is what updateAPBalance and
	// updatePOPaymentStatus follow.
	bill, err := s.findPurchaseByUUID(f.ctx, f.company.ID, billUUID)
	is.NoErr(err)
	is.True(bill.Source != nil, "the bill records the order it came from")
	is.Equal(bill.Source.ID, poUUID)

	// The order is untouched: no bare {"target": ...} was written onto it.
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM purchases WHERE uuid = $1 AND source IS NULL`, poUUID), 1)
}

// TestVendorBillFromReceipt_StampsForwardLink: a receipt does have a source, so the
// forward link is still written and the receipt list view keeps working.
func TestVendorBillFromReceipt_StampsForwardLink(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	vendorID, _ := newVendor(t, f, g).Build()
	itemID, _ := mkItem(t, f, 100, 60)

	poUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseOrder).
		WithLine(itemID, 5, 100, 18).Build()
	receiptUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseReceipt).
		FromSource(&PurchaseSource{Type: PurchaseTransactionKinds.PurchaseOrder, ID: poUUID}).
		WithLine(itemID, 5, 100, 18).Build()

	billUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		FromSource(&PurchaseSource{Type: PurchaseTransactionKinds.PurchaseReceipt, ID: receiptUUID}).
		WithLine(itemID, 5, 100, 18).Build()

	// The receipt keeps its own source and gains a forward link to the bill.
	is.Equal(scalarString(t, s.db,
		`SELECT source->>'type' FROM purchases WHERE uuid = $1`, receiptUUID), "purchase_order")
	is.Equal(scalarString(t, s.db,
		`SELECT source->'target'->>'id' FROM purchases WHERE uuid = $1`, receiptUUID), billUUID)
	is.Equal(scalarString(t, s.db,
		`SELECT source->'target'->>'type' FROM purchases WHERE uuid = $1`, receiptUUID), "vendor_bill")
}

// TestCanCarryForwardLink covers the shapes the check constraint accepts.
func TestCanCarryForwardLink(t *testing.T) {
	is := newIs(t)

	is.True(!canCarryForwardLink(nil), "nil source cannot carry a target")
	is.True(!canCarryForwardLink(map[string]any{}), "empty source cannot carry a target")
	is.True(!canCarryForwardLink(map[string]any{"type": "purchase_order"}), "id is required")
	is.True(!canCarryForwardLink(map[string]any{"id": "abc"}), "type is required")
	is.True(!canCarryForwardLink(map[string]any{"type": "", "id": "abc"}), "type must be non-empty")
	is.True(canCarryForwardLink(map[string]any{"type": "purchase_order", "id": "abc"}),
		"type + id is enough")
}
