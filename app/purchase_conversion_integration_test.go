//go:build integration

package app

import (
	"database/sql"
	"testing"
	"time"
)

// seedPurchaseSequences installs the per-kind code sequences that
// storePurchase reads from companies_settings.
func seedPurchaseSequences(t *testing.T, db *sql.DB, companyID int) {
	t.Helper()
	seq := `{
		"purchase_order":   {"next": 1, "prefix": "PO-", "padding": 4},
		"purchase_receipt": {"next": 1, "prefix": "PR-", "padding": 4},
		"vendor_bill":      {"next": 1, "prefix": "VB-", "padding": 4}
	}`
	must(t, exec(db,
		`INSERT INTO companies_settings (company_id, sequences) VALUES ($1, $2::jsonb)`,
		companyID, seq))
}

func purchaseForm(kind PurchaseTransactionKind, f inventoryFixtures, vendorID, unitID int, source *PurchaseSource) *StorePurchaseForm {
	form := &StorePurchaseForm{
		VendorID: vendorID,
		Date:     time.Now().UTC(),
		Terms:    "net0",
		Kind:     kind,
		Source:   source,
		Lines: []*Line{
			{ID: f.ItemID, Unit: unitID, Qty: 12, Price: 5, WarehouseID: f.WHFrom},
		},
	}
	form.Compute()
	return form
}

// Converting a purchase order into a receipt links the two, and confirming the
// receipt posts inventory movements.
func TestIntegration_ConvertPurchaseOrderToReceipt(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)
	ctx := companyCtx(f.CompanyID)
	seedPurchaseSequences(t, db, f.CompanyID)

	var vendorID int
	must(t, db.QueryRow(
		`INSERT INTO vendors (company_id, name) VALUES ($1, 'Acme Supply') RETURNING id`,
		f.CompanyID).Scan(&vendorID))
	var unitID int
	must(t, db.QueryRow(
		`INSERT INTO units (company_id, name, base_qty) VALUES ($1, 'unit', 1) RETURNING id`,
		f.CompanyID).Scan(&unitID))

	// 1. Create the purchase order (no inventory movement yet).
	poUUID, err := srv.storePurchase(ctx, purchaseForm(PurchaseTransactionKinds.PurchaseOrder, f, vendorID, unitID, nil))
	if err != nil {
		t.Fatalf("store purchase order: %v", err)
	}
	if n := movementCountByKind(t, db, f.CompanyID, "purchase_order"); n != 0 {
		t.Errorf("a purchase order should record no movements, got %d", n)
	}

	// 2. Convert it to a receipt sourced from the order.
	receiptUUID, err := srv.storePurchase(ctx, purchaseForm(
		PurchaseTransactionKinds.PurchaseReceipt, f, vendorID, unitID,
		&PurchaseSource{Type: PurchaseTransactionKinds.PurchaseOrder, ID: poUUID}))
	if err != nil {
		t.Fatalf("store receipt from order: %v", err)
	}

	// The order → receipt link is queryable.
	linked, err := srv.findLinkedReceiptsForOrder(ctx, f.CompanyID, poUUID)
	if err != nil {
		t.Fatalf("findLinkedReceiptsForOrder: %v", err)
	}
	if len(linked) != 1 {
		t.Errorf("expected 1 linked receipt for the order, got %d", len(linked))
	}

	// 3. Confirm the receipt → stock is posted.
	if err := srv.confirmPurchase(ctx, receiptUUID); err != nil {
		t.Fatalf("confirm receipt: %v", err)
	}
	if qty, _ := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHFrom); qty != 12 {
		t.Errorf("balance after receipt confirm: want 12, got %v", qty)
	}
	if n := movementCountByKind(t, db, f.CompanyID, "purchase_receipt"); n != 1 {
		t.Errorf("purchase_receipt movements: want 1, got %d", n)
	}
}

func movementCountByKind(t *testing.T, db *sql.DB, companyID int, kind string) int {
	t.Helper()
	var n int
	must(t, db.QueryRow(
		`SELECT count(*) FROM inventory_movements WHERE company_id=$1 AND transaction_kind=$2`,
		companyID, kind).Scan(&n))
	return n
}
