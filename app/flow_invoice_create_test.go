package app

import (
	"testing"
	"time"
)

// TestFlowCreateInvoicePersistsColumns verifies the playsql-backed invoice INSERT
// round-trips id + the DB-generated uuid (read back via RawScalar) and writes the
// tricky columns exactly: due date, tax-receipt pointers, and JSON payment/
// discount.
func TestFlowCreateInvoicePersistsColumns(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Credit("net30").Build()
	uuid := newInvoice(t, f, g).ForCustomer(custID).Credit("net30").
		WithLine(itemID, 1, 100, 18).Build()

	// The returned uuid is non-empty and matches the stored row.
	is.True(uuid != "", "storeInvoice should return a uuid")
	var id int
	var storedUUID string
	is.NoErr(s.db.QueryRow(
		`SELECT id, uuid FROM invoices WHERE company_id = $1 AND uuid = $2`, f.company.ID, uuid,
	).Scan(&id, &storedUUID))
	is.Equal(storedUUID, uuid)

	// Credit invoice: due_on set, tax receipt stamped (NCF), discount JSON present.
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM invoices WHERE id = $1 AND due_on IS NOT NULL`, id), 1)
	is.Equal(scalarInt(t, s.db, `SELECT COALESCE(tax_receipt_id,0) FROM invoices WHERE id = $1`, id), f.taxReceiptID)
	is.True(scalarString(t, s.db, `SELECT COALESCE(tax_number,'') FROM invoices WHERE id = $1`, id) != "",
		"credit invoice must carry a tax (NCF) number")
	is.True(scalarString(t, s.db, `SELECT COALESCE(discount::text,'') FROM invoices WHERE id = $1`, id) != "",
		"discount JSON should be persisted")
}

// TestFlowCreateCashInvoiceNoDueOn: a cash invoice leaves due_on NULL through the
// playsql insert (the pointer column is written as NULL).
func TestFlowCreateCashInvoiceNoDueOn(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Build()
	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithLine(itemID, 1, 100, 18).Build()

	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM invoices WHERE uuid = $1 AND due_on IS NULL`, uuid), 1)
}

// TestFlowCreateTemplatePersistsRecurrence: a recurring template round-trips its
// recurrence JSON and tax_receipt_id through the playsql insert (both are *[]byte
// / pointer columns that must survive encoding).
func TestFlowCreateTemplatePersistsRecurrence(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Credit("net30").Build()

	now := time.Now()
	form := &StoreInvoiceForm{
		CustomerID: custID, Date: now, Terms: "net30", TaxReceipt: f.taxReceiptID,
		Discount: Discount{Type: "percentage"}, Kind: TransactionKinds.Template,
		Recurrence: &Recurrence{
			Enabled: true, Name: "Monthly", Type: "schedule",
			Frequency: "monthly", Interval: 1, Timezone: "America/Santo_Domingo",
			DayOfMonth: now.Day(), StartDate: &now,
		},
		Lines: []*Line{mkLine(itemID, f.unitID, f.warehouseID, 1, 100, 18)},
	}
	form.Compute()
	uuid, err := f.s.storeInvoice(f.ctx, form)
	is.NoErr(err)

	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM invoices WHERE uuid = $1 AND recurrence IS NOT NULL`, uuid), 1)
	is.Equal(scalarInt(t, s.db,
		`SELECT COALESCE(tax_receipt_id,0) FROM invoices WHERE uuid = $1`, uuid), f.taxReceiptID)
}
