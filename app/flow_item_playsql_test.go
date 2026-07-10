package app

import (
	"errors"
	"testing"
)

// item-repository converted to playsql. The two by-id/list reads dropped the
// LEFT JOIN LATERAL that resolved the item's unit; the two search reads stay raw.

func unitNameOf(t *testing.T, f *fixture, unitID int) string {
	t.Helper()
	return scalarString(t, f.s.db, `SELECT name FROM units WHERE id = $1`, unitID)
}

// ─── reads ────────────────────────────────────────────────────────────────────

// TestFindItemByID_LoadsTaxAndUnit: the tax arrives through a constrained belongsTo
// and the unit through hasOne -> belongsTo, replacing an INNER JOIN and a lateral.
func TestFindItemByID_LoadsTaxAndUnit(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, _ := mkItem(t, f, 250, 100)

	i, err := s.findItemByID(f.ctx, itemID)
	is.NoErr(err)
	is.Equal(i.ID, itemID)
	is.EqualFloat(i.Price, 250)
	is.Equal(i.ItemType, "product")
	is.True(i.UUID != "", "uuid is DB-generated")
	is.Equal(string(i.Status), "enabled")

	is.Equal(i.Tax.ID, int64(f.taxID))
	is.Equal(i.Tax.Name, scalarString(t, s.db, `SELECT name FROM taxes WHERE id = $1`, f.taxID))
	is.EqualFloat(i.Tax.Rate, scalarFloat(t, s.db, `SELECT rate FROM taxes WHERE id = $1`, f.taxID))

	is.True(i.Unit.ID != nil, "the item's unit is eager-loaded")
	is.Equal(*i.Unit.ID, f.unitID)
	is.Equal(*i.Unit.Name, unitNameOf(t, f, f.unitID))
}

// TestFindItemByID_TaxProjectionIsNarrow pins withItemTax. The old query selected
// only i.tax_id, t.name and t.rate, so the item responses have never carried the
// tax's uuid or timestamps. Loading the whole tax row would start leaking them.
func TestFindItemByID_TaxProjectionIsNarrow(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, _ := mkItem(t, f, 100, 60)

	i, err := s.findItemByID(f.ctx, itemID)
	is.NoErr(err)
	is.Equal(i.Tax.UUID, "")
	is.True(i.Tax.CreatedAt == nil, "the tax's timestamps are not part of an item response")
	is.True(i.Tax.UpdatedAt == nil, "the tax's timestamps are not part of an item response")
}

// TestFindItem_WithoutUnit: the lateral was a LEFT join, so an item with no
// items_units row still came back, with a nil unit. hasOne keeps that.
func TestFindItem_WithoutUnit(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, _ := mkItem(t, f, 100, 60)
	_, err := s.db.Exec(`DELETE FROM items_units WHERE item_id = $1`, itemID)
	is.NoErr(err)

	i, err := s.findItemByID(f.ctx, itemID)
	is.NoErr(err)
	is.Equal(i.ID, itemID)
	is.True(i.Unit.ID == nil, "an item with no unit row has a nil unit id")
	is.True(i.Unit.Name == nil, "an item with no unit row has a nil unit name")

	// And it is still listed, not dropped.
	all, err := s.findItems(f.ctx, ItemTypeAll)
	is.NoErr(err)
	is.Equal(len(all), 1)
	is.True(all[0].Unit.ID == nil, "the list read left-joins the unit too")
}

// TestFindItems_FiltersByTypeAndOrdersByName: `($2 = 'all' OR i.item_type = $2)` is a
// filter that disables itself for the sentinel, now written as Unless.
func TestFindItems_FiltersByTypeAndOrdersByName(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	mkNamed := func(name, itemType string) int {
		t.Helper()
		is.NoErr(s.storeItem(f.ctx, &StoreItemForm{
			Name: name, Price: 10, Description: "d",
			TaxID: f.taxID, UnitID: f.unitID, ItemType: itemType,
		}))
		return scalarInt(t, s.db,
			`SELECT id FROM items WHERE company_id = $1 AND name = $2`, f.company.ID, name)
	}

	mkNamed("Zeta", "product")
	mkNamed("Alpha", "service")
	mkNamed("Mid", "product")

	all, err := s.findItems(f.ctx, ItemTypeAll)
	is.NoErr(err)
	is.Equal(len(all), 3)
	is.Equal(all[0].Name, "Alpha") // ORDER BY name
	is.Equal(all[1].Name, "Mid")
	is.Equal(all[2].Name, "Zeta")

	products, err := s.findItems(f.ctx, ItemTypeProduct)
	is.NoErr(err)
	is.Equal(len(products), 2)
	is.Equal(products[0].Name, "Mid")
	is.Equal(products[1].Name, "Zeta")

	services, err := s.findItems(f.ctx, ItemTypeService)
	is.NoErr(err)
	is.Equal(len(services), 1)
	is.Equal(services[0].Name, "Alpha")

	// Every listed item carries its tax and unit.
	for _, i := range all {
		is.Equal(i.Tax.ID, int64(f.taxID))
		is.True(i.Unit.ID != nil, "each listed item resolves its unit")
		is.Equal(*i.Unit.ID, f.unitID)
	}
}

