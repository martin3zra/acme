//go:build integration

package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/martin3zra/forge/auth"
	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/support"
)

// Storing a vendor bill creates the linked accounts_payable entry and bumps the
// vendor's outstanding payable. Built through ParseRequest so the form carries
// the authenticated user that createAPForVendorBill records as created_by.
func TestIntegration_VendorBill_CreatesAccountsPayable(t *testing.T) {
	db, _, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)
	seedPurchaseSequences(t, db, f.CompanyID)

	var vendorID int
	must(t, db.QueryRow(
		`INSERT INTO vendors (company_id, name) VALUES ($1, 'Acme Supply') RETURNING id`,
		f.CompanyID).Scan(&vendorID))
	var unitID int
	must(t, db.QueryRow(
		`INSERT INTO units (company_id, name, base_qty) VALUES ($1, 'unit', 1) RETURNING id`,
		f.CompanyID).Scan(&unitID))

	date := time.Now().UTC().Format(time.RFC3339)
	body := fmt.Sprintf(`{
		"vendor_id": %d,
		"kind": "vendor_bill",
		"date": %q,
		"terms": "net0",
		"invoice_number": "VB-INV-001",
		"discount": {"type": "fixed", "value": 0},
		"lines": [
			{"id": %d, "unit": %d, "qty": 5, "price": 5, "rate": 0, "action": "added", "warehouse_id": %d}
		]
	}`, vendorID, date, f.ItemID, unitID, f.WHFrom)

	req := httptest.NewRequest(http.MethodPost, "/purchases", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	_, req = testSessionManager.Start(req)
	ctx := req.Context()
	ctx = context.WithValue(ctx, database.ConnectionKey{}, db)
	ctx = context.WithValue(ctx, auth.ContextUserID{}, map[string]any{"id": f.UserID, "role": "admin"})
	ctx = context.WithValue(ctx, CompanyKey{}, &Company{ID: f.CompanyID})
	req = req.WithContext(ctx)

	var form StorePurchaseForm
	if err := support.ParseRequest(req, &form); err != nil {
		t.Fatalf("ParseRequest: %v", err)
	}

	if _, err := srv.storePurchase(ctx, &form); err != nil {
		t.Fatalf("storePurchase: %v", err)
	}

	// One AP entry for the vendor, amount = 5 x 5 = 25, recorded by the user.
	var count int
	var amount float64
	var createdBy int
	var invoiceNumber string
	must(t, db.QueryRow(
		`SELECT count(*), coalesce(max(amount_total),0), coalesce(max(created_by),0),
		        coalesce(max(invoice_number),'')
		   FROM accounts_payable WHERE company_id=$1 AND vendor_id=$2`,
		f.CompanyID, vendorID).Scan(&count, &amount, &createdBy, &invoiceNumber))

	if count != 1 {
		t.Fatalf("accounts_payable rows: want 1, got %d", count)
	}
	if amount != 25 {
		t.Errorf("amount_total: want 25, got %v", amount)
	}
	if createdBy != f.UserID {
		t.Errorf("created_by: want %d, got %d", f.UserID, createdBy)
	}
	if invoiceNumber != "VB-INV-001" {
		t.Errorf("invoice_number: want VB-INV-001, got %q", invoiceNumber)
	}

	// The vendor's outstanding payable reflects the bill.
	var payable float64
	must(t, db.QueryRow(
		`SELECT amount_payable FROM vendors WHERE id=$1`, vendorID).Scan(&payable))
	if payable != 25 {
		t.Errorf("vendor amount_payable: want 25, got %v", payable)
	}
}
