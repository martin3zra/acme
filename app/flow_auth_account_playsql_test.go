package app

import (
	"context"
	"errors"
	"testing"

	"github.com/martin3zra/forge/auth"
	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
)

// The auth/account cluster converted to playsql: the two credential resolvers and
// MarkPasswordAsChanged in auth-user.go, plus the five reads and writes in types.go.

// connCtx carries only the connection, which is all auth.NewAuth requires. The
// harness's authCtx also stamps the tenant, which these resolvers never read.
func connCtx(s *Server) context.Context {
	return context.WithValue(context.Background(), database.ConnectionKey{}, s.db)
}

func userEmail(t *testing.T, s *Server, userID int) string {
	t.Helper()
	return scalarString(t, s.db, `SELECT email FROM users WHERE id = $1`, userID)
}

// setPassword gives a user a known password through the production path.
func setPassword(t *testing.T, s *Server, userID int, password string) {
	t.Helper()
	u := &AuthUser{Id: userID}
	if err := u.MarkPasswordAsChanged(s.db, password); err != nil {
		t.Fatalf("MarkPasswordAsChanged: %v", err)
	}
}

// ─── credential resolver ──────────────────────────────────────────────────────

// TestCredentialResolver_MapsColumnsByName is the regression test for the login break.
//
// The resolver used to SELECT an explicit column list and Scan positionally. Appending
// remember_token to the users table shifted nothing here, but a SELECT * did, and the
// explicit list was the patch. playsql maps by column name, so the projection and the
// destination cannot drift apart.
//
// Every field the identity carries is asserted, because a positional scan that is off
// by one still populates every field — with the neighbouring column's value.
func TestCredentialResolver_MapsColumnsByName(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	setPassword(t, s, f.user.Id, "s3cret-password")
	email := userEmail(t, s, f.user.Id)

	a := auth.NewAuth(connCtx(s))
	got, err := a.Authenticate(email, "s3cret-password")
	is.NoErr(err)

	u, ok := got.(*AuthUser)
	is.True(ok, "the resolver returns an *AuthUser")
	is.Equal(u.Id, f.user.Id)
	is.Equal(u.Email, email)
	is.Equal(u.Name, "Owner")
	is.Equal(u.Status, "enabled")
	is.True(u.UUID != "", "uuid is populated")
	is.True(u.Password != "", "the password hash is selected — Authenticate compares it")
	is.True(u.EmailVerifiedAt != nil, "the factory verifies the owner")
	is.True(u.LastPasswordReset != nil, "MarkPasswordAsChanged stamped it")
	is.True(!u.MustChangePassword, "MarkPasswordAsChanged cleared the flag")
	is.True(u.PendingEmail == nil, "no pending email")
	is.True(u.CreatedAt != nil, "timestamps are populated")
	is.True(u.DeletedAt == nil, "a live user has no deleted_at")

	// A wrong password is rejected by Authenticate, not by the resolver.
	_, err = a.Authenticate(email, "wrong")
	is.Err(err, "a wrong password does not authenticate")

	// An unknown email finds no user.
	_, err = a.Authenticate("nobody@test.local", "s3cret-password")
	is.Err(err, "an unknown email finds no user")
}

// TestCredentialResolver_ResolvesByIdColumn: the resolver takes the column name from
// its caller. LoginUsingId passes "id", Authenticate passes "email". The name is now
// quoted as an identifier rather than interpolated into the statement.
func TestCredentialResolver_ResolvesByIdColumn(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	a := auth.NewAuth(connCtx(s))
	got, err := a.LoginUsingId(f.user.Id)
	is.NoErr(err)
	is.Equal(got.GetAuthIdentifier(), f.user.Id)

	_, err = a.LoginUsingId(99999999)
	is.Err(err, "an unknown id finds no user")
}

// TestPasswordResolver_ReturnsTheHash.
func TestPasswordResolver_ReturnsTheHash(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	setPassword(t, s, f.user.Id, "another-password")

	a := auth.NewAuth(connCtx(s))
	hash, err := a.GetCurrentPassword(f.user.Id)
	is.NoErr(err)
	is.True(hash != "", "the hash comes back")
	is.True(hash != "another-password", "and it is hashed, not the plaintext")
	is.True(foundation.NewHashable().Check("another-password", hash), "the hash verifies")

	_, err = a.GetCurrentPassword(99999999)
	is.Err(err, "an unknown id has no password")
}

