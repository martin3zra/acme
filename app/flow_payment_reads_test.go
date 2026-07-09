package app

import (
	"testing"
	"time"
)

// Customer-payment reads and the update path, converted from raw database/sql to
// playsql. Only storePayment and voidPayment had coverage before; the list read,
// the detail read, findPaymentLines and processPaymentLines had none.

// mkCreditInvoice bills a customer on credit and returns the customer id/uuid and
// the invoice uuid.
func mkCreditInvoice(t *testing.T, f *fixture, unitPrice float64, qty int) (int, string, string) {
	t.Helper()
	g := newFaker(t)
	itemID, _ := mkItem(t, f, unitPrice, unitPrice*0.6)
	custID, custUUID := newCustomer(t, f, g).Credit("net30").Build()
	invUUID := newInvoice(t, f, g).ForCustomer(custID).Credit("net30").
		WithLine(itemID, qty, unitPrice, 18).Build()
	return custID, custUUID, invUUID
}

func paymentUUID(t *testing.T, f *fixture) string {
	t.Helper()
	return scalarString(t, f.s.db,
		`SELECT uuid::text FROM receivables_income WHERE company_id = $1 ORDER BY id DESC LIMIT 1`,
		f.company.ID)
}

// TestFindPayments_CountsInvoicesAndLoadsCustomer: the old correlated
// `(select count(*) …) as invoices` becomes WithCount, and the INNER JOIN on
// customers becomes a belongsTo eager load.
func TestFindPayments_CountsInvoicesAndLoadsCustomer(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	custID, custUUID, invUUID := mkCreditInvoice(t, f, 100, 1)

	is.NoErr(f.s.storePayment(f.ctx, &StorePaymentForm{
		CustomerID: custUUID, Date: time.Now(), Amount: 118, Notes: "first",
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 118}}},
		Lines:   []*PaymentLine{{Uuid: invUUID, AmountDue: 118, Payment: 118}},
	}))

	payments, err := f.s.findPayments(f.ctx)
	is.NoErr(err)
	is.Equal(len(payments), 1)

	p := payments[0]
	is.EqualFloat(p.Amount, 118)
	is.Equal(p.Notes, "first")
	is.Equal(p.Invoices, 1) // WithCount over receivables_income_items
	is.Equal(p.Customer.UUID, custUUID)
	is.Equal(p.Customer.ID, custID)
	is.True(p.Customer.Name != "", "customer name should be eager-loaded")
	is.EqualFloat(p.Payment.Cash.Amount, 118)
	is.Equal(string(p.Status), "completed")
	is.True(p.Code != "", "payment code should be populated")
}

// TestFindPayments_HidesTrashedButDetailDoesNot: the list carried
// `deleted_at IS NULL`, the detail query did not — so the detail read opts into
// trashed rows with WithTrashed.
func TestFindPayments_HidesTrashedButDetailDoesNot(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, custUUID, invUUID := mkCreditInvoice(t, f, 100, 1)
	is.NoErr(f.s.storePayment(f.ctx, &StorePaymentForm{
		CustomerID: custUUID, Date: time.Now(), Amount: 118,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 118}}},
		Lines:   []*PaymentLine{{Uuid: invUUID, AmountDue: 118, Payment: 118}},
	}))
	uuid := paymentUUID(t, f)

	_, err := s.db.Exec(`UPDATE receivables_income SET deleted_at = now() WHERE uuid = $1`, uuid)
	is.NoErr(err)

	payments, err := f.s.findPayments(f.ctx)
	is.NoErr(err)
	is.Equal(len(payments), 0)

	p, err := f.s.findPaymentByUUID(f.ctx, uuid)
	is.NoErr(err)
	is.Equal(p.UUID, uuid)
	is.Equal(p.Invoices, 1)
}

