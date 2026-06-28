//go:build integration

package app

import "testing"

// Recording invoice (sale) movements draws stock down; reversing them — the
// path a voided invoice takes — restores it as sale_return entries.
func TestIntegration_InvoiceSaleMovementsAndReversal(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)
	ctx := companyCtx(f.CompanyID)

	var unitID int
	must(t, db.QueryRow(
		`INSERT INTO units (company_id, name, base_qty) VALUES ($1, 'unit', 1) RETURNING id`,
		f.CompanyID).Scan(&unitID))

	// Start with 20 on hand.
	must(t, srv.storeAdjustment(ctx, &StoreAdjustmentForm{
		VariantID: f.VariantID, WarehouseID: f.WHFrom, Qty: 20, Reason: "seed"}))

	const invoiceID = 1
	lines := []*Line{
		{ID: f.ItemID, Unit: unitID, Qty: 5, Price: 5, WarehouseID: f.WHFrom},
	}

	// Sale draws stock down: 20 - 5 = 15.
	tx, err := db.Begin()
	must(t, err)
	if err := srv.recordInvoiceMovements(tx, f.CompanyID, invoiceID, lines); err != nil {
		tx.Rollback()
		t.Fatalf("recordInvoiceMovements: %v", err)
	}
	must(t, tx.Commit())

	if qty, _ := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHFrom); qty != 15 {
		t.Fatalf("balance after sale: want 15, got %v", qty)
	}
	if n := movementCountByKind(t, db, f.CompanyID, "sale"); n != 1 {
		t.Errorf("sale movements: want 1, got %d", n)
	}

	// Void/reversal restores stock: 15 + 5 = 20.
	tx2, err := db.Begin()
	must(t, err)
	if err := srv.reverseMovements(tx2, f.CompanyID, "invoice", invoiceID, InventoryMovementKinds.SaleReturn); err != nil {
		tx2.Rollback()
		t.Fatalf("reverseMovements: %v", err)
	}
	must(t, tx2.Commit())

	if qty, _ := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHFrom); qty != 20 {
		t.Errorf("balance after reversal: want 20, got %v", qty)
	}
	if n := movementCountByKind(t, db, f.CompanyID, "sale_return"); n != 1 {
		t.Errorf("sale_return movements: want 1, got %d", n)
	}
}
