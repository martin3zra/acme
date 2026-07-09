package app

import "testing"

// The user repository had no coverage at all, which is why adding the
// remember_token column broke login and user creation without turning the suite
// red. These tests were written against the raw-SQL hotfix (53c9d58) and now guard
// the playsql conversion, which selects and scans by column name.

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
	is.True(u.UUID != "", "uuid should map by name, not by scan position")
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

// TestStoreUser covers the insert, the uuid read-back and both pivot rows.
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
	is.True(u.UUID != "", "the DB-generated uuid should be read back")
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

func companyUUIDOf(t *testing.T, f *fixture) string {
	t.Helper()
	return scalarString(t, f.s.db, `SELECT uuid::text FROM companies WHERE id = $1`, f.company.ID)
}

// TestFindUsers_LinkedCountAndAccountScope: the correlated COUNT becomes WithCount
// over the companies_users pivot, and the EXISTS(accounts_users) predicate becomes
// WhereRelation — so a second account's users must not leak in.
func TestFindUsers_LinkedCountAndAccountScope(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	other := mkAccountCompany(t, s) // a second tenant

	added, err := s.storeUser(f.ctx, &StoreUserForm{
		Name:      "Nina",
		Email:     uniq("nina") + "@test.local",
		Companies: []CompanyRole{{Company: companyUUIDOf(t, f), Role: "admin"}},
	})
	is.NoErr(err)

	users, err := s.findUsers(f.ctx)
	is.NoErr(err)
	is.Equal(len(users), 2) // the owner and Nina; the other account's owner is excluded

	byID := map[int]*User{}
	for _, u := range users {
		byID[u.Id] = u
		is.True(u.Id != other.user.Id, "another account's user must not appear")
	}
	is.Equal(byID[f.user.Id].Linked, 1) // owner is linked to the one company
	is.Equal(byID[added.Id].Linked, 1)  // Nina too
	is.Equal(byID[added.Id].Name, "Nina")
	is.Equal(byID[added.Id].Password, "") // the projection must not carry the hash
}

// TestFindUserLinkedCompanies: the link rows are read directly and each company
// arrives through a belongsTo, scoped to the current account.
func TestFindUserLinkedCompanies(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	linked, err := s.findUserLinkedCompanies(f.ctx, f.user.Id)
	is.NoErr(err)
	is.Equal(len(linked), 1)
	is.Equal(linked[0].ID, f.company.ID)
	is.Equal(linked[0].UUID, companyUUIDOf(t, f))
	is.Equal(linked[0].Role, "owner")
}

func updateUserForm(t *testing.T, uuid string, form *StoreUserForm) *StoreUserForm {
	t.Helper()
	form.SetPathParams(map[string]string{"id": uuid})
	return form
}

// TestUpdateUser: renames the user and replaces the company links.
func TestUpdateUser(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	companyUUID := companyUUIDOf(t, f)
	added, err := s.storeUser(f.ctx, &StoreUserForm{
		Name:      "Nina",
		Email:     uniq("nina") + "@test.local",
		Companies: []CompanyRole{{Company: companyUUID, Role: "admin"}},
	})
	is.NoErr(err)

	is.NoErr(s.updateUser(f.ctx, updateUserForm(t, added.UUID, &StoreUserForm{
		Name:      "Nina Renamed",
		Companies: []CompanyRole{{Company: companyUUID, Role: "standard"}},
	})))

	reloaded, err := s.findUserByUUID(added.UUID)
	is.NoErr(err)
	is.Equal(reloaded.Name, "Nina Renamed")

	// The old link was hard-deleted and replaced, not duplicated.
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM companies_users WHERE user_id = $1`, added.Id), 1)
	is.Equal(scalarString(t, s.db,
		`SELECT role::text FROM companies_users WHERE user_id = $1`, added.Id), "standard")
}

// TestUpdateUser_UnknownCompanyRollsBack: the links are deleted before being
// reinserted, so a bad company uuid must roll the whole transaction back rather
// than leave the user unlinked.
func TestUpdateUser_UnknownCompanyRollsBack(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	added, err := s.storeUser(f.ctx, &StoreUserForm{
		Name:      "Nina",
		Email:     uniq("nina") + "@test.local",
		Companies: []CompanyRole{{Company: companyUUIDOf(t, f), Role: "admin"}},
	})
	is.NoErr(err)

	err = s.updateUser(f.ctx, updateUserForm(t, added.UUID, &StoreUserForm{
		Name:      "Nina Renamed",
		Companies: []CompanyRole{{Company: "00000000-0000-0000-0000-000000000000", Role: "admin"}},
	}))
	is.Err(err, "an unknown company uuid should fail")

	reloaded, err := s.findUserByUUID(added.UUID)
	is.NoErr(err)
	is.Equal(reloaded.Name, "Nina") // rename rolled back
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM companies_users WHERE user_id = $1`, added.Id), 1)
}

// TestRememberToken covers the store / lookup / clear trio, including the
// status = 'enabled' guard and the nil-writes-NULL behaviour of clear.
func TestRememberToken(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	is.NoErr(s.storeRememberToken(f.user.Id, "deadbeef"))

	id, err := s.findUserIDByRememberToken("deadbeef")
	is.NoErr(err)
	is.Equal(id, f.user.Id)

	// Unknown hash resolves to 0 with no error.
	id, err = s.findUserIDByRememberToken("nope")
	is.NoErr(err)
	is.Equal(id, 0)

	// A disabled user is not resolvable even with a matching hash.
	_, err = s.db.Exec(`UPDATE users SET status = 'disabled' WHERE id = $1`, f.user.Id)
	is.NoErr(err)
	id, err = s.findUserIDByRememberToken("deadbeef")
	is.NoErr(err)
	is.Equal(id, 0)

	_, err = s.db.Exec(`UPDATE users SET status = 'enabled' WHERE id = $1`, f.user.Id)
	is.NoErr(err)

	is.NoErr(s.clearRememberToken(f.user.Id))
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM users WHERE id = $1 AND remember_token IS NULL`, f.user.Id), 1)
}
