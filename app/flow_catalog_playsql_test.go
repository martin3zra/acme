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

// TestFindTaxReceiptsForSetup_CatalogJoin: findTaxReceiptsForSetup lists the full
// shared_tax_receipts catalog left-joined against this company's own tax_receipts,
// so the settings taxSequences tab can show every DGII comprobante type with a
// checkbox — configured ones report their real range, unconfigured ones report
// zero. findTaxesReceipts, by contrast, only ever lists what the company has
// actually configured; it has no notion of the catalog at all.
func TestFindTaxReceiptsForSetup_CatalogJoin(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	// shared_tax_receipts.id has no identity/default — DGII comprobante codes get
	// fixed, manually-assigned ids so tax_receipts.shared_tax_receipt_id references
	// stay stable across environments, unlike shared_taxes/shared_warehouses/etc.
	// Sentinel ids well clear of any real seeded catalog range.
	const configuredCatalogID, unconfiguredCatalogID = 900001, 900002
	_, err := s.db.Exec(
		`INSERT INTO shared_tax_receipts (id, name, serie, type) VALUES ($1, 'Fiscal', 'B01', 'fiscal')`,
		configuredCatalogID,
	)
	is.NoErr(err)
	_, err = s.db.Exec(
		`INSERT INTO shared_tax_receipts (id, name, serie, type) VALUES ($1, 'Consumo', 'B02', 'fiscal')`,
		unconfiguredCatalogID,
	)
	is.NoErr(err)

	// Link the fixture's existing tax_receipts row back to the "configured" catalog
	// entry — this is what upsert_tax_receipts does when a user activates a type.
	_, err = s.db.Exec(
		`UPDATE tax_receipts SET shared_tax_receipt_id = $1 WHERE id = $2`,
		configuredCatalogID, f.taxReceiptID,
	)
	is.NoErr(err)

	forSetup, err := s.findTaxReceiptsForSetup(f.ctx)
	is.NoErr(err)
	// >= 2, not ==: the real seed migration also populates shared_tax_receipts, so
	// this asserts our two sentinel rows are present among the whole catalog rather
	// than assuming the catalog is otherwise empty.
	is.True(len(forSetup) >= 2, "catalog includes at least our two sentinel rows")

	var configuredRow, unconfiguredRow *taxReceipt
	for _, r := range forSetup {
		if r.ID == configuredCatalogID {
			configuredRow = r
		}
		if r.ID == unconfiguredCatalogID {
			unconfiguredRow = r
		}
	}
	is.True(configuredRow != nil, "configured catalog entry present")
	is.True(unconfiguredRow != nil, "unconfigured catalog entry present")

	is.Equal(configuredRow.SequenceStart, 1)
	is.Equal(configuredRow.SequenceEnd, 1000)
	is.Equal(unconfiguredRow.SequenceStart, 0)
	is.Equal(unconfiguredRow.SequenceEnd, 0)

	listed, err := s.findTaxesReceipts(f.ctx)
	is.NoErr(err)
	is.Equal(len(listed), 1) // only what the company actually has, catalog or not
	is.Equal(listed[0].ID, f.taxReceiptID)

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

// TestUpsertTaxReceipts_ReactivatingUpdatesInPlace: upsert_tax_receipts' conflict
// target used to be (company_id, id) — dead code, since id is a fresh serial value
// on every INSERT and can never collide. Saving the taxSequences form twice for the
// same catalog entry (e.g. editing an already-activated type's range) inserted a
// second tax_receipts row instead of updating the first. The real key is
// (company_id, shared_tax_receipt_id), enforced by a partial unique index since
// shared_tax_receipt_id is nullable for rows created outside this flow.
func TestUpsertTaxReceipts_ReactivatingUpdatesInPlace(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	const catalogID = 900101
	_, err := s.db.Exec(
		`INSERT INTO shared_tax_receipts (id, name, serie, type) VALUES ($1, 'Fiscal', 'B01', 'fiscal')`,
		catalogID,
	)
	is.NoErr(err)

	is.NoErr(s.upsertTaxReceipts(f.ctx, &TaxReceiptsForm{
		Receipts: []TaxReceiptSequenceForm{{ID: catalogID, Start: 1, End: 500}},
	}))
	is.NoErr(s.upsertTaxReceipts(f.ctx, &TaxReceiptsForm{
		Receipts: []TaxReceiptSequenceForm{{ID: catalogID, Start: 1, End: 600}},
	}))

	var count, sequenceEnd int
	is.NoErr(s.db.QueryRow(
		`SELECT count(*) FROM tax_receipts WHERE company_id = $1 AND shared_tax_receipt_id = $2`,
		f.company.ID, catalogID,
	).Scan(&count))
	is.Equal(count, 1) // updated in place, not duplicated

	is.NoErr(s.db.QueryRow(
		`SELECT sequence_end FROM tax_receipts WHERE company_id = $1 AND shared_tax_receipt_id = $2`,
		f.company.ID, catalogID,
	).Scan(&sequenceEnd))
	is.Equal(sequenceEnd, 600) // the second save's range won
}
