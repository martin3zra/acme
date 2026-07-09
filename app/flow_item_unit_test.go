package app

import "testing"

// An item has exactly one unit. attachItemUnit used to conflict on `id`, the serial
// primary key the insert never supplies, so its ON CONFLICT branch could never fire
// and no unique index existed to target. Every call appended a row, and because both
// item reads pick the unit through `LEFT JOIN LATERAL (... LIMIT 1)` with no
// ORDER BY, editing an item's unit silently kept the old one.

func mkUnit(t *testing.T, f *fixture, name string, baseQty int) int {
	t.Helper()
	var id int
	if err := f.s.db.QueryRow(
		`INSERT INTO units (company_id, name, base_qty) VALUES ($1, $2, $3) RETURNING id`,
		f.company.ID, name, baseQty,
	).Scan(&id); err != nil {
		t.Fatalf("insert unit %q: %v", name, err)
	}
	return id
}

// TestUpdateItem_SwitchesUnit is the regression test: the edit must actually change
// the unit, and must not leave a second link row behind.
func TestUpdateItem_SwitchesUnit(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	box := mkUnit(t, f, "Box", 12)
	itemID, _ := mkItem(t, f, 100, 60)

	before, err := s.findItemByID(f.ctx, itemID)
	is.NoErr(err)
	is.Equal(*before.Unit.ID, f.unitID)
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_units WHERE item_id = $1`, itemID), 1)

	is.NoErr(s.updateItem(f.ctx, itemID, &UpdateItemForm{
		Name: "Renamed", Description: "d", Price: 100, TaxID: f.taxID,
		ItemType: "product", UnitID: box,
	}))

	after, err := s.findItemByID(f.ctx, itemID)
	is.NoErr(err)
	is.Equal(*after.Unit.ID, box)
	is.Equal(*after.Unit.Name, "Box")
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_units WHERE item_id = $1`, itemID), 1)
	is.Equal(after.Name, "Renamed")
}

// TestAttachItemUnit_Idempotent: re-attaching the same unit neither duplicates the
// row nor errors on the new unique constraint.
func TestAttachItemUnit_Idempotent(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, _ := mkItem(t, f, 100, 60)

	for i := 0; i < 3; i++ {
		is.NoErr(s.updateItem(f.ctx, itemID, &UpdateItemForm{
			Name: "Stable", Description: "d", Price: 100, TaxID: f.taxID,
			ItemType: "product", UnitID: f.unitID,
		}))
	}

	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_units WHERE item_id = $1`, itemID), 1)
	item, err := s.findItemByID(f.ctx, itemID)
	is.NoErr(err)
	is.Equal(*item.Unit.ID, f.unitID)
}

// TestItemsUnits_OnePerItem: the constraint is scoped per company+item, so two items
// may each hold their own link row.
func TestItemsUnits_OnePerItem(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	first, _ := mkItem(t, f, 100, 60)
	second, _ := mkItem(t, f, 200, 120)

	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_units WHERE item_id = $1`, first), 1)
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_units WHERE item_id = $1`, second), 1)

	// Switching one item's unit leaves the other alone.
	box := mkUnit(t, f, "Box", 12)
	is.NoErr(s.updateItem(f.ctx, first, &UpdateItemForm{
		Name: "First", Description: "d", Price: 100, TaxID: f.taxID,
		ItemType: "product", UnitID: box,
	}))

	one, err := s.findItemByID(f.ctx, first)
	is.NoErr(err)
	is.Equal(*one.Unit.ID, box)

	two, err := s.findItemByID(f.ctx, second)
	is.NoErr(err)
	is.Equal(*two.Unit.ID, f.unitID)
}