// TestFindItems_ScopedToCompanyAndSkipsDeleted.
func TestFindItems_ScopedToCompanyAndSkipsDeleted(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	first := mkAccountCompany(t, s)
	second := mkAccountCompany(t, s)

	firstItem, _ := mkItem(t, first, 100, 60)
	mkItem(t, second, 200, 120)

	rows, err := s.findItems(first.ctx, ItemTypeAll)
	is.NoErr(err)
	is.Equal(len(rows), 1)
	is.Equal(rows[0].ID, firstItem)

	// Another tenant's item is not readable by id either.
	_, err = s.findItemByID(second.ctx, firstItem)
	is.Err(err, "another tenant's item is not findable")

	// Nor deletable: without the company_id predicate this would soft-delete it.
	err = s.deleteItem(second.ctx, firstItem)
	is.True(errors.Is(err, ErrRecordNotFound), "another tenant's item is not deletable")
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM items WHERE id = $1 AND deleted_at IS NULL`, firstItem), 1)

	is.NoErr(s.deleteItem(first.ctx, firstItem))
	rows, err = s.findItems(first.ctx, ItemTypeAll)
	is.NoErr(err)
	is.Equal(len(rows), 0)

	_, err = s.findItemByID(first.ctx, firstItem)
	is.Err(err, "a soft-deleted item is not findable")
}

// TestFindItem_UnmarshalsIdentifiers: identifiers is jsonb, scanned raw and decoded
// in toItem because ItemIdentifiers is a struct and playsql reads a struct field as
// a relation.
func TestFindItem_UnmarshalsIdentifiers(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	ref, barcode := "REF-1", "BAR-1"
	is.NoErr(s.storeItem(f.ctx, &StoreItemForm{
		Name: "Tagged", Price: 10, Description: "d",
		TaxID: f.taxID, UnitID: f.unitID, ItemType: "product",
		Identifiers: ItemIdentifiers{Reference: &ref, Barcode: &barcode},
	}))
	id := scalarInt(t, s.db, `SELECT id FROM items WHERE company_id = $1 AND name = 'Tagged'`, f.company.ID)

	i, err := s.findItemByID(f.ctx, id)
	is.NoErr(err)
	is.True(i.Identifiers.Reference != nil, "reference round-trips through jsonb")
	is.Equal(*i.Identifiers.Reference, ref)
	is.True(i.Identifiers.Barcode != nil, "barcode round-trips through jsonb")
	is.Equal(*i.Identifiers.Barcode, barcode)
	is.True(i.Identifiers.SKU == nil, "an absent identifier stays nil")
}

// ─── attachItemUnit ───────────────────────────────────────────────────────────

// TestAttachItemUnit_UpsertsInPlace pins the bug the migration fixed. The old
// statement conflicted on `id`, which the insert never supplied, so every call
// appended a row and the lateral's LIMIT 1 kept returning the first one. Changing an
// item's unit looked like it worked and did nothing.
func TestAttachItemUnit_UpsertsInPlace(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, _ := mkItem(t, f, 100, 60)
	caja := mkUnit(t, f, "Caja", 12)

	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_units WHERE item_id = $1`, itemID), 1)
	before := scalarString(t, s.db,
		`SELECT updated_at::text FROM items_units WHERE item_id = $1`, itemID)

	is.NoErr(s.updateItem(f.ctx, itemID, &UpdateItemForm{
		Name: "Renamed", Price: 150, Description: "d",
		TaxID: f.taxID, UnitID: caja, ItemType: "product",
	}))

	// One link row, and it points at the new unit.
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_units WHERE item_id = $1`, itemID), 1)
	is.Equal(scalarInt(t, s.db, `SELECT unit_id FROM items_units WHERE item_id = $1`, itemID), caja)

	// The read agrees.
	i, err := s.findItemByID(f.ctx, itemID)
	is.NoErr(err)
	is.Equal(i.Name, "Renamed")
	is.EqualFloat(i.Price, 150)
	is.True(i.Unit.ID != nil, "the item resolves a unit")
	is.Equal(*i.Unit.ID, caja)
	is.Equal(*i.Unit.Name, "Caja")

	// The DO UPDATE bumps updated_at, as `updated_at = now()` did. It is compared
	// against its own previous value, not against created_at: created_at is written by
	// the DB default (CURRENT_TIMESTAMP, UTC) while a mapped timestamp is stamped with
	// Go's time.Now() (local). On a machine that is not on UTC those differ by the
	// offset, so `updated_at > created_at` can hold while updated_at never moved.
	after := scalarString(t, s.db,
		`SELECT updated_at::text FROM items_units WHERE item_id = $1`, itemID)
	is.True(after > before, "the upsert advanced updated_at: "+before+" -> "+after)
}

// TestStoreItemWithVariants_SharesTheItemInsert: storeItemInternal and
// insertVariantItem ran identical INSERTs; they share insertItemRow now. The variant
// path still creates no default variant.
func TestStoreItemWithVariants_SharesTheItemInsert(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, variantIDs := mkVariantItem(t, f, 500)
	is.Equal(len(variantIDs), 2)

	// The item row and its unit link were written by the shared insert.
	i, err := s.findItemByID(f.ctx, itemID)
	is.NoErr(err)
	is.EqualFloat(i.Price, 500)
	is.Equal(i.Tax.ID, int64(f.taxID))
	is.True(i.Unit.ID != nil, "the variant path attaches a unit too")
	is.Equal(*i.Unit.ID, f.unitID)

	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_units WHERE item_id = $1`, itemID), 1)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM items_variants WHERE item_id = $1 AND is_default = true`, itemID), 0)

	// The plain path does create one.
	plainID, _ := mkItem(t, f, 100, 60)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM items_variants WHERE item_id = $1 AND is_default = true`, plainID), 1)
}

