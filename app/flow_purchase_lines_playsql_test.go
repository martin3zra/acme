package app

import (
	"testing"
	"time"
)

// purchase-repository line reads and writes converted to playsql: findPurchaseLines,
// the variant-owner and item-tax lookups, the three line writes, the AP re-sync, the
// PO status update and recordPurchaseMovements' line read.

func purchaseIDOf(t *testing.T, f *fixture, uuid string) int {
	t.Helper()
	return scalarInt(t, f.s.db, `SELECT id FROM purchases WHERE uuid = $1`, uuid)
}

// updateFormFor rebuilds an UpdatePurchaseForm from a purchase's current lines.
func updateFormFor(t *testing.T, f *fixture, vendorID int, kind PurchaseTransactionKind,
	invoiceNumber string, lines []*Line) *UpdatePurchaseForm {
	t.Helper()
	form := &UpdatePurchaseForm{StorePurchaseForm{
		VendorID:      vendorID,
		Date:          time.Now(),
		Terms:         "net30",
		Discount:      Discount{Type: "percentage"},
		Lines:         lines,
		Kind:          kind,
		InvoiceNumber: invoiceNumber,
	}}
	form.SetContext(f.ctx)
	form.Compute()
	return form
}

// ─── findPurchaseLines ────────────────────────────────────────────────────────

// TestFindPurchaseLines_LoadsVariantItemTaxAndUnit covers the projection. The item is
// reached through the variant, since purchase_items has no item_id column.
func TestFindPurchaseLines_LoadsVariantItemTaxAndUnit(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	vendorID, _, _ := mkVendorBill(t, f, 250)
	_ = vendorID

	purchaseID := scalarInt(t, s.db,
		`SELECT id FROM purchases WHERE company_id = $1 ORDER BY id DESC LIMIT 1`, f.company.ID)

	lines, err := s.findPurchaseLines(f.ctx, f.company.ID, purchaseID)
	is.NoErr(err)
	is.Equal(len(lines), 1)

	l := lines[0]

	// line.ID is the *item* id, reached through the variant — purchase_items stores
	// only variant_id.
	variantID := scalarInt(t, s.db,
		`SELECT variant_id FROM purchase_items WHERE purchase_id = $1`, purchaseID)
	itemID := scalarInt(t, s.db, `SELECT item_id FROM items_variants WHERE id = $1`, variantID)
	is.Equal(l.ID, int64(itemID))
	is.Equal(l.VariantID, int64(variantID))

	is.Equal(l.Qty, int64(1))
	is.EqualFloat(l.Price, 250)
	is.Equal(l.Action, UNCHANGED)
	is.Equal(l.Name, scalarString(t, s.db, `SELECT name FROM items WHERE id = $1`, itemID))

	// amount is computed (qty * unit_price); total is the stored line_total.
	is.EqualFloat(l.Amount, 250)
	is.EqualFloat(l.Total, scalarFloat(t, s.db,
		`SELECT line_total FROM purchase_items WHERE purchase_id = $1`, purchaseID))

	// The unit falls back to the item's, since the line stored one explicitly.
	is.Equal(l.Unit.ID, int64(f.unitID))
	is.Equal(l.Unit.Name, scalarString(t, s.db, `SELECT name FROM units WHERE id = $1`, f.unitID))

	// The tax falls back to the item's, and its rate comes off the tax row.
	is.Equal(l.Tax.ID, int64(f.taxID))
	is.EqualFloat(l.Tax.Rate, scalarFloat(t, s.db, `SELECT rate FROM taxes WHERE id = $1`, f.taxID))
}