// TestFindPaymentByUUID_SoftDeletedCustomer: customerRead is softdelete-tagged but
// the old INNER JOIN never filtered deleted_at. voidPayment needs Customer.ID to
// restore the balance, so the eager load must include trashed customers.
func TestFindPaymentByUUID_SoftDeletedCustomer(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	custID, custUUID, invUUID := mkCreditInvoice(t, f, 100, 1)
	is.NoErr(f.s.storePayment(f.ctx, &StorePaymentForm{
		CustomerID: custUUID, Date: time.Now(), Amount: 118,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 118}}},
		Lines:   []*PaymentLine{{Uuid: invUUID, AmountDue: 118, Payment: 118}},
	}))
	uuid := paymentUUID(t, f)

	is.NoErr(f.s.deleteCustomer(f.ctx, custID))

	p, err := f.s.findPaymentByUUID(f.ctx, uuid)
	is.NoErr(err)
	is.Equal(p.Customer.ID, custID)

	// The void path depends on that customer id.
	is.NoErr(f.s.voidPayment(f.ctx, uuid))
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), 118)

	// A nil map value must write SQL NULL, not an empty jsonb blob.
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM receivables_income WHERE uuid = $1 AND payment IS NULL`, uuid), 1)
}

// TestFindPaymentLines_PaidStatusAndInvoiceFields: the old CASE (an exact `= 0`
// comparison, not `<= 0`) is computed in Go; the invoice columns come from the
// eager-loaded invoice.
func TestFindPaymentLines_PaidStatusAndInvoiceFields(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, custUUID, invUUID := mkCreditInvoice(t, f, 100, 1)
	is.NoErr(f.s.storePayment(f.ctx, &StorePaymentForm{
		CustomerID: custUUID, Date: time.Now(), Amount: 50,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 50}}},
		Lines:   []*PaymentLine{{Uuid: invUUID, AmountDue: 118, Payment: 50}},
	}))

	p, err := f.s.findPaymentByUUID(f.ctx, paymentUUID(t, f))
	is.NoErr(err)

	lines, err := f.s.findPaymentLines(f.ctx, p.ID)
	is.NoErr(err)
	is.Equal(len(lines), 1)

	l := lines[0]
	is.EqualFloat(l.Payment, 50)
	is.EqualFloat(l.Invoice.AmountDue, 118)
	is.Equal(string(l.Invoice.PaidStatus), "partial") // 118 - 50 != 0
	is.Equal(l.Invoice.UUID, invUUID)
	is.True(l.Invoice.Code != "", "invoice code should be eager-loaded")
	is.EqualFloat(l.Invoice.Amount, 118) // invoices.total

	// An exactly-settled line reports paid.
	_, err = s.db.Exec(`UPDATE receivables_income_items SET amount_due = payment_amount WHERE id = $1`, l.ID)
	is.NoErr(err)
	lines, err = f.s.findPaymentLines(f.ctx, p.ID)
	is.NoErr(err)
	is.Equal(string(lines[0].Invoice.PaidStatus), "paid")
}

// TestFindPaymentLines_InvoiceWithoutTaxReceipt: the old query's INNER JOIN on
// tax_receipts silently dropped lines whose invoice has no tax receipt, because the
// column is nullable. That join is gone — the line is returned with an empty NCF.
func TestFindPaymentLines_InvoiceWithoutTaxReceipt(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, custUUID, invUUID := mkCreditInvoice(t, f, 100, 1)
	is.NoErr(f.s.storePayment(f.ctx, &StorePaymentForm{
		CustomerID: custUUID, Date: time.Now(), Amount: 118,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 118}}},
		Lines:   []*PaymentLine{{Uuid: invUUID, AmountDue: 118, Payment: 118}},
	}))

	p, err := f.s.findPaymentByUUID(f.ctx, paymentUUID(t, f))
	is.NoErr(err)

	_, err = s.db.Exec(`UPDATE invoices SET tax_receipt_id = NULL, tax_number = NULL WHERE uuid = $1`, invUUID)
	is.NoErr(err)

	lines, err := f.s.findPaymentLines(f.ctx, p.ID)
	is.NoErr(err)
	is.Equal(len(lines), 1)
	is.Equal(lines[0].Invoice.UUID, invUUID)
	is.Equal(lines[0].Invoice.NCF, "")
}

// TestVoidPayment_InvoiceWithoutTaxReceipt is the reason the join had to go:
// voidPayment walks findPaymentLines to give the money back, so while the line was
// being dropped, voiding left the invoice and the customer still charged.
func TestVoidPayment_InvoiceWithoutTaxReceipt(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	custID, custUUID, invUUID := mkCreditInvoice(t, f, 100, 1)
	is.NoErr(f.s.storePayment(f.ctx, &StorePaymentForm{
		CustomerID: custUUID, Date: time.Now(), Amount: 118,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 118}}},
		Lines:   []*PaymentLine{{Uuid: invUUID, AmountDue: 118, Payment: 118}},
	}))
	uuid := paymentUUID(t, f)
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), 0)

	_, err := s.db.Exec(`UPDATE invoices SET tax_receipt_id = NULL WHERE uuid = $1`, invUUID)
	is.NoErr(err)

	is.NoErr(f.s.voidPayment(f.ctx, uuid))

	is.Equal(scalarString(t, s.db, `SELECT status FROM receivables_income WHERE uuid = $1`, uuid), "void")
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM invoices WHERE uuid = $1`, invUUID), 118)
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), 118)
}

