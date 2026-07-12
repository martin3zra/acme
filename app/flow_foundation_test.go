package app

import (
	"strings"
	"testing"
)

// TestFlowCompanySetup verifies a new tenant is provisioned with the membership
// row and default settings/sequences the rest of the app depends on.
func TestFlowCompanySetup(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	assertRow(t, s.db, "companies", map[string]any{"id": f.company.ID, "account_id": f.accountID})
	assertRow(t, s.db, "companies_users", map[string]any{
		"company_id": f.company.ID, "user_id": f.user.Id, "current": true, "role": "owner",
	})
	assertRow(t, s.db, "companies_settings", map[string]any{"company_id": f.company.ID})

	// Default + patched sequences are present.
	seq := scalarString(t, s.db,
		`SELECT sequences::text FROM companies_settings WHERE company_id = $1`, f.company.ID)
	for _, key := range []string{"invoice", "customer", "estimate", "order", "purchase_order", "vendor_payment"} {
		is.True(strings.Contains(seq, key), "sequences missing key: "+key)
	}

	// copy_shared_data seeds the standard tax and a General warehouse; the fixture
	// reuses those seeded rows (see mkAccountCompany), so exactly one of each exists.
	assertRow(t, s.db, "taxes", map[string]any{"company_id": f.company.ID, "name": "ITBIS"})
	assertRow(t, s.db, "warehouses", map[string]any{"company_id": f.company.ID, "name": "General"})

	// linkCompanyDefaults pins the seeded tax + warehouse ids into settings.defaults
	// so the product and stock forms can pre-select them.
	wantTax := scalarInt(t, s.db, `SELECT id FROM taxes WHERE company_id = $1 ORDER BY id LIMIT 1`, f.company.ID)
	wantWarehouse := scalarInt(t, s.db, `SELECT id FROM warehouses WHERE company_id = $1 ORDER BY id LIMIT 1`, f.company.ID)
	gotTax := scalarInt(t, s.db, `SELECT (defaults->>'tax_id')::int FROM companies_settings WHERE company_id = $1`, f.company.ID)
	gotWarehouse := scalarInt(t, s.db, `SELECT (defaults->>'warehouse_id')::int FROM companies_settings WHERE company_id = $1`, f.company.ID)
	is.Equal(gotTax, wantTax)
	is.Equal(gotWarehouse, wantWarehouse)
}

// TestFlowCompanySeedsStarterData exercises storeCompany directly (bypassing the
// mkAccountCompany fixture, which reuses/strips seeded rows) to prove copy_shared_data
// pre-populates the tax, warehouse, and starter expense categories a new company
// needs to reach the first invoice quickly.
func TestFlowCompanySeedsStarterData(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)

	var userID int
	err := s.db.QueryRow(
		`INSERT INTO users (name, email, password, status, must_change_password, email_verified_at)
		 VALUES ('Owner', $1, 'x', 'enabled', false, now()) RETURNING id`,
		uniq("seed")+"@test.local",
	).Scan(&userID)
	is.NoErr(err)

	var accountID int
	is.NoErr(s.db.QueryRow(
		`INSERT INTO accounts (owner_id, status, verified_at) VALUES ($1, 'enabled', now()) RETURNING id`,
		userID,
	).Scan(&accountID))

	is.NoErr(s.storeCompany(accountID, userID, StoreCompanyForm{
		Name: uniq("Seeded"), RNC: "131000001", City: "Santo Domingo", Address: "Calle 2",
	}))

	companyID := scalarInt(t, s.db,
		`SELECT id FROM companies WHERE account_id = $1 ORDER BY id DESC LIMIT 1`, accountID)

	// One standard tax and one warehouse, and the starter expense categories.
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM taxes WHERE company_id = $1 AND name = 'ITBIS'`, companyID), 1)
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM warehouses WHERE company_id = $1 AND name = 'General'`, companyID), 1)
	is.True(
		scalarInt(t, s.db, `SELECT count(*) FROM expenses_categories WHERE company_id = $1`, companyID) > 0,
		"starter expense categories not seeded",
	)

	// defaults resolve to the seeded tax + warehouse.
	is.Equal(
		scalarInt(t, s.db, `SELECT (defaults->>'tax_id')::int FROM companies_settings WHERE company_id = $1`, companyID),
		scalarInt(t, s.db, `SELECT id FROM taxes WHERE company_id = $1`, companyID),
	)
	is.Equal(
		scalarInt(t, s.db, `SELECT (defaults->>'warehouse_id')::int FROM companies_settings WHERE company_id = $1`, companyID),
		scalarInt(t, s.db, `SELECT id FROM warehouses WHERE company_id = $1`, companyID),
	)
}

// TestFlowPrerequisitesPass confirms that with an empty resource_prerequisites
// table the gate passes (the default/happy path).
func TestFlowPrerequisitesPass(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, err := CheckResourcePrerequisites(f.ctx, "invoice", f.company.ID)
	is.NoErr(err)
}
