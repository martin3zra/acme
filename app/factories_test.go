package app

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
)

// mkLine builds an invoice/order line for an item.
func mkLine(itemID, unitID, warehouseID, qty int, price, rate float64) *Line {
	return &Line{
		ID: itemID, Unit: unitID, WarehouseID: warehouseID,
		Qty: qty, Price: price, Rate: rate, Action: LineAction("added"),
	}
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
	// storeCompany now seeds one tax (ITBIS) and one warehouse (General) via
	// copy_shared_data. Reuse the seeded rows instead of inserting duplicates, so
	// tests that count taxes/warehouses keep their one-row baseline.
	if err := s.db.QueryRow(
		`SELECT id FROM taxes WHERE company_id = $1 ORDER BY id LIMIT 1`, companyID,
	).Scan(&f.taxID); err != nil {
		t.Fatalf("seeded tax: %v", err)
	}
	if err := s.db.QueryRow(
		`SELECT id FROM warehouses WHERE company_id = $1 ORDER BY id LIMIT 1`, companyID,
	).Scan(&f.warehouseID); err != nil {
		t.Fatalf("seeded warehouse: %v", err)
	}
	// copy_shared_data also seeds starter expense categories; drop them so tests
	// that assume an empty categories table keep their baseline.
	if _, err := s.db.Exec(`DELETE FROM expenses_categories WHERE company_id = $1`, companyID); err != nil {
		t.Fatalf("clear seeded expense categories: %v", err)
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
	// storeItem now creates the default variant; load it and set the test cost.
	if err := f.s.db.QueryRow(
		`SELECT id FROM items_variants WHERE company_id = $1 AND item_id = $2 AND is_default = true`,
		f.company.ID, itemID,
	).Scan(&variantID); err != nil {
		t.Fatalf("load default variant: %v", err)
	}
	if _, err := f.s.db.Exec(`UPDATE items_variants SET cost_price = $2 WHERE id = $1`, variantID, cost); err != nil {
		t.Fatalf("set variant cost: %v", err)
	}
	return itemID, variantID
}

// mkVariantItem creates a has_variants product with a Color attribute (two
// values) and returns the item id plus its two variant ids ordered by id. The
// item has NO default variant, matching the wired variant-matrix behaviour.
func mkVariantItem(t *testing.T, f *fixture, price float64) (itemID int, variantIDs []int) {
	t.Helper()
	color := uniq("Color")
	if err := f.s.storeAttribute(f.ctx, &StoreAttributeForm{Name: color, Type: "select", DisplayName: color}); err != nil {
		t.Fatalf("storeAttribute: %v", err)
	}
	attrs, err := f.s.findAttributes(f.ctx)
	if err != nil {
		t.Fatalf("findAttributes: %v", err)
	}
	var colorID int
	for _, a := range attrs {
		if a.Name == color {
			colorID = a.ID
		}
	}
	for _, v := range []string{"Red", "Blue"} {
		if err := f.s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{AttributeID: colorID, Value: v, DisplayName: v}); err != nil {
			t.Fatalf("storeAttributeValue: %v", err)
		}
	}
	vals, err := f.s.findAttributeValuesByAttribute(f.ctx, colorID)
	if err != nil {
		t.Fatalf("findAttributeValuesByAttribute: %v", err)
	}

	name := uniq("VariantItem")
	form := &StoreItemWithAttributesForm{
		Name: name, Price: price, TaxID: f.taxID, UnitID: f.unitID, ItemType: "product",
		AttributeIDs: []int{colorID},
		VariantCombos: []VariantCombo{
			{AttributeValueIDs: map[int]int{colorID: vals[0].ID}},
			{AttributeValueIDs: map[int]int{colorID: vals[1].ID}},
		},
	}
	if err := f.s.storeItemWithVariants(f.ctx, form); err != nil {
		t.Fatalf("storeItemWithVariants: %v", err)
	}
	if err := f.s.db.QueryRow(
		`SELECT id FROM items WHERE company_id = $1 AND name = $2`, f.company.ID, name,
	).Scan(&itemID); err != nil {
		t.Fatalf("load variant item: %v", err)
	}
	// Combos carry no explicit price (they coalesce to 0); price every variant so
	// search/line assertions see a realistic per-variant price.
	if _, err := f.s.db.Exec(
		`UPDATE items_variants SET price = $2 WHERE company_id = $1 AND item_id = $3`,
		f.company.ID, price, itemID,
	); err != nil {
		t.Fatalf("price variants: %v", err)
	}
	rows, err := f.s.db.Query(
		`SELECT id FROM items_variants WHERE company_id = $1 AND item_id = $2 ORDER BY id`,
		f.company.ID, itemID,
	)
	if err != nil {
		t.Fatalf("load variants: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan variant: %v", err)
		}
		variantIDs = append(variantIDs, id)
	}
	return itemID, variantIDs
}