// TestFindPurchaseLines_RateFollowsTheTaxRow is the mirror of the invoice-line test.
//
// An invoice line freezes the rate it was billed at on invoices_items. A purchase line
// has no rate column: the old query selected taxes.rate, so changing the tax restates
// the line. Preserved, not "fixed" — the two documents genuinely differ.
func TestFindPurchaseLines_RateFollowsTheTaxRow(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	mkVendorBill(t, f, 100)
	purchaseID := scalarInt(t, s.db,
		`SELECT id FROM purchases WHERE company_id = $1 ORDER BY id DESC LIMIT 1`, f.company.ID)

	_, err := s.db.Exec(`UPDATE taxes SET rate = 42 WHERE id = $1`, f.taxID)
	is.NoErr(err)

	lines, err := s.findPurchaseLines(f.ctx, f.company.ID, purchaseID)
	is.NoErr(err)
	is.EqualFloat(lines[0].Tax.Rate, 42)

	// tax_amount, though, is the stored per-line value.
	is.EqualFloat(lines[0].Tax.Amount, scalarFloat(t, s.db,
		`SELECT tax_amount FROM purchase_items WHERE purchase_id = $1`, purchaseID))
}

// TestFindPurchaseLines_LineOverridesBeatItemDefaults pins the two COALESCEs.
//
// purchase_items.unit_id and purchase_items.tax_id are nullable overrides. When set
// they win; when NULL the line falls back to the item's unit and the item's tax.
func TestFindPurchaseLines_LineOverridesBeatItemDefaults(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	mkVendorBill(t, f, 100)
	purchaseID := scalarInt(t, s.db,
		`SELECT id FROM purchases WHERE company_id = $1 ORDER BY id DESC LIMIT 1`, f.company.ID)

	otherUnit := mkUnit(t, f, "Caja", 12)
	is.NoErr(s.storeTax(f.ctx, &StoreTaxForm{Name: "ITBIS 8", Rate: 8}))
	otherTax := scalarInt(t, s.db,
		`SELECT id FROM taxes WHERE company_id = $1 AND name = 'ITBIS 8'`, f.company.ID)

	// With overrides set, the line's own unit and tax win.
	_, err := s.db.Exec(
		`UPDATE purchase_items SET unit_id = $2, tax_id = $3 WHERE purchase_id = $1`,
		purchaseID, otherUnit, otherTax)
	is.NoErr(err)

	lines, err := s.findPurchaseLines(f.ctx, f.company.ID, purchaseID)
	is.NoErr(err)
	is.Equal(lines[0].Unit.ID, int64(otherUnit))
	is.Equal(lines[0].Unit.Name, "Caja")
	is.Equal(lines[0].Tax.ID, int64(otherTax))
	is.EqualFloat(lines[0].Tax.Rate, 8)

	// With both NULL, it falls back to the item's unit and the item's tax.
	_, err = s.db.Exec(
		`UPDATE purchase_items SET unit_id = NULL, tax_id = NULL WHERE purchase_id = $1`, purchaseID)
	is.NoErr(err)

	lines, err = s.findPurchaseLines(f.ctx, f.company.ID, purchaseID)
	is.NoErr(err)
	is.Equal(lines[0].Unit.ID, int64(f.unitID))
	is.True(lines[0].Unit.Name != "", "the item's unit name comes through the hasOne")
	is.Equal(lines[0].Tax.ID, int64(f.taxID))
}

// TestFindPurchaseLines_NamesNonDefaultVariants pins the CASE.
func TestFindPurchaseLines_NamesNonDefaultVariants(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	varItemID, variantIDs := mkVariantItem(t, f, 500)
	itemName := scalarString(t, s.db, `SELECT name FROM items WHERE id = $1`, varItemID)
	variantName := scalarString(t, s.db, `SELECT name FROM items_variants WHERE id = $1`, variantIDs[0])

	plainID, _ := mkItem(t, f, 100, 60)
	plainName := scalarString(t, s.db, `SELECT name FROM items WHERE id = $1`, plainID)

	vendorID, _ := newVendor(t, f, g).Build()
	uuid := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		WithVariantLine(varItemID, variantIDs[0], 1, 500, 0).
		WithLine(plainID, 2, 100, 0).
		Build()

	lines, err := s.findPurchaseLines(f.ctx, f.company.ID, purchaseIDOf(t, f, uuid))
	is.NoErr(err)
	is.Equal(len(lines), 2)

	byItem := map[int64]*line{}
	for _, l := range lines {
		byItem[l.ID] = l
	}
	is.Equal(byItem[int64(varItemID)].Name, itemName+" — "+variantName)
	is.Equal(byItem[int64(plainID)].Name, plainName)
}

