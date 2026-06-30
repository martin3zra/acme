package app

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// mkLine builds an invoice/order line for an item.
func mkLine(itemID, unitID, warehouseID, qty int, price, rate float64) *Line {
	return &Line{
		ID: itemID, Unit: unitID, WarehouseID: warehouseID,
		Qty: qty, Price: price, Rate: rate, Action: LineAction("added"),
	}
}

// mkInvoice creates a document (invoice/estimate/order) through the real
// storeInvoice path. terms "pia" = cash, "net30" = credit. Returns the new uuid.
func mkInvoice(t *testing.T, f *fixture, customerID int, kind TransactionKind, terms string, src *TransactionSource, lines ...*Line) string {
	t.Helper()
	form := &StoreInvoiceForm{
		CustomerID: customerID,
		Date:       time.Now(),
		Terms:      terms,
		TaxReceipt: f.taxReceiptID,
		Discount:   Discount{Type: "percentage"},
		Lines:      lines,
		Kind:       kind,
		Source:     src,
	}
	form.Compute() // populate protected fields (HTTP layer normally does this)
	if kind == TransactionKinds.Invoice && terms == "pia" {
		form.Payment.Cash.Amount = form.total
	}
	uuid, err := f.s.storeInvoice(f.ctx, form)
	if err != nil {
		t.Fatalf("storeInvoice(%s,%s): %v", kind, terms, err)
	}
	return uuid
}

var seqN int64

func uniq(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, atomic.AddInt64(&seqN, 1))
}

// fixture is a ready-to-use tenant: an enabled account + owner user + company,
// with the prerequisite catalog rows (unit, tax, warehouse, tax receipt) and an
// auth context wired for repository calls.
type fixture struct {
	s            *Server
	ctx          context.Context
	company      *Company
	accountID    int
	user         *AuthUser
	unitID       int
	taxID        int
	warehouseID  int
	taxReceiptID int
}

