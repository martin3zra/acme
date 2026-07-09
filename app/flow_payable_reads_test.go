package app

import (
	"testing"
	"time"
)

// Payable read paths converted from raw database/sql to playsql. The flow suite
// only reaches findVendorPaymentByUUID/findVendorPaymentLines (through
// voidVendorPayment), so the list reads, the relation mappings and the filters
// they dropped joins for are exercised directly here.

// mkVendorBill registers one vendor bill and returns the vendor id/uuid plus the
// accounts_payable uuid it created.
func mkVendorBill(t *testing.T, f *fixture, amount float64) (int, string, string) {
	t.Helper()
	g := newFaker(t)
	vendorID, vendorUUID := newVendor(t, f, g).Build()
	itemID, _ := mkItem(t, f, 100, 60)
	newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		WithLine(itemID, 1, amount, 0).Build()
	return vendorID, vendorUUID, apUUID(t, f, vendorID)
}

// TestFindPayables_JoinsRegisterRow: the AP entry is the query root and the
// payables cross-reference row arrives through the Register relation, so
// Payable.ID/UUID must be the payables row and InvoiceID/UUID the AP row.
func TestFindPayables_JoinsRegisterRow(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, _, ap := mkVendorBill(t, f, 100)

	payables, err := f.s.findPayables(f.ctx)
	is.NoErr(err)
	is.Equal(len(payables), 1)

	p := payables[0]
	is.Equal(p.InvoiceUUID, ap)
	is.EqualFloat(p.AmountPayable, 100)
	is.EqualFloat(p.AmountPaid, 0)
	is.Equal(string(p.PaidStatus), "unpaid")
	is.True(p.Notes != nil, "notes should be non-nil (COALESCE'd to empty)")

	// ID/UUID identify the payables register row, not the AP entry.
	wantID := scalarInt(t, s.db,
		`SELECT id FROM payables WHERE company_id = $1 AND accounts_payable_id = $2`,
		f.company.ID, p.InvoiceID)
	is.Equal(int(p.ID), wantID)
	is.Equal(p.UUID, scalarString(t, s.db, `SELECT uuid::text FROM payables WHERE id = $1`, wantID))
	is.True(p.UUID != p.InvoiceUUID, "payables uuid and accounts_payable uuid must differ")
}

// TestFindPayables_OrdersByDueDate: ordering moved onto the AP root's own column.
func TestFindPayables_OrdersByDueDate(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, _, first := mkVendorBill(t, f, 100)
	_, _, second := mkVendorBill(t, f, 200)

	// Push the first bill's due date out so the second must sort ahead of it.
	_, err := s.db.Exec(`UPDATE accounts_payable SET due_date = due_date + interval '30 days' WHERE uuid = $1`, first)
	is.NoErr(err)

	payables, err := f.s.findPayables(f.ctx)
	is.NoErr(err)
	is.Equal(len(payables), 2)
	is.Equal(payables[0].InvoiceUUID, second)
	is.Equal(payables[1].InvoiceUUID, first)
}

// TestFindPayables_ExcludesVoidAndPaid: the two status filters survive the
// rewrite (they are != comparisons against Postgres enum columns).
func TestFindPayables_ExcludesVoidAndPaid(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, vendorUUID, paid := mkVendorBill(t, f, 100)
	_, _, voided := mkVendorBill(t, f, 200)
	_, _, open := mkVendorBill(t, f, 300)

	is.NoErr(f.s.storeVendorPayment(f.ctx, &StoreVendorPaymentForm{
		VendorID: vendorUUID, Date: time.Now(), Amount: 100,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 100}}},
		Lines:   []*VendorPaymentLine{{UUID: paid, AmountDue: 100, Payment: 100}},
	}))

	_, err := s.db.Exec(`UPDATE accounts_payable SET status = 'void' WHERE uuid = $1`, voided)
	is.NoErr(err)

	payables, err := f.s.findPayables(f.ctx)
	is.NoErr(err)
	is.Equal(len(payables), 1)
	is.Equal(payables[0].InvoiceUUID, open)
}