// ─── MarkPasswordAsChanged ────────────────────────────────────────────────────

// TestMarkPasswordAsChanged_WritesAndGuards: the raw statement discarded the affected
// row count, so resetting the password of a user id that no longer exists reported
// success.
func TestMarkPasswordAsChanged_WritesAndGuards(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	// last_password_reset is NOT NULL with a CURRENT_TIMESTAMP default, so it can only
	// be observed changing, not appearing.
	_, err := s.db.Exec(`UPDATE users SET must_change_password = true WHERE id = $1`, f.user.Id)
	is.NoErr(err)
	before := scalarString(t, s.db,
		`SELECT last_password_reset::text FROM users WHERE id = $1`, f.user.Id)

	u := &AuthUser{Id: f.user.Id}
	is.NoErr(u.MarkPasswordAsChanged(s.db, "brand-new"))

	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM users WHERE id = $1 AND must_change_password = false`, f.user.Id), 1)

	// Compared for change, not ordering: the column default writes CURRENT_TIMESTAMP
	// (UTC) while playsql stamps Go's time.Now() (process-local).
	after := scalarString(t, s.db,
		`SELECT last_password_reset::text FROM users WHERE id = $1`, f.user.Id)
	is.True(after != before, "the reset stamp moved: "+before+" -> "+after)

	hash := scalarString(t, s.db, `SELECT password FROM users WHERE id = $1`, f.user.Id)
	is.True(foundation.NewHashable().Check("brand-new", hash), "the stored hash verifies")

	// An unknown user is an error rather than a silent success.
	ghost := &AuthUser{Id: 99999999}
	err = ghost.MarkPasswordAsChanged(s.db, "irrelevant")
	is.True(errors.Is(err, ErrRecordNotFound), "resetting a missing user's password is not-found")
}

// ─── MarkEmailAsVerified ──────────────────────────────────────────────────────

// TestMarkEmailAsVerified_GuardsMissingUser: the raw statement returned `err == nil`,
// so a user id matching no row reported verified. mustAffectRows makes it false.
func TestMarkEmailAsVerified_GuardsMissingUser(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, err := s.db.Exec(`UPDATE users SET email_verified_at = NULL WHERE id = $1`, f.user.Id)
	is.NoErr(err)

	u := &User{AuthUser: AuthUser{Id: f.user.Id}}
	is.True(u.MarkEmailAsVerified(s.db), "verifying a real user succeeds")
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM users WHERE id = $1 AND email_verified_at IS NOT NULL`, f.user.Id), 1)

	ghost := &User{AuthUser: AuthUser{Id: 99999999}}
	is.True(!ghost.MarkEmailAsVerified(s.db), "verifying a missing user must not report success")
}

// ─── Account / OwnedBy ────────────────────────────────────────────────────────

// TestUserAccount_OwnerAndNonOwner: Account keys on accounts.owner_id.
func TestUserAccount_OwnerAndNonOwner(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	owner := &User{AuthUser: AuthUser{Id: f.user.Id}}
	a := owner.Account(s.db)
	is.True(a != nil, "the owner resolves their account")
	is.Equal(a.ID, f.accountID)
	is.Equal(a.Owner.ID, f.user.Id)
	is.Equal(a.Status, "enabled")
	is.True(a.VerifiedAt != nil, "the factory verifies the account")
	is.True(a.UUID != "", "uuid is populated")
	is.True(owner.IsOwner(s.db), "IsOwner agrees")

	// A user who owns nothing.
	var strangerID int
	is.NoErr(s.db.QueryRow(
		`INSERT INTO users (name, email, password, status) VALUES ('Stranger', $1, 'x', 'enabled')
		 RETURNING id`, uniq("stranger")+"@test.local").Scan(&strangerID))

	stranger := &User{AuthUser: AuthUser{Id: strangerID}}
	is.True(stranger.Account(s.db) == nil, "a non-owner resolves no account")
	is.True(stranger.IsNotOwner(s.db), "IsNotOwner agrees")
}

