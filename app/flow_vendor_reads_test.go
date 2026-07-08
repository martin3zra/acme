package app

import "testing"

// Vendor read paths (findVendorByID / findVendors / findVendorsBySearchCriteria)
// converted from raw database/sql to playsql, mirroring the customer reads.
// These exercise the converted paths directly: enum WhereEq filter, ILIKE
// search, softdelete scope, and the address mapping.

func TestFindVendorByID_Playsql(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	id, _ := newVendor(t, f, g).Named("Global Supplies SRL").WithAddress("Zona Franca 12").Build()

	v, err := s.findVendorByID(f.ctx, id)
	is.NoErr(err)
	is.Equal(v.ID, id)
	is.Equal(v.Name, "Global Supplies SRL")
	is.Equal(v.Address, "Zona Franca 12")
}

func TestFindVendors_EnumFilterAndSoftDelete_Playsql(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	bizID, _ := newVendor(t, f, g).Build()          // business
	_, _ = newVendor(t, f, g).Individual().Build()  // individual
	delID, _ := newVendor(t, f, g).Build()          // business, to be soft-deleted

	if err := s.deleteVendor(f.ctx, delID); err != nil {
		t.Fatalf("deleteVendor: %v", err)
	}

	all, err := s.findVendors(f.ctx, "all")
	is.NoErr(err)
	is.Equal(len(all), 2) // deleted row excluded

	biz, err := s.findVendors(f.ctx, "business")
	is.NoErr(err)
	is.Equal(len(biz), 1)
	is.Equal(biz[0].ID, bizID)
	is.Equal(biz[0].VendorType, "business")

	individuals, err := s.findVendors(f.ctx, "individual")
	is.NoErr(err)
	is.Equal(len(individuals), 1)
	is.Equal(individuals[0].VendorType, "individual")
}

func TestFindVendorsBySearchCriteria_ILIKE_Playsql(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	hitID, _ := newVendor(t, f, g).Named("Ferreteria Lopez").WithAddress("Calle Mella 8").Build()
	_, _ = newVendor(t, f, g).Named("Distribuidora Sol").Build()

	res, err := s.findVendorsBySearchCriteria(f.ctx, "lopez")
	is.NoErr(err)
	is.Equal(len(res), 1)
	is.Equal(res[0].ID, hitID)
	is.Equal(res[0].Address, "Calle Mella 8")

	_, err = s.findVendorsBySearchCriteria(f.ctx, "   ")
	is.Err(err, "blank search term should error")
}
