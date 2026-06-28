//go:build integration

package app

import (
	"database/sql"
	"encoding/csv"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// seedImport inserts an imports row (required by processRows for progress
// updates and row-issue FKs) and returns its id.
func seedImport(t *testing.T, db *sql.DB, source string) string {
	t.Helper()
	id := uuid.New().String()
	_, err := db.Exec(
		`INSERT INTO imports (id, upload_id, user_id, source, status)
		 VALUES ($1, $2, $3, $4, 'queued')`,
		id, uuid.New().String(), uuid.New().String(), source)
	must(t, err)
	return id
}

func tableCount(t *testing.T, db *sql.DB, table string, companyID int) int {
	t.Helper()
	var n int
	must(t, db.QueryRow(
		"SELECT count(*) FROM "+table+" WHERE company_id=$1", companyID).Scan(&n))
	return n
}

// Proves the import-dispatch fix: source=vendors rows are stored in `vendors`,
// not `customers`, with the CSV code honored.
func TestIntegration_ImportVendors_LandInVendorsNotCustomers(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db) // provides a company
	srv := testServer(db)
	importID := seedImport(t, db, "vendors")

	// Headers must match the server-side vendor column map; payment_terms
	// values are valid enum buckets (net30 / net0).
	content := strings.Join([]string{
		"NOMBRE,CORREO,TELEFONO,PAGO,CONDICIONES,TIEMPO_ENTREGA,CODIGO,TIPO",
		"Distribuidora Andina,ventas@andina.com,8095550123,ck,30,7,PRV-001,Individual",
		"Suministros Lopez,ana@lopez.com,8295559876,bt,0,3,PRV-002,Negocio",
	}, "\n")

	reader := csv.NewReader(strings.NewReader(content))
	headers, err := reader.Read()
	must(t, err)
	columnMap, err := mapHeaders(headers, "vendors")
	must(t, err)

	if err := srv.processRows(f.CompanyID, importID, "vendors", reader, columnMap, 2); err != nil {
		t.Fatalf("processRows: %v", err)
	}

	if got := tableCount(t, db, "vendors", f.CompanyID); got != 2 {
		t.Errorf("vendors: want 2, got %d", got)
	}
	if got := tableCount(t, db, "customers", f.CompanyID); got != 0 {
		t.Errorf("customers: want 0 (vendors must not leak into customers), got %d", got)
	}

	// CSV codes are honored.
	var codes []string
	rows, err := db.Query(`SELECT code FROM vendors WHERE company_id=$1 ORDER BY code`, f.CompanyID)
	must(t, err)
	defer rows.Close()
	for rows.Next() {
		var c string
		must(t, rows.Scan(&c))
		codes = append(codes, c)
	}
	if len(codes) != 2 || codes[0] != "PRV-001" || codes[1] != "PRV-002" {
		t.Errorf("vendor codes: want [PRV-001 PRV-002], got %v", codes)
	}
}

// A blank-but-present required column (name) is rejected per-row and recorded
// as an import issue, without aborting the whole import.
func TestIntegration_ImportVendors_BadRowRecordsIssue(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)
	importID := seedImport(t, db, "vendors")

	content := strings.Join([]string{
		"NOMBRE,CORREO,PAGO,CONDICIONES,CODIGO,TIPO",
		",ventas@andina.com,ck,30,PRV-001,Individual", // missing name → bad row
		"Suministros Lopez,ana@lopez.com,bt,0,PRV-002,Negocio",
	}, "\n")

	reader := csv.NewReader(strings.NewReader(content))
	headers, _ := reader.Read()
	columnMap, _ := mapHeaders(headers, "vendors")

	if err := srv.processRows(f.CompanyID, importID, "vendors", reader, columnMap, 2); err != nil {
		t.Fatalf("processRows: %v", err)
	}

	if got := tableCount(t, db, "vendors", f.CompanyID); got != 1 {
		t.Errorf("vendors: want 1 valid row imported, got %d", got)
	}

	var issues int
	must(t, db.QueryRow(
		`SELECT count(*) FROM import_row_issues WHERE import_id=$1`, importID).Scan(&issues))
	if issues == 0 {
		t.Error("expected at least one import row issue for the bad row")
	}
}
