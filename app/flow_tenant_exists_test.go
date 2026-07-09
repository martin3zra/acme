package app

import (
	"strconv"
	"testing"

	"github.com/martin3zra/forge/validator"
)

// The bare `exists:customers,id` rule only asks whether a row with that id exists
// anywhere in the table. Every table these rules point at is tenant-owned, so a
// request could name another company's customer, item, warehouse or tax receipt and
// pass validation. tenantExists binds the lookup to the requesting company.
//
// These run the real validator against the real database: the test harness context
// already carries both the company and the connection the rule resolves.

type existsProbe struct {
	ID int `json:"id"`
}

// validates runs one rule against one value, exactly as a form would.
func validates(t *testing.T, f *fixture, table string, id int) bool {
	t.Helper()
	v := &validator.Validator{}
	return v.Validate(f.ctx, existsProbe{ID: id}, map[string]any{
		"id": []any{"required", tenantExists(f.ctx, table, "id")},
	})
}

// TestTenantExists_RejectsForeignRows: a real row belonging to another company must
// not satisfy an exists rule evaluated for this one.
func TestTenantExists_RejectsForeignRows(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	other := mkAccountCompany(t, s)
	g := newFaker(t)

	// One of each, owned by each tenant.
	mineCustomer, _ := newCustomer(t, f, g).Build()
	foreignCustomer, _ := newCustomer(t, other, g).Build()
	mineItem, _ := mkItem(t, f, 100, 60)
	foreignItem, _ := mkItem(t, other, 100, 60)
	mineVendor, _ := newVendor(t, f, g).Build()
	foreignVendor, _ := newVendor(t, other, g).Build()

	cases := []struct {
		table   string
		mine    int
		foreign int
	}{
		{"customers", mineCustomer, foreignCustomer},
		{"items", mineItem, foreignItem},
		{"vendors", mineVendor, foreignVendor},
		{"warehouses", f.warehouseID, other.warehouseID},
		{"units", f.unitID, other.unitID},
		{"taxes", f.taxID, other.taxID},
		{"tax_receipts", f.taxReceiptID, other.taxReceiptID},
	}

	for _, tc := range cases {
		if !validates(t, f, tc.table, tc.mine) {
			t.Errorf("%s: own row id=%d should validate", tc.table, tc.mine)
		}
		if validates(t, f, tc.table, tc.foreign) {
			t.Errorf("%s: another company's row id=%d must NOT validate", tc.table, tc.foreign)
		}
	}

	// A row that exists nowhere is rejected too, as it always was.
	is.True(!validates(t, f, "customers", 99999999), "an unknown id must not validate")
}

// TestTenantExists_UnscopedWithoutCompany: a form validated outside a company
// context — account signup, company creation — has no company to scope to and falls
// back to the unscoped rule rather than rejecting everything.
func TestTenantExists_UnscopedWithoutCompany(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	rule := tenantExists(f.ctx, "tax_receipts", "id")
	is.Equal(rule.Constraints(),
		"exists:tax_receipts,id,company_id__"+strconv.Itoa(f.company.ID))

	// No company in context: the rule is emitted unscoped.
	bare := tenantExists(t.Context(), "tax_receipts", "id")
	is.Equal(bare.Constraints(), "exists:tax_receipts,id")
}
