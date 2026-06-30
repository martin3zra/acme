package app

import (
	"strings"
	"testing"
	"time"
)

// TestFlowItemCreatesVariant: an item is created with its unit link and a
// default, inventory-tracked variant (the factory supplies the variant the app
// never creates on its own).
func TestFlowItemCreatesVariant(t *testing.T) {
	s := newTestServer(t)
	f := mkAccountCompany(t, s)
	itemID, variantID := mkItem(t, f, 100, 60)

	assertRow(t, s.db, "items", map[string]any{"id": itemID, "company_id": f.company.ID})
	assertRow(t, s.db, "items_units", map[string]any{"item_id": itemID, "unit_id": f.unitID})
	assertRow(t, s.db, "items_variants", map[string]any{
		"id": variantID, "item_id": itemID, "is_default": true, "track_inventory": true,
	})
}

// TestFlowCustomerHasCode: a customer gets a sequence code and zero balance.
func TestFlowCustomerHasCode(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	id, _ := mkCustomer(t, f, "net30")

	code := scalarString(t, s.db, `SELECT code FROM customers WHERE id = $1`, id)
	is.True(strings.HasPrefix(code, "CUST"), "customer code should start with CUST, got "+code)
	is.EqualFloat(scalarFloat(t, s.db, `SELECT amount_due FROM customers WHERE id = $1`, id), 0)
}

// TestFlowCustomerOpeningBalance: a customer created with an opening balance
// produces an opening invoice + receivable and carries that balance.
func TestFlowCustomerOpeningBalance(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	form := &StoreCustomerForm{
		Name: uniq("Opening"), Email: uniq("open") + "@test.local",
		PaymentMethod: "cash", PaymentTerms: "net30", CustomerType: "business",
		TaxReceipt: f.taxReceiptID, OpenBalance: 500, OpenBalanceAsOf: time.Now(),
	}
	is.NoErr(f.s.storeCustomer(f.ctx, form))

	id := 0
	var balance float64
	row := s.db.QueryRow(`SELECT id, amount_due FROM customers WHERE company_id = $1 AND email = $2`, f.company.ID, form.Email)
	is.NoErr(row.Scan(&id, &balance))
	is.EqualFloat(balance, 500)

	assertRow(t, s.db, "invoices", map[string]any{"customer_id": id, "type": "opening"})
	assertRow(t, s.db, "receivables", map[string]any{"customer_id": id, "company_id": f.company.ID})
}
