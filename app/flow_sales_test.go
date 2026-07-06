package app

import "testing"

// TestFlowCashInvoice: a cash (pia) invoice closes immediately, creates no
// receivable, leaves the customer balance untouched, and ships stock.
func TestFlowCashInvoice(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, variantID := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Build() // cash (pia) by default

	// Line price is explicit: the totals below are asserted exactly, so it must
	// not come from faker. The builder still handles customer + document wiring.
	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithLine(itemID, 2, 100, 18).Build()

	var status, paid, termType string
	var total, amountDue float64
	var movementRecorded bool
	row := s.db.QueryRow(
		`SELECT status, paid_status, type, total, amount_due, movement_recorded FROM invoices WHERE uuid = $1`, uuid)
	is.NoErr(row.Scan(&status, &paid, &termType, &total, &amountDue, &movementRecorded))

	is.Equal(status, "closed")
	is.Equal(paid, "paid")
	is.Equal(termType, "cash")
	is.EqualFloat(total, 236) // 2*100 + 18% tax(36)
	is.EqualFloat(amountDue, 0)
	is.True(movementRecorded, "cash invoice should record stock movement")

	// No receivable for a cash sale; customer balance unchanged.
	assertNoRow(t, s.db, "receivables", map[string]any{"customer_id": custID})
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), 0)

	// Stock left the warehouse.
	assertRow(t, s.db, "inventory_movements", map[string]any{
		"variant_id": variantID, "warehouse_id": f.warehouseID, "transaction_kind": "sale",
	})
	qty := scalarFloat(t, s.db,
		`SELECT quantity FROM inventory_balances WHERE company_id = $1 AND variant_id = $2 AND warehouse_id = $3`,
		f.company.ID, variantID, f.warehouseID)
	is.EqualFloat(qty, -2)
}

// TestFlowCreditInvoice: a net30 invoice stays open, registers a receivable,
// raises the customer balance, sets a due date, and still ships stock.
func TestFlowCreditInvoice(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, variantID := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Credit("net30").Build()

	uuid := newInvoice(t, f, g).ForCustomer(custID).Credit("net30").
		WithLine(itemID, 1, 100, 18).Build()

	var status, paid, termType string
	var total, amountDue float64
	var dueOn *string
	row := s.db.QueryRow(
		`SELECT status, paid_status, type, total, amount_due, due_on FROM invoices WHERE uuid = $1`, uuid)
	is.NoErr(row.Scan(&status, &paid, &termType, &total, &amountDue, &dueOn))

	is.Equal(status, "sent")
	is.Equal(paid, "unpaid")
	is.Equal(termType, "credit")
	is.EqualFloat(total, 118) // 100 + 18% tax
	is.EqualFloat(amountDue, 118)
	is.True(dueOn != nil, "credit invoice must have a due date")

	assertRow(t, s.db, "receivables", map[string]any{"customer_id": custID, "company_id": f.company.ID})
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), 118)

	// Credit sale still moves stock (status sent).
	qty := scalarFloat(t, s.db,
		`SELECT quantity FROM inventory_balances WHERE company_id = $1 AND variant_id = $2 AND warehouse_id = $3`,
		f.company.ID, variantID, f.warehouseID)
	is.EqualFloat(qty, -1)
}
