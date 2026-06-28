//go:build integration

// Integration test harness. Builds a throwaway Postgres database from the Camel
// schema baseline (db/migrations) so repository code runs against the real
// schema — enum types, constraints, triggers and all.
//
// Run with:
//   set -a; . ./.env; set +a
//   go test -tags integration ./app/ -run Integration -v
//
// Requires the acme Postgres (per .env DB_*) to be reachable; the admin
// connection uses the default `postgres` database to create/drop test DBs.
package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/martin3zra/camel"
	"github.com/martin3zra/forge/i18n"

	_ "github.com/lib/pq"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func dsnFor(dbname string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		envOr("DB_USERNAME", "postgres"), os.Getenv("DB_PASSWORD"),
		envOr("DB_HOST", "localhost"), envOr("DB_PORT", "5433"),
		dbname, envOr("DB_SSLMODE", "disable"))
}

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(file)) // app/ -> repo root
}

// newTestDB creates a uniquely-named database, applies the Camel baseline to
// it, and returns a connection plus a cleanup that drops the database.
func newTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	admin, err := sql.Open("postgres", dsnFor("postgres"))
	if err != nil {
		t.Fatalf("open admin: %v", err)
	}
	if err := admin.Ping(); err != nil {
		admin.Close()
		t.Skipf("acme Postgres unreachable (%v) — is the DB up?", err)
	}

	name := fmt.Sprintf("acme_it_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		admin.Close()
		t.Fatalf("create database: %v", err)
	}

	cfg := camel.DefaultConfig()
	cfg.DB.Driver = "postgres"
	cfg.DB.Source = dsnFor(name)
	cfg.Migration.Directory = "db/migrations"
	cfg.Migration.Pattern = "*.yaml"

	runner, err := camel.NewRunner(cfg, repoRoot())
	if err != nil {
		t.Fatalf("camel runner: %v", err)
	}
	if err := runner.Migrate(false); err != nil {
		runner.Close()
		t.Fatalf("camel migrate: %v", err)
	}
	runner.Close()

	db, err := sql.Open("postgres", dsnFor(name))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	cleanup := func() {
		db.Close()
		// Terminate stragglers so DROP DATABASE succeeds.
		admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1`, name)
		if _, err := admin.Exec("DROP DATABASE IF EXISTS " + name); err != nil {
			t.Logf("drop database %s: %v", name, err)
		}
		admin.Close()
	}
	return db, cleanup
}

func testServer(db *sql.DB) *Server {
	return &Server{
		db:         db,
		translator: i18n.NewTranslator(localesFS, defaultLang, fallbackLang, trans("global")),
	}
}

func companyCtx(companyID int) context.Context {
	return context.WithValue(context.Background(), CompanyKey{}, &Company{ID: companyID})
}

// inventoryFixtures holds the seeded ids used by inventory tests.
type inventoryFixtures struct {
	CompanyID int
	ItemID    int
	VariantID int
	WHFrom    int
	WHTo      int
}

// seedInventory inserts the minimal rows needed to exercise inventory movement
// and transfer logic: an account + company, a tracked item variant, and two
// warehouses.
func seedInventory(t *testing.T, db *sql.DB) inventoryFixtures {
	t.Helper()
	var f inventoryFixtures

	var userID int
	must(t, db.QueryRow(
		`INSERT INTO users (name, email, password)
		 VALUES ('Tester', 'tester@example.com', 'x') RETURNING id`).Scan(&userID))

	var accountID int
	must(t, db.QueryRow(
		`INSERT INTO accounts (owner_id) VALUES ($1) RETURNING id`, userID).Scan(&accountID))

	must(t, db.QueryRow(
		`INSERT INTO companies (name, city, address, account_id)
		 VALUES ('Test Co', 'SD', 'Calle 1', $1) RETURNING id`, accountID).Scan(&f.CompanyID))

	var taxID int
	must(t, db.QueryRow(
		`INSERT INTO taxes (company_id, name) VALUES ($1, 'ITBIS') RETURNING id`,
		f.CompanyID).Scan(&taxID))

	must(t, db.QueryRow(
		`INSERT INTO items (company_id, name, description, tax_id)
		 VALUES ($1, 'Widget', 'A widget', $2) RETURNING id`, f.CompanyID, taxID).Scan(&f.ItemID))

	// track_inventory defaults to true.
	must(t, db.QueryRow(
		`INSERT INTO items_variants (company_id, item_id, name)
		 VALUES ($1, $2, 'Default') RETURNING id`, f.CompanyID, f.ItemID).Scan(&f.VariantID))

	must(t, db.QueryRow(
		`INSERT INTO warehouses (company_id, name) VALUES ($1, 'Main') RETURNING id`,
		f.CompanyID).Scan(&f.WHFrom))
	must(t, db.QueryRow(
		`INSERT INTO warehouses (company_id, name) VALUES ($1, 'Branch') RETURNING id`,
		f.CompanyID).Scan(&f.WHTo))

	return f
}

func balanceQty(t *testing.T, db *sql.DB, companyID, variantID, warehouseID int) (float64, bool) {
	t.Helper()
	var qty float64
	err := db.QueryRow(
		`SELECT quantity FROM inventory_balances
		  WHERE company_id=$1 AND variant_id=$2 AND warehouse_id=$3`,
		companyID, variantID, warehouseID).Scan(&qty)
	if err == sql.ErrNoRows {
		return 0, false
	}
	must(t, err)
	return qty, true
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
}

func exec(db *sql.DB, q string, args ...any) error {
	_, err := db.Exec(q, args...)
	return err
}
