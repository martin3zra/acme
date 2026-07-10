package app

import (
	"errors"
	"testing"
)

// customer-repository and vendor-repository converted to playsql. The two by-uuid
// reads replaced a LEFT JOIN onto a filtered subquery with a constrained hasOne; the
// two list reads dropped four joins each; the CRUD writes picked up the softdelete
// predicate their raw statements never had.

// markInvoiceOpening retypes an invoice as an opening balance. No builder produces
// one, and the columns invoices requires are many; storeInvoice then a retype gets a
// realistic row without duplicating its insert.
func markInvoiceOpening(t *testing.T, f *fixture, invoiceUUID string) {
	t.Helper()
	if _, err := f.s.db.Exec(
		`UPDATE invoices SET type = 'opening'::invoice_terms WHERE uuid = $1`, invoiceUUID); err != nil {
		t.Fatalf("mark invoice opening: %v", err)
	}
}

// markPayableDraft retypes an accounts_payable row as a draft, which is what
// findVendorByUUID treats as the vendor's opening balance.
func markPayableDraft(t *testing.T, f *fixture, apUUID string) {
	t.Helper()
	if _, err := f.s.db.Exec(
		`UPDATE accounts_payable SET status = 'draft' WHERE uuid = $1`, apUUID); err != nil {
		t.Fatalf("mark payable draft: %v", err)
	}
}

// ─── customer by uuid ─────────────────────────────────────────────────────────

// TestFindCustomerByUUID_OpeningBalance covers all three states the old LEFT JOIN
// produced: no opening invoice, one opening invoice, and a non-opening invoice that
// must not be mistaken for one.
func TestFindCustomerByUUID_OpeningBalance(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	custID, custUUID := newCustomer(t, f, g).Credit("net30").Build()

	// No opening invoice: the LEFT JOIN produced NULLs, so OpenBalance is present
	// with nil fields. Callers dereference c.OpenBalance unconditionally.
	c, err := s.findCustomeByUUID(f.ctx, custUUID)
	is.NoErr(err)
	is.Equal(c.ID, custID)
	is.True(c.OpenBalance != nil, "OpenBalance is always non-nil")
	is.True(c.OpenBalance.InvoiceID == nil, "no opening invoice means a nil invoice id")
	is.True(c.OpenBalance.Date == nil, "no opening invoice means a nil date")
	is.True(c.OpenBalance.Amount == nil, "no opening invoice means a nil amount")

	// An ordinary credit invoice is not an opening balance.
	itemID, _ := mkItem(t, f, 100, 60)
	newInvoice(t, f, g).ForCustomer(custID).Credit("net30").WithLine(itemID, 1, 100, 18).Build()

	c, err = s.findCustomeByUUID(f.ctx, custUUID)
	is.NoErr(err)
	is.True(c.OpenBalance.InvoiceID == nil, "a credit invoice is not an opening balance")

	// Now one that is.
	openingUUID := newInvoice(t, f, g).ForCustomer(custID).Credit("net30").
		WithLine(itemID, 2, 250, 0).Build()
	markInvoiceOpening(t, f, openingUUID)

	openingID := scalarInt(t, s.db, `SELECT id FROM invoices WHERE uuid = $1`, openingUUID)
	openingAmount := scalarFloat(t, s.db, `SELECT amount FROM invoices WHERE uuid = $1`, openingUUID)

	c, err = s.findCustomeByUUID(f.ctx, custUUID)
	is.NoErr(err)
	is.True(c.OpenBalance.InvoiceID != nil, "the opening invoice is eager-loaded")
	is.Equal(*c.OpenBalance.InvoiceID, openingID)
	is.EqualFloat(*c.OpenBalance.Amount, openingAmount)
	is.True(c.OpenBalance.Date != nil, "the opening invoice carries its date")
}

