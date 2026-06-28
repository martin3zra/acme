//go:build integration

package app

import (
	"database/sql"
	"testing"
)

// Confirming a draft purchase receipt posts inventory movements for each line
// and transitions the purchase to received.
func TestIntegration_ConfirmPurchase_ReceiptRecordsMovements(t *testing.T) {
	db, _, cleanup := newTestDB(t)
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

// Deleting a confirmed receipt reverses its inventory movements, returning the
// balance to zero and recording purchase_return entries.
func TestIntegration_DestroyConfirmedReceipt_ReversesMovements(t *testing.T) {
	db, _, cleanup := newTestDB(t)
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
	if qty, _ := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHFrom); qty != 12 {
		t.Fatalf("precondition: balance should be 12 after confirm, got %v", qty)
	}

	if err := srv.destroyPurchase(ctx, purchaseUUID); err != nil {
		t.Fatalf("destroyPurchase: %v", err)
	}

	// Reversal returns the balance to zero.
	if qty, _ := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHFrom); qty != 0 {
		t.Errorf("balance after reversal: want 0, got %v", qty)
	}

	// A purchase_return movement was recorded.
	var returns int
	must(t, db.QueryRow(
		`SELECT count(*) FROM inventory_movements
		  WHERE company_id=$1 AND transaction_kind='purchase_return'`, f.CompanyID).Scan(&returns))
	if returns == 0 {
		t.Error("expected at least one purchase_return movement")
	}

	// The purchase is soft-deleted.
	var deletedAt sql.NullTime
	must(t, db.QueryRow(
		`SELECT deleted_at FROM purchases WHERE id=$1`, purchaseID).Scan(&deletedAt))
	if !deletedAt.Valid {
		t.Error("purchase should be soft-deleted")
	}
}

// A receipt already confirmed cannot be confirmed again.
func TestIntegration_ConfirmPurchase_RejectsDoubleConfirm(t *testing.T) {
	db, _, cleanup := newTestDB(t)
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

	var purchaseID int
	var uuid string
	must(t, db.QueryRow(
		`INSERT INTO purchases (company_id, vendor_id, warehouse_id, code, transaction_kind)
		 VALUES ($1, $2, $3, 'PR-001', 'purchase_receipt') RETURNING id, uuid`,
		f.CompanyID, vendorID, f.WHFrom).Scan(&purchaseID, &uuid))
	must(t, exec(db,
		`INSERT INTO purchase_items (company_id, purchase_id, variant_id, qty, unit_price, unit_id)
		 VALUES ($1, $2, $3, 4, 5, $4)`, f.CompanyID, purchaseID, f.VariantID, unitID))

	if err := srv.confirmPurchase(ctx, uuid); err != nil {
		t.Fatalf("first confirm: %v", err)
	}
	if err := srv.confirmPurchase(ctx, uuid); err == nil {
		t.Fatal("second confirm should be rejected (not in draft status)")
	}

	// Movements were posted exactly once.
	var n int
	must(t, db.QueryRow(
		`SELECT count(*) FROM inventory_movements
		  WHERE company_id=$1 AND transaction_kind='purchase_receipt'`, f.CompanyID).Scan(&n))
	if n != 1 {
		t.Errorf("purchase_receipt movements: want 1 (no double posting), got %d", n)
	}
}

// A purchase order (not a receipt or vendor bill) cannot be confirmed.
func TestIntegration_ConfirmPurchase_RejectsPurchaseOrder(t *testing.T) {
	db, _, cleanup := newTestDB(t)
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
