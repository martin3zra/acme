//go:build integration

package app

import "testing"

// The warehouse read path is served by playsql: active warehouses for the
// company, ordered by name, with soft-deleted rows excluded.
func TestIntegration_FindWarehouses_ViaPlaysql(t *testing.T) {
	db, pdb, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db) // seeds warehouses 'Main' and 'Branch'

	srv := testServer(db)
	srv.play = pdb
	ctx := companyCtx(f.CompanyID)

	// A soft-deleted warehouse must not appear.
	must(t, exec(db,
		`INSERT INTO warehouses (company_id, name, deleted_at) VALUES ($1, 'Zeta', NOW())`,
		f.CompanyID))

	whs, err := srv.findWarehouses(ctx)
	if err != nil {
		t.Fatalf("findWarehouses: %v", err)
	}
	if len(whs) != 2 {
		t.Fatalf("active warehouses: want 2 (soft-deleted excluded), got %d", len(whs))
	}
	if whs[0].Name != "Branch" || whs[1].Name != "Main" {
		t.Errorf("order by name: got [%s, %s]", whs[0].Name, whs[1].Name)
	}
	if string(whs[0].Status) == "" {
		t.Errorf("status should be populated, got empty")
	}
}
