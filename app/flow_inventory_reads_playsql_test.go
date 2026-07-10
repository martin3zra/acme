package app

import (
	"testing"
)

// findMovements, findAdjustments and findTransferLines converted to playsql. None of
// the three had direct coverage. The remaining reads in inventory-repository.go are
// aggregates or order by joined columns and stay raw.

// dbTimeOf renders a movement's created_at exactly as the old TO_CHAR did, so the
// converted Go formatting can be compared against the database's own rendering.
func dbTimeOf(t *testing.T, s *Server, movementID int64) string {
	t.Helper()
	return scalarString(t, s.db,
		`SELECT TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI') FROM inventory_movements WHERE id = $1`,
		movementID)
}

// mkAdjustment records one manual adjustment and returns the movement id.
func mkAdjustment(t *testing.T, f *fixture, variantID, warehouseID int, qty float64, reason string) int64 {
	t.Helper()
	if err := f.s.storeAdjustment(f.ctx, &StoreAdjustmentForm{
		VariantID: variantID, WarehouseID: warehouseID, Qty: qty, Reason: reason,
	}); err != nil {
		t.Fatalf("storeAdjustment: %v", err)
	}
	var id int64
	if err := f.s.db.QueryRow(
		`SELECT id FROM inventory_movements WHERE company_id = $1 ORDER BY id DESC LIMIT 1`,
		f.company.ID).Scan(&id); err != nil {
		t.Fatalf("load movement: %v", err)
	}
	return id
}

// ─── findMovements ────────────────────────────────────────────────────────────

// TestFindMovements_LoadsRelationsAndFormatsTime: the three INNER JOINs become
// relations, and TO_CHAR becomes a Go format of the scanned value.
func TestFindMovements_LoadsRelationsAndFormatsTime(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, variantID := mkItem(t, f, 100, 60)
	id := mkAdjustment(t, f, variantID, f.warehouseID, 5, "recount")

	rows, err := s.findMovements(f.ctx, f.company.ID)
	is.NoErr(err)
	is.Equal(len(rows), 1)

	m := rows[0]
	is.Equal(m.ID, id)
	is.Equal(m.VariantID, int64(variantID))
	is.Equal(m.WarehouseID, int64(f.warehouseID))
	is.Equal(m.Kind, "adjustment")
	is.EqualFloat(m.Qty, 5)
	is.Equal(m.ReferenceType, "recount")
	is.Equal(m.ReferenceID, int64(0)) // NULL/0 scans to zero — the old COALESCE

	is.Equal(m.ItemName, scalarString(t, s.db, `SELECT name FROM items WHERE id = $1`, itemID))
	is.Equal(m.VariantName, scalarString(t, s.db,
		`SELECT name FROM items_variants WHERE id = $1`, variantID))
	is.Equal(m.Warehouse, scalarString(t, s.db,
		`SELECT name FROM warehouses WHERE id = $1`, f.warehouseID))

	// The Go format renders the same wall clock the database's TO_CHAR did.
	is.Equal(m.CreatedAt, dbTimeOf(t, s, id))
	is.Equal(len(m.CreatedAt), len("2006-01-02 15:04"))
}

// TestFindMovements_OrdersNewestFirst: ORDER BY created_at DESC on the movement's own
// column. recordMovement stamps created_at with Go's clock, so the rows differ.
func TestFindMovements_OrdersNewestFirst(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, variantID := mkItem(t, f, 100, 60)
	first := mkAdjustment(t, f, variantID, f.warehouseID, 1, "first")
	second := mkAdjustment(t, f, variantID, f.warehouseID, 2, "second")

	rows, err := s.findMovements(f.ctx, f.company.ID)
	is.NoErr(err)
	is.Equal(len(rows), 2)
	is.Equal(rows[0].ID, second)
	is.Equal(rows[1].ID, first)
}

