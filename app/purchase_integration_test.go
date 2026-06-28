//go:build integration

package app

import "testing"

// Confirming a draft purchase receipt posts inventory movements for each line
// and transitions the purchase to received.
func TestIntegration_ConfirmPurchase_ReceiptRecordsMovements(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)
	ctx := companyCtx(f.CompanyID)

	var vendorID int
	must(t, db.QueryRow(
		`INSERT INTO vendors (company_id, name) VALUES ($1, 'Acme Supply') RETURNING id`,
		f.CompanyID).Scan(&vendorID))

	var unitID int
	must(t, db.QueryRow(
		`INSERT INTO units (company_id, name, base_qty) VALUES ($1, 'unit', 1) RETURNING id`,
		f.CompanyID).Scan(&unitID))

	// Draft purchase receipt landing stock in WHFrom.
	var purchaseID int
	var purchaseUUID string
	must(t, db.QueryRow(
		`INSERT INTO purchases (company_id, vendor_id, warehouse_id, code, transaction_kind)
		 VALUES ($1, $2, $3, 'PR-001', 'purchase_receipt') RETURNING id, uuid`,
		f.CompanyID, vendorID, f.WHFrom).Scan(&purchaseID, &purchaseUUID))

	must(t, exec(db,
		`INSERT INTO purchase_items (company_id, purchase_id, variant_id, qty, unit_price, unit_id)
		 VALUES ($1, $2, $3, 12, 5, $4)`,
		f.CompanyID, purchaseID, f.VariantID, unitID))

	if err := srv.confirmPurchase(ctx, purchaseUUID); err != nil {
		t.Fatalf("confirmPurchase: %v", err)
	}

	// Stock posted to the receipt warehouse.
	if qty, ok := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHFrom); !ok || qty != 12 {
		t.Errorf("balance: want 12, got %v (exists=%v)", qty, ok)
	}

	// One purchase_receipt movement recorded for the line.
	var movements int
	must(t, db.QueryRow(
		`SELECT count(*) FROM inventory_movements
		  WHERE company_id=$1 AND transaction_kind='purchase_receipt'`, f.CompanyID).Scan(&movements))
	if movements != 1 {
		t.Errorf("purchase_receipt movements: want 1, got %d", movements)
	}

	// Status transitioned and movement recorded flag set.
	var status string
	var recorded bool
	must(t, db.QueryRow(
		`SELECT purchase_status, movement_recorded FROM purchases WHERE id=$1`, purchaseID).
		Scan(&status, &recorded))
	if status != string(PurchaseStatuses.Received) {
		t.Errorf("status: want received, got %q", status)
	}
	if !recorded {
		t.Error("movement_recorded should be true after confirm")
	}
}

// A purchase order (not a receipt or vendor bill) cannot be confirmed.
func TestIntegration_ConfirmPurchase_RejectsPurchaseOrder(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)
	ctx := companyCtx(f.CompanyID)

	var vendorID int
	must(t, db.QueryRow(
		`INSERT INTO vendors (company_id, name) VALUES ($1, 'Acme Supply') RETURNING id`,
		f.CompanyID).Scan(&vendorID))

	var uuid string
	must(t, db.QueryRow(
		`INSERT INTO purchases (company_id, vendor_id, warehouse_id, code, transaction_kind)
		 VALUES ($1, $2, $3, 'PO-001', 'purchase_order') RETURNING uuid`,
		f.CompanyID, vendorID, f.WHFrom).Scan(&uuid))

	if err := srv.confirmPurchase(ctx, uuid); err == nil {
		t.Fatal("expected confirm of a purchase_order to be rejected")
	}
}
