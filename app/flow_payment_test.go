package app

import (
	"testing"
	"time"
)

// TestFlowCollectCreditInvoice: paying a credit invoice in full zeroes the
// invoice balance, closes it, and clears the customer's amount due.
func TestFlowCollectCreditInvoice(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	g := newFaker(t)
	itemID, _ := mkItem(t, f, 100, 60)
	custID, custUUID := newCustomer(t, f, g).Credit("net30").Build()

	invUUID := newInvoice(t, f, g).ForCustomer(custID).Credit("net30").
		WithLine(itemID, 1, 100, 18).Build()
	const total = 118.0

	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), total)

	form := &StorePaymentForm{
		CustomerID: custUUID,
		Date:       time.Now(),
		Amount:     total,
		Payment:    Payment{Cash: Cash{PaymentAmount{Amount: total}}},
		Lines:      []*PaymentLine{{Uuid: invUUID, AmountDue: total, Payment: total}},
	}
	is.NoErr(f.s.storePayment(f.ctx, form))

	// Invoice fully paid + closed.
	var paid, status string
	var amountDue float64
	row := s.db.QueryRow(`SELECT paid_status, status, amount_due FROM invoices WHERE uuid = $1`, invUUID)
	is.NoErr(row.Scan(&paid, &status, &amountDue))
	is.Equal(paid, "paid")
	is.Equal(status, "closed")
	is.EqualFloat(amountDue, 0)

	// Customer balance cleared; payment records persisted.
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), 0)
	assertRow(t, s.db, "receivables_income", map[string]any{"customer_id": custID, "company_id": f.company.ID})
}

// TestFlowPartialPayment: a partial payment leaves the invoice partial with the
// remaining balance on both invoice and customer.
func TestFlowPartialPayment(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	g := newFaker(t)
	itemID, _ := mkItem(t, f, 100, 60)
	custID, custUUID := newCustomer(t, f, g).Credit("net30").Build()
	invUUID := newInvoice(t, f, g).ForCustomer(custID).Credit("net30").
		WithLine(itemID, 1, 100, 18).Build()

	form := &StorePaymentForm{
		CustomerID: custUUID, Date: time.Now(), Amount: 50,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 50}}},
		Lines:   []*PaymentLine{{Uuid: invUUID, AmountDue: 118, Payment: 50}},
	}
	is.NoErr(f.s.storePayment(f.ctx, form))

	var paid string
	var amountDue float64
	is.NoErr(s.db.QueryRow(`SELECT paid_status, amount_due FROM invoices WHERE uuid = $1`, invUUID).Scan(&paid, &amountDue))
	is.Equal(paid, "partial")
	is.EqualFloat(amountDue, 68)
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), 68)
}
