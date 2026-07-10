package app

import (
	"errors"
	"testing"
)

// The catalog cluster — warehouses, taxes, units and tax receipts — converted to
// playsql. All four files had reads and writes that never had direct coverage.

// ─── warehouses ──────────────────────────────────────────────────────────────

// TestWarehouseCRUD: store, read, update, toggle. The list is ordered by name and
// excludes soft-deleted rows via warehouseRead's softdelete tag.
func TestWarehouseCRUD(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	// The factory already created "General".
	is.NoErr(s.storeWarehouse(f.ctx, &StoreWarehouseForm{Name: "Boutique", Location: "Piantini"}))
	is.NoErr(s.storeWarehouse(f.ctx, &StoreWarehouseForm{Name: "Almacen", Location: ""}))

	all, err := s.findWarehouses(f.ctx)
	is.NoErr(err)
	is.Equal(len(all), 3)
	is.Equal(all[0].Name, "Almacen") // ORDER BY name
	is.Equal(all[1].Name, "Boutique")
	is.Equal(all[2].Name, "General")

	// An empty location stores NULL and reads back as "" — the old COALESCE.
	is.Equal(all[0].Location, "")
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM warehouses WHERE name = 'Almacen' AND location IS NULL`), 1)
	is.Equal(all[1].Location, "Piantini")

	boutique := all[1]
	one, err := s.findWarehouseByID(f.ctx, boutique.ID)
	is.NoErr(err)
	is.Equal(one.Name, "Boutique")
	is.True(one.UUID != "", "uuid is DB-generated")
	is.Equal(string(one.Status), "enabled")

	is.NoErr(s.updateWarehouse(f.ctx, boutique.ID, &UpdateWarehouseForm{Name: "Boutique 2", Location: ""}))
	one, err = s.findWarehouseByID(f.ctx, boutique.ID)
	is.NoErr(err)
	is.Equal(one.Name, "Boutique 2")
	is.Equal(one.Location, "") // cleared to NULL
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM warehouses WHERE id = $1 AND location IS NULL`, boutique.ID), 1)

	is.NoErr(s.toggleWarehouseStatus(f.ctx, one))
	one, err = s.findWarehouseByID(f.ctx, boutique.ID)
	is.NoErr(err)
	is.Equal(string(one.Status), "disabled")
}