// TestFindVendorPayments_Playsql: the list read decodes the payment jsonb blob
// and carries the vendor loaded through the belongsTo relation.
func TestFindVendorPayments_Playsql(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	vendorID, vendorUUID, ap := mkVendorBill(t, f, 100)

	is.NoErr(f.s.storeVendorPayment(f.ctx, &StoreVendorPaymentForm{
		VendorID: vendorUUID, Date: time.Now(), Amount: 40, Notes: "partial settlement",
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 40}}},
		Lines:   []*VendorPaymentLine{{UUID: ap, AmountDue: 100, Payment: 40}},
	}))

	payments, err := f.s.findVendorPayments(f.ctx)
	is.NoErr(err)
	is.Equal(len(payments), 1)

	p := payments[0]
	is.EqualFloat(p.Amount, 40)
	is.Equal(p.Notes, "partial settlement")
	is.Equal(p.Status, "completed")
	is.Equal(p.Vendor.ID, vendorID)
	is.Equal(p.Vendor.UUID, vendorUUID)
	is.True(p.Payment != nil, "payment jsonb should decode")
	is.EqualFloat(p.Payment.Cash.Amount, 40)
}

// TestFindVendorPaymentLines_PaidStatus: the old CASE expression is now computed
// in Go, and the bill columns come from the eager-loaded AP entry.
func TestFindVendorPaymentLines_PaidStatus(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, vendorUUID, ap := mkVendorBill(t, f, 100)

	is.NoErr(f.s.storeVendorPayment(f.ctx, &StoreVendorPaymentForm{
		VendorID: vendorUUID, Date: time.Now(), Amount: 40,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 40}}},
		Lines:   []*VendorPaymentLine{{UUID: ap, AmountDue: 100, Payment: 40}},
	}))

	payment, err := f.s.findVendorPaymentByUUID(f.ctx,
		scalarString(t, s.db, `SELECT uuid::text FROM vendor_payments WHERE company_id = $1`, f.company.ID))
	is.NoErr(err)

	lines, err := f.s.findVendorPaymentLines(f.ctx, payment.ID)
	is.NoErr(err)
	is.Equal(len(lines), 1)

	l := lines[0]
	is.Equal(l.APUUID, ap)
	is.EqualFloat(l.AmountDue, 100)
	is.EqualFloat(l.Payment, 40)
	is.Equal(l.PaidStatus, "partial")
	is.True(l.BillNumber != "", "bill number should come from the AP entry")
}

// TestFindVendorPaymentByUUID_SoftDeletedVendor: vendorRead is soft-deletable, so
// the eager load has to opt into trashed rows — the old INNER JOIN never filtered
// deleted_at, and voidVendorPayment needs Vendor.ID to restore the balance.
func TestFindVendorPaymentByUUID_SoftDeletedVendor(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	vendorID, vendorUUID, ap := mkVendorBill(t, f, 100)

	is.NoErr(f.s.storeVendorPayment(f.ctx, &StoreVendorPaymentForm{
		VendorID: vendorUUID, Date: time.Now(), Amount: 100,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 100}}},
		Lines:   []*VendorPaymentLine{{UUID: ap, AmountDue: 100, Payment: 100}},
	}))

	vpUUID := scalarString(t, s.db, `SELECT uuid::text FROM vendor_payments WHERE company_id = $1`, f.company.ID)
	is.NoErr(f.s.deleteVendor(f.ctx, vendorID))

	p, err := f.s.findVendorPaymentByUUID(f.ctx, vpUUID)
	is.NoErr(err)
	is.Equal(p.Vendor.ID, vendorID)

	// The void path depends on that vendor id to give the balance back.
	is.NoErr(f.s.voidVendorPayment(f.ctx, vpUUID))
	is.Equal(scalarString(t, s.db, `SELECT status FROM vendor_payments WHERE uuid = $1`, vpUUID), "void")
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_payable FROM vendors WHERE id = $1`, vendorID), 100)
}
