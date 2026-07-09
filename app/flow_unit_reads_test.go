package app

import "testing"

// TestFindUnitByBasedQty: the playsql read returns the company's base unit
// (the factory creates one unit with base_qty = 1).
func TestFindUnitByBasedQty(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	id, err := f.s.findUnitByBasedQty(f.company.ID)
	is.NoErr(err)
	is.Equal(id, f.unitID)
}

// TestFindUnits: the playsql list read returns the company's units, including
// the factory-created base unit.
func TestFindUnits(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	units, err := f.s.findUnits(f.ctx)
	is.NoErr(err)

	found := false
	for _, u := range units {
		if int(u.ID) == f.unitID {
			found = true
			is.Equal(u.Name, "Unit")
			is.Equal(u.BaseQty, 1)
		}
	}
	is.True(found, "findUnits should return the factory base unit")
}
