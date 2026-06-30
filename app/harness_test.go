package app

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/martin3zra/forge/auth"
	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/store"
)

var txCounter int64

// newTestServer returns a *Server wired to a fresh txdb connection (its own
// transaction). The connection is closed on cleanup, rolling back everything the
// test wrote — no truncation needed, full isolation.
func newTestServer(t *testing.T) *Server {
	t.Helper()
	name := fmt.Sprintf("tx_%d", atomic.AddInt64(&txCounter, 1))
	db, err := sql.Open("txdb", name)
	if err != nil {
		t.Fatalf("open txdb: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	qs, err := store.NewQueryStore(sqlQueriesFS, "sql/")
	if err != nil {
		t.Fatalf("query store: %v", err)
	}
	return &Server{db: db, qs: qs}
}

// authCtx builds a request-equivalent context: tenant company, account and the
// authenticated user, plus the DB handle — exactly what the repositories read.
func authCtx(s *Server, company *Company, accountID int, user *AuthUser) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, database.ConnectionKey{}, s.db)
	ctx = context.WithValue(ctx, CompanyKey{}, company)
	ctx = context.WithValue(ctx, AccountKey{}, map[string]any{
		"id":   accountID,
		"uuid": "",
	})
	ctx = context.WithValue(ctx, auth.ContextUserID{}, map[string]any{
		"id":    user.Id,
		"uuid":  user.UUID,
		"email": user.Email,
		"role":  user.Role,
	})
	return ctx
}

// ── is-style assertions (inspired by matryer/is; stdlib only) ──────────────────

type is struct{ t *testing.T }

func newIs(t *testing.T) *is { return &is{t: t} }

func (i *is) NoErr(err error) {
	i.t.Helper()
	if err != nil {
		i.t.Fatalf("unexpected error: %v", err)
	}
}

func (i *is) Err(err error, msg string) {
	i.t.Helper()
	if err == nil {
		i.t.Fatalf("expected an error: %s", msg)
	}
}

func (i *is) True(cond bool, msg string) {
	i.t.Helper()
	if !cond {
		i.t.Fatalf("expected true: %s", msg)
	}
}

func (i *is) Equal(got, want any) {
	i.t.Helper()
	if !reflect.DeepEqual(got, want) {
		i.t.Fatalf("got %v (%T), want %v (%T)", got, got, want, want)
	}
}

// EqualFloat compares money/quantity values within a small epsilon.
func (i *is) EqualFloat(got, want float64) {
	i.t.Helper()
	if math.Abs(got-want) > 0.001 {
		i.t.Fatalf("got %.4f, want %.4f", got, want)
	}
}

// ── DB assertion helpers (port of gest's AssertDatabaseHas/Missing) ────────────

func whereClause(where map[string]any) (string, []any) {
	if len(where) == 0 {
		return "", nil
	}
	cols := make([]string, 0, len(where))
	args := make([]any, 0, len(where))
	i := 1
	for col, val := range where {
		cols = append(cols, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}
	return " WHERE " + strings.Join(cols, " AND "), args
}

func countRows(t *testing.T, db *sql.DB, table string, where map[string]any) int {
	t.Helper()
	clause, args := whereClause(where)
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM "+table+clause, args...).Scan(&n); err != nil {
		t.Fatalf("countRows %s: %v", table, err)
	}
	return n
}

func assertRow(t *testing.T, db *sql.DB, table string, where map[string]any) {
	t.Helper()
	if countRows(t, db, table, where) == 0 {
		t.Fatalf("expected a row in %s matching %v, found none", table, where)
	}
}

func assertNoRow(t *testing.T, db *sql.DB, table string, where map[string]any) {
	t.Helper()
	if n := countRows(t, db, table, where); n != 0 {
		t.Fatalf("expected no row in %s matching %v, found %d", table, where, n)
	}
}

func assertCount(t *testing.T, db *sql.DB, table string, where map[string]any, want int) {
	t.Helper()
	if got := countRows(t, db, table, where); got != want {
		t.Fatalf("expected %d rows in %s matching %v, got %d", want, table, where, got)
	}
}

func scalarFloat(t *testing.T, db *sql.DB, query string, args ...any) float64 {
	t.Helper()
	var v float64
	if err := db.QueryRow(query, args...).Scan(&v); err != nil {
		t.Fatalf("scalarFloat: %v", err)
	}
	return v
}

func scalarString(t *testing.T, db *sql.DB, query string, args ...any) string {
	t.Helper()
	var v string
	if err := db.QueryRow(query, args...).Scan(&v); err != nil {
		t.Fatalf("scalarString: %v", err)
	}
	return v
}
