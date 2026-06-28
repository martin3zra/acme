//go:build integration

package app

import (
	"errors"
	"testing"
)

func TestIntegration_RecordMovement_UpsertsBalance(t *testing.T) {
	db, _, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)

	// Two inbound movements accumulate on the same balance row.
	for _, qty := range []float64{10, 5} {
		tx, err := db.Begin()
		must(t, err)
		if err := srv.recordMovement(tx, f.CompanyID, f.VariantID, f.WHFrom, 0,
			qty, 0, InventoryMovementKinds.Adjustment, "seed", 0); err != nil {
			tx.Rollback()
			t.Fatalf("recordMovement: %v", err)
		}
		must(t, tx.Commit())
	}

	got, ok := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHFrom)
	if !ok || got != 15 {
		t.Fatalf("balance: want 15, got %v (exists=%v)", got, ok)
	}

	var movements int
	must(t, db.QueryRow(
		`SELECT count(*) FROM inventory_movements WHERE company_id=$1`, f.CompanyID).Scan(&movements))
	if movements != 2 {
		t.Errorf("movements: want 2, got %d", movements)
	}
}

func TestIntegration_RecordMovement_SkipsUntrackedVariant(t *testing.T) {
	db, _, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)

	// A second variant on the same item with inventory tracking disabled.
	var untracked int
	must(t, db.QueryRow(
		`INSERT INTO items_variants (company_id, item_id, name, track_inventory)
		 VALUES ($1, $2, 'Untracked', FALSE) RETURNING id`, f.CompanyID, f.ItemID).Scan(&untracked))

	tx, err := db.Begin()
	must(t, err)
	if err := srv.recordMovement(tx, f.CompanyID, untracked, f.WHFrom, 0,
		7, 0, InventoryMovementKinds.Adjustment, "seed", 0); err != nil {
		tx.Rollback()
		t.Fatalf("recordMovement: %v", err)
	}
	must(t, tx.Commit())

	// No movement and no balance row for an untracked variant.
	if _, ok := balanceQty(t, db, f.CompanyID, untracked, f.WHFrom); ok {
		t.Error("untracked variant should not produce a balance row")
	}
	var movements int
	must(t, db.QueryRow(
		`SELECT count(*) FROM inventory_movements WHERE company_id=$1 AND variant_id=$2`,
		f.CompanyID, untracked).Scan(&movements))
	if movements != 0 {
		t.Errorf("untracked variant should record no movements, got %d", movements)
	}
}

func TestIntegration_RecordMovement_AppliesUnitConversion(t *testing.T) {
	db, _, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)

	// A unit of 3 base units each.
	var unitID int
	must(t, db.QueryRow(
		`INSERT INTO units (company_id, name, base_qty) VALUES ($1, 'box', 3) RETURNING id`,
		f.CompanyID).Scan(&unitID))

	tx, err := db.Begin()
	must(t, err)
	// 4 boxes × 3 base units = 12 on hand.
	if err := srv.recordMovement(tx, f.CompanyID, f.VariantID, f.WHFrom, unitID,
		4, 0, InventoryMovementKinds.Adjustment, "seed", 0); err != nil {
		tx.Rollback()
		t.Fatalf("recordMovement: %v", err)
	}
	must(t, tx.Commit())

	if qty, _ := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHFrom); qty != 12 {
		t.Errorf("balance after unit conversion: want 12, got %v", qty)
	}
}

func TestIntegration_StoreAdjustment_PositiveAndNegative(t *testing.T) {
	db, _, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)
	ctx := companyCtx(f.CompanyID)

	steps := []struct {
		qty  float64
		want float64
	}{
		{10, 10},
		{-4, 6},
		{-10, -4}, // adjustments may drive the balance negative
	}
	for _, s := range steps {
		form := &StoreAdjustmentForm{
			VariantID:   f.VariantID,
			WarehouseID: f.WHFrom,
			Qty:         s.qty,
			Reason:      "count correction",
		}
		if err := srv.storeAdjustment(ctx, form); err != nil {
			t.Fatalf("storeAdjustment(%v): %v", s.qty, err)
		}
		if got, _ := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHFrom); got != s.want {
			t.Errorf("after qty %v: balance want %v, got %v", s.qty, s.want, got)
		}
	}

	var n int
	must(t, db.QueryRow(
		`SELECT count(*) FROM inventory_movements
		  WHERE company_id=$1 AND transaction_kind='adjustment'`, f.CompanyID).Scan(&n))
	if n != 3 {
		t.Errorf("adjustment movements: want 3, got %d", n)
	}
}

