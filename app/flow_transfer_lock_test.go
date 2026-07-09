package app

import (
	"database/sql"
	"testing"
	"time"
)

// loadTransferForUpdate and lockedBalance take FOR UPDATE row locks. Those locks are
// what stop two concurrent dispatches reading the same status (or the same on-hand
// quantity) and both moving stock.
//
// playsql v0.3.0's LockForUpdate replaced the raw SQL. A conversion that dropped the
// clause would still pass every other test in the suite — the lock only shows itself
// under real concurrency, which txdb cannot express because it multiplexes a single
// transaction.

// realDBTransfer seeds a throwaway company with two warehouses and one requested
// transfer, on the base test database.
func realDBTransfer(t *testing.T) (*sql.DB, func(), int, string) {
	t.Helper()

	db, err := sql.Open("postgres", testDSN)
	if err != nil {
		t.Fatalf("open base db: %v", err)
	}

	var userID, accountID, companyID, fromWh, toWh, transferID int
	var uuid string
	must := func(err error) {
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	must(db.QueryRow(
		`INSERT INTO users (name, email, password, status) VALUES ('lock', $1, 'x', 'enabled') RETURNING id`,
		uniq("lockprobe")+"@test.local").Scan(&userID))
	must(db.QueryRow(
		`INSERT INTO accounts (owner_id, status) VALUES ($1, 'enabled') RETURNING id`, userID).Scan(&accountID))
	must(db.QueryRow(
		`INSERT INTO companies (account_id, name, identifier, city, address)
		 VALUES ($1, 'LockProbe', '1', 'SD', 'x') RETURNING id`, accountID).Scan(&companyID))
	must(db.QueryRow(
		`INSERT INTO warehouses (company_id, name) VALUES ($1, 'From') RETURNING id`, companyID).Scan(&fromWh))
	must(db.QueryRow(
		`INSERT INTO warehouses (company_id, name) VALUES ($1, 'To') RETURNING id`, companyID).Scan(&toWh))
	must(db.QueryRow(
		`INSERT INTO inventory_transfers (company_id, from_warehouse_id, to_warehouse_id, status, created_at)
		 VALUES ($1, $2, $3, 'requested', now()) RETURNING id, uuid::text`,
		companyID, fromWh, toWh).Scan(&transferID, &uuid))

	cleanup := func() {
		db.Exec(`DELETE FROM inventory_transfers WHERE id = $1`, transferID)
		db.Exec(`DELETE FROM warehouses WHERE company_id = $1`, companyID)
		db.Exec(`DELETE FROM companies WHERE id = $1`, companyID)
		db.Exec(`DELETE FROM accounts WHERE id = $1`, accountID)
		db.Exec(`DELETE FROM users WHERE id = $1`, userID)
		db.Close()
	}
	return db, cleanup, companyID, uuid
}

// TestLoadTransferForUpdate_BlocksSecondReader: the second transaction must wait on
// the first's row lock. Drop LockForUpdate and it returns immediately.
func TestLoadTransferForUpdate_BlocksSecondReader(t *testing.T) {
	db, cleanup, companyID, uuid := realDBTransfer(t)
	defer cleanup()

	is := newIs(t)

	tx1, err := db.Begin()
	is.NoErr(err)
	defer tx1.Rollback()
	tx2, err := db.Begin()
	is.NoErr(err)
	defer tx2.Rollback()

	// tx1 holds the lock.
	h, err := loadTransferForUpdate(tx1, companyID, uuid)
	is.NoErr(err)
	is.Equal(string(h.Status), "requested")

	done := make(chan error, 1)
	go func() {
		_, err := loadTransferForUpdate(tx2, companyID, uuid)
		done <- err
	}()

	// While tx1 is open, tx2 must not get through.
	select {
	case err := <-done:
		t.Fatalf("second reader was not blocked (lock missing): err=%v", err)
	case <-time.After(200 * time.Millisecond):
	}

	is.NoErr(tx1.Commit())

	select {
	case err := <-done:
		is.NoErr(err) // proceeds once the lock is released
	case <-time.After(3 * time.Second):
		t.Fatal("second reader never unblocked after commit")
	}
	is.NoErr(tx2.Commit())
}

// TestLockedBalance_BlocksSecondReader: same property for the on-hand check that
// guards a dispatch against double-spending stock.
func TestLockedBalance_BlocksSecondReader(t *testing.T) {
	db, cleanup, companyID, _ := realDBTransfer(t)
	defer cleanup()

	is := newIs(t)

	var warehouseID int
	is.NoErr(db.QueryRow(
		`SELECT id FROM warehouses WHERE company_id = $1 ORDER BY id LIMIT 1`, companyID).Scan(&warehouseID))

	// inventory_balances.variant_id is a foreign key, so the variant has to be real.
	var taxID, itemID, variantID int
	is.NoErr(db.QueryRow(
		`INSERT INTO taxes (company_id, name, rate) VALUES ($1, 'ITBIS', 18) RETURNING id`, companyID).Scan(&taxID))
	is.NoErr(db.QueryRow(
		`INSERT INTO items (company_id, name, description, price, tax_id)
		 VALUES ($1, 'LockItem', 'd', 100, $2) RETURNING id`, companyID, taxID).Scan(&itemID))
	is.NoErr(db.QueryRow(
		`INSERT INTO items_variants (company_id, item_id, name, sku, combination_signature, is_default, price, cost_price, track_inventory)
		 VALUES ($1, $2, 'LockItem', 'SKU-LOCK', 'default', true, 100, 60, true) RETURNING id`,
		companyID, itemID).Scan(&variantID))
	defer func() {
		db.Exec(`DELETE FROM inventory_balances WHERE company_id = $1`, companyID)
		db.Exec(`DELETE FROM items_variants WHERE company_id = $1`, companyID)
		db.Exec(`DELETE FROM items WHERE company_id = $1`, companyID)
		db.Exec(`DELETE FROM taxes WHERE company_id = $1`, companyID)
	}()

	_, err := db.Exec(
		`INSERT INTO inventory_balances (company_id, variant_id, warehouse_id, quantity, updated_at)
		 VALUES ($1, $2, $3, 7, now())`, companyID, variantID, warehouseID)
	is.NoErr(err)

	tx1, err := db.Begin()
	is.NoErr(err)
	defer tx1.Rollback()
	tx2, err := db.Begin()
	is.NoErr(err)
	defer tx2.Rollback()

	qty, err := lockedBalance(tx1, companyID, variantID, warehouseID)
	is.NoErr(err)
	is.EqualFloat(qty, 7)

	done := make(chan error, 1)
	go func() {
		_, err := lockedBalance(tx2, companyID, variantID, warehouseID)
		done <- err
	}()

	select {
	case err := <-done:
		t.Fatalf("second reader was not blocked (lock missing): err=%v", err)
	case <-time.After(200 * time.Millisecond):
	}

	is.NoErr(tx1.Commit())
	select {
	case err := <-done:
		is.NoErr(err)
	case <-time.After(3 * time.Second):
		t.Fatal("second reader never unblocked after commit")
	}
	is.NoErr(tx2.Commit())
}

// TestLockedBalance_MissingRowIsZero: a variant never stocked in that warehouse has
// no row, which is zero on hand rather than an error.
func TestLockedBalance_MissingRowIsZero(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	tx, err := s.db.Begin()
	is.NoErr(err)
	defer tx.Rollback()

	qty, err := lockedBalance(tx, f.company.ID, 999999, f.warehouseID)
	is.NoErr(err)
	is.EqualFloat(qty, 0)
}