// ─── writes ───────────────────────────────────────────────────────────────────

// TestItemWrites_SoftDeleteNarrowing pins a deliberate narrowing: the raw UPDATEs had
// no deleted_at predicate, so a soft-deleted item could still be edited, toggled and
// re-deleted, each reporting success. itemRead's softdelete tag makes all three
// not-found.
func TestItemWrites_SoftDeleteNarrowing(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, _ := mkItem(t, f, 100, 60)

	live, err := s.findItemByID(f.ctx, itemID)
	is.NoErr(err)
	is.NoErr(s.toggleItemStatus(f.ctx, live))
	is.Equal(scalarString(t, s.db, `SELECT status FROM items WHERE id = $1`, itemID), "disabled")

	is.NoErr(s.deleteItem(f.ctx, itemID))
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM items WHERE id = $1 AND deleted_at IS NOT NULL AND updated_at >= deleted_at`,
		itemID), 1)

	err = s.updateItem(f.ctx, itemID, &UpdateItemForm{
		Name: "Zombie", Price: 1, Description: "d",
		TaxID: f.taxID, UnitID: f.unitID, ItemType: "product",
	})
	is.True(errors.Is(err, ErrRecordNotFound), "updating a deleted item is not-found")
	is.True(scalarString(t, s.db, `SELECT name FROM items WHERE id = $1`, itemID) != "Zombie",
		"and the update did not land")

	live.Status = "disabled"
	err = s.toggleItemStatus(f.ctx, live)
	is.True(errors.Is(err, ErrRecordNotFound), "toggling a deleted item is not-found")

	err = s.deleteItem(f.ctx, itemID)
	is.True(errors.Is(err, ErrRecordNotFound), "deleting an already-deleted item is not-found")
}

// TestUpdateItem_RollsBackTheUnitOnFailure: updateItem runs the item UPDATE and the
// unit upsert in one transaction, so a not-found item must not leave a rewritten
// items_units row behind.
func TestUpdateItem_RollsBackTheUnitOnFailure(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	first := mkAccountCompany(t, s)
	second := mkAccountCompany(t, s)

	itemID, _ := mkItem(t, first, 100, 60)
	otherUnit := mkUnit(t, second, "Caja", 12)

	// Another tenant's id: the UPDATE matches nothing, so the whole transaction fails.
	err := s.updateItem(second.ctx, itemID, &UpdateItemForm{
		Name: "Stolen", Price: 1, Description: "d",
		TaxID: second.taxID, UnitID: otherUnit, ItemType: "product",
	})
	is.True(errors.Is(err, ErrRecordNotFound), "another tenant's item is not updatable")

	is.True(scalarString(t, s.db, `SELECT name FROM items WHERE id = $1`, itemID) != "Stolen",
		"the item row is untouched")
	is.Equal(scalarInt(t, s.db, `SELECT unit_id FROM items_units WHERE item_id = $1`, itemID), first.unitID)
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_units WHERE item_id = $1`, itemID), 1)
}