func TestIntegration_StoreTransfer_HappyPath(t *testing.T) {
	db, _, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)
	ctx := companyCtx(f.CompanyID)

	// Seed 20 units at the source.
	seedTx, _ := db.Begin()
	must(t, srv.recordMovement(seedTx, f.CompanyID, f.VariantID, f.WHFrom, 0,
		20, 0, InventoryMovementKinds.Adjustment, "seed", 0))
	must(t, seedTx.Commit())

	form := &StoreTransferForm{
		VariantID:       f.VariantID,
		FromWarehouseID: f.WHFrom,
		ToWarehouseID:   f.WHTo,
		Qty:             8,
	}
	if err := srv.storeTransfer(ctx, form); err != nil {
		t.Fatalf("storeTransfer: %v", err)
	}

	if src, _ := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHFrom); src != 12 {
		t.Errorf("source balance: want 12, got %v", src)
	}
	if dst, _ := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHTo); dst != 8 {
		t.Errorf("destination balance: want 8, got %v", dst)
	}

	// Two transfer movements (out + in) were recorded.
	var n int
	must(t, db.QueryRow(
		`SELECT count(*) FROM inventory_movements
		  WHERE company_id=$1 AND transaction_kind='transfer'`, f.CompanyID).Scan(&n))
	if n != 2 {
		t.Errorf("transfer movements: want 2, got %d", n)
	}
}

func TestIntegration_StoreTransfer_ExactAvailableSucceeds(t *testing.T) {
	db, _, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)
	ctx := companyCtx(f.CompanyID)

	seedTx, _ := db.Begin()
	must(t, srv.recordMovement(seedTx, f.CompanyID, f.VariantID, f.WHFrom, 0,
		5, 0, InventoryMovementKinds.Adjustment, "seed", 0))
	must(t, seedTx.Commit())

	// Transferring exactly what's on hand is allowed (the guard is available < qty).
	form := &StoreTransferForm{
		VariantID:       f.VariantID,
		FromWarehouseID: f.WHFrom,
		ToWarehouseID:   f.WHTo,
		Qty:             5,
	}
	if err := srv.storeTransfer(ctx, form); err != nil {
		t.Fatalf("storeTransfer of exact available qty: %v", err)
	}
	if src, _ := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHFrom); src != 0 {
		t.Errorf("source balance: want 0, got %v", src)
	}
	if dst, _ := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHTo); dst != 5 {
		t.Errorf("destination balance: want 5, got %v", dst)
	}
}

func TestIntegration_StoreTransfer_InsufficientStock(t *testing.T) {
	db, _, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)
	ctx := companyCtx(f.CompanyID)

	// Source has only 3 units.
	seedTx, _ := db.Begin()
	must(t, srv.recordMovement(seedTx, f.CompanyID, f.VariantID, f.WHFrom, 0,
		3, 0, InventoryMovementKinds.Adjustment, "seed", 0))
	must(t, seedTx.Commit())

	form := &StoreTransferForm{
		VariantID:       f.VariantID,
		FromWarehouseID: f.WHFrom,
		ToWarehouseID:   f.WHTo,
		Qty:             10,
	}
	err := srv.storeTransfer(ctx, form)
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("want ErrInsufficientStock, got %v", err)
	}

	// Nothing moved: source unchanged, destination has no row.
	if src, _ := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHFrom); src != 3 {
		t.Errorf("source balance should be untouched at 3, got %v", src)
	}
	if _, ok := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHTo); ok {
		t.Error("destination balance should not exist after a blocked transfer")
	}
}

func TestIntegration_StoreTransfer_SameWarehouseRejected(t *testing.T) {
	db, _, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)
	ctx := companyCtx(f.CompanyID)

	form := &StoreTransferForm{
		VariantID:       f.VariantID,
		FromWarehouseID: f.WHFrom,
		ToWarehouseID:   f.WHFrom,
		Qty:             1,
	}
	if err := srv.storeTransfer(ctx, form); !errors.Is(err, ErrSameWarehouse) {
		t.Fatalf("want ErrSameWarehouse, got %v", err)
	}
}
