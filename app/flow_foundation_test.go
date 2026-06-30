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
