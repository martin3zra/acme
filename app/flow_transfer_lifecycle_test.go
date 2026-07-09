package app

import (
	"testing"
	"time"
)

// The transfer write paths, converted to playsql. storeTransfer's line insert is
// already exercised by flow_transfer_variant_lines_test.go; the status transitions,
// the NULLIF-empty behaviour and reverseMovements' exclusion were untested.

func transferUUID(t *testing.T, f *fixture) string {
	t.Helper()
	return scalarString(t, f.s.db,
		`SELECT uuid::text FROM inventory_transfers WHERE company_id = $1 ORDER BY id DESC LIMIT 1`,
		f.company.ID)
}

// mkStockedTransfer stocks the source warehouse and requests a transfer of qty.
func mkStockedTransfer(t *testing.T, f *fixture, dest, qty int) (string, int) {
	t.Helper()
	g := newFaker(t)
	itemID, variantID := mkItem(t, f, 100, 60)

	// Bring stock in via a confirmed vendor bill.
	vendorID, _ := newVendor(t, f, g).Build()
	billUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		WithLine(itemID, qty, 100, 18).Build()
	if err := f.s.confirmPurchase(f.ctx, billUUID); err != nil {
		t.Fatalf("confirmPurchase: %v", err)
	}

	if err := f.s.storeTransfer(f.ctx, &StoreTransferForm{
		FromWarehouseID: f.warehouseID, ToWarehouseID: dest,
		Lines: []TransferLineInput{{ID: itemID, VariantID: variantID, Qty: float64(qty)}},
	}); err != nil {
		t.Fatalf("storeTransfer: %v", err)
	}
	return transferUUID(t, f), variantID
}

// TestStoreTransfer_NullifEmpty: a nil map value writes SQL NULL, which is what the
// old NULLIF-against-empty-string produced for an empty note or description.
func TestStoreTransfer_NullifEmpty(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	dest := mkWarehouse(t, f, "Dest")

	itemID, variantID := mkItem(t, f, 100, 60)
	is.NoErr(s.storeTransfer(f.ctx, &StoreTransferForm{
		FromWarehouseID: f.warehouseID, ToWarehouseID: dest, Notes: "",
		Lines: []TransferLineInput{{ID: itemID, VariantID: variantID, Qty: 2, Description: ""}},
	}))

	uuid := transferUUID(t, f)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM inventory_transfers WHERE uuid = $1 AND notes IS NULL`, uuid), 1)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM inventory_transfer_lines tl
		   JOIN inventory_transfers t ON t.id = tl.transfer_id
		  WHERE t.uuid = $1 AND tl.description IS NULL`, uuid), 1)

	// A non-empty note is stored as given.
	is.NoErr(s.storeTransfer(f.ctx, &StoreTransferForm{
		FromWarehouseID: f.warehouseID, ToWarehouseID: dest, Notes: "urgent",
		Lines: []TransferLineInput{{ID: itemID, VariantID: variantID, Qty: 1, Description: "top shelf"}},
	}))
	uuid = transferUUID(t, f)
	is.Equal(scalarString(t, s.db, `SELECT notes FROM inventory_transfers WHERE uuid = $1`, uuid), "urgent")
}

