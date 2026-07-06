package app

import "testing"

// TestFlowVoidReversesStock: voiding a sale returns the stock (sale_return
// movement) and restores the balance, and marks the invoice void.
func TestFlowVoidReversesStock(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	g := newFaker(t)
	itemID, variantID := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Build()
	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithLine(itemID, 2, 100, 18).Build()

	bal := func() float64 {
		return scalarFloat(t, s.db,
			`SELECT quantity FROM inventory_balances WHERE company_id = $1 AND variant_id = $2 AND warehouse_id = $3`,
			f.company.ID, variantID, f.warehouseID)
	}
	is.EqualFloat(bal(), -2)

	is.NoErr(f.s.voidInvoice(f.ctx, TransactionKinds.Invoice, uuid))

	is.Equal(scalarString(t, s.db, `SELECT status FROM invoices WHERE uuid = $1`, uuid), "void")
	is.EqualFloat(bal(), 0) // reversed back to zero
	assertRow(t, s.db, "inventory_movements", map[string]any{
		"variant_id": variantID, "transaction_kind": "sale_return",
	})
}

// TestFlowEstimateDoesNotMoveStock: estimates never touch inventory.
func TestFlowEstimateDoesNotMoveStock(t *testing.T) {
	s := newTestServer(t)
	f := mkAccountCompany(t, s)

	g := newFaker(t)
	itemID, variantID := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Build()
	newInvoice(t, f, g).ForCustomer(custID).Kind(TransactionKinds.Estimate).
		WithLine(itemID, 3, 100, 18).Build()

	assertNoRow(t, s.db, "inventory_movements", map[string]any{"variant_id": variantID})
	assertNoRow(t, s.db, "inventory_balances", map[string]any{"variant_id": variantID})
}
