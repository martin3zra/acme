package app

import (
	"strings"
	"testing"
	"time"
)

// apUUID returns the most recent accounts_payable uuid for a vendor.
func apUUID(t *testing.T, f *fixture, vendorID int) string {
	t.Helper()
	return scalarString(t, f.s.db,
		`SELECT uuid::text FROM accounts_payable WHERE company_id = $1 AND vendor_id = $2 ORDER BY id DESC LIMIT 1`,
		f.company.ID, vendorID)
}

// TestFlowVendorBillCreatesPayable: a vendor bill registers accounts payable and
// raises the vendor balance, but does not move stock yet (deferred to confirm).
func TestFlowVendorBillCreatesPayable(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	vendorID, _ := mkVendor(t, f, "net30")
	itemID, variantID := mkItem(t, f, 100, 60)

	uuid := mkPurchase(t, f, vendorID, PurchaseTransactionKinds.VendorBill, "net30", uniq("BILL"), nil,
		mkLine(itemID, f.unitID, f.warehouseID, 1, 100, 18))

	var kind, status string
	var total float64
	is.NoErr(s.db.QueryRow(
		`SELECT transaction_kind, purchase_status, total FROM purchases WHERE uuid = $1`, uuid).Scan(&kind, &status, &total))
	is.Equal(kind, "vendor_bill")
	is.Equal(status, "draft")
	is.EqualFloat(total, 118)

	assertRow(t, s.db, "accounts_payable", map[string]any{"vendor_id": vendorID, "company_id": f.company.ID})
	assertRow(t, s.db, "payables", map[string]any{"company_id": f.company.ID})
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_payable FROM vendors WHERE id = $1`, vendorID), 118)

	// Stock not moved until confirmed.
	assertNoRow(t, s.db, "inventory_balances", map[string]any{"variant_id": variantID})
}

// TestFlowConfirmVendorBillMovesStock: confirming a standalone vendor bill posts
// it and brings stock into the warehouse.
func TestFlowConfirmVendorBillMovesStock(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	vendorID, _ := mkVendor(t, f, "net30")
	itemID, variantID := mkItem(t, f, 100, 60)
	uuid := mkPurchase(t, f, vendorID, PurchaseTransactionKinds.VendorBill, "net30", uniq("BILL"), nil,
		mkLine(itemID, f.unitID, f.warehouseID, 5, 100, 18))

	is.NoErr(f.s.confirmPurchase(f.ctx, uuid))

	is.Equal(scalarString(t, s.db, `SELECT purchase_status FROM purchases WHERE uuid = $1`, uuid), "posted")
	assertRow(t, s.db, "inventory_movements", map[string]any{
		"variant_id": variantID, "transaction_kind": "vendor_bill",
	})
	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT quantity FROM inventory_balances WHERE company_id = $1 AND variant_id = $2 AND warehouse_id = $3`,
		f.company.ID, variantID, f.warehouseID), 5)
}

// TestFlowPayVendorBill: paying a vendor bill in full clears the payable and the
// vendor balance.
func TestFlowPayVendorBill(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	vendorID, vendorUUID := mkVendor(t, f, "net30")
	itemID, _ := mkItem(t, f, 100, 60)
	mkPurchase(t, f, vendorID, PurchaseTransactionKinds.VendorBill, "net30", uniq("BILL"), nil,
		mkLine(itemID, f.unitID, f.warehouseID, 1, 100, 18))

	ap := apUUID(t, f, vendorID)
	form := &StoreVendorPaymentForm{
		VendorID: vendorUUID, Date: time.Now(), Amount: 118,
		Payment: Payment{Cash: Cash{PaymentAmount{Amount: 118}}},
		Lines:   []*VendorPaymentLine{{UUID: ap, AmountDue: 118, Payment: 118}},
	}
	is.NoErr(f.s.storeVendorPayment(f.ctx, form))

	is.Equal(scalarString(t, s.db, `SELECT paid_status FROM accounts_payable WHERE uuid = $1`, ap), "paid")
	is.EqualFloat(scalarFloat(t, s.db, `SELECT balance_due FROM accounts_payable WHERE uuid = $1`, ap), 0)
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_payable FROM vendors WHERE id = $1`, vendorID), 0)
	assertRow(t, s.db, "vendor_payments", map[string]any{"vendor_id": vendorID, "company_id": f.company.ID})
}

// TestFlowVendorOverpaymentBlocked: a second full payment is rejected by the
// trg_prevent_ap_overpayment trigger.
func TestFlowVendorOverpaymentBlocked(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	vendorID, vendorUUID := mkVendor(t, f, "net30")
	itemID, _ := mkItem(t, f, 100, 60)
	mkPurchase(t, f, vendorID, PurchaseTransactionKinds.VendorBill, "net30", uniq("BILL"), nil,
		mkLine(itemID, f.unitID, f.warehouseID, 1, 100, 18))
	ap := apUUID(t, f, vendorID)

	pay := func() error {
		return f.s.storeVendorPayment(f.ctx, &StoreVendorPaymentForm{
			VendorID: vendorUUID, Date: time.Now(), Amount: 118,
			Payment: Payment{Cash: Cash{PaymentAmount{Amount: 118}}},
			Lines:   []*VendorPaymentLine{{UUID: ap, AmountDue: 118, Payment: 118}},
		})
	}
	is.NoErr(pay())
	is.Err(pay(), "overpaying a settled payable must be blocked")
}

// TestFlowPurchaseOrderToReceipt: a receipt converted from a purchase order links
// back to it, and confirming the receipt brings stock in.
func TestFlowPurchaseOrderToReceipt(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	vendorID, _ := mkVendor(t, f, "net30")
	itemID, variantID := mkItem(t, f, 100, 60)

	poUUID := mkPurchase(t, f, vendorID, PurchaseTransactionKinds.PurchaseOrder, "net30", "", nil,
		mkLine(itemID, f.unitID, f.warehouseID, 2, 100, 18))
	rcUUID := mkPurchase(t, f, vendorID, PurchaseTransactionKinds.PurchaseReceipt, "net30", uniq("PR"),
		&PurchaseSource{Type: PurchaseTransactionKinds.PurchaseOrder, ID: poUUID},
		mkLine(itemID, f.unitID, f.warehouseID, 2, 100, 18))

	rcSrc := scalarString(t, s.db, `SELECT COALESCE(source::text, '') FROM purchases WHERE uuid = $1`, rcUUID)
	is.True(strings.Contains(rcSrc, poUUID), "receipt.source should reference the purchase order")

	is.NoErr(f.s.confirmPurchase(f.ctx, rcUUID))
	is.Equal(scalarString(t, s.db, `SELECT purchase_status FROM purchases WHERE uuid = $1`, rcUUID), "received")
	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT quantity FROM inventory_balances WHERE company_id = $1 AND variant_id = $2 AND warehouse_id = $3`,
		f.company.ID, variantID, f.warehouseID), 2)
}
