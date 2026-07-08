package app

import "testing"

// mkWarehouse creates an extra warehouse for the fixture's company (transfers
// need a distinct source and destination).
func mkWarehouse(t *testing.T, f *fixture, name string) int {
	t.Helper()
	var id int
	if err := f.s.db.QueryRow(
		`INSERT INTO warehouses (company_id, name) VALUES ($1, $2) RETURNING id`, f.company.ID, name,
	).Scan(&id); err != nil {
		t.Fatalf("insert warehouse: %v", err)
	}
	return id
}

// TestFlowTransferVariantExplicit: a transfer line naming a variant persists that
// exact variant on inventory_transfer_lines — not a guessed one.
func TestFlowTransferVariantExplicit(t *testing.T) {
	s := newTestServer(t)
	f := mkAccountCompany(t, s)
	dest := mkWarehouse(t, f, "Dest")

	itemID, variantIDs := mkVariantItem(t, f, 100)
	blue := variantIDs[1]

	if err := f.s.storeTransfer(f.ctx, &StoreTransferForm{
		FromWarehouseID: f.warehouseID, ToWarehouseID: dest,
		Lines: []TransferLineInput{{ID: itemID, VariantID: blue, Qty: 2}},
	}); err != nil {
		t.Fatalf("storeTransfer: %v", err)
	}
	assertRow(t, s.db, "inventory_transfer_lines", map[string]any{"variant_id": blue, "qty": 2})
}

// TestFlowTransferVariantRequired: a has_variants item on a line with no variant
// is rejected — transfers no longer silently default to a variant of a variant
// product (the old resolveDefaultVariantIDs behaviour).
func TestFlowTransferVariantRequired(t *testing.T) {
	s := newTestServer(t)
	f := mkAccountCompany(t, s)
	dest := mkWarehouse(t, f, "Dest")

	itemID, _ := mkVariantItem(t, f, 100)

	err := f.s.storeTransfer(f.ctx, &StoreTransferForm{
		FromWarehouseID: f.warehouseID, ToWarehouseID: dest,
		Lines: []TransferLineInput{{ID: itemID, Qty: 1}}, // VariantID left 0
	})
	if err == nil {
		t.Fatal("expected error: variant item transferred without a variant")
	}
	// Nothing persisted (transaction rolled back).
	assertNoRow(t, s.db, "inventory_transfer_lines", map[string]any{"company_id": f.company.ID})
}

// TestFlowTransferPlainDefaults: a plain item with no variant resolves to its
// default variant, unchanged behaviour for the common case.
func TestFlowTransferPlainDefaults(t *testing.T) {
	s := newTestServer(t)
	f := mkAccountCompany(t, s)
	dest := mkWarehouse(t, f, "Dest")

	itemID, defaultVariant := mkItem(t, f, 100, 60)

	if err := f.s.storeTransfer(f.ctx, &StoreTransferForm{
		FromWarehouseID: f.warehouseID, ToWarehouseID: dest,
		Lines: []TransferLineInput{{ID: itemID, Qty: 3}},
	}); err != nil {
		t.Fatalf("storeTransfer: %v", err)
	}
	assertRow(t, s.db, "inventory_transfer_lines", map[string]any{"variant_id": defaultVariant, "qty": 3})
}

// TestFlowTransferCrossItemVariantRejected: a line may not reference a variant
// that belongs to a different item.
func TestFlowTransferCrossItemVariantRejected(t *testing.T) {
	s := newTestServer(t)
	f := mkAccountCompany(t, s)
	dest := mkWarehouse(t, f, "Dest")

	_, variantIDs := mkVariantItem(t, f, 100)
	plainItem, _ := mkItem(t, f, 50, 30)

	err := f.s.storeTransfer(f.ctx, &StoreTransferForm{
		FromWarehouseID: f.warehouseID, ToWarehouseID: dest,
		Lines: []TransferLineInput{{ID: plainItem, VariantID: variantIDs[0], Qty: 1}},
	})
	if err == nil {
		t.Fatalf("expected error: variant %d does not belong to item %d", variantIDs[0], plainItem)
	}
}

// TestFlowTransferSearchExpandsPerVariant: the transfer item search returns one
// row per trackable variant, each carrying its own variant id, so the picker can
// land on a specific variant.
func TestFlowTransferSearchExpandsPerVariant(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, variantIDs := mkVariantItem(t, f, 100)

	opts, err := f.s.findTransferItems(f.ctx, f.company.ID, "")
	is.NoErr(err)

	got := map[int64]bool{}
	for _, o := range opts {
		if int(o.ID) == itemID {
			got[o.VariantID] = true
		}
	}
	for _, v := range variantIDs {
		if !got[int64(v)] {
			t.Fatalf("variant %d missing from transfer search results", v)
		}
	}
}
