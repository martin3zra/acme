package app

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	txdb "github.com/DATA-DOG/go-txdb"
	"github.com/joho/godotenv"
)

// testDSN is the connection string for the dedicated acme_test database, built in
// TestMain. Each test opens a "txdb" connection (a transaction) against it.
var testDSN string

// TestMain bootstraps the integration test database:
//   - loads .env.test (never the dev .env)
//   - ensures the acme_test database exists
//   - runs `camel migrate` to build the schema (idempotent)
//   - registers the "txdb" driver so every test runs in a rolled-back transaction
func TestMain(m *testing.M) {
	loadTestEnv()

	dbName := getenv("DB_NAME", "acme_test")
	if dbName == "acme" {
		fmt.Println("refusing to run tests against the dev database 'acme'; set DB_NAME=acme_test in .env.test")
		os.Exit(1)
	}

	testDSN = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		getenv("DB_USERNAME", "postgres"), getenv("DB_PASSWORD", ""),
		getenv("DB_HOST", "localhost"), getenv("DB_PORT", "5432"),
		dbName, getenv("DB_SSLMODE", "disable"),
	)

	if err := ensureTestDatabase(dbName); err != nil {
		fmt.Printf("ensure test database: %v\n", err)
		os.Exit(1)
	}
	if err := runMigrations(testDSN); err != nil {
		fmt.Printf("run migrations: %v\n", err)
		os.Exit(1)
	}

	// Each opened "txdb" connection is its own transaction; closing it rolls back.
	txdb.Register("txdb", "postgres", testDSN)

	os.Exit(m.Run())
}

// loadTestEnv loads .env.test from the repo root (tests run from the app/ dir).
func loadTestEnv() {
	for _, p := range []string{".env.test", "../.env.test"} {
		if _, err := os.Stat(p); err == nil {
			_ = godotenv.Load(p)
			return
		}
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ensureTestDatabase creates the test database if it does not already exist.
func ensureTestDatabase(dbName string) error {
	admin := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/postgres?sslmode=%s",
		getenv("DB_USERNAME", "postgres"), getenv("DB_PASSWORD", ""),
		getenv("DB_HOST", "localhost"), getenv("DB_PORT", "5432"), getenv("DB_SSLMODE", "disable"),
	)
	db, err := sql.Open("postgres", admin)
	if err != nil {
		return err
	}
	defer db.Close()

	var exists bool
	if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = db.Exec("CREATE DATABASE " + pqIdent(dbName))
	return err
}

// pqIdent quotes a Postgres identifier (db name) for the non-parameterizable
// CREATE DATABASE statement.
func pqIdent(name string) string {
	return `"` + name + `"`
}

// runMigrations applies the Camel migrations to the test DB. camel reads .env
// first, so DB_SOURCE is passed explicitly to override it, and the command runs
// from the repo root where camel.yaml and db/migrations live.
func runMigrations(dsn string) error {
	camel, err := exec.LookPath("camel")
	if err != nil {
		camel = filepath.Join(os.Getenv("HOME"), "go", "bin", "camel")
	}
	cmd := exec.Command(camel, "migrate")
	cmd.Dir = ".." // repo root
	cmd.Env = append(os.Environ(), "DB_SOURCE="+dsn)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %s", err, string(out))
	}
	return nil
}
