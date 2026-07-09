package app

import "testing"

// Invoice header reads converted from hand-concatenated SQL to playsql. The flow
// suite hits findInvoicesByUUID heavily but never the list read, the jsonb columns
// or the two joins the conversion dropped.

// TestFindInvoices_ListLoadsCustomerAndJSON: the INNER JOIN on customers becomes a
// belongsTo eager load, and discount/payment/source decode from jsonb.
func TestFindInvoices_ListLoadsCustomerAndJSON(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, custUUID := newCustomer(t, f, g).Build()
	invUUID := newInvoice(t, f, g).ForCustomer(custID).Cash().WithLine(itemID, 1, 100, 18).Build()

	_, err := s.db.Exec(
		`UPDATE invoices SET discount = '{"value":10,"type":"percentage"}'::jsonb WHERE uuid = $1`, invUUID)
	is.NoErr(err)

	invoices, err := s.findInvoices(f.ctx, TransactionKinds.Invoice, InvoiceTypeAll)
	is.NoErr(err)
	is.Equal(len(invoices), 1)

	i := invoices[0]
	is.Equal(i.UUID, invUUID)
	is.Equal(i.Customer.UUID, custUUID)
	is.Equal(i.Customer.ID, custID)
	is.True(i.Customer.Name != "", "customer name should be eager-loaded")
	is.EqualFloat(i.Discount.Val, 10)
	is.Equal(i.Discount.Type, "percentage")
	is.EqualFloat(i.Payment.Cash.Amount, 118) // a cash sale records its payment blob
	is.True(i.Source == nil, "a standalone invoice has no source")

	// The list projection never carried company_id, and still does not.
	is.Equal(i.CompanyID, 0)
	// Terms is only derived by the detail reads.
	is.Equal(i.Terms, "")
}

// TestFindInvoices_FiltersByKindAndType covers the old
// `($2 != 'invoice' OR $3 = 'all' OR invoices.type = $3)` predicate, now expressed
// in Go: the type filter applies only to invoices, and only when narrowed.
func TestFindInvoices_FiltersByKindAndType(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Credit("net30").Build()

	cashUUID := newInvoice(t, f, g).ForCustomer(custID).Cash().WithLine(itemID, 1, 100, 18).Build()
	creditUUID := newInvoice(t, f, g).ForCustomer(custID).Credit("net30").WithLine(itemID, 1, 100, 18).Build()
	estUUID := newInvoice(t, f, g).ForCustomer(custID).Kind(TransactionKinds.Estimate).
		WithLine(itemID, 1, 100, 18).Build()

	all, err := s.findInvoices(f.ctx, TransactionKinds.Invoice, InvoiceTypeAll)
	is.NoErr(err)
	is.Equal(len(all), 2)             // both invoices, not the estimate
	is.Equal(all[0].UUID, creditUUID) // ORDER BY id DESC

	cash, err := s.findInvoices(f.ctx, TransactionKinds.Invoice, InvoiceTypeCash)
	is.NoErr(err)
	is.Equal(len(cash), 1)
	is.Equal(cash[0].UUID, cashUUID)

	credit, err := s.findInvoices(f.ctx, TransactionKinds.Invoice, InvoiceTypeCredit)
	is.NoErr(err)
	is.Equal(len(credit), 1)
	is.Equal(credit[0].UUID, creditUUID)

	// A non-invoice kind ignores the type narrowing entirely.
	estimates, err := s.findInvoices(f.ctx, TransactionKinds.Estimate, InvoiceTypeCash)
	is.NoErr(err)
	is.Equal(len(estimates), 1)
	is.Equal(estimates[0].UUID, estUUID)
}

// TestFindInvoicesByUUID_DerivesTerms: the detail reads label the invoice from the
// gap between date and due_on.
func TestFindInvoicesByUUID_DerivesTerms(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Credit("net30").Build()

	creditUUID := newInvoice(t, f, g).ForCustomer(custID).Credit("net30").WithLine(itemID, 1, 100, 18).Build()
	cashUUID := newInvoice(t, f, g).ForCustomer(custID).Cash().WithLine(itemID, 1, 100, 18).Build()

	credit, err := s.findInvoicesByUUID(f.ctx, TransactionKinds.Invoice, f.company.ID, creditUUID)
	is.NoErr(err)
	is.Equal(credit.Terms, "net30")
	is.Equal(credit.CompanyID, f.company.ID) // the detail projection does carry it
	is.True(credit.NCF != nil, "a posted invoice gets a tax number")

	cash, err := s.findInvoicesByUUID(f.ctx, TransactionKinds.Invoice, f.company.ID, cashUUID)
	is.NoErr(err)
	is.Equal(cash.Terms, "pia") // no due date

	// The kind filter still applies.
	_, err = s.findInvoicesByUUID(f.ctx, TransactionKinds.Estimate, f.company.ID, cashUUID)
	is.Err(err, "an invoice must not be findable as an estimate")
}

// TestFindInvoicesByID_DocumentWithoutTaxReceipt: the dropped LEFT JOIN on
// tax_receipts contributed no columns and could not filter, so a document with a
// null tax_receipt_id is still findable. findInvoicesByID also applies no kind
// filter, which the recurrence generator depends on.
//
// An estimate is the document with a null tax_receipt_id: only kind=invoice
// consumes a receipt sequence, and a template still stores the receipt id it will
// stamp on each generated invoice.
func TestFindInvoicesByID_DocumentWithoutTaxReceipt(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Build()
	estUUID := newInvoice(t, f, g).ForCustomer(custID).Kind(TransactionKinds.Estimate).
		WithLine(itemID, 1, 100, 18).Build()

	id := scalarInt(t, s.db, `SELECT id FROM invoices WHERE uuid = $1`, estUUID)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM invoices WHERE id = $1 AND tax_receipt_id IS NULL`, id), 1)

	est, err := s.findInvoicesByID(f.company.ID, id)
	is.NoErr(err)
	is.Equal(est.UUID, estUUID)
	is.Equal(string(est.Kind), "estimate") // found despite no kind filter
	is.True(est.NCF == nil, "an estimate has no tax number")
	is.True(est.TaxReceiptID == nil, "and no tax receipt")
	is.True(est.Customer.UUID != "", "customer should still be eager-loaded")
}

// TestFindInvoices_SoftDeletedCustomer: customerRead is softdelete-tagged, but the
// old INNER JOIN never filtered deleted_at, so the invoice must still render.
func TestFindInvoices_SoftDeletedCustomer(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, custUUID := newCustomer(t, f, g).Build()
	invUUID := newInvoice(t, f, g).ForCustomer(custID).Cash().WithLine(itemID, 1, 100, 18).Build()

	is.NoErr(s.deleteCustomer(f.ctx, custID))

	invoices, err := s.findInvoices(f.ctx, TransactionKinds.Invoice, InvoiceTypeAll)
	is.NoErr(err)
	is.Equal(len(invoices), 1)
	is.Equal(invoices[0].Customer.UUID, custUUID)

	detail, err := s.findInvoicesByUUID(f.ctx, TransactionKinds.Invoice, f.company.ID, invUUID)
	is.NoErr(err)
	is.Equal(detail.Customer.UUID, custUUID)
}
