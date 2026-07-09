package app

import "testing"

// These reads scan a fixed column list positionally. They had no coverage, which
// is why adding the remember_token column broke login and user creation without
// turning the suite red. Each test below fails against a SELECT */RETURNING *
// projection once the users table grows a column.

func ownerEmail(t *testing.T, f *fixture) string {
	t.Helper()
	return scalarString(t, f.s.db, `SELECT email FROM users WHERE id = $1`, f.user.Id)
}

// TestFindUserByEmail backs the login handler.
func TestFindUserByEmail(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	u, err := s.findUserByEmail(ownerEmail(t, f))
	is.NoErr(err)
	is.Equal(u.Id, f.user.Id)
	is.Equal(u.Name, "Owner")
	is.True(u.UUID != "", "uuid should be scanned, not shifted onto another column")
	is.True(u.Password != "", "password should be scanned")
}

func TestFindUserByUUID(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	byEmail, err := s.findUserByEmail(ownerEmail(t, f))
	is.NoErr(err)

	u, err := s.findUserByUUID(byEmail.UUID)
	is.NoErr(err)
	is.Equal(u.Id, f.user.Id)
	is.Equal(u.UUID, byEmail.UUID)
}

func TestFindUserByAccountUUID(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	accountUUID := scalarString(t, s.db, `SELECT uuid::text FROM accounts WHERE id = $1`, f.accountID)

	u, err := s.findUserByAccountUUID(accountUUID)
	is.NoErr(err)
	is.Equal(u.Id, f.user.Id) // the account's owner
}

// TestStoreUser covers the INSERT ... RETURNING projection and the pivot rows.
func TestStoreUser(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	companyUUID := scalarString(t, s.db, `SELECT uuid::text FROM companies WHERE id = $1`, f.company.ID)

	u, err := s.storeUser(f.ctx, &StoreUserForm{
		Name:      "Nina",
		Email:     uniq("nina") + "@test.local",
		Companies: []CompanyRole{{Company: companyUUID, Role: "admin"}},
	})
	is.NoErr(err)
	is.True(u.Id != 0, "the new user id should be returned")
	is.Equal(u.Name, "Nina")
	is.True(u.UUID != "", "uuid should be scanned from RETURNING")
	is.Equal(string(u.Status), "disabled")

	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM accounts_users WHERE account_id = $1 AND user_id = $2`, f.accountID, u.Id), 1)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM companies_users WHERE company_id = $1 AND user_id = $2 AND role = 'admin' AND current`,
		f.company.ID, u.Id), 1)
}

// TestStoreUser_UnknownCompany: an unmatched company uuid inserts no pivot row and
// must abort the transaction rather than create a half-linked user.
func TestStoreUser_UnknownCompany(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	email := uniq("ghost") + "@test.local"
	_, err := s.storeUser(f.ctx, &StoreUserForm{
		Name:      "Ghost",
		Email:     email,
		Companies: []CompanyRole{{Company: "00000000-0000-0000-0000-000000000000", Role: "admin"}},
	})
	is.Err(err, "an unknown company uuid should fail")
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM users WHERE email = $1`, email), 0)
}