// TestFindCustomerByUUID_OpeningBalanceIsNotAnotherCustomers pins the FK match.
//
// The eager loader groups the invoices it fetched by their customer_id *field*, so
// withOpeningInvoice must put customer_id in its projection. Drop it and every child
// keys on 0, matching no parent — the relation silently comes back nil. Two customers
// with opening invoices catch a mismatched key even if it were nonzero.
func TestFindCustomerByUUID_OpeningBalanceIsNotAnotherCustomers(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)

	firstID, firstUUID := newCustomer(t, f, g).Credit("net30").Build()
	secondID, secondUUID := newCustomer(t, f, g).Credit("net30").Build()

	firstOpening := newInvoice(t, f, g).ForCustomer(firstID).Credit("net30").
		WithLine(itemID, 1, 111, 0).Build()
	secondOpening := newInvoice(t, f, g).ForCustomer(secondID).Credit("net30").
		WithLine(itemID, 1, 222, 0).Build()
	markInvoiceOpening(t, f, firstOpening)
	markInvoiceOpening(t, f, secondOpening)

	firstInvoiceID := scalarInt(t, s.db, `SELECT id FROM invoices WHERE uuid = $1`, firstOpening)
	secondInvoiceID := scalarInt(t, s.db, `SELECT id FROM invoices WHERE uuid = $1`, secondOpening)

	a, err := s.findCustomeByUUID(f.ctx, firstUUID)
	is.NoErr(err)
	is.True(a.OpenBalance.InvoiceID != nil, "first customer's opening invoice is loaded")
	is.Equal(*a.OpenBalance.InvoiceID, firstInvoiceID)

	b, err := s.findCustomeByUUID(f.ctx, secondUUID)
	is.NoErr(err)
	is.True(b.OpenBalance.InvoiceID != nil, "second customer's opening invoice is loaded")
	is.Equal(*b.OpenBalance.InvoiceID, secondInvoiceID)
}

// TestFindCustomerByUUID_SkipsSoftDeleted: the read filtered deleted_at before, and
// customerModel's softdelete tag keeps doing it.
func TestFindCustomerByUUID_SkipsSoftDeleted(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	id, uuid := newCustomer(t, f, g).Build()
	is.NoErr(s.deleteCustomer(f.ctx, id))

	_, err := s.findCustomeByUUID(f.ctx, uuid)
	is.Err(err, "a soft-deleted customer is not findable")
}

// ─── customer writes ──────────────────────────────────────────────────────────

// TestCustomerWrites_SoftDeleteNarrowing pins a deliberate narrowing. The raw
// UPDATEs had no deleted_at predicate, so a soft-deleted customer could still be
// edited, toggled, and re-deleted, each reporting success. customerModel's softdelete
// tag makes all three not-found.
func TestCustomerWrites_SoftDeleteNarrowing(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	id, uuid := newCustomer(t, f, g).Named("Before").Build()

	// While live, all three writes land.
	is.NoErr(s.updateCustomer(f.ctx, id, &UpdateCustomerForm{
		Name: "After", CustomerType: "business", PaymentMethod: "cash", PaymentTerms: "net30",
		TaxReceipt: f.taxReceiptID,
	}))
	is.Equal(scalarString(t, s.db, `SELECT name FROM customers WHERE id = $1`, id), "After")

	live, err := s.findCustomeByUUID(f.ctx, uuid)
	is.NoErr(err)
	is.Equal(string(live.Status), "enabled")
	is.NoErr(s.toggleCustomerStatus(f.ctx, live))
	is.Equal(scalarString(t, s.db, `SELECT status FROM customers WHERE id = $1`, id), "disabled")

	is.NoErr(s.deleteCustomer(f.ctx, id))
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM customers WHERE id = $1 AND deleted_at IS NOT NULL`, id), 1)

	// deleteCustomer bumps updated_at as well as deleted_at, as the raw statement did.
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM customers WHERE id = $1 AND updated_at >= deleted_at`, id), 1)

	// Once deleted, none of the three touch it.
	err = s.updateCustomer(f.ctx, id, &UpdateCustomerForm{
		Name: "Zombie", CustomerType: "business", PaymentMethod: "cash", PaymentTerms: "net30",
		TaxReceipt: f.taxReceiptID,
	})
	is.True(errors.Is(err, ErrRecordNotFound), "updating a deleted customer is not-found")
	is.Equal(scalarString(t, s.db, `SELECT name FROM customers WHERE id = $1`, id), "After")

	live.Status = "disabled"
	err = s.toggleCustomerStatus(f.ctx, live)
	is.True(errors.Is(err, ErrRecordNotFound), "toggling a deleted customer is not-found")

	err = s.deleteCustomer(f.ctx, id)
	is.True(errors.Is(err, ErrRecordNotFound), "deleting an already-deleted customer is not-found")
}

