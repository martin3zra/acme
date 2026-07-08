package app

import "testing"

// The customer read paths (findCustomeByID / findCustomers /
// findCustomersBySearchCriteria) were converted from raw database/sql to
// playsql as the reference for the wider migration. These tests exercise the
// converted paths directly so the conversion is proven, not just compiled:
// the enum WhereEq filter, the ILIKE search, the softdelete scope, and the
// mapping back onto the customer response struct (incl. the address column).

func TestFindCustomerByID_Playsql(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	id, _ := newCustomer(t, f, g).Named("Acme Retail SRL").WithAddress("Calle Duarte 45").Build()

	c, err := s.findCustomeByID(f.ctx, id)
	is.NoErr(err)
	is.Equal(c.ID, id)
	is.Equal(c.Name, "Acme Retail SRL")
	is.Equal(c.Address, "Calle Duarte 45")
}

func TestFindCustomers_EnumFilterAndSoftDelete_Playsql(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	bizID, _ := newCustomer(t, f, g).Build()                  // business
	_, _ = newCustomer(t, f, g).Individual(g).Build()         // individual
	delID, _ := newCustomer(t, f, g).Build()                  // business, to be soft-deleted

	// Soft-delete one row; the softdelete scope must exclude it.
	if err := s.deleteCustomer(f.ctx, delID); err != nil {
		t.Fatalf("deleteCustomer: %v", err)
	}

	all, err := s.findCustomers(f.ctx, "all")
	is.NoErr(err)
	is.Equal(len(all), 2) // deleted row excluded

	// Enum WhereEq: only the business customer comes back.
	biz, err := s.findCustomers(f.ctx, "business")
	is.NoErr(err)
	is.Equal(len(biz), 1)
	is.Equal(biz[0].ID, bizID)
	is.Equal(biz[0].CustomerType, "business")

	individuals, err := s.findCustomers(f.ctx, "individual")
	is.NoErr(err)
	is.Equal(len(individuals), 1)
	is.Equal(individuals[0].CustomerType, "individual")
}

func TestFindCustomersBySearchCriteria_ILIKE_Playsql(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	hitID, _ := newCustomer(t, f, g).Named("Zamora Distributors").WithAddress("Av. 27 de Febrero").Build()
	_, _ = newCustomer(t, f, g).Named("Northwind Foods").Build()

	// Case-insensitive partial match on name.
	res, err := s.findCustomersBySearchCriteria(f.ctx, "zamora")
	is.NoErr(err)
	is.Equal(len(res), 1)
	is.Equal(res[0].ID, hitID)
	is.Equal(res[0].Address, "Av. 27 de Febrero")

	// Empty term is rejected before touching the DB.
	_, err = s.findCustomersBySearchCriteria(f.ctx, "   ")
	is.Err(err, "blank search term should error")
}
