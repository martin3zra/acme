package app

import (
	"testing"

	"github.com/martin3zra/forge/foundation"
)

// createAccountOwner, extracted from the SetupAccount console command and converted to
// playsql. The command itself is interactive; the two inserts it ran are not, and had
// no coverage.

// TestCreateAccountOwner_InsertsOwnerAndAccount.
func TestCreateAccountOwner_InsertsOwnerAndAccount(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()

	email := uniq("setup") + "@test.local"
	accountID, err := createAccountOwner(tx, "Setup Owner", email)
	is.NoErr(err)
	is.True(accountID > 0, "the account id comes back from the insert")

	var ownerID int
	var status string
	is.NoErr(tx.QueryRow(
		`SELECT owner_id, status FROM accounts WHERE id = $1`, accountID).Scan(&ownerID, &status))
	is.Equal(status, "disabled") // the column default

	var name, gotEmail, password, userStatus string
	is.NoErr(tx.QueryRow(
		`SELECT name, email, password, status FROM users WHERE id = $1`, ownerID).
		Scan(&name, &gotEmail, &password, &userStatus))
	is.Equal(name, "Setup Owner")
	is.Equal(gotEmail, email)
	is.Equal(userStatus, "disabled")

	// The placeholder password is hashed, not stored as plaintext.
	is.True(password != "password", "the placeholder password is not stored in the clear")
	is.True(foundation.NewHashable().Check("password", password), "and it is the hash of it")

	// uuid comes from the column default.
	var uuid string
	is.NoErr(tx.QueryRow(`SELECT uuid::text FROM accounts WHERE id = $1`, accountID).Scan(&uuid))
	is.True(uuid != "", "uuid is DB-generated")
}

// TestCreateAccountOwner_ReportsTheUserInsertFailure: users.email is unique, so a
// duplicate must come back as an error rather than a zero account id.
func TestCreateAccountOwner_ReportsTheUserInsertFailure(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()

	email := uniq("setup") + "@test.local"
	_, err = createAccountOwner(tx, "First", email)
	is.NoErr(err)

	accountID, err := createAccountOwner(tx, "Second", email)
	is.Err(err, "a duplicate email must be reported, not swallowed")
	is.Equal(accountID, 0)
}

// TestCreateAccountOwner_ReportsTheAccountInsertFailure pins the error the original
// code discarded.
//
// It prepared the accounts insert and called stmt.QueryRow *before* checking Prepare's
// error — a failed Prepare dereferenced nil — then dropped that QueryRow's own Scan
// error on the floor. A failed accounts insert therefore returned accountID = 0 with
// no error, and the caller went on to read account 0, reporting a confusing not-found
// instead of the real cause.
//
// The user insert must succeed for this to bite, so the accounts table is made
// un-insertable for the length of the transaction. The DDL rolls back with it.
func TestCreateAccountOwner_ReportsTheAccountInsertFailure(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()

	// NOT VALID leaves the existing rows alone and still rejects every new insert.
	_, err = tx.Exec(`ALTER TABLE accounts ADD CONSTRAINT tmp_reject_inserts CHECK (false) NOT VALID`)
	is.NoErr(err)

	accountID, err := createAccountOwner(tx, "Owner", uniq("setup")+"@test.local")
	is.Err(err, "a failed accounts insert must be reported, not swallowed")
	is.Equal(accountID, 0)
}
