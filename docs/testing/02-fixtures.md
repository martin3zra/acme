# Fixtures & Harness

The per-test scaffolding lives in `app/harness_test.go`. It gives each test an
isolated server, a transaction that rolls back, and a small assertion vocabulary.

## Isolation: one transaction per test

`newTestServer(t)` opens a fresh `txdb` connection and returns a `*Server` whose
`db` is that connection:

```go
func newTestServer(t *testing.T) *Server {
	name := fmt.Sprintf("tx_%d", atomic.AddInt64(&txCounter, 1))
	db, _ := sql.Open("txdb", name)
	t.Cleanup(func() { _ = db.Close() })   // Close = ROLLBACK
	pdb, _ := playsql.Use(db, "postgres")  // mirrors Boot's configurePlan
	return &Server{db: db, plan: pdb}
}
```

Because every `txdb` handle is a single transaction, a test's writes are visible
to itself but invisible to other tests, and vanish on `Close`. No truncation, no
shared-state bleed. The `plan` field is set here the same way `Boot` sets it
(via `playsql.Use`), so `s.play()` returns a live executor in tests.

## Two server flavours

- **`newTestServer(t)`** — repository/flow tests. Only `db` + `plan` wired.
- **`newHandlerServer(t)`** — HTTP-handler tests. Adds a translator (flash
  messages) and a session manager (`BackWith`/`Flash`/`Errors`).

## Driving a handler

`handlerCtx(t, s, f, method, path, body)` assembles a `*routing.Context` with a
JSON body and the tenant/auth/db context values the repos and validator read,
plus a fresh session. It returns the context, the session (to assert flashes and
errors), and an `httptest.ResponseRecorder`.

`authCtx(s, company, accountID, user)` builds the same context values for direct
repository calls (no HTTP) — this is what the `fixture` uses (see
[03-test-data.md](03-test-data.md)).

Both put `database.ConnectionKey{} = s.db`, `CompanyKey{}`, `AccountKey{}`, and
the authenticated user into the context — exactly what production middleware sets.

## Assertions (`is`-style, stdlib only)

A tiny matcher inspired by matryer/is:

```go
is := newIs(t)
is.NoErr(err)
is.Equal(got, want)          // reflect.DeepEqual
is.EqualFloat(got, want)     // money/qty within epsilon
is.True(cond, "message")
is.Err(err, "expected an error")
```

## Database assertions

Port of gest's `AssertDatabaseHas/Missing`, for checking rows directly:

```go
assertRow(t, db, "customers", map[string]any{"email": e})
assertNoRow(t, db, "invoices", map[string]any{"id": id})
assertCount(t, db, "invoices_items", map[string]any{"invoice_id": id}, 3)

scalarInt(t, db, "SELECT count(*) FROM …", args...)
scalarFloat(t, db, "SELECT amount_due FROM …", args...)
scalarString(t, db, "SELECT uuid::text FROM …", args...)
```

These read the raw table, bypassing playsql — useful to assert exactly what was
persisted (e.g. a DB-generated `uuid`, a stamped `updated_at`).

## Related files

- `app/harness_test.go` — `newTestServer`, `newHandlerServer`, `handlerCtx`,
  `authCtx`, `is`, `assertRow`/`assertCount`, `scalar*`
- `app/main_test.go` — `txdb` registration
- `app/playsql_models.go` — `(*Server).play`, `configurePlan`