// TestFindMovements_ExcludesSoftDeletedVariantOrItem pins liveVariantAndItem. Neither
// itemVariantRead nor lineItemRead carries a softdelete tag, so the two
// `deleted_at IS NULL` predicates the INNER JOINs held are written out.
func TestFindMovements_ExcludesSoftDeletedVariantOrItem(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, variantID := mkItem(t, f, 100, 60)
	mkAdjustment(t, f, variantID, f.warehouseID, 5, "recount")

	rows, err := s.findMovements(f.ctx, f.company.ID)
	is.NoErr(err)
	is.Equal(len(rows), 1)

	// A soft-deleted variant hides its movements.
	_, err = s.db.Exec(`UPDATE items_variants SET deleted_at = now() WHERE id = $1`, variantID)
	is.NoErr(err)
	rows, err = s.findMovements(f.ctx, f.company.ID)
	is.NoErr(err)
	is.Equal(len(rows), 0)

	// Restore the variant, delete the item instead: same result.
	_, err = s.db.Exec(`UPDATE items_variants SET deleted_at = NULL WHERE id = $1`, variantID)
	is.NoErr(err)
	rows, err = s.findMovements(f.ctx, f.company.ID)
	is.NoErr(err)
	is.Equal(len(rows), 1)

	is.NoErr(s.deleteItem(f.ctx, itemID))
	rows, err = s.findMovements(f.ctx, f.company.ID)
	is.NoErr(err)
	is.Equal(len(rows), 0)
}

// TestFindMovements_KeepsRetiredWarehouse pins the WithTrashed on the warehouse.
//
// warehouseRead is softdelete-tagged, but `JOIN warehouses` carried no deleted_at
// predicate. Without WithTrashed the eager load would drop the warehouse (blanking its
// name) and the WhereHas would drop the movement entirely — a retired warehouse would
// erase its own history.
func TestFindMovements_KeepsRetiredWarehouse(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, variantID := mkItem(t, f, 100, 60)
	mkAdjustment(t, f, variantID, f.warehouseID, 5, "recount")

	name := scalarString(t, s.db, `SELECT name FROM warehouses WHERE id = $1`, f.warehouseID)
	is.NoErr(s.deleteWarehouse(f.ctx, f.warehouseID))

	rows, err := s.findMovements(f.ctx, f.company.ID)
	is.NoErr(err)
	is.Equal(len(rows), 1)
	is.Equal(rows[0].Warehouse, name)
}

// TestFindMovements_ScopedToCompany.
func TestFindMovements_ScopedToCompany(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	first := mkAccountCompany(t, s)
	second := mkAccountCompany(t, s)

	_, v1 := mkItem(t, first, 100, 60)
	mkAdjustment(t, first, v1, first.warehouseID, 5, "recount")

	rows, err := s.findMovements(second.ctx, second.company.ID)
	is.NoErr(err)
	is.Equal(len(rows), 0)

	rows, err = s.findMovements(first.ctx, first.company.ID)
	is.NoErr(err)
	is.Equal(len(rows), 1)
}

// ─── findAdjustments ──────────────────────────────────────────────────────────

// TestFindAdjustments_OnlyTheAdjustmentKind: the sale movements a confirmed bill
// records must not appear in the adjustments list.
func TestFindAdjustments_OnlyTheAdjustmentKind(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	// A purchase records an `in` movement; the adjustment records an `adjustment` one.
	_, variantID := mkStockedTransfer(t, f, mkWarehouse(t, f, "Dest"), 10)
	id := mkAdjustment(t, f, variantID, f.warehouseID, -2, "damaged")

	all, err := s.findMovements(f.ctx, f.company.ID)
	is.NoErr(err)
	is.True(len(all) > 1, "the purchase and the adjustment both recorded movements")

	adjustments, err := s.findAdjustments(f.ctx, f.company.ID)
	is.NoErr(err)
	is.Equal(len(adjustments), 1)

	a := adjustments[0]
	is.Equal(a.ID, id)
	is.EqualFloat(a.Qty, -2)
	is.Equal(a.Reason, "damaged") // reference_type
	is.Equal(a.Notes, "")         // the projection's literal ''
	is.Equal(a.CreatedAt, dbTimeOf(t, s, id))
	is.True(a.ItemName != "", "the item is eager-loaded")
	is.True(a.Warehouse != "", "the warehouse is eager-loaded")
}

