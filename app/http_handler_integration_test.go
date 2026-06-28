//go:build integration

package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/martin3zra/forge/auth"
	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/routing"
	"github.com/martin3zra/forge/session"
)

// testSessionManager mints fresh sessions for cookieless requests without
// touching a store. The far-future GC interval ensures the background ticker
// never fires (and so never dereferences the nil store) during a test run.
var testSessionManager = session.NewSessionManager(
	nil, time.Hour, time.Hour, time.Hour, "acme_test_session", "", false, true)

// transferRequest builds a routing.Context for POST /inventories/transfers with
// the request context the middleware would normally populate: a session (read
// by ParseRequest), the DB connection (for `exists` validation), the
// authenticated user (role drives ACL), and the current company.
func transferRequest(db *sql.DB, companyID int, role, body string) (*routing.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/inventories/transfers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Attach a fresh session, then layer the auth/company/db context values.
	_, req = testSessionManager.Start(req)
	ctx := req.Context()
	ctx = context.WithValue(ctx, database.ConnectionKey{}, db)
	ctx = context.WithValue(ctx, auth.ContextUserID{}, map[string]any{"id": 1, "role": role})
	ctx = context.WithValue(ctx, CompanyKey{}, &Company{ID: companyID})
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	return &routing.Context{Response: rec, Request: req, Params: map[string]string{}}, rec
}

func movementCount(t *testing.T, db *sql.DB, companyID int) int {
	t.Helper()
	var n int
	must(t, db.QueryRow(
		`SELECT count(*) FROM inventory_movements WHERE company_id=$1`, companyID).Scan(&n))
	return n
}

// A valid admin request transfers stock end-to-end through the handler:
// binding, validation, ACL, and storeTransfer.
func TestIntegration_TransferHandler_HappyPath(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)

	// Seed 20 units at the source.
	seedTx, _ := db.Begin()
	must(t, srv.recordMovement(seedTx, f.CompanyID, f.VariantID, f.WHFrom, 0,
		20, 0, InventoryMovementKinds.Adjustment, "seed", 0))
	must(t, seedTx.Commit())

	body := fmt.Sprintf(
		`{"variant_id":%d,"from_warehouse_id":%d,"to_warehouse_id":%d,"qty":8}`,
		f.VariantID, f.WHFrom, f.WHTo)
	ctx, rec := transferRequest(db, f.CompanyID, "admin", body)

	srv.storeTransferHandler()(ctx)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status: want 201, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	if src, _ := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHFrom); src != 12 {
		t.Errorf("source balance: want 12, got %v", src)
	}
	if dst, _ := balanceQty(t, db, f.CompanyID, f.VariantID, f.WHTo); dst != 8 {
		t.Errorf("destination balance: want 8, got %v", dst)
	}
}

// A valid admin request that references a non-existent warehouse fails the
// `exists` validation at the HTTP boundary (422) and persists nothing.
func TestIntegration_TransferHandler_RejectsUnknownWarehouse(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)

	body := fmt.Sprintf(
		`{"variant_id":%d,"from_warehouse_id":%d,"to_warehouse_id":999999,"qty":5}`,
		f.VariantID, f.WHFrom)
	ctx, rec := transferRequest(db, f.CompanyID, "admin", body)

	srv.storeTransferHandler()(ctx)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status: want 422, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	if n := movementCount(t, db, f.CompanyID); n != 0 {
		t.Errorf("no movements should be recorded on validation failure, got %d", n)
	}
}

// A role without create:transfer is rejected by Authorize before the handler
// body runs, and nothing is persisted.
func TestIntegration_TransferHandler_RejectsUnauthorizedRole(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()
	f := seedInventory(t, db)
	srv := testServer(db)

	body := fmt.Sprintf(
		`{"variant_id":%d,"from_warehouse_id":%d,"to_warehouse_id":%d,"qty":5}`,
		f.VariantID, f.WHFrom, f.WHTo)
	// "standard" role has no create:transfer ability.
	ctx, rec := transferRequest(db, f.CompanyID, "standard", body)

	srv.storeTransferHandler()(ctx)

	if rec.Code < 400 {
		t.Errorf("status: want a client error for unauthorized role, got %d", rec.Code)
	}
	if n := movementCount(t, db, f.CompanyID); n != 0 {
		t.Errorf("no movements should be recorded for an unauthorized request, got %d", n)
	}
}
