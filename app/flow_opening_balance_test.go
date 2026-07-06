package app

import (
	"testing"
	"time"
)

// TestFlowVendorOpeningBalance: creating a vendor with an opening balance writes
// an opening AP entry (via the playsql openingPayableInsert) plus its payables
// cross-reference, and the vendor carries the balance. Also guards the created_by
// fix: the AP entry references the authenticated user, not a hardcoded 0.
func TestFlowVendorOpeningBalance(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	form := &StoreVendorForm{
		Name: uniq("OpenVend"), Email: uniq("ov") + "@test.local",
		PaymentMethod: "cash", PaymentTerms: "net30", VendorType: "business",
		TaxReceipt: f.taxReceiptID, OpenBalance: 750, OpenBalanceAsOf: time.Now(),
	}
	is.NoErr(f.s.storeVendor(f.ctx, form))

	var vendorID int
	is.NoErr(s.db.QueryRow(
		`SELECT id FROM vendors WHERE company_id = $1 AND email = $2`, f.company.ID, form.Email,
	).Scan(&vendorID))

	// Opening AP entry + its generated amount_payable (amount_total 750, no tax),
	// stamped with the authenticated user.
	assertRow(t, s.db, "accounts_payable", map[string]any{
		"vendor_id": vendorID, "notes": "Saldo inicial", "created_by": f.user.Id,
	})
	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT amount_payable FROM accounts_payable WHERE vendor_id = $1`, vendorID), 750)
	assertRow(t, s.db, "payables", map[string]any{"vendor_id": vendorID, "company_id": f.company.ID})
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_payable FROM vendors WHERE id = $1`, vendorID), 750)
}
