package app

import (
	"errors"
	"testing"
)

// Warehouse ids arrive straight from the request on invoice lines and on a
// transfer's from/to. Nothing downstream checked them: recordMovement took the id on
// trust, so a foreign warehouse produced an inventory_balances row keyed to the
// caller's company_id and someone else's warehouse_id. findStocks joins
// `w.company_id = ib.company_id`, so that row leaked nothing — it simply vanished,
// leaving stock recorded and invisible to everyone.

// TestRecordMovement_RejectsForeignWarehouse: the choke point for every stock write.
func TestRecordMovement_RejectsForeignWarehouse(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	other := mkAccountCompany(t, s)

	_, variantID := mkItem(t, f, 100, 60)

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()

	err = s.recordMovement(tx, f.company.ID, variantID, other.warehouseID, f.unitID,
		5, 60, InventoryMovementKinds.VendorBill, "purchase", 1)
	is.True(errors.Is(err, ErrWarehouseNotInCompany), "another company's warehouse must be rejected")

	// Own warehouse still works, and the stock lands where it can be seen.
	is.NoErr(s.recordMovement(tx, f.company.ID, variantID, f.warehouseID, f.unitID,
		5, 60, InventoryMovementKinds.VendorBill, "purchase", 1))
	is.NoErr(tx.Commit())

	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM inventory_balances WHERE company_id = $1 AND warehouse_id = $2`,
		f.company.ID, other.warehouseID), 0)

	stocks, err := s.findStocks(f.ctx, f.company.ID)
	is.NoErr(err)
	is.Equal(len(stocks), 1) // visible, not vanished
}

// TestRecordMovement_RejectsDeletedWarehouse: warehouseRead is softdelete-tagged, so
// a deleted warehouse does not count as owned.
func TestRecordMovement_RejectsDeletedWarehouse(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, variantID := mkItem(t, f, 100, 60)
	gone := mkWarehouse(t, f, "Retired")
	is.NoErr(s.deleteWarehouse(f.ctx, gone))

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()

	err = s.recordMovement(tx, f.company.ID, variantID, gone, f.unitID,
		5, 60, InventoryMovementKinds.VendorBill, "purchase", 1)
	is.True(errors.Is(err, ErrWarehouseNotInCompany), "a deleted warehouse must be rejected")
}

// TestStoreTransfer_RejectsForeignWarehouse: from/to come straight from the form.
func TestStoreTransfer_RejectsForeignWarehouse(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	other := mkAccountCompany(t, s)

	itemID, variantID := mkItem(t, f, 100, 60)
	dest := mkWarehouse(t, f, "Dest")

	before := scalarInt(t, s.db,
		`SELECT count(*) FROM inventory_transfers WHERE company_id = $1`, f.company.ID)

	// Destination belongs to another tenant.
	err := s.storeTransfer(f.ctx, &StoreTransferForm{
		FromWarehouseID: f.warehouseID, ToWarehouseID: other.warehouseID,
		Lines: []TransferLineInput{{ID: itemID, VariantID: variantID, Qty: 1}},
	})
	is.True(errors.Is(err, ErrWarehouseNotInCompany), "a foreign destination must be rejected")

	// Source belongs to another tenant.
	err = s.storeTransfer(f.ctx, &StoreTransferForm{
		FromWarehouseID: other.warehouseID, ToWarehouseID: dest,
		Lines: []TransferLineInput{{ID: itemID, VariantID: variantID, Qty: 1}},
	})
	is.True(errors.Is(err, ErrWarehouseNotInCompany), "a foreign source must be rejected")

	// Nothing was written.
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM inventory_transfers WHERE company_id = $1`, f.company.ID), before)

	// Both warehouses owned: accepted.
	is.NoErr(s.storeTransfer(f.ctx, &StoreTransferForm{
		FromWarehouseID: f.warehouseID, ToWarehouseID: dest,
		Lines: []TransferLineInput{{ID: itemID, VariantID: variantID, Qty: 1}},
	}))
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM inventory_transfers WHERE company_id = $1`, f.company.ID), before+1)
}

// TestAttachInvoiceLines_RejectsForeignWarehouse: a draft writes its lines without
// moving stock, so the check cannot live only in recordMovement.
func TestAttachInvoiceLines_RejectsForeignWarehouse(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	other := mkAccountCompany(t, s)

	itemID, _ := mkItem(t, f, 100, 60)

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()

	foreign := &StoreInvoiceForm{Lines: []*Line{
		mkLine(itemID, f.unitID, other.warehouseID, 1, 100, 18),
	}}
	err = s.attachInvoiceLines(tx, f.company.ID, 1, foreign)
	is.True(errors.Is(err, ErrWarehouseNotInCompany), "a foreign line warehouse must be rejected")

	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM invoices_items WHERE company_id = $1`, f.company.ID), 0)
}