// TestFindPurchaseLines_KeepsLinesOfDeletedItems: the item joins never filtered
// deleted_at, so lineItemRead (untagged) is used rather than itemRead.
func TestFindPurchaseLines_KeepsLinesOfDeletedItems(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	mkVendorBill(t, f, 100)
	purchaseID := scalarInt(t, s.db,
		`SELECT id FROM purchases WHERE company_id = $1 ORDER BY id DESC LIMIT 1`, f.company.ID)
	variantID := scalarInt(t, s.db,
		`SELECT variant_id FROM purchase_items WHERE purchase_id = $1`, purchaseID)
	itemID := scalarInt(t, s.db, `SELECT item_id FROM items_variants WHERE id = $1`, variantID)
	itemName := scalarString(t, s.db, `SELECT name FROM items WHERE id = $1`, itemID)

	is.NoErr(s.deleteItem(f.ctx, itemID))

	lines, err := s.findPurchaseLines(f.ctx, f.company.ID, purchaseID)
	is.NoErr(err)
	is.Equal(len(lines), 1)
	is.Equal(lines[0].Name, itemName)
	is.Equal(lines[0].Tax.ID, int64(f.taxID))
	is.Equal(lines[0].Unit.ID, int64(f.unitID))
}

// TestFindPurchaseLines_SkipsSoftDeletedLines: purchaseLineRead is softdelete-tagged,
// matching the old `pi.deleted_at IS NULL`.
func TestFindPurchaseLines_SkipsSoftDeletedLines(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	mkVendorBill(t, f, 100)
	purchaseID := scalarInt(t, s.db,
		`SELECT id FROM purchases WHERE company_id = $1 ORDER BY id DESC LIMIT 1`, f.company.ID)

	lines, err := s.findPurchaseLines(f.ctx, f.company.ID, purchaseID)
	is.NoErr(err)
	is.Equal(len(lines), 1)

	_, err = s.db.Exec(`UPDATE purchase_items SET deleted_at = now() WHERE purchase_id = $1`, purchaseID)
	is.NoErr(err)

	lines, err = s.findPurchaseLines(f.ctx, f.company.ID, purchaseID)
	is.NoErr(err)
	is.Equal(len(lines), 0)
}

// ─── line writes ──────────────────────────────────────────────────────────────