// TestUserOwnedBy_ResolvesThroughThePivotAndCaches: OwnedBy reaches the account through
// accounts_users, and memoises it on the receiver.
func TestUserOwnedBy_ResolvesThroughThePivotAndCaches(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	u := &User{AuthUser: AuthUser{Id: f.user.Id}}
	a, err := u.OwnedBy(s.db)
	is.NoErr(err)
	is.Equal(a.ID, f.accountID)
	is.Equal(a.Owner.ID, f.user.Id)
	is.True(u.IsOwned(s.db), "IsOwned agrees")

	// The second call is served from the cached pointer.
	again, err := u.OwnedBy(s.db)
	is.NoErr(err)
	is.True(again == a, "OwnedBy memoises the account on the receiver")

	// A user with no accounts_users row is orphaned.
	var strangerID int
	is.NoErr(s.db.QueryRow(
		`INSERT INTO users (name, email, password, status) VALUES ('Stranger', $1, 'x', 'enabled')
		 RETURNING id`, uniq("stranger")+"@test.local").Scan(&strangerID))

	stranger := &User{AuthUser: AuthUser{Id: strangerID}}
	_, err = stranger.OwnedBy(s.db)
	is.Err(err, "a user with no membership row belongs to no account")
	is.True(stranger.IsOrphan(s.db), "and is an orphan")
}

// TestUserOwnedBy_MembershipWithoutAccountIsNotFound pins Has("Account").
//
// The old INNER JOIN dropped a membership row whose account was gone. A belongsTo alone
// would return the pivot row with a nil relation; Has keeps the join's semantics.
func TestUserOwnedBy_MembershipWithoutAccountIsNotFound(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	// Point the membership row at an account that does not exist. The FK is dropped for
	// the length of the transaction so the dangling row can be written at all.
	_, err := s.db.Exec(`ALTER TABLE accounts_users DROP CONSTRAINT IF EXISTS fk_accounts_users_accounts`)
	is.NoErr(err)
	_, err = s.db.Exec(
		`UPDATE accounts_users SET account_id = 99999999 WHERE user_id = $1`, f.user.Id)
	is.NoErr(err)

	u := &User{AuthUser: AuthUser{Id: f.user.Id}}
	_, err = u.OwnedBy(s.db)
	is.Err(err, "a membership row whose account is gone resolves to nothing")

	// Has("Account") excludes the dangling row in SQL, so the read finds nothing at all.
	// Drop it and the pivot row comes back with a nil relation, which the Go-side guard
	// turns into sql.ErrNoRows instead. Both are errors; only the first reproduces the
	// INNER JOIN, and the error value is what tells them apart.
	is.True(errors.Is(err, playsql.ErrNotFound),
		"the dangling membership is filtered in SQL, not caught in Go; got: "+errString(err))
}

// ─── currentCompany ───────────────────────────────────────────────────────────

// TestCurrentCompany_CarriesTheRoleFromThePivot: the role lives on companies_users, so
// the read is rooted there and the company arrives through the belongsTo.
func TestCurrentCompany_CarriesTheRoleFromThePivot(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	u := &User{AuthUser: AuthUser{Id: f.user.Id}}
	c, err := u.currentCompany(s.db)
	is.NoErr(err)
	is.Equal(c.ID, f.company.ID)
	is.Equal(c.Name, scalarString(t, s.db, `SELECT name FROM companies WHERE id = $1`, f.company.ID))
	is.True(c.Identifier != "", "identifier is populated")
	is.Equal(c.City, "Santo Domingo")
	is.Equal(c.Address, "Calle 1")
	is.True(c.CreatedAt != nil, "timestamps are populated")

	role := scalarString(t, s.db,
		`SELECT role FROM companies_users WHERE user_id = $1 AND current = true`, f.user.Id)
	is.Equal(c.UserRole, role)
	is.True(c.UserRole != "", "the role comes off the pivot, not the company")
}

// TestCurrentCompany_OnlyTheCurrentMembership: `current = true` selects among several.
func TestCurrentCompany_OnlyTheCurrentMembership(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	u := &User{AuthUser: AuthUser{Id: f.user.Id}}

	// No current membership: nothing resolves.
	_, err := s.db.Exec(`UPDATE companies_users SET current = false WHERE user_id = $1`, f.user.Id)
	is.NoErr(err)
	_, err = u.currentCompany(s.db)
	is.Err(err, "with no current membership there is no current company")

	// Restore it and it resolves again.
	_, err = s.db.Exec(
		`UPDATE companies_users SET current = true WHERE user_id = $1 AND company_id = $2`,
		f.user.Id, f.company.ID)
	is.NoErr(err)

	c, err := u.currentCompany(s.db)
	is.NoErr(err)
	is.Equal(c.ID, f.company.ID)
}
