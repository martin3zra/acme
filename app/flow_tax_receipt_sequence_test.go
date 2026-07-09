package app

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

// grabTaxReceiptSequence must hand every caller a distinct fiscal sequence.
//
// The implementation this replaced read `current` in one query and wrote
// `current + 1` from the value it had read, with no row lock. Under READ COMMITTED
// two concurrent invoices both read the same `current`, both issued the same NCF,
// and the counter advanced only once. Duplicate fiscal numbers are a compliance
// problem.

// TestGrabTaxReceiptSequence_Increments: consecutive calls consume consecutive
// sequences and advance the stored counter each time.
func TestGrabTaxReceiptSequence_Increments(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()

	first, err := s.grabTaxReceiptSequence(tx, f.company.ID, f.taxReceiptID)
	is.NoErr(err)
	second, err := s.grabTaxReceiptSequence(tx, f.company.ID, f.taxReceiptID)
	is.NoErr(err)

	is.Equal(int(second.Seq), int(first.Seq)+1)
	is.True(first.Number != second.Number, "two issues must not share a fiscal number")

	var current int64
	is.NoErr(tx.QueryRow(
		`SELECT current FROM tax_receipts WHERE id = $1`, f.taxReceiptID).Scan(&current))
	is.Equal(int(current), int(second.Seq)+1)
}

// TestGrabTaxReceiptSequence_Exhausted: when current reaches sequence_end the range
// is spent and no further number is issued. The UPDATE matches no row, and the cause
// is reported as exhaustion rather than silently succeeding.
func TestGrabTaxReceiptSequence_Exhausted(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	// Park the counter one short of the end: exactly one number remains.
	_, err := s.db.Exec(
		`UPDATE tax_receipts SET current = sequence_end - 1 WHERE id = $1`, f.taxReceiptID)
	is.NoErr(err)

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()

	last, err := s.grabTaxReceiptSequence(tx, f.company.ID, f.taxReceiptID)
	is.NoErr(err)
	is.True(last.Number != "", "the final number in the range is still issued")

	_, err = s.grabTaxReceiptSequence(tx, f.company.ID, f.taxReceiptID)
	is.True(errors.Is(err, ErrTaxReceiptReachEnd), "an exhausted range reports exhaustion")

	// The counter did not run past the end.
	var current, end int64
	is.NoErr(tx.QueryRow(
		`SELECT current, sequence_end FROM tax_receipts WHERE id = $1`, f.taxReceiptID).Scan(&current, &end))
	is.Equal(int(current), int(end))
}

// TestGrabTaxReceiptSequence_UnknownReceipt: updating no row is ambiguous, so an
// unknown id must not be reported as an exhausted range.
func TestGrabTaxReceiptSequence_UnknownReceipt(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()

	_, err = s.grabTaxReceiptSequence(tx, f.company.ID, 99999999)
	is.True(errors.Is(err, ErrTaxReceiptNotFound), "an unknown receipt reports not-found")
	is.True(!errors.Is(err, ErrTaxReceiptReachEnd), "and not exhaustion")
}

// TestGrabTaxReceiptSequence_ForeignCompany: StoreInvoiceForm validates tax_receipt
// with `exists:tax_receipts,id`, which is not scoped to a company — so a request can
// name another tenant's receipt id. The company_id predicate is what stops it being
// consumed. This pins that: the foreign receipt's counter must not move.
func TestGrabTaxReceiptSequence_ForeignCompany(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	other := mkAccountCompany(t, s)

	before := scalarInt(t, s.db, `SELECT current FROM tax_receipts WHERE id = $1`, other.taxReceiptID)

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()

	_, err = s.grabTaxReceiptSequence(tx, f.company.ID, other.taxReceiptID)
	is.True(errors.Is(err, ErrTaxReceiptNotFound), "another company's receipt is not found")
	is.NoErr(tx.Rollback())

	after := scalarInt(t, s.db, `SELECT current FROM tax_receipts WHERE id = $1`, other.taxReceiptID)
	is.Equal(after, before) // the foreign counter never moved
}

// TestGrabTaxReceiptSequence_SoftDeleted: a retired receipt must not keep issuing
// fiscal numbers.
//
// Nothing in the codebase soft-deletes a tax receipt today — no handler, no
// migration, no stored procedure writes tax_receipts.deleted_at — so this pins
// intended behaviour rather than fixing a live bug. The test sets the column by hand.
func TestGrabTaxReceiptSequence_SoftDeleted(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	before := scalarInt(t, s.db, `SELECT current FROM tax_receipts WHERE id = $1`, f.taxReceiptID)

	_, err := s.db.Exec(`UPDATE tax_receipts SET deleted_at = NOW() WHERE id = $1`, f.taxReceiptID)
	is.NoErr(err)

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()

	_, err = s.grabTaxReceiptSequence(tx, f.company.ID, f.taxReceiptID)
	is.True(errors.Is(err, ErrTaxReceiptNotFound), "a soft-deleted receipt is not found")
	is.True(!errors.Is(err, ErrTaxReceiptReachEnd), "and not exhaustion")
	is.NoErr(tx.Rollback())

	after := scalarInt(t, s.db, `SELECT current FROM tax_receipts WHERE id = $1`, f.taxReceiptID)
	is.Equal(after, before) // the counter never moved
}

