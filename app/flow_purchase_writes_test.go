package app

import (
	"database/sql"
	"testing"
)

// purchases-table writes converted to playsql. The header INSERT used to build its
// column list and renumber placeholders by hand so that purchase_status and
// invoice_number could fall through to their database defaults; a map insert says
// the same thing by omitting the key.

// TestStorePurchase_OmittedColumnsKeepDefaults: a purchase order names neither
// purchase_status nor invoice_number, so both keep their database defaults.
func TestStorePurchase_OmittedColumnsKeepDefaults(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	vendorID, _ := newVendor(t, f, g).Build()
	itemID, _ := mkItem(t, f, 100, 60)

	orderUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseOrder).
		WithLine(itemID, 1, 100, 18).Build()

	// invoice_number was never supplied: NULL, not an empty string.
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM purchases WHERE uuid = $1 AND invoice_number IS NULL`, orderUUID), 1)
	// purchase_status came from the column default, not from the insert.
	is.True(scalarString(t, s.db,
		`SELECT purchase_status::text FROM purchases WHERE uuid = $1`, orderUUID) != "",
		"purchase_status is NOT NULL and defaulted")
	is.True(scalarString(t, s.db, `SELECT uuid::text FROM purchases WHERE uuid = $1`, orderUUID) != "",
		"uuid is DB-generated and read back")

	// A receipt is created in draft, explicitly.
	receiptUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseReceipt).
		WithLine(itemID, 1, 100, 18).Build()
	is.Equal(scalarString(t, s.db,
		`SELECT purchase_status::text FROM purchases WHERE uuid = $1`, receiptUUID), "draft")

	// A vendor bill with an invoice number stores it.
	billUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		InvoiceNumber("V-1001").WithLine(itemID, 1, 100, 18).Build()
	is.Equal(scalarString(t, s.db,
		`SELECT invoice_number FROM purchases WHERE uuid = $1`, billUUID), "V-1001")
	is.Equal(scalarString(t, s.db,
		`SELECT purchase_status::text FROM purchases WHERE uuid = $1`, billUUID), "draft")
}

// poPaymentStatus drives updatePOPaymentStatus on its own transaction.
func poPaymentStatus(t *testing.T, s *Server, companyID int, poUUID string) {
	t.Helper()
	tx, err := s.db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := updatePOPaymentStatus(tx, companyID, poUUID); err != nil {
		tx.Rollback()
		t.Fatalf("updatePOPaymentStatus: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

// TestUpdatePOPaymentStatus_Branches: the three raw statements collapsed into one
// builder with two When clauses. Each branch keeps its own guard.
func TestUpdatePOPaymentStatus_Branches(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	vendorID, _ := newVendor(t, f, g).Build()
	itemID, _ := mkItem(t, f, 100, 60)
	poUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseOrder).
		WithLine(itemID, 1, 100, 18).Build()

	status := func() (string, string) {
		var pay, pur string
		is.NoErr(s.db.QueryRow(
			`SELECT payment_status::text, purchase_status::text FROM purchases WHERE uuid = $1`,
			poUUID).Scan(&pay, &pur))
		return pay, pur
	}

	// No linked bills: unpaid, and purchase_status untouched.
	_, before := status()
	poPaymentStatus(t, s, f.company.ID, poUUID)
	pay, pur := status()
	is.Equal(pay, "unpaid")
	is.Equal(pur, before)

	// A bill linked to the PO, half paid -> partial + partially_paid.
	//
	// The bill is created standalone and its source is stamped afterwards, rather
	// than passed to storePurchase. Converting a PO *directly* into a vendor bill
	// fails on purchases_source_check today — the target stamp writes a source object
	// carrying only "target", with no "type" or "id". That is a pre-existing bug (the
	// raw code's COALESCE(source,'{}') produced the same empty map), unrelated to
	// this conversion, and out of scope here.
	billUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		WithLine(itemID, 1, 100, 18).Build()
	_, err := s.db.Exec(
		`UPDATE purchases SET source = jsonb_build_object('type','purchase_order','id',$2::text)
		  WHERE uuid = $1`, billUUID, poUUID)
	is.NoErr(err)

	billID := scalarInt(t, s.db, `SELECT id FROM purchases WHERE uuid = $1`, billUUID)
	_, err = s.db.Exec(
		`UPDATE accounts_payable SET amount_paid = 50 WHERE purchase_id = $1`, billID)
	is.NoErr(err)

	poPaymentStatus(t, s, f.company.ID, poUUID)
	pay, pur = status()
	is.Equal(pay, "partial")
	is.Equal(pur, "partially_paid")

	// Fully paid -> paid + closed.
	_, err = s.db.Exec(
		`UPDATE accounts_payable SET amount_paid = amount_payable WHERE purchase_id = $1`, billID)
	is.NoErr(err)
	poPaymentStatus(t, s, f.company.ID, poUUID)
	pay, pur = status()
	is.Equal(pay, "paid")
	is.Equal(pur, "closed")

	// Back to partial: the PO is closed, so purchase_status must NOT be demoted.
	// (The old `purchase_status NOT IN ('closed')` guard, now a When clause.)
	_, err = s.db.Exec(`UPDATE accounts_payable SET amount_paid = 50 WHERE purchase_id = $1`, billID)
	is.NoErr(err)
	poPaymentStatus(t, s, f.company.ID, poUUID)
	pay, pur = status()
	is.Equal(pur, "closed")  // guard held: no demotion
	is.Equal(pay, "paid")    // and the whole update matched no row, so payment_status stands
}

// TestUpdatePOPaymentStatus_SkipsSoftDeletedPO pins a deliberate narrowing.
//
// The raw statements had no deleted_at predicate, so they would happily rewrite a
// soft-deleted purchase order's status. purchaseRead carries play:"softdelete", so
// the converted update skips it.
func TestUpdatePOPaymentStatus_SkipsSoftDeletedPO(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	vendorID, _ := newVendor(t, f, g).Build()
	itemID, _ := mkItem(t, f, 100, 60)
	poUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.PurchaseOrder).
		WithLine(itemID, 1, 100, 18).Build()

	before := scalarString(t, s.db, `SELECT payment_status::text FROM purchases WHERE uuid = $1`, poUUID)

	_, err := s.db.Exec(`UPDATE purchases SET deleted_at = now() WHERE uuid = $1`, poUUID)
	is.NoErr(err)

	// A soft-deleted PO is not rewritten, and the call is not an error.
	poPaymentStatus(t, s, f.company.ID, poUUID)

	after := scalarString(t, s.db, `SELECT payment_status::text FROM purchases WHERE uuid = $1`, poUUID)
	is.Equal(after, before)
}

// TestConfirmPurchase_GuardsMissingRow: confirmPurchase's header update targets an id
// that came from a company-scoped read, so a zero-row result is a bug, not a no-op.
func TestConfirmPurchase_GuardsMissingRow(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	vendorID, _ := newVendor(t, f, g).Build()
	itemID, _ := mkItem(t, f, 100, 60)
	billUUID := newPurchase(t, f).ForVendor(vendorID).Kind(PurchaseTransactionKinds.VendorBill).
		WithLine(itemID, 2, 100, 18).Build()

	// The happy path still confirms and posts.
	is.NoErr(s.confirmPurchase(f.ctx, billUUID))
	is.Equal(scalarString(t, s.db, `SELECT purchase_status::text FROM purchases WHERE uuid = $1`, billUUID), "posted")
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM purchases WHERE uuid = $1 AND movement_recorded`, billUUID), 1)
}

var _ = sql.ErrNoRows