// TestUpdatePayment_AddedLine: allocating the payment across a second invoice. The
// ADDED branch was unreachable — processPaymentLines resolved the stored line by
// line.ID before the switch, and a new line has no id yet.
func TestUpdatePayment_AddedLine(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	custID, custUUID, firstInv := mkCreditInvoice(t, f, 100, 1)
	itemID, _ := mkItem(t, f, 50, 30)
	secondInv := newInvoice(t, f, g).ForCustomer(custID).Credit("net30").
		WithLine(itemID, 1, 50, 18).Build()

	is.NoErr(f.s.storePayment(f.ctx, &StorePaymentForm{
		CustomerID: custUUID, Date: time.Now(), Amount: 118,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 118}}},
		Lines:   []*PaymentLine{{Uuid: firstInv, AmountDue: 118, Payment: 118}},
	}))
	uuid := paymentUUID(t, f)
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), 59)

	// Allocate 59 more against the second invoice.
	is.NoErr(f.s.updatePayment(f.ctx, uuid, &UpdatePaymentForm{
		CustomerID: custUUID, Date: time.Now(), Amount: 177,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 177}}},
		Lines: []*PaymentLine{{
			Uuid: secondInv, AmountDue: 59, Payment: 59, Action: ADDED,
		}},
	}))

	p, err := f.s.findPaymentByUUID(f.ctx, uuid)
	is.NoErr(err)
	is.Equal(p.Invoices, 2)

	lines, err := f.s.findPaymentLines(f.ctx, p.ID)
	is.NoErr(err)
	is.Equal(len(lines), 2)
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM invoices WHERE uuid = $1`, secondInv), 0)
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), 0)
}

// TestUpdatePayment_UpdatedLine: the UPDATED action rewrites the line and pushes
// the difference back onto the invoice and the customer.
func TestUpdatePayment_UpdatedLine(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	custID, custUUID, invUUID := mkCreditInvoice(t, f, 100, 1)
	is.NoErr(f.s.storePayment(f.ctx, &StorePaymentForm{
		CustomerID: custUUID, Date: time.Now(), Amount: 118,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 118}}},
		Lines:   []*PaymentLine{{Uuid: invUUID, AmountDue: 118, Payment: 118}},
	}))
	uuid := paymentUUID(t, f)
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), 0)

	p, err := f.s.findPaymentByUUID(f.ctx, uuid)
	is.NoErr(err)
	lines, err := f.s.findPaymentLines(f.ctx, p.ID)
	is.NoErr(err)

	// Drop the payment from 118 to 100: 18 goes back onto the invoice.
	is.NoErr(f.s.updatePayment(f.ctx, uuid, &UpdatePaymentForm{
		CustomerID: custUUID, Date: time.Now(), Amount: 100, Notes: "revised",
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 100}}},
		Lines: []*PaymentLine{{
			ID: lines[0].ID, Uuid: invUUID, AmountDue: 118, Payment: 100, Action: UPDATED,
		}},
	}))

	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount FROM receivables_income WHERE uuid = $1`, uuid), 100)
	is.Equal(scalarString(t, s.db, `SELECT notes FROM receivables_income WHERE uuid = $1`, uuid), "revised")
	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT payment_amount FROM receivables_income_items WHERE id = $1`, lines[0].ID), 100)
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM invoices WHERE uuid = $1`, invUUID), 18)
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), 18)
}

// TestUpdatePayment_DeletedLine: the DELETED action hard-deletes the row —
// paymentItem carries no softdelete tag, so Builder.Delete issues a real DELETE.
func TestUpdatePayment_DeletedLine(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	custID, custUUID, invUUID := mkCreditInvoice(t, f, 100, 1)
	is.NoErr(f.s.storePayment(f.ctx, &StorePaymentForm{
		CustomerID: custUUID, Date: time.Now(), Amount: 118,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 118}}},
		Lines:   []*PaymentLine{{Uuid: invUUID, AmountDue: 118, Payment: 118}},
	}))
	uuid := paymentUUID(t, f)

	p, err := f.s.findPaymentByUUID(f.ctx, uuid)
	is.NoErr(err)
	lines, err := f.s.findPaymentLines(f.ctx, p.ID)
	is.NoErr(err)

	is.NoErr(f.s.updatePayment(f.ctx, uuid, &UpdatePaymentForm{
		CustomerID: custUUID, Date: time.Now(), Amount: 0,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 0}}},
		Lines: []*PaymentLine{{
			ID: lines[0].ID, Uuid: invUUID, AmountDue: 118, Payment: 118, Action: DELETED,
		}},
	}))

	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM receivables_income_items WHERE id = $1`, lines[0].ID), 0)
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM invoices WHERE uuid = $1`, invUUID), 118)
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), 118)
}
