package app

import (
	"testing"
)

// findInvoiceLines and reconcileInvoiceStock's line read converted to playsql. The
// five joins the line read carried are now relations; updateInvoiceBalance, the
// GROUP BY/HAVING aggregate and the balance increment stay raw.

// ─── findInvoiceLines ─────────────────────────────────────────────────────────

// TestFindInvoiceLines_LoadsItemVariantTaxAndUnit covers the whole projection: the
// item, its tax, its unit (through the old lateral) and the variant.
func TestFindInvoiceLines_LoadsItemVariantTaxAndUnit(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	custID, _, invUUID := mkCreditInvoice(t, f, 100, 3)
	_ = custID

	invoiceID := scalarInt(t, s.db, `SELECT id FROM invoices WHERE uuid = $1`, invUUID)

	lines, err := s.findInvoiceLines(f.ctx, f.company.ID, invoiceID)
	is.NoErr(err)
	is.Equal(len(lines), 1)

	l := lines[0]

	// line.ID is the *item* id, not the invoices_items id — that is what the old
	// projection selected into it, and updateInvoiceLines round-trips it as such.
	itemID := scalarInt(t, s.db,
		`SELECT item_id FROM invoices_items WHERE invoice_id = $1`, invoiceID)
	is.Equal(l.ID, int64(itemID))
	is.True(l.ID != int64(scalarInt(t, s.db,
		`SELECT id FROM invoices_items WHERE invoice_id = $1`, invoiceID)),
		"line.ID is the item id, deliberately not the line's own id")

	is.Equal(l.Qty, int64(3))
	is.EqualFloat(l.Price, 100)
	is.Equal(l.Action, UNCHANGED)

	// The item's own columns.
	is.Equal(l.Name, scalarString(t, s.db, `SELECT name FROM items WHERE id = $1`, itemID))
	is.Equal(l.Description, "test item")

	// The tax: id and name off the tax row, rate and amount frozen on the line.
	is.Equal(l.Tax.ID, int64(f.taxID))
	is.Equal(l.Tax.Name, scalarString(t, s.db, `SELECT name FROM taxes WHERE id = $1`, f.taxID))
	is.EqualFloat(l.Tax.Rate, scalarFloat(t, s.db,
		`SELECT rate FROM invoices_items WHERE invoice_id = $1`, invoiceID))
	is.EqualFloat(l.Tax.Amount, scalarFloat(t, s.db,
		`SELECT tax FROM invoices_items WHERE invoice_id = $1`, invoiceID))

	// The unit came through the lateral, now a hasOne on the item.
	is.Equal(l.Unit.ID, int64(f.unitID))
	is.Equal(l.Unit.Name, scalarString(t, s.db, `SELECT name FROM units WHERE id = $1`, f.unitID))

	// The variant.
	is.Equal(l.VariantID, int64(scalarInt(t, s.db,
		`SELECT variant_id FROM invoices_items WHERE invoice_id = $1`, invoiceID)))
	is.True(l.VariantName != "", "the variant's name is loaded")
}

// TestFindInvoiceLines_FreezesTheBilledRate: a line's rate and tax amount live on
// invoices_items, so moving the tax rate afterwards must not restate old invoices.
//
// This does not pin withInvoiceLineTax's projection width. toLine copies only the
// tax's id and name by hand, so widening the SELECT would fetch more columns and
// change nothing — the constraint is a projection narrowing, not a guard.
func TestFindInvoiceLines_FreezesTheBilledRate(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, _, invUUID := mkCreditInvoice(t, f, 100, 1)
	invoiceID := scalarInt(t, s.db, `SELECT id FROM invoices WHERE uuid = $1`, invUUID)

	// Move the tax rate after billing. The line keeps the rate it was billed at.
	billedRate := scalarFloat(t, s.db,
		`SELECT rate FROM invoices_items WHERE invoice_id = $1`, invoiceID)
	_, err := s.db.Exec(`UPDATE taxes SET rate = 99 WHERE id = $1`, f.taxID)
	is.NoErr(err)

	lines, err := s.findInvoiceLines(f.ctx, f.company.ID, invoiceID)
	is.NoErr(err)
	is.Equal(len(lines), 1)

	is.EqualFloat(lines[0].Tax.Rate, billedRate)
	is.True(lines[0].Tax.Rate != 99, "the line's rate is frozen, not read off the tax row")
}

// TestFindInvoiceLines_NamesNonDefaultVariants pins the CASE the projection carried:
// a default variant shows the item's name, a named variant shows "Item — Variant".
func TestFindInvoiceLines_NamesNonDefaultVariants(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	// A variant item: two non-default variants.
	varItemID, variantIDs := mkVariantItem(t, f, 500)
	itemName := scalarString(t, s.db, `SELECT name FROM items WHERE id = $1`, varItemID)
	variantName := scalarString(t, s.db, `SELECT name FROM items_variants WHERE id = $1`, variantIDs[0])

	// And a plain item, which has a default variant.
	plainID, _ := mkItem(t, f, 100, 60)
	plainName := scalarString(t, s.db, `SELECT name FROM items WHERE id = $1`, plainID)

	custID, _ := newCustomer(t, f, g).Credit("net30").Build()
	invUUID := newInvoice(t, f, g).ForCustomer(custID).Credit("net30").
		WithVariantLine(varItemID, variantIDs[0], 1, 500, 18).
		WithLine(plainID, 2, 100, 18).
		Build()
	invoiceID := scalarInt(t, s.db, `SELECT id FROM invoices WHERE uuid = $1`, invUUID)

	lines, err := s.findInvoiceLines(f.ctx, f.company.ID, invoiceID)
	is.NoErr(err)
	is.Equal(len(lines), 2)

	byItem := map[int64]*line{}
	for _, l := range lines {
		byItem[l.ID] = l
	}

	is.Equal(byItem[int64(varItemID)].Name, itemName+" — "+variantName)
	is.Equal(byItem[int64(plainID)].Name, plainName)
}