// TestGrabTaxReceiptSequence_ConcurrentTxDistinct is the regression test.
//
// It runs the real code on two genuinely concurrent transactions against the base
// test database — txdb multiplexes a single transaction, so it cannot express this.
//
// The interleaving is the one that broke: tx2 reaches the receipt while tx1 has
// consumed a number but not yet committed.
//
//   - Old code: tx2's plain SELECT does not block, so it read the same `current` tx1
//     read, and both issued the same NCF. Only tx2's UPDATE blocked, and it then
//     wrote the stale `current + 1` it had computed.
//   - New code: tx2's UPDATE blocks on tx1's row lock, and Postgres re-evaluates
//     `current = current + 1` against the committed row, so tx2 gets the next number.
func TestGrabTaxReceiptSequence_ConcurrentTxDistinct(t *testing.T) {
	db, cleanup, companyID, taxReceiptID := realDBTaxReceipt(t)
	defer cleanup()

	is := newIs(t)
	s := &Server{db: db}

	tx1, err := db.Begin()
	is.NoErr(err)
	defer tx1.Rollback()
	tx2, err := db.Begin()
	is.NoErr(err)
	defer tx2.Rollback()

	// tx1 consumes a number and holds the row lock, uncommitted.
	a, err := s.grabTaxReceiptSequence(tx1, companyID, taxReceiptID)
	is.NoErr(err)

	// tx2 reaches the receipt while tx1 is still open. Under the old code its SELECT
	// sails through and reads the pre-tx1 value; under the new one its UPDATE blocks.
	type result struct {
		seq *taxReceiptSeq
		err error
	}
	done := make(chan result, 1)
	go func() {
		seq, err := s.grabTaxReceiptSequence(tx2, companyID, taxReceiptID)
		done <- result{seq, err}
	}()

	// Give tx2 time to reach (and, in the old code, sail past) the read.
	time.Sleep(150 * time.Millisecond)
	is.NoErr(tx1.Commit())

	r := <-done
	is.NoErr(r.err)
	is.NoErr(tx2.Commit())
	b := r.seq

	is.True(a.Seq != b.Seq, "concurrent issues must not share a sequence")
	is.True(a.Number != b.Number, "concurrent issues must not share a fiscal number")
	is.Equal(int(b.Seq), int(a.Seq)+1)

	var current int64
	is.NoErr(db.QueryRow(`SELECT current FROM tax_receipts WHERE id = $1`, taxReceiptID).Scan(&current))
	is.Equal(int(current), int(b.Seq)+1) // the counter advanced once per issue
}

// realDBTaxReceipt opens the base test database (not the txdb harness, which shares a
// single transaction and so cannot exercise row locking) and seeds a throwaway
// company with one tax receipt.
func realDBTaxReceipt(t *testing.T) (*sql.DB, func(), int, int) {
	t.Helper()

	db, err := sql.Open("postgres", testDSN)
	if err != nil {
		t.Fatalf("open base db: %v", err)
	}

	var userID, accountID, companyID, taxReceiptID int
	must := func(err error) {
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	email := uniq("seqprobe") + "@test.local"
	must(db.QueryRow(
		`INSERT INTO users (name, email, password, status) VALUES ('seq', $1, 'x', 'enabled') RETURNING id`,
		email).Scan(&userID))
	must(db.QueryRow(
		`INSERT INTO accounts (owner_id, status) VALUES ($1, 'enabled') RETURNING id`, userID).Scan(&accountID))
	must(db.QueryRow(
		`INSERT INTO companies (account_id, name, identifier, city, address)
		 VALUES ($1, 'SeqProbe', '1', 'SD', 'x') RETURNING id`, accountID).Scan(&companyID))
	must(db.QueryRow(
		`INSERT INTO tax_receipts (company_id, name, serie, type, sequence_start, sequence_end, current)
		 VALUES ($1, 'SeqProbe', 'B99', 'fiscal', 1, 1000, 1) RETURNING id`, companyID).Scan(&taxReceiptID))

	cleanup := func() {
		db.Exec(`DELETE FROM tax_receipts WHERE id = $1`, taxReceiptID)
		db.Exec(`DELETE FROM companies WHERE id = $1`, companyID)
		db.Exec(`DELETE FROM accounts WHERE id = $1`, accountID)
		db.Exec(`DELETE FROM users WHERE id = $1`, userID)
		db.Close()
	}
	return db, cleanup, companyID, taxReceiptID
}