// TestTransferLifecycle: requested -> in_transit -> received, moving stock between
// warehouses and stamping each transition's timestamp plus updated_at.
func TestTransferLifecycle(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	dest := mkWarehouse(t, f, "Dest")

	uuid, variantID := mkStockedTransfer(t, f, dest, 10)

	qtyAt := func(wh int) float64 {
		return scalarFloat(t, s.db,
			`SELECT COALESCE(SUM(quantity),0) FROM inventory_balances
			  WHERE company_id = $1 AND variant_id = $2 AND warehouse_id = $3`,
			f.company.ID, variantID, wh)
	}
	is.EqualFloat(qtyAt(f.warehouseID), 10)
	is.EqualFloat(qtyAt(dest), 0)

	before := scalarString(t, s.db, `SELECT updated_at::text FROM inventory_transfers WHERE uuid = $1`, uuid)
	time.Sleep(2 * time.Millisecond)

	is.NoErr(s.dispatchTransfer(f.ctx, uuid))
	is.Equal(scalarString(t, s.db, `SELECT status::text FROM inventory_transfers WHERE uuid = $1`, uuid), "in_transit")
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM inventory_transfers WHERE uuid = $1 AND dispatched_at IS NOT NULL`, uuid), 1)
	is.True(scalarString(t, s.db, `SELECT updated_at::text FROM inventory_transfers WHERE uuid = $1`, uuid) != before,
		"dispatch should bump updated_at")
	is.EqualFloat(qtyAt(f.warehouseID), 0) // stock left the source
	is.EqualFloat(qtyAt(dest), 0)          // not arrived yet

	is.NoErr(s.receiveTransfer(f.ctx, uuid))
	is.Equal(scalarString(t, s.db, `SELECT status::text FROM inventory_transfers WHERE uuid = $1`, uuid), "received")
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM inventory_transfers WHERE uuid = $1 AND received_at IS NOT NULL`, uuid), 1)
	is.EqualFloat(qtyAt(dest), 10)
}

// TestDispatchTransfer_InsufficientStock: the FOR UPDATE balance check stays raw,
// and still refuses to move more than is on hand.
func TestDispatchTransfer_InsufficientStock(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	dest := mkWarehouse(t, f, "Dest")

	itemID, variantID := mkItem(t, f, 100, 60)
	is.NoErr(s.storeTransfer(f.ctx, &StoreTransferForm{
		FromWarehouseID: f.warehouseID, ToWarehouseID: dest,
		Lines: []TransferLineInput{{ID: itemID, VariantID: variantID, Qty: 5}},
	}))
	uuid := transferUUID(t, f)

	err := s.dispatchTransfer(f.ctx, uuid)
	is.Err(err, "dispatching more than is on hand must fail")

	// Status untouched, no movement recorded.
	is.Equal(scalarString(t, s.db, `SELECT status::text FROM inventory_transfers WHERE uuid = $1`, uuid), "requested")
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM inventory_movements WHERE company_id = $1 AND variant_id = $2`,
		f.company.ID, variantID), 0)
}

// TestCancelTransfer: requested -> cancelled, and only before dispatch.
func TestCancelTransfer(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	dest := mkWarehouse(t, f, "Dest")

	uuid, _ := mkStockedTransfer(t, f, dest, 4)

	is.NoErr(s.cancelTransfer(f.ctx, uuid))
	is.Equal(scalarString(t, s.db, `SELECT status::text FROM inventory_transfers WHERE uuid = $1`, uuid), "cancelled")

	// A cancelled transfer cannot then be dispatched.
	is.Err(s.dispatchTransfer(f.ctx, uuid), "a cancelled transfer must not dispatch")

	// And a dispatched one cannot be cancelled.
	other, _ := mkStockedTransfer(t, f, dest, 3)
	is.NoErr(s.dispatchTransfer(f.ctx, other))
	is.Err(s.cancelTransfer(f.ctx, other), "an in-transit transfer must not cancel")
}

// TestReverseMovements_SkipsReturns: reverseMovements excludes rows already booked
// as returns, so voiding twice does not double the reversal.
func TestReverseMovements_SkipsReturns(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, variantID := mkItem(t, f, 100, 60)
	vendorID, _ := newVendor(t, f, g).Build()
	billUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		WithLine(itemID, 6, 100, 18).Build()
	is.NoErr(s.confirmPurchase(f.ctx, billUUID))

	onHand := func() float64 {
		return scalarFloat(t, s.db,
			`SELECT COALESCE(SUM(quantity),0) FROM inventory_balances
			  WHERE company_id = $1 AND variant_id = $2`, f.company.ID, variantID)
	}
	is.EqualFloat(onHand(), 6)

	// Deleting the bill reverses its movements exactly once.
	is.NoErr(s.destroyPurchase(f.ctx, billUUID))
	is.EqualFloat(onHand(), 0)

	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM inventory_movements
		  WHERE company_id = $1 AND variant_id = $2 AND transaction_kind = 'purchase_return'`,
		f.company.ID, variantID), 1)
}