// TestDeleteWarehouse_SecondDeleteIsNotFound pins a deliberate narrowing: the raw
// statement had no deleted_at predicate, so deleting twice reported success both
// times. The softdelete tag makes the second a not-found.
func TestDeleteWarehouse_SecondDeleteIsNotFound(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	id := mkWarehouse(t, f, "Temp")

	is.NoErr(s.deleteWarehouse(f.ctx, id))
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM warehouses WHERE id = $1 AND deleted_at IS NOT NULL`, id), 1)

	err := s.deleteWarehouse(f.ctx, id)
	is.True(errors.Is(err, ErrRecordNotFound), "deleting an already-deleted warehouse is not-found")

	// And it is gone from both reads.
	_, err = s.findWarehouseByID(f.ctx, id)
	is.Err(err, "a soft-deleted warehouse is not findable")

	all, err := s.findWarehouses(f.ctx)
	is.NoErr(err)
	for _, w := range all {
		is.True(w.ID != id, "a soft-deleted warehouse must not be listed")
	}
}

// ─── taxes ───────────────────────────────────────────────────────────────────

// TestTaxCRUD: findTaxes deliberately does not filter deleted_at, matching the raw
// query. updateTax is keyed by uuid.
func TestTaxCRUD(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	is.NoErr(s.storeTax(f.ctx, &StoreTaxForm{Name: "ITBIS Reducido", Rate: 8}))

	taxes, err := s.findTaxes(f.ctx)
	is.NoErr(err)
	is.Equal(len(taxes), 2) // the factory's ITBIS plus this one

	var target *tax
	for _, tx := range taxes {
		if tx.Name == "ITBIS Reducido" {
			target = tx
		}
	}
	is.True(target != nil, "the new tax is listed")
	is.EqualFloat(target.Rate, 8)
	is.True(target.UUID != "", "uuid is DB-generated")

	is.NoErr(s.updateTax(f.ctx, target.UUID, &StoreTaxForm{Name: "ITBIS 16", Rate: 16}))
	is.Equal(scalarString(t, s.db, `SELECT name FROM taxes WHERE uuid = $1`, target.UUID), "ITBIS 16")
	is.EqualFloat(scalarFloat(t, s.db, `SELECT rate FROM taxes WHERE uuid = $1`, target.UUID), 16)

	// The read does not hide soft-deleted taxes, as it never did.
	_, err = s.db.Exec(`UPDATE taxes SET deleted_at = now() WHERE uuid = $1`, target.UUID)
	is.NoErr(err)
	taxes, err = s.findTaxes(f.ctx)
	is.NoErr(err)
	is.Equal(len(taxes), 2)
}

// ─── units ───────────────────────────────────────────────────────────────────

func TestUnitCRUD(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	is.NoErr(s.storeUnit(f.ctx, &StoreUnitForm{Name: "Caja", BaseQty: 12}))

	units, err := s.findUnits(f.ctx)
	is.NoErr(err)
	is.Equal(len(units), 2) // the factory's base unit plus this one

	var caja *unit
	for _, u := range units {
		if u.Name == "Caja" {
			caja = u
		}
	}
	is.True(caja != nil, "the new unit is listed")
	is.Equal(caja.BaseQty, 12)

	is.NoErr(s.updateUnit(f.ctx, int(caja.ID), &StoreUnitForm{Name: "Caja x24", BaseQty: 24}))
	is.Equal(scalarString(t, s.db, `SELECT name FROM units WHERE id = $1`, caja.ID), "Caja x24")
	is.Equal(scalarInt(t, s.db, `SELECT base_qty FROM units WHERE id = $1`, caja.ID), 24)

	// The base unit is still the one with base_qty = 1.
	baseID, err := s.findUnitByBasedQty(f.company.ID)
	is.NoErr(err)
	is.Equal(baseID, f.unitID)
}

// ─── tax receipts ────────────────────────────────────────────────────────────

// TestFindTaxReceipts_DuplicateEntryPoints: findTaxReceiptsForSetup was a duplicate of
// findTaxesReceipts once the dead COALESCEs are removed. Both return the same rows.
//
// The sequence columns are NOT NULL, so the old COALESCE(..., 0) never fired.
func TestFindTaxReceipts_DuplicateEntryPoints(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	forSetup, err := s.findTaxReceiptsForSetup(f.ctx)
	is.NoErr(err)
	listed, err := s.findTaxesReceipts(f.ctx)
	is.NoErr(err)

	is.Equal(len(forSetup), 1)
	is.Equal(len(listed), len(forSetup))
	is.Equal(forSetup[0].ID, listed[0].ID)
	is.Equal(forSetup[0].Serie, listed[0].Serie)
	is.Equal(forSetup[0].SequenceStart, listed[0].SequenceStart)
	is.Equal(forSetup[0].SequenceEnd, listed[0].SequenceEnd)
	is.Equal(forSetup[0].Current, listed[0].Current)

	is.Equal(forSetup[0].ID, f.taxReceiptID)
	is.Equal(forSetup[0].SequenceStart, 1)
	is.Equal(forSetup[0].SequenceEnd, 1000)

	// Neither read filters deleted_at, so a retired receipt still shows up. That is
	// existing behaviour; grabTaxReceiptSequence is the one that excludes it.
	_, err = s.db.Exec(`UPDATE tax_receipts SET deleted_at = now() WHERE id = $1`, f.taxReceiptID)
	is.NoErr(err)

	listed, err = s.findTaxesReceipts(f.ctx)
	is.NoErr(err)
	is.Equal(len(listed), 1)

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()
	_, err = s.grabTaxReceiptSequence(tx, f.company.ID, f.taxReceiptID)
	is.True(errors.Is(err, ErrTaxReceiptNotFound), "but a retired receipt issues no numbers")
}
