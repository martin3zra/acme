package app

import "testing"

// account-repository converted to playsql. The three reads were the same query with a
// different WHERE; each drops `INNER JOIN users` for a belongsTo eager load.
// MarkAccountAsVerified's owner update was an UPDATE ... FROM join.

func accountUUIDOf(t *testing.T, f *fixture) string {
	t.Helper()
	return scalarString(t, f.s.db, `SELECT uuid::text FROM accounts WHERE id = $1`, f.accountID)
}

// TestFindAccount_LoadsOwner: all three reads return the account with its owner.
func TestFindAccount_LoadsOwner(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	uuid := accountUUIDOf(t, f)
	email := scalarString(t, s.db, `SELECT email FROM users WHERE id = $1`, f.user.Id)

	byUUID, err := s.findAccountByUUID(uuid)
	is.NoErr(err)
	is.Equal(byUUID.ID, f.accountID)
	is.Equal(byUUID.Owner.ID, f.user.Id)
	is.Equal(byUUID.Owner.Email, email)
	is.Equal(byUUID.Owner.Name, "Owner")

	byEmail, err := s.findAccountByOwnerEmailAddress(email)
	is.NoErr(err)
	is.Equal(byEmail.ID, f.accountID)
	is.Equal(byEmail.Owner.ID, f.user.Id)

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()
	byID, err := s.findAccountByIDUsingTx(tx, f.accountID)
	is.NoErr(err)
	is.Equal(byID.UUID, uuid)
	is.Equal(byID.Owner.Email, email)
}

// TestFindAccountByOwnerEmailAddress_ScopesToItsOwner: the WHERE moved onto a related
// table, so a second account's owner must not match the first.
func TestFindAccountByOwnerEmailAddress_ScopesToItsOwner(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	first := mkAccountCompany(t, s)
	second := mkAccountCompany(t, s)

	firstEmail := scalarString(t, s.db, `SELECT email FROM users WHERE id = $1`, first.user.Id)
	secondEmail := scalarString(t, s.db, `SELECT email FROM users WHERE id = $1`, second.user.Id)

	a, err := s.findAccountByOwnerEmailAddress(firstEmail)
	is.NoErr(err)
	is.Equal(a.ID, first.accountID)

	b, err := s.findAccountByOwnerEmailAddress(secondEmail)
	is.NoErr(err)
	is.Equal(b.ID, second.accountID)

	_, err = s.findAccountByOwnerEmailAddress("nobody@test.local")
	is.Err(err, "an unknown owner email finds no account")
}

// TestMarkAccountAsVerified: enables the account and its owner in one transaction.
func TestMarkAccountAsVerified(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	// The factory creates a verified account; start from unverified.
	_, err := s.db.Exec(
		`UPDATE accounts SET verified_at = NULL, status = 'disabled' WHERE id = $1`, f.accountID)
	is.NoErr(err)
	_, err = s.db.Exec(
		`UPDATE users SET email_verified_at = NULL, status = 'disabled' WHERE id = $1`, f.user.Id)
	is.NoErr(err)

	a, err := s.findAccountByUUID(accountUUIDOf(t, f))
	is.NoErr(err)
	is.True(!a.HasVerifiedAccount(), "starts unverified")

	is.True(a.MarkAccountAsVerified(s.db), "verification succeeds")

	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM accounts WHERE id = $1 AND verified_at IS NOT NULL AND status = 'enabled'`,
		f.accountID), 1)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM users WHERE id = $1 AND email_verified_at IS NOT NULL AND status = 'enabled'`,
		f.user.Id), 1)

	// Re-verifying is idempotent: both rows still match, so both updates affect one row.
	reloaded, err := s.findAccountByUUID(accountUUIDOf(t, f))
	is.NoErr(err)
	is.True(reloaded.MarkAccountAsVerified(s.db), "re-verification is a no-op that still succeeds")
}

// TestMarkAccountAsVerified_RollsBackWithoutOwner pins the guard.
//
// The owner update used to be an UPDATE ... FROM join that resolved owner_id in SQL.
// It now uses the owner id the caller already loaded, so an account struct built
// without one would silently skip verifying the user. mustAffectRows turns that into
// a failed transaction instead: the account is not verified either.
func TestMarkAccountAsVerified_RollsBackWithoutOwner(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, err := s.db.Exec(
		`UPDATE accounts SET verified_at = NULL, status = 'disabled' WHERE id = $1`, f.accountID)
	is.NoErr(err)

	// An account with no owner loaded.
	orphan := account{ID: f.accountID}
	is.True(!orphan.MarkAccountAsVerified(s.db), "a missing owner fails the whole transaction")

	// The accounts row was rolled back, not left half-verified.
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM accounts WHERE id = $1 AND verified_at IS NULL`, f.accountID), 1)
}