// ─── findTransferLines ────────────────────────────────────────────────────────

// TestFindTransferLines_LoadsVariantItemAndUnit.
func TestFindTransferLines_LoadsVariantItemAndUnit(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	dest := mkWarehouse(t, f, "Dest")

	uuid, variantID := mkStockedTransfer(t, f, dest, 4)
	transferID := int64(scalarInt(t, s.db, `SELECT id FROM inventory_transfers WHERE uuid = $1`, uuid))

	lines, err := s.findTransferLines(f.ctx, f.company.ID, transferID)
	is.NoErr(err)
	is.Equal(len(lines), 1)

	l := lines[0]
	is.Equal(l.VariantID, int64(variantID))
	is.EqualFloat(l.Qty, 4)
	is.EqualFloat(l.LineTotal, l.Qty*l.UnitCost) // computed in the mapper
	is.True(l.ItemName != "", "the item arrives through Variant.Item")
	is.True(l.VariantName != "", "the variant is eager-loaded")

	// storeTransfer writes no unit_id and no description for this line.
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM inventory_transfer_lines WHERE transfer_id = $1 AND unit_id IS NULL`,
		transferID), 1)
	is.Equal(l.Unit, "")        // the LEFT JOIN's outer arm
	is.Equal(l.Description, "") // COALESCE(tl.description, '')
	is.Equal(l.Reference, "")   // the item carries no identifiers.reference
}

// TestFindTransferLines_ResolvesUnitAndReference: a line with a unit, and an item with
// an identifiers.reference, fill the two fields the previous test found empty.
func TestFindTransferLines_ResolvesUnitAndReference(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	dest := mkWarehouse(t, f, "Dest")

	uuid, variantID := mkStockedTransfer(t, f, dest, 4)
	transferID := int64(scalarInt(t, s.db, `SELECT id FROM inventory_transfers WHERE uuid = $1`, uuid))
	itemID := scalarInt(t, s.db, `SELECT item_id FROM items_variants WHERE id = $1`, variantID)

	caja := mkUnit(t, f, "Caja", 12)
	_, err := s.db.Exec(
		`UPDATE inventory_transfer_lines SET unit_id = $2, description = 'handle with care'
		  WHERE transfer_id = $1`, transferID, caja)
	is.NoErr(err)
	_, err = s.db.Exec(
		`UPDATE items SET identifiers = '{"reference":"REF-9"}'::jsonb WHERE id = $1`, itemID)
	is.NoErr(err)

	lines, err := s.findTransferLines(f.ctx, f.company.ID, transferID)
	is.NoErr(err)
	is.Equal(len(lines), 1)
	is.Equal(lines[0].Unit, "Caja")
	is.Equal(lines[0].Description, "handle with care")
	is.Equal(lines[0].Reference, "REF-9")
}

// TestFindTransferLines_OrdersByIdAndScopesToTransfer.
func TestFindTransferLines_OrdersByIdAndScopesToTransfer(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	dest := mkWarehouse(t, f, "Dest")

	firstUUID, _ := mkStockedTransfer(t, f, dest, 3)
	secondUUID, _ := mkStockedTransfer(t, f, dest, 7)

	firstID := int64(scalarInt(t, s.db, `SELECT id FROM inventory_transfers WHERE uuid = $1`, firstUUID))
	secondID := int64(scalarInt(t, s.db, `SELECT id FROM inventory_transfers WHERE uuid = $1`, secondUUID))

	lines, err := s.findTransferLines(f.ctx, f.company.ID, firstID)
	is.NoErr(err)
	is.Equal(len(lines), 1)
	is.EqualFloat(lines[0].Qty, 3)

	lines, err = s.findTransferLines(f.ctx, f.company.ID, secondID)
	is.NoErr(err)
	is.Equal(len(lines), 1)
	is.EqualFloat(lines[0].Qty, 7)

	// Another tenant sees nothing, even naming the right transfer id.
	other := mkAccountCompany(t, s)
	lines, err = s.findTransferLines(other.ctx, other.company.ID, firstID)
	is.NoErr(err)
	is.Equal(len(lines), 0)
}
