//go:build integration

package app

import "testing"

// The vendor read path is served by playsql: it returns active vendors for the
// company ordered by name, excludes soft-deleted rows, and the (Go-side)
// vendor_type filter narrows the result.
func TestIntegration_FindVendors_ViaPlaysql(t *testing.T) {
	db, pdb, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)

	srv := testServer(db)
	srv.play = pdb
	ctx := companyCtx(f.CompanyID)

	// Two active vendors (one of each type) plus a soft-deleted one.
	must(t, exec(db,
		`INSERT INTO vendors (company_id, name, vendor_type, code) VALUES
		 ($1, 'Beta',  'business',   'V-B'),
		 ($1, 'Alpha', 'individual', 'V-A')`, f.CompanyID))
	must(t, exec(db,
		`INSERT INTO vendors (company_id, name, vendor_type, code, deleted_at)
		 VALUES ($1, 'Zeta', 'business', 'V-Z', NOW())`, f.CompanyID))

	// "all": active only, ordered by name.
	all, err := srv.findVendors(ctx, "all")
	if err != nil {
		t.Fatalf("findVendors(all): %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("active vendors: want 2 (soft-deleted excluded), got %d", len(all))
	}
	if all[0].Name != "Alpha" || all[1].Name != "Beta" {
		t.Errorf("order by name: got [%s, %s]", all[0].Name, all[1].Name)
	}
	// toVendor mapping populates derived fields.
	if all[0].Address == "" || string(all[0].Status) == "" {
		t.Errorf("mapped vendor missing derived fields: %+v", all[0])
	}

	// Type filter narrows the set.
	biz, err := srv.findVendors(ctx, VendorType("business"))
	if err != nil {
		t.Fatalf("findVendors(business): %v", err)
	}
	if len(biz) != 1 || biz[0].Name != "Beta" {
		t.Errorf("business filter: want [Beta], got %v", names(biz))
	}

	ind, err := srv.findVendors(ctx, VendorType("individual"))
	if err != nil {
		t.Fatalf("findVendors(individual): %v", err)
	}
	if len(ind) != 1 || ind[0].Name != "Alpha" {
		t.Errorf("individual filter: want [Alpha], got %v", names(ind))
	}
}

func names(vs []*vendor) []string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = v.Name
	}
	return out
}