// TestProcessPurchaseLines_ReviveNeedsWithTrashed pins the trap in the UPDATED arm.
//
// Re-adding a previously removed line arrives as UPDATED with `deleted_at = NULL`.
// purchaseLineRead is softdelete-tagged, so without WithTrashed the default scope adds
// `deleted_at IS NULL` and the update matches zero rows: the line stays deleted and
// the edit silently loses it.
func TestProcessPurchaseLines_ReviveNeedsWithTrashed(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	vendorID, _ := newVendor(t, f, g).Build()
	invNo := uniq("PINV")
	uuid := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		InvoiceNumber(invNo).WithLine(itemID, 2, 100, 0).Build()
	purchaseID := purchaseIDOf(t, f, uuid)

	// Remove the line.
	del := mkLine(itemID, f.unitID, f.warehouseID, 2, 100, 0)
	del.Action = DELETED
	is.NoErr(s.updatePurchase(f.ctx, uuid,
		updateFormFor(t, f, vendorID, PurchaseTransactionKinds.VendorBill, invNo, []*Line{del})))

	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM purchase_items WHERE purchase_id = $1 AND deleted_at IS NOT NULL`,
		purchaseID), 1)

	lines, err := s.findPurchaseLines(f.ctx, f.company.ID, purchaseID)
	is.NoErr(err)
	is.Equal(len(lines), 0)

	// Re-add it: arrives as UPDATED, and must revive the soft-deleted row.
	revive := mkLine(itemID, f.unitID, f.warehouseID, 5, 120, 0)
	revive.Action = UPDATED
	is.NoErr(s.updatePurchase(f.ctx, uuid,
		updateFormFor(t, f, vendorID, PurchaseTransactionKinds.VendorBill, invNo, []*Line{revive})))

	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM purchase_items WHERE purchase_id = $1 AND deleted_at IS NULL`,
		purchaseID), 1)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM purchase_items WHERE purchase_id = $1`, purchaseID), 1)

	lines, err = s.findPurchaseLines(f.ctx, f.company.ID, purchaseID)
	is.NoErr(err)
	is.Equal(len(lines), 1)
	is.Equal(lines[0].Qty, int64(5))
	is.EqualFloat(lines[0].Price, 120)
}

// TestProcessPurchaseLines_DeleteStampsBothTimestamps: the DELETED arm uses Update, not
// Delete. Builder.Delete on a softdelete model stamps deleted_at alone, and the raw
// statement bumped updated_at too.
func TestProcessPurchaseLines_DeleteStampsBothTimestamps(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	vendorID, _ := newVendor(t, f, g).Build()
	invNo := uniq("PINV")
	uuid := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		InvoiceNumber(invNo).WithLine(itemID, 2, 100, 0).Build()
	purchaseID := purchaseIDOf(t, f, uuid)

	before := scalarString(t, s.db,
		`SELECT updated_at::text FROM purchase_items WHERE purchase_id = $1`, purchaseID)

	del := mkLine(itemID, f.unitID, f.warehouseID, 2, 100, 0)
	del.Action = DELETED
	is.NoErr(s.updatePurchase(f.ctx, uuid,
		updateFormFor(t, f, vendorID, PurchaseTransactionKinds.VendorBill, invNo, []*Line{del})))

	after := scalarString(t, s.db,
		`SELECT updated_at::text FROM purchase_items WHERE purchase_id = $1`, purchaseID)

	// Compared for change, not for ordering. The row was inserted with the column's
	// CURRENT_TIMESTAMP default (UTC) and rewritten with a playsql stamp (Go's
	// time.Now(), process-local), so on a host that is not on UTC the new value can sort
	// *before* the old one while still being the newer write.
	is.True(after != before, "the soft delete rewrote updated_at: "+before+" -> "+after)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM purchase_items WHERE purchase_id = $1 AND deleted_at IS NOT NULL`,
		purchaseID), 1)
}

// TestProcessPurchaseLines_AddsAndUpdates covers the ADDED and UPDATED arms.
func TestProcessPurchaseLines_AddsAndUpdates(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	first, _ := mkItem(t, f, 100, 60)
	second, _ := mkItem(t, f, 200, 120)
	vendorID, _ := newVendor(t, f, g).Build()
	invNo := uniq("PINV")
	uuid := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		InvoiceNumber(invNo).WithLine(first, 1, 100, 0).Build()
	purchaseID := purchaseIDOf(t, f, uuid)

	upd := mkLine(first, f.unitID, f.warehouseID, 7, 111, 0)
	upd.Action = UPDATED
	add := mkLine(second, f.unitID, f.warehouseID, 3, 222, 0)
	add.Action = ADDED

	is.NoErr(s.updatePurchase(f.ctx, uuid,
		updateFormFor(t, f, vendorID, PurchaseTransactionKinds.VendorBill, invNo, []*Line{upd, add})))

	lines, err := s.findPurchaseLines(f.ctx, f.company.ID, purchaseID)
	is.NoErr(err)
	is.Equal(len(lines), 2)

	byItem := map[int64]*line{}
	for _, l := range lines {
		byItem[l.ID] = l
	}
	is.Equal(byItem[int64(first)].Qty, int64(7))
	is.EqualFloat(byItem[int64(first)].Price, 111)
	is.Equal(byItem[int64(second)].Qty, int64(3))
	is.EqualFloat(byItem[int64(second)].Price, 222)

	// The added line carries the item's tax id, resolved by resolveItemTaxIDs.
	is.Equal(scalarInt(t, s.db,
		`SELECT tax_id FROM purchase_items WHERE purchase_id = $1 AND variant_id =
		   (SELECT id FROM items_variants WHERE item_id = $2 AND is_default)`,
		purchaseID, second), f.taxID)
}

