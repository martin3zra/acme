package app

import "testing"

// Purchase header reads converted from hand-concatenated SQL to playsql. The flow
// suite drives storePurchase/confirmPurchase/destroyPurchase, but never the list
// read, the linked-receipt lookup, or the fields the scan loops derived by hand.

// TestFindPurchases_ListLoadsVendorAndDerivedFields: the INNER JOIN on vendors
// becomes a belongsTo eager load, and Discount/AmountDue/Terms are derived in
// toPurchase rather than in a scan loop.
func TestFindPurchases_ListLoadsVendorAndDerivedFields(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	vendorID, vendorUUID := newVendor(t, f, g).Named("Ferreteria Lopez").Build()
	itemID, _ := mkItem(t, f, 100, 60)
	uuid := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		WithLine(itemID, 2, 100, 18).Build()

	purchases, err := s.findPurchases(f.ctx, PurchaseTransactionKinds.VendorBill)
	is.NoErr(err)
	is.Equal(len(purchases), 1)

	p := purchases[0]
	is.Equal(p.UUID, uuid)
	is.Equal(p.Vendor.UUID, vendorUUID)
	is.Equal(p.Vendor.Name, "Ferreteria Lopez")
	is.EqualFloat(p.Total, 236)
	is.EqualFloat(p.AmountDue, 236) // unpaid, so amount_due == total
	is.Equal(p.Discount.Type, "fixed")
	is.EqualFloat(p.Discount.Val, 0)
	is.True(p.Status != "", "purchase_status is NOT NULL and always populated")

	// The list projection never carried company_id, and still does not.
	is.Equal(p.CompanyID, 0)
}

// TestFindPurchases_FiltersByKindAndHidesDeleted: purchaseRead's softdelete tag
// supplies the `deleted_at IS NULL` both header reads carried.
func TestFindPurchases_FiltersByKindAndHidesDeleted(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	vendorID, _ := newVendor(t, f, g).Build()
	itemID, _ := mkItem(t, f, 100, 60)

	billUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		WithLine(itemID, 1, 100, 18).Build()
	orderUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseOrder).
		WithLine(itemID, 1, 100, 18).Build()

	bills, err := s.findPurchases(f.ctx, PurchaseTransactionKinds.VendorBill)
	is.NoErr(err)
	is.Equal(len(bills), 1)
	is.Equal(bills[0].UUID, billUUID)

	orders, err := s.findPurchases(f.ctx, PurchaseTransactionKinds.PurchaseOrder)
	is.NoErr(err)
	is.Equal(len(orders), 1)
	is.Equal(orders[0].UUID, orderUUID)

	// Soft-delete the bill; it drops out of both the list and the detail read.
	is.NoErr(s.destroyPurchase(f.ctx, billUUID))

	bills, err = s.findPurchases(f.ctx, PurchaseTransactionKinds.VendorBill)
	is.NoErr(err)
	is.Equal(len(bills), 0)

	_, err = s.findPurchaseByUUID(f.ctx, f.company.ID, billUUID)
	is.Err(err, "a soft-deleted purchase must not be findable")
}

