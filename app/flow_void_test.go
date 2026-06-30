package app

import (
	"testing"
	"time"
)

// TestFlowVoidPaymentRestoresBalance: voiding a customer payment re-opens the
// invoice and restores the customer's amount due.
func TestFlowVoidPaymentRestoresBalance(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, custUUID := mkCustomer(t, f, "net30")
	invUUID := mkInvoice(t, f, custID, TransactionKinds.Invoice, "net30", nil,
		mkLine(itemID, f.unitID, f.warehouseID, 1, 100, 18))

	is.NoErr(f.s.storePayment(f.ctx, &StorePaymentForm{
		CustomerID: custUUID, Date: time.Now(), Amount: 118,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 118}}},
		Lines:   []*PaymentLine{{Uuid: invUUID, AmountDue: 118, Payment: 118}},
	}))
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), 0)

	payUUID := scalarString(t, s.db, `SELECT uuid::text FROM receivables_income WHERE company_id = $1 ORDER BY id DESC LIMIT 1`, f.company.ID)
	is.NoErr(f.s.voidPayment(f.ctx, payUUID))

	is.Equal(scalarString(t, s.db, `SELECT status FROM receivables_income WHERE uuid = $1`, payUUID), "void")
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM invoices WHERE uuid = $1`, invUUID), 118)
	is.Equal(scalarString(t, s.db, `SELECT paid_status FROM invoices WHERE uuid = $1`, invUUID), "unpaid")
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, custID), 118)
}

// TestFlowVoidVendorPaymentRestoresPayable: voiding a vendor payment re-opens the
// payable and restores the vendor balance.
func TestFlowVoidVendorPaymentRestoresPayable(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	vendorID, vendorUUID := mkVendor(t, f, "net30")
	itemID, _ := mkItem(t, f, 100, 60)
	mkPurchase(t, f, vendorID, PurchaseTransactionKinds.VendorBill, "net30", uniq("BILL"), nil,
		mkLine(itemID, f.unitID, f.warehouseID, 1, 100, 18))
	ap := apUUID(t, f, vendorID)

	is.NoErr(f.s.storeVendorPayment(f.ctx, &StoreVendorPaymentForm{
		VendorID: vendorUUID, Date: time.Now(), Amount: 118,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 118}}},
		Lines:   []*VendorPaymentLine{{UUID: ap, AmountDue: 118, Payment: 118}},
	}))
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_payable FROM vendors WHERE id = $1`, vendorID), 0)

	vpUUID := scalarString(t, s.db, `SELECT uuid::text FROM vendor_payments WHERE company_id = $1 ORDER BY id DESC LIMIT 1`, f.company.ID)
	is.NoErr(f.s.voidVendorPayment(f.ctx, vpUUID))

	is.Equal(scalarString(t, s.db, `SELECT status FROM vendor_payments WHERE uuid = $1`, vpUUID), "void")
	is.EqualFloat(scalarFloat(t, s.db, `SELECT balance_due FROM accounts_payable WHERE uuid = $1`, ap), 118)
	is.Equal(scalarString(t, s.db, `SELECT paid_status FROM accounts_payable WHERE uuid = $1`, ap), "unpaid")
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_payable FROM vendors WHERE id = $1`, vendorID), 118)
}
