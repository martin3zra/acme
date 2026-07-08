package app

import (
	"testing"
	"time"
)

// storePurchaseForVariantTest runs a minimal vendor-bill form through the real
// storePurchase path, returning the error (unlike the builder, which fails the
// test on error) so the guard paths can be asserted.
func storePurchaseForVariantTest(t *testing.T, f *fixture, vendorID int, lines []*Line) (string, error) {
	t.Helper()
	form := &StorePurchaseForm{
		VendorID:      vendorID,
		Date:          time.Now(),
		Terms:         "net30",
		Discount:      Discount{Type: "percentage"},
		Lines:         lines,
		Kind:          PurchaseTransactionKinds.VendorBill,
		InvoiceNumber: uniq("PINV"),
	}
	form.SetContext(f.ctx)
	form.Compute()
	return f.s.storePurchase(f.ctx, form)
}

// TestFlowPurchaseVariantLinePersists: a purchase line naming an explicit variant
// stores that exact variant on purchase_items, not a guessed one.
func TestFlowPurchaseVariantLinePersists(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, variantIDs := mkVariantItem(t, f, 100)
	blue := variantIDs[1]
	vendorID, _ := newVendor(t, f, g).Build()

	uuid := newPurchase(t, f).ForVendor(vendorID).
		Kind(PurchaseTransactionKinds.VendorBill).
		WithVariantLine(itemID, blue, 3, 40, 18).Build()

	var purchaseID int
	is.NoErr(s.db.QueryRow(`SELECT id FROM purchases WHERE uuid = $1`, uuid).Scan(&purchaseID))
	assertRow(t, s.db, "purchase_items", map[string]any{"purchase_id": purchaseID, "variant_id": blue, "qty": 3})
}

// TestFlowPurchaseVariantRequired: a has_variants item purchased without a variant
// is rejected.
func TestFlowPurchaseVariantRequired(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkVariantItem(t, f, 100)
	vendorID, _ := newVendor(t, f, g).Build()

	line := mkLine(itemID, f.unitID, f.warehouseID, 1, 40, 18) // VariantID left 0
	_, err := storePurchaseForVariantTest(t, f, vendorID, []*Line{line})
	if err == nil {
		t.Fatal("expected error: variant item purchased without a variant")
	}
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM purchases WHERE company_id = $1`, f.company.ID), 0)
}