// TestFindInvoiceLines_KeepsLinesOfDeletedItems pins why the item side uses
// lineItemRead rather than itemRead.
//
// The old `INNER JOIN items` had no deleted_at predicate. itemRead carries a
// softdelete tag and playsql excludes soft-deleted rows from an eager load, so
// reusing it would blank out the name and description of a line whose item was later
// deleted — a historical invoice would render empty rows.
func TestFindInvoiceLines_KeepsLinesOfDeletedItems(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, _, invUUID := mkCreditInvoice(t, f, 100, 1)
	invoiceID := scalarInt(t, s.db, `SELECT id FROM invoices WHERE uuid = $1`, invUUID)
	itemID := scalarInt(t, s.db, `SELECT item_id FROM invoices_items WHERE invoice_id = $1`, invoiceID)
	itemName := scalarString(t, s.db, `SELECT name FROM items WHERE id = $1`, itemID)

	is.NoErr(s.deleteItem(f.ctx, itemID))

	lines, err := s.findInvoiceLines(f.ctx, f.company.ID, invoiceID)
	is.NoErr(err)
	is.Equal(len(lines), 1)
	is.Equal(lines[0].Name, itemName)
	is.Equal(lines[0].Description, "test item")
	is.Equal(lines[0].Tax.ID, int64(f.taxID))
	is.Equal(lines[0].Unit.ID, int64(f.unitID))
}

// TestFindInvoiceLines_LoadsUnitAndTaxTogether pins the single-traversal eager load.
//
// Written as two root paths — With("Item.ItemUnit.Unit") plus
// WithConstraint("Item.Tax", ...) — the second reloads Item and replaces the structs
// the first had filled in, so ItemUnit goes back to nil and Unit.ID reads 0. That is
// not a cosmetic loss: the recurring-invoice flow writes this straight into
// invoices_items.unit_id, where it trips a foreign key.
func TestFindInvoiceLines_LoadsUnitAndTaxTogether(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, _, invUUID := mkCreditInvoice(t, f, 100, 1)
	invoiceID := scalarInt(t, s.db, `SELECT id FROM invoices WHERE uuid = $1`, invUUID)

	lines, err := s.findInvoiceLines(f.ctx, f.company.ID, invoiceID)
	is.NoErr(err)
	is.Equal(len(lines), 1)

	// Both branches of the item's eager load survive the same query.
	is.True(lines[0].Unit.ID != 0, "the unit must not be zeroed by the tax load")
	is.Equal(lines[0].Unit.ID, int64(f.unitID))
	is.True(lines[0].Tax.ID != 0, "the tax must not be zeroed by the unit load")
	is.Equal(lines[0].Tax.ID, int64(f.taxID))
	is.True(lines[0].Unit.Name != "", "the unit's name comes from the nested belongsTo")
}

// TestFindInvoiceLines_ScopedToInvoiceAndCompany.
func TestFindInvoiceLines_ScopedToInvoiceAndCompany(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	other := mkAccountCompany(t, s)

	_, _, firstUUID := mkCreditInvoice(t, f, 100, 1)
	_, _, secondUUID := mkCreditInvoice(t, f, 200, 2)
	firstID := scalarInt(t, s.db, `SELECT id FROM invoices WHERE uuid = $1`, firstUUID)
	secondID := scalarInt(t, s.db, `SELECT id FROM invoices WHERE uuid = $1`, secondUUID)

	lines, err := s.findInvoiceLines(f.ctx, f.company.ID, firstID)
	is.NoErr(err)
	is.Equal(len(lines), 1)
	is.EqualFloat(lines[0].Price, 100)

	lines, err = s.findInvoiceLines(f.ctx, f.company.ID, secondID)
	is.NoErr(err)
	is.Equal(len(lines), 1)
	is.EqualFloat(lines[0].Price, 200)

	// Another tenant sees nothing, even naming the right invoice id.
	lines, err = s.findInvoiceLines(other.ctx, other.company.ID, firstID)
	is.NoErr(err)
	is.Equal(len(lines), 0)
}

// TestFindInvoiceLines_UnmarshalsItemIdentifiers: identifiers is jsonb on the item.
func TestFindInvoiceLines_UnmarshalsItemIdentifiers(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	ref := "REF-INV"
	is.NoErr(s.storeItem(f.ctx, &StoreItemForm{
		Name: "Tagged", Price: 50, Description: "d",
		TaxID: f.taxID, UnitID: f.unitID, ItemType: "product",
		Identifiers: ItemIdentifiers{Reference: &ref},
	}))
	itemID := scalarInt(t, s.db, `SELECT id FROM items WHERE company_id = $1 AND name = 'Tagged'`, f.company.ID)

	custID, _ := newCustomer(t, f, g).Credit("net30").Build()
	invUUID := newInvoice(t, f, g).ForCustomer(custID).Credit("net30").
		WithLine(itemID, 1, 50, 18).Build()
	invoiceID := scalarInt(t, s.db, `SELECT id FROM invoices WHERE uuid = $1`, invUUID)

	lines, err := s.findInvoiceLines(f.ctx, f.company.ID, invoiceID)
	is.NoErr(err)
	is.Equal(len(lines), 1)
	is.True(lines[0].Identifier.Reference != nil, "the item's identifiers round-trip")
	is.Equal(*lines[0].Identifier.Reference, ref)
}