// TestUpdateCustomer_ScopedToCompany: the company_id predicate is what stops one
// tenant editing another's customer. Without it the write would land silently.
func TestUpdateCustomer_ScopedToCompany(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	first := mkAccountCompany(t, s)
	second := mkAccountCompany(t, s)
	g := newFaker(t)

	id, _ := newCustomer(t, first, g).Named("Theirs").Build()

	err := s.updateCustomer(second.ctx, id, &UpdateCustomerForm{
		Name: "Stolen", CustomerType: "business", PaymentMethod: "cash", PaymentTerms: "net30",
		TaxReceipt: second.taxReceiptID,
	})
	is.True(errors.Is(err, ErrRecordNotFound), "another tenant's customer is not updatable")
	is.Equal(scalarString(t, s.db, `SELECT name FROM customers WHERE id = $1`, id), "Theirs")

	err = s.deleteCustomer(second.ctx, id)
	is.True(errors.Is(err, ErrRecordNotFound), "another tenant's customer is not deletable")
}

// ─── customer receivables ─────────────────────────────────────────────────────

// TestFindCustomeReceivables_FiltersAndMaps: unpaid only, this customer only, live
// rows only, with the invoice's columns mapped through the belongsTo relation.
func TestFindCustomeReceivables_FiltersAndMaps(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	custID, custUUID, invUUID := mkCreditInvoice(t, f, 100, 1)

	rows, err := s.findCustomeReceivables(f.ctx, custUUID)
	is.NoErr(err)
	is.Equal(len(rows), 1)

	r := rows[0]
	is.True(r.UUID != "", "the receivables row's own uuid")
	is.Equal(r.Invoice.UUID, invUUID)
	is.Equal(r.Invoice.ID, scalarInt(t, s.db, `SELECT id FROM invoices WHERE uuid = $1`, invUUID))
	is.Equal(r.Invoice.Number, scalarString(t, s.db, `SELECT code FROM invoices WHERE uuid = $1`, invUUID))
	is.EqualFloat(r.Invoice.Total, 118)
	is.EqualFloat(r.Invoice.AmountDue, 118)
	is.Equal(string(r.Invoice.PaidStatus), "unpaid")
	is.True(r.Invoice.NCF != nil, "the NCF comes off invoices.tax_number, not the dropped tax_receipts join")

	// The receivables row belongs to this customer, and to no other.
	other := newFaker(t)
	_, otherUUID := newCustomer(t, f, other).Credit("net30").Build()
	otherRows, err := s.findCustomeReceivables(f.ctx, otherUUID)
	is.NoErr(err)
	is.Equal(len(otherRows), 0)

	// A paid invoice drops out: the filter lives on the related invoice.
	_, err = s.db.Exec(`UPDATE invoices SET paid_status = 'paid' WHERE uuid = $1`, invUUID)
	is.NoErr(err)
	rows, err = s.findCustomeReceivables(f.ctx, custUUID)
	is.NoErr(err)
	is.Equal(len(rows), 0)

	// And a soft-deleted receivable drops out even while its invoice is unpaid.
	_, err = s.db.Exec(`UPDATE invoices SET paid_status = 'unpaid' WHERE uuid = $1`, invUUID)
	is.NoErr(err)
	_, err = s.db.Exec(`UPDATE receivables SET deleted_at = now() WHERE customer_id = $1`, custID)
	is.NoErr(err)
	rows, err = s.findCustomeReceivables(f.ctx, custUUID)
	is.NoErr(err)
	is.Equal(len(rows), 0)
}