// TestProcessPurchaseLines_PersistsDiscount covers the discount column, which the
// ADDED and UPDATED maps assign but purchaseLineRead does not map — an Update or Insert
// map's keys go to the statement rather than being filtered against the struct.
func TestProcessPurchaseLines_PersistsDiscount(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	vendorID, _ := newVendor(t, f, g).Build()
	invNo := uniq("PINV")
	uuid := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		InvoiceNumber(invNo).WithLine(itemID, 1, 100, 0).Build()
	purchaseID := purchaseIDOf(t, f, uuid)

	// A form-level percentage is what populates each line's discount (types.go:1168).
	upd := mkLine(itemID, f.unitID, f.warehouseID, 2, 100, 0)
	upd.Action = UPDATED

	form := &UpdatePurchaseForm{StorePurchaseForm{
		VendorID:      vendorID,
		Date:          time.Now(),
		Terms:         "net30",
		Discount:      Discount{Type: "percentage", Val: 10},
		Lines:         []*Line{upd},
		Kind:          PurchaseTransactionKinds.VendorBill,
		InvoiceNumber: invNo,
	}}
	form.SetContext(f.ctx)
	form.Compute()

	is.NoErr(s.updatePurchase(f.ctx, uuid, form))

	// 10% of a 200.00 line.
	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT discount FROM purchase_items WHERE purchase_id = $1`, purchaseID), 20)
}

// ─── resolveVariantsForLines ──────────────────────────────────────────────────

// TestResolveVariantsForLines_RejectsForeignVariant pins the variant-owner read: an
// explicitly named variant must belong to the line's item.
func TestResolveVariantsForLines_RejectsForeignVariant(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	mine, _ := mkItem(t, f, 100, 60)
	_, theirVariants := mkVariantItem(t, f, 500)

	vendorID, _ := newVendor(t, f, g).Build()

	// A line for `mine` naming a variant of the other item.
	bad := mkLine(mine, f.unitID, f.warehouseID, 1, 100, 0)
	bad.VariantID = theirVariants[0]

	form := &StorePurchaseForm{
		VendorID: vendorID, Date: time.Now(), Terms: "net30",
		Discount: Discount{Type: "percentage"},
		Lines:    []*Line{bad},
		Kind:     PurchaseTransactionKinds.VendorBill, InvoiceNumber: uniq("PINV"),
	}
	form.SetContext(f.ctx)
	form.Compute()

	_, err := s.storePurchase(f.ctx, form)
	is.Err(err, "a variant of another item must be rejected")
}

// TestResolveVariantsForLines_SkipsSoftDeletedVariants: the variant-owner read filters
// deleted_at explicitly, since itemVariantRead carries no softdelete tag. A
// soft-deleted variant must not be transactable.
func TestResolveVariantsForLines_SkipsSoftDeletedVariants(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, variantIDs := mkVariantItem(t, f, 500)
	_, err := s.db.Exec(`UPDATE items_variants SET deleted_at = now() WHERE id = $1`, variantIDs[0])
	is.NoErr(err)

	vendorID, _ := newVendor(t, f, g).Build()
	l := mkLine(itemID, f.unitID, f.warehouseID, 1, 500, 0)
	l.VariantID = variantIDs[0]

	form := &StorePurchaseForm{
		VendorID: vendorID, Date: time.Now(), Terms: "net30",
		Discount: Discount{Type: "percentage"},
		Lines:    []*Line{l},
		Kind:     PurchaseTransactionKinds.VendorBill, InvoiceNumber: uniq("PINV"),
	}
	form.SetContext(f.ctx)
	form.Compute()

	_, err = s.storePurchase(f.ctx, form)
	is.Err(err, "a soft-deleted variant must not be transactable")
}

// ─── AP re-sync ───────────────────────────────────────────────────────────────

// TestUpdatePurchase_ResyncsLinkedPayable: editing a vendor bill rewrites its
// accounts_payable row and bumps updated_at (accountsPayableModel does not map the
// column, so the update sets it explicitly).
func TestUpdatePurchase_ResyncsLinkedPayable(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	vendorID, _ := newVendor(t, f, g).Build()
	invNo := uniq("PINV")
	uuid := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		InvoiceNumber(invNo).WithLine(itemID, 1, 100, 0).Build()
	purchaseID := purchaseIDOf(t, f, uuid)

	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT amount_total FROM accounts_payable WHERE purchase_id = $1`, purchaseID), 100)
	beforeAP := scalarString(t, s.db,
		`SELECT updated_at::text FROM accounts_payable WHERE purchase_id = $1`, purchaseID)

	newInvNo := uniq("PINV2")
	upd := mkLine(itemID, f.unitID, f.warehouseID, 4, 100, 0)
	upd.Action = UPDATED
	is.NoErr(s.updatePurchase(f.ctx, uuid,
		updateFormFor(t, f, vendorID, PurchaseTransactionKinds.VendorBill, newInvNo, []*Line{upd})))

	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT amount_total FROM accounts_payable WHERE purchase_id = $1`, purchaseID), 400)
	is.Equal(scalarString(t, s.db,
		`SELECT invoice_number FROM accounts_payable WHERE purchase_id = $1`, purchaseID), newInvNo)

	// tax_amount lands even though accountsPayableModel does not map the column: an
	// Update map's keys are passed through to the statement, not filtered against the
	// struct. (A column that does not exist fails loudly at the database.)
	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT tax_amount FROM accounts_payable WHERE purchase_id = $1`, purchaseID),
		scalarFloat(t, s.db, `SELECT tax_amount FROM purchases WHERE id = $1`, purchaseID))

	// updated_at is deliberately not asserted. accounts_payable has a BEFORE UPDATE
	// trigger (trg_ap_updated_at) that stamps it with now(), which under the txdb
	// harness is the enclosing transaction's start time and never advances.
	_ = beforeAP

	// The vendor's payable balance moved by the delta, not to the new total.
	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT amount_payable FROM vendors WHERE id = $1`, vendorID), 400)
}

// ─── resolveItemTaxIDs ────────────────────────────────────────────────────────

// TestResolveItemTaxIDs_AlwaysResolves: items.tax_id is NOT NULL, so the nil arm the
// old scan carried was dead. Every requested item maps to a non-nil tax id.
func TestResolveItemTaxIDs_AlwaysResolves(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	first, _ := mkItem(t, f, 100, 60)
	second, _ := mkItem(t, f, 200, 120)

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()

	m, err := resolveItemTaxIDs(tx, f.company.ID, []int{first, second})
	is.NoErr(err)
	is.Equal(len(m), 2)
	is.True(m[first] != nil, "items.tax_id is NOT NULL")
	is.Equal(*m[first], f.taxID)
	is.Equal(*m[second], f.taxID)

	// Another tenant's item resolves to nothing.
	other := mkAccountCompany(t, s)
	m, err = resolveItemTaxIDs(tx, other.company.ID, []int{first})
	is.NoErr(err)
	is.Equal(len(m), 0)
}
