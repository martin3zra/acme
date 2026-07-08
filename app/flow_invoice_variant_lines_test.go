package app

import (
	"strings"
	"testing"
	"time"
)

// storeInvoiceForVariantTest assembles a minimal cash-invoice form and runs it
// through the real storeInvoice path, returning the error (unlike the builder,
// which fails the test on error) so the guard paths can be asserted.
func storeInvoiceForVariantTest(t *testing.T, f *fixture, custID int, lines []*Line) (string, error) {
	t.Helper()
	form := &StoreInvoiceForm{
		CustomerID: custID,
		Date:       time.Now(),
		Terms:      "pia",
		TaxReceipt: f.taxReceiptID,
		Discount:   Discount{Type: "percentage"},
		Lines:      lines,
		Kind:       TransactionKinds.Invoice,
	}
	form.Compute()
	form.Payment.Cash.Amount = form.total
	return f.s.storeInvoice(f.ctx, form)
}

// TestFlowInvoiceVariantLinePersists: a line naming an explicit variant persists
// that exact variant on invoices_items — not a guessed one.
func TestFlowInvoiceVariantLinePersists(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, variantIDs := mkVariantItem(t, f, 100)
	blue := variantIDs[1]
	custID, _ := newCustomer(t, f, g).Build()

	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithVariantLine(itemID, blue, 2, 100, 18).Build()

	var invID int
	is.NoErr(s.db.QueryRow(`SELECT id FROM invoices WHERE uuid = $1`, uuid).Scan(&invID))
	assertRow(t, s.db, "invoices_items", map[string]any{"invoice_id": invID, "item_id": itemID, "variant_id": blue})
}

// TestFlowInvoiceVariantMovementHitsChosenVariant: the stock OUT movement debits
// the exact variant sold — not another variant of the same item.
func TestFlowInvoiceVariantMovementHitsChosenVariant(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, variantIDs := mkVariantItem(t, f, 100)
	red, blue := variantIDs[0], variantIDs[1]
	custID, _ := newCustomer(t, f, g).Build()

	newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithVariantLine(itemID, blue, 2, 100, 18).Build()

	// Movement + balance land on blue …
	assertRow(t, s.db, "inventory_movements", map[string]any{
		"variant_id": blue, "warehouse_id": f.warehouseID, "transaction_kind": "sale",
	})
	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT quantity FROM inventory_balances WHERE company_id = $1 AND variant_id = $2 AND warehouse_id = $3`,
		f.company.ID, blue, f.warehouseID), -2)
	// … and never on red.
	assertNoRow(t, s.db, "inventory_movements", map[string]any{"variant_id": red})
}

// TestFlowInvoiceLinesCarryVariant: the read path (findInvoiceLines, backing
// Show/Edit) returns the line's variant id and name so an edit round-trips it.
func TestFlowInvoiceLinesCarryVariant(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, variantIDs := mkVariantItem(t, f, 100)
	blue := variantIDs[1]
	custID, _ := newCustomer(t, f, g).Build()

	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithVariantLine(itemID, blue, 2, 100, 18).Build()

	var invID int
	is.NoErr(s.db.QueryRow(`SELECT id FROM invoices WHERE uuid = $1`, uuid).Scan(&invID))

	var blueName string
	is.NoErr(s.db.QueryRow(`SELECT name FROM items_variants WHERE id = $1`, blue).Scan(&blueName))

	lines, err := f.s.findInvoiceLines(f.ctx, f.company.ID, invID)
	is.NoErr(err)
	is.Equal(len(lines), 1)
	is.Equal(int(lines[0].VariantID), blue)
	is.Equal(lines[0].VariantName, blueName)
	// Display name composes item + variant so Show/Edit read like the search rows.
	if !strings.Contains(lines[0].Name, blueName) || !strings.Contains(lines[0].Name, "—") {
		t.Fatalf("expected composed line name, got %q", lines[0].Name)
	}
}

// TestFlowInvoiceVariantRequired: a has_variants item on a line with no variant
// is rejected — no silent fallback across its variants.
func TestFlowInvoiceVariantRequired(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkVariantItem(t, f, 100)
	custID, _ := newCustomer(t, f, g).Build()

	line := mkLine(itemID, f.unitID, f.warehouseID, 1, 100, 18) // VariantID left 0
	_, err := storeInvoiceForVariantTest(t, f, custID, []*Line{line})
	if err == nil {
		t.Fatal("expected error: variant item invoiced without a variant")
	}
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM invoices WHERE company_id = $1`, f.company.ID), 0)
}

// TestFlowInvoiceNonVariantDefaultsToDefaultVariant: a plain item resolves to its
// default variant, unchanged behaviour for the common case.
func TestFlowInvoiceNonVariantDefaultsToDefaultVariant(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, defaultVariant := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Build()

	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithLine(itemID, 1, 100, 18).Build()

	var invID int
	is.NoErr(s.db.QueryRow(`SELECT id FROM invoices WHERE uuid = $1`, uuid).Scan(&invID))
	assertRow(t, s.db, "invoices_items", map[string]any{"invoice_id": invID, "item_id": itemID, "variant_id": defaultVariant})
}

// TestFlowInvoiceCrossItemVariantRejected: a line may not reference a variant that
// belongs to a different item.
func TestFlowInvoiceCrossItemVariantRejected(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	variantItem, variantIDs := mkVariantItem(t, f, 100)
	plainItem, _ := mkItem(t, f, 50, 30)
	custID, _ := newCustomer(t, f, g).Build()

	// plainItem line, but pointing at the variant item's variant.
	line := mkLine(plainItem, f.unitID, f.warehouseID, 1, 50, 18)
	line.VariantID = variantIDs[0]
	_, err := storeInvoiceForVariantTest(t, f, custID, []*Line{line})
	if err == nil {
		t.Fatalf("expected error: variant %d does not belong to item %d", variantIDs[0], plainItem)
	}
	_ = variantItem
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM invoices WHERE company_id = $1`, f.company.ID), 0)
}