// TestFindCustomeReceivables_ListsASoftDeletedCustomersRows pins existing behaviour:
// the old query joined customers without a deleted_at predicate, so a deleted
// customer's receivables were still listed. WithTrashed on the uuid lookup keeps that.
func TestFindCustomeReceivables_ListsASoftDeletedCustomersRows(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	custID, custUUID, _ := mkCreditInvoice(t, f, 100, 1)
	is.NoErr(s.deleteCustomer(f.ctx, custID))

	rows, err := s.findCustomeReceivables(f.ctx, custUUID)
	is.NoErr(err)
	is.Equal(len(rows), 1)
}

// ─── vendor by uuid ───────────────────────────────────────────────────────────

// TestFindVendorByUUID_OpeningBalance mirrors the customer test: the vendor's opening
// balance is its draft accounts_payable row.
func TestFindVendorByUUID_OpeningBalance(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	vendorID, vendorUUID, ap := mkVendorBill(t, f, 300)

	// The bill is not a draft, so there is no opening balance yet.
	v, err := s.findVendorByUUID(f.ctx, vendorUUID)
	is.NoErr(err)
	is.Equal(v.ID, vendorID)
	is.True(v.OpenBalance != nil, "OpenBalance is always non-nil")
	is.True(v.OpenBalance.InvoiceID == nil, "a posted bill is not an opening balance")

	markPayableDraft(t, f, ap)

	apID := scalarInt(t, s.db, `SELECT id FROM accounts_payable WHERE uuid = $1`, ap)
	apTotal := scalarFloat(t, s.db, `SELECT amount_total FROM accounts_payable WHERE uuid = $1`, ap)

	v, err = s.findVendorByUUID(f.ctx, vendorUUID)
	is.NoErr(err)
	is.True(v.OpenBalance.InvoiceID != nil, "the draft payable is eager-loaded")
	is.Equal(*v.OpenBalance.InvoiceID, apID)
	is.EqualFloat(*v.OpenBalance.Amount, apTotal)
	is.True(v.OpenBalance.Date != nil, "the draft payable carries its invoice_date")
}

// TestFindVendorByUUID_OpeningBalanceIsNotAnotherVendors pins the FK match, and with
// it the field type. vendorModel.ID is an int; accounts_payable.vendor_id used to map
// as an int64. The loader groups children by that field's Go value, so an int64 key
// never matches an int parent and the relation comes back nil for every vendor.
func TestFindVendorByUUID_OpeningBalanceIsNotAnotherVendors(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, firstUUID, firstAP := mkVendorBill(t, f, 100)
	_, secondUUID, secondAP := mkVendorBill(t, f, 200)
	markPayableDraft(t, f, firstAP)
	markPayableDraft(t, f, secondAP)

	a, err := s.findVendorByUUID(f.ctx, firstUUID)
	is.NoErr(err)
	is.True(a.OpenBalance.InvoiceID != nil, "first vendor's draft payable is loaded")
	is.EqualFloat(*a.OpenBalance.Amount, 100)

	b, err := s.findVendorByUUID(f.ctx, secondUUID)
	is.NoErr(err)
	is.True(b.OpenBalance.InvoiceID != nil, "second vendor's draft payable is loaded")
	is.EqualFloat(*b.OpenBalance.Amount, 200)
}

// ─── vendor writes ────────────────────────────────────────────────────────────