// mkAccountCompany builds a fully-provisioned tenant. It reuses the production
// storeCompany path (sequences + shared-data copy) and backfills the catalog
// prerequisites the seeder omits.
func mkAccountCompany(t *testing.T, s *Server) *fixture {
	t.Helper()

	// User + account (enabled, verified).
	var userID int
	var userUUID string
	email := uniq("owner") + "@test.local"
	err := s.db.QueryRow(
		`INSERT INTO users (name, email, password, status, must_change_password, email_verified_at)
		 VALUES ($1, $2, 'x', 'enabled', false, now()) RETURNING id, uuid`,
		"Owner", email,
	).Scan(&userID, &userUUID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	var accountID int
	if err := s.db.QueryRow(
		`INSERT INTO accounts (owner_id, status, verified_at) VALUES ($1, 'enabled', now()) RETURNING id`,
		userID,
	).Scan(&accountID); err != nil {
		t.Fatalf("insert account: %v", err)
	}
	if _, err := s.db.Exec(`INSERT INTO accounts_users (account_id, user_id) VALUES ($1, $2)`, accountID, userID); err != nil {
		t.Fatalf("insert accounts_users: %v", err)
	}

	// Company via the real path (companies_users + copy_shared_data + sequences).
	if err := s.storeCompany(accountID, userID, StoreCompanyForm{
		Name: uniq("Acme"), RNC: "131000000", City: "Santo Domingo", Address: "Calle 1",
	}); err != nil {
		t.Fatalf("storeCompany: %v", err)
	}

	var companyID int
	var companyUUID string
	if err := s.db.QueryRow(
		`SELECT id, uuid FROM companies WHERE account_id = $1 ORDER BY id DESC LIMIT 1`, accountID,
	).Scan(&companyID, &companyUUID); err != nil {
		t.Fatalf("load company: %v", err)
	}

	company := &Company{ID: companyID, UUID: companyUUID, UserRole: "owner"}
	user := &AuthUser{Id: userID, UUID: userUUID, Email: email, Role: "owner"}

	f := &fixture{s: s, company: company, accountID: accountID, user: user}
	f.ctx = authCtx(s, company, accountID, user)

	// Catalog prerequisites.
	if err := s.db.QueryRow(
		`INSERT INTO units (company_id, name, base_qty) VALUES ($1, 'Unit', 1) RETURNING id`, companyID,
	).Scan(&f.unitID); err != nil {
		t.Fatalf("insert unit: %v", err)
	}
	if err := s.db.QueryRow(
		`INSERT INTO taxes (company_id, name, rate) VALUES ($1, 'ITBIS', 18) RETURNING id`, companyID,
	).Scan(&f.taxID); err != nil {
		t.Fatalf("insert tax: %v", err)
	}
	if err := s.db.QueryRow(
		`INSERT INTO warehouses (company_id, name) VALUES ($1, 'General') RETURNING id`, companyID,
	).Scan(&f.warehouseID); err != nil {
		t.Fatalf("insert warehouse: %v", err)
	}
	if err := s.db.QueryRow(
		`INSERT INTO tax_receipts (company_id, name, serie, type, sequence_start, sequence_end, current)
		 VALUES ($1, 'Fiscal', 'B01', 'fiscal', 1, 1000, 1) RETURNING id`, companyID,
	).Scan(&f.taxReceiptID); err != nil {
		t.Fatalf("insert tax_receipt: %v", err)
	}

	// Add purchase / vendor-payment sequence keys the seeder omits (for the
	// purchases batch). jsonb || merges top-level keys.
	if _, err := s.db.Exec(`
		UPDATE companies_settings
		   SET sequences = sequences || $2::jsonb
		 WHERE company_id = $1`, companyID, purchaseSeqJSON,
	); err != nil {
		t.Fatalf("patch sequences: %v", err)
	}

	return f
}

const purchaseSeqJSON = `{
  "purchase_order":   {"prefix": "PO-",  "next": 1, "padding": 4, "format": "{prefix}-{year}-{seq}"},
  "purchase_receipt": {"prefix": "PR-",  "next": 1, "padding": 4, "format": "{prefix}-{year}-{seq}"},
  "vendor_bill":      {"prefix": "VB-",  "next": 1, "padding": 4, "format": "{prefix}-{year}-{seq}"},
  "vendor_payment":   {"prefix": "VP-",  "next": 1, "padding": 4, "format": "{prefix}-{year}-{seq}"}
}`

// mkItem creates an item (real storeItem) plus its default, inventory-tracked
// variant (which the app never creates on its own). Returns item id + variant id.
func mkItem(t *testing.T, f *fixture, price, cost float64) (itemID, variantID int) {
	t.Helper()
	name := uniq("Item")
	form := &StoreItemForm{
		Name: name, Price: price, Description: "test item",
		TaxID: f.taxID, UnitID: f.unitID, ItemType: "product",
	}
	if err := f.s.storeItem(f.ctx, form); err != nil {
		t.Fatalf("storeItem: %v", err)
	}
	if err := f.s.db.QueryRow(
		`SELECT id FROM items WHERE company_id = $1 AND name = $2 ORDER BY id DESC LIMIT 1`,
		f.company.ID, name,
	).Scan(&itemID); err != nil {
		t.Fatalf("load item: %v", err)
	}
	if err := f.s.db.QueryRow(
		`INSERT INTO items_variants (company_id, item_id, name, combination_signature, price, cost_price, is_default, track_inventory)
		 VALUES ($1, $2, 'default', 'default', $3, $4, true, true) RETURNING id`,
		f.company.ID, itemID, price, cost,
	).Scan(&variantID); err != nil {
		t.Fatalf("insert variant: %v", err)
	}
	return itemID, variantID
}

// mkCustomer creates a customer via the real storeCustomer path. terms e.g.
// "pia" (cash) or "net30" (credit). Returns id + uuid.
func mkCustomer(t *testing.T, f *fixture, terms string) (id int, uuid string) {
	t.Helper()
	email := uniq("cust") + "@test.local"
	form := &StoreCustomerForm{
		Name: uniq("Customer"), Email: email, PaymentMethod: "cash",
		PaymentTerms: terms, CreditLimited: false, CreditLimit: 0,
		CustomerType: "business", TaxReceipt: f.taxReceiptID,
	}
	if err := f.s.storeCustomer(f.ctx, form); err != nil {
		t.Fatalf("storeCustomer: %v", err)
	}
	if err := f.s.db.QueryRow(
		`SELECT id, uuid FROM customers WHERE company_id = $1 AND email = $2`, f.company.ID, email,
	).Scan(&id, &uuid); err != nil {
		t.Fatalf("load customer: %v", err)
	}
	return id, uuid
}