// TestDestroyPurchase_StampsBothColumns: Builder.Delete would only set deleted_at,
// so the soft delete goes through Update to keep bumping updated_at.
func TestDestroyPurchase_StampsBothColumns(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	vendorID, _ := newVendor(t, f, g).Build()
	itemID, _ := mkItem(t, f, 100, 60)
	uuid := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseOrder).
		WithLine(itemID, 1, 100, 18).Build()

	before := scalarString(t, s.db, `SELECT updated_at::text FROM purchases WHERE uuid = $1`, uuid)
	is.NoErr(s.destroyPurchase(f.ctx, uuid))

	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM purchases WHERE uuid = $1 AND deleted_at IS NOT NULL`, uuid), 1)
	after := scalarString(t, s.db, `SELECT updated_at::text FROM purchases WHERE uuid = $1`, uuid)
	is.True(after != before, "destroyPurchase should bump updated_at")
}

// TestDestroyVendorBill_VoidsPayable: the accounts_payable lookup and its void are
// both on playsql now, and only the unpaid portion comes off the vendor balance.
func TestDestroyVendorBill_VoidsPayable(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	vendorID, _ := newVendor(t, f, g).Build()
	itemID, _ := mkItem(t, f, 100, 60)
	uuid := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		WithLine(itemID, 1, 100, 18).Build()

	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_payable FROM vendors WHERE id = $1`, vendorID), 118)

	is.NoErr(s.destroyPurchase(f.ctx, uuid))

	is.Equal(scalarString(t, s.db,
		`SELECT ap.status::text FROM accounts_payable ap
		 JOIN purchases p ON ap.purchase_id = p.id WHERE p.uuid = $1`, uuid), "void")
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_payable FROM vendors WHERE id = $1`, vendorID), 0)
}

// TestFindLinkedReceiptsForOrder: the `source->>'id' = $2` filter becomes WhereJSON,
// which renders the equivalent `source #>> '{id}'`.
func TestFindLinkedReceiptsForOrder(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	vendorID, _ := newVendor(t, f, g).Build()
	itemID, _ := mkItem(t, f, 100, 60)

	orderUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseOrder).
		WithLine(itemID, 10, 100, 18).Build()

	// Two receipts against the order, plus an unrelated standalone receipt.
	src := &PurchaseSource{Type: PurchaseTransactionKinds.PurchaseOrder, ID: orderUUID}
	r1 := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseReceipt).
		FromSource(src).WithLine(itemID, 4, 100, 18).Build()
	r2 := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseReceipt).
		FromSource(src).WithLine(itemID, 3, 100, 18).Build()
	_ = newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseReceipt).
		WithLine(itemID, 1, 100, 18).Build()

	linked, err := s.findLinkedReceiptsForOrder(f.ctx, f.company.ID, orderUUID)
	is.NoErr(err)
	is.Equal(len(linked), 2) // the sourceless receipt is excluded
	is.Equal(linked[0].UUID, r1)
	is.Equal(linked[1].UUID, r2) // ORDER BY id ASC
	is.True(linked[0].Number != "", "code should be projected")

	// An unknown order uuid matches nothing rather than erroring.
	none, err := s.findLinkedReceiptsForOrder(f.ctx, f.company.ID, "00000000-0000-0000-0000-000000000000")
	is.NoErr(err)
	is.Equal(len(none), 0)
}

// TestFindPurchaseByUUID_SoftDeletedVendor: vendorRead is softdelete-tagged, but the
// old INNER JOIN never filtered deleted_at, so the purchase must still render.
func TestFindPurchaseByUUID_SoftDeletedVendor(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	vendorID, vendorUUID := newVendor(t, f, g).Build()
	itemID, _ := mkItem(t, f, 100, 60)
	uuid := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseOrder).
		WithLine(itemID, 1, 100, 18).Build()

	is.NoErr(s.deleteVendor(f.ctx, vendorID))

	p, err := s.findPurchaseByUUID(f.ctx, f.company.ID, uuid)
	is.NoErr(err)
	is.Equal(p.Vendor.UUID, vendorUUID)
	is.Equal(p.CompanyID, f.company.ID) // the detail projection does carry it

	list, err := s.findPurchases(f.ctx, PurchaseTransactionKinds.PurchaseOrder)
	is.NoErr(err)
	is.Equal(len(list), 1)
	is.Equal(list[0].Vendor.UUID, vendorUUID)
}

// TestFindPurchases_SourceDecodes: source is jsonb, decoded in toPurchase because
// purchase.Source is a pointer-to-struct that playsql would read as a relation.
//
// A standalone document stores SQL NULL there. Converting a receipt from an order
// writes the link both ways: the receipt points back at the order, and the order is
// updated to point forward at the receipt.
func TestFindPurchases_SourceDecodes(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	vendorID, _ := newVendor(t, f, g).Build()
	itemID, _ := mkItem(t, f, 100, 60)

	orderUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseOrder).
		WithLine(itemID, 5, 100, 18).Build()

	// Before any conversion the order's source is NULL.
	order, err := s.findPurchaseByUUID(f.ctx, f.company.ID, orderUUID)
	is.NoErr(err)
	is.True(order.Source == nil, "a standalone order has no source")

	receiptUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseReceipt).
		FromSource(&PurchaseSource{Type: PurchaseTransactionKinds.PurchaseOrder, ID: orderUUID}).
		WithLine(itemID, 5, 100, 18).Build()

	receipt, err := s.findPurchaseByUUID(f.ctx, f.company.ID, receiptUUID)
	is.NoErr(err)
	is.True(receipt.Source != nil, "the receipt should carry its source")
	is.Equal(receipt.Source.ID, orderUUID)
	is.Equal(string(receipt.Source.Type), "purchase_order")

	// The conversion back-fills the order's source with a pointer to the receipt.
	order, err = s.findPurchaseByUUID(f.ctx, f.company.ID, orderUUID)
	is.NoErr(err)
	is.True(order.Source != nil, "the converted order gains a back-reference")
	is.Equal(order.Source.ID, receiptUUID)
	is.Equal(string(order.Source.Type), "purchase_receipt")
}