// TestVendorWrites_SoftDeleteNarrowing is the vendor half of the same narrowing.
func TestVendorWrites_SoftDeleteNarrowing(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	id, uuid := newVendor(t, f, g).Named("Before").Build()

	is.NoErr(s.updateVendor(f.ctx, id, &UpdateVendorForm{
		Name: "After", VendorType: "business", PaymentMethod: "cash", PaymentTerms: "net30",
	}))
	is.Equal(scalarString(t, s.db, `SELECT name FROM vendors WHERE id = $1`, id), "After")

	live, err := s.findVendorByUUID(f.ctx, uuid)
	is.NoErr(err)
	is.NoErr(s.toggleVendorStatus(f.ctx, live))
	is.Equal(scalarString(t, s.db, `SELECT status FROM vendors WHERE id = $1`, id), "disabled")

	is.NoErr(s.deleteVendor(f.ctx, id))
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM vendors WHERE id = $1 AND deleted_at IS NOT NULL AND updated_at >= deleted_at`, id), 1)

	err = s.updateVendor(f.ctx, id, &UpdateVendorForm{
		Name: "Zombie", VendorType: "business", PaymentMethod: "cash", PaymentTerms: "net30",
	})
	is.True(errors.Is(err, ErrRecordNotFound), "updating a deleted vendor is not-found")
	is.Equal(scalarString(t, s.db, `SELECT name FROM vendors WHERE id = $1`, id), "After")

	live.Status = "disabled"
	err = s.toggleVendorStatus(f.ctx, live)
	is.True(errors.Is(err, ErrRecordNotFound), "toggling a deleted vendor is not-found")

	err = s.deleteVendor(f.ctx, id)
	is.True(errors.Is(err, ErrRecordNotFound), "deleting an already-deleted vendor is not-found")

	_, err = s.findVendorByUUID(f.ctx, uuid)
	is.Err(err, "a soft-deleted vendor is not findable")
}

// ─── vendor payables ──────────────────────────────────────────────────────────

// TestFindVendorPayables_FiltersAndMaps: unpaid only, this vendor only, with
// Payable.ID/UUID coming from the payables register row and InvoiceID/UUID from the
// accounts_payable row.
func TestFindVendorPayables_FiltersAndMaps(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, vendorUUID, ap := mkVendorBill(t, f, 400)

	rows, err := s.findVendorPayables(f.ctx, vendorUUID)
	is.NoErr(err)
	is.Equal(len(rows), 1)

	p := rows[0]
	is.Equal(p.InvoiceUUID, ap)
	is.Equal(p.InvoiceID, int64(scalarInt(t, s.db, `SELECT id FROM accounts_payable WHERE uuid = $1`, ap)))
	is.Equal(p.UUID, scalarString(t, s.db,
		`SELECT payables.uuid::text FROM payables
		   JOIN accounts_payable ap ON ap.id = payables.accounts_payable_id
		  WHERE ap.uuid = $1`, ap))
	is.EqualFloat(p.AmountTotal, 400)
	is.Equal(string(p.PaidStatus), "unpaid")

	// A second vendor's bill does not appear in the first vendor's list.
	_, secondUUID, _ := mkVendorBill(t, f, 900)
	rows, err = s.findVendorPayables(f.ctx, vendorUUID)
	is.NoErr(err)
	is.Equal(len(rows), 1)
	is.EqualFloat(rows[0].AmountTotal, 400)

	second, err := s.findVendorPayables(f.ctx, secondUUID)
	is.NoErr(err)
	is.Equal(len(second), 1)
	is.EqualFloat(second[0].AmountTotal, 900)

	// A paid bill drops out.
	_, err = s.db.Exec(`UPDATE accounts_payable SET paid_status = 'paid' WHERE uuid = $1`, ap)
	is.NoErr(err)
	rows, err = s.findVendorPayables(f.ctx, vendorUUID)
	is.NoErr(err)
	is.Equal(len(rows), 0)
}

// TestFindVendorPayables_RequiresRegisterRow: the old query INNER JOINed payables, so
// an AP entry with no register row was never listed. Has("Register") is what keeps
// that true now that accounts_payable is the query root.
func TestFindVendorPayables_RequiresRegisterRow(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, vendorUUID, ap := mkVendorBill(t, f, 500)

	rows, err := s.findVendorPayables(f.ctx, vendorUUID)
	is.NoErr(err)
	is.Equal(len(rows), 1)

	_, err = s.db.Exec(
		`DELETE FROM payables WHERE accounts_payable_id = (SELECT id FROM accounts_payable WHERE uuid = $1)`, ap)
	is.NoErr(err)

	rows, err = s.findVendorPayables(f.ctx, vendorUUID)
	is.NoErr(err)
	is.Equal(len(rows), 0)
}
