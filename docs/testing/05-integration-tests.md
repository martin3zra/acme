# Integration (Flow) Tests

The dominant test style — the `flow_*_test.go` files. Each drives real
repository/handler code against the migrated `acme_test` schema inside a
rolled-back transaction, and asserts both the returned values and the persisted
rows.

## Anatomy of a flow test

```go
func TestUpdateExpense(t *testing.T) {
	s := newTestServer(t)          // txdb server + plan (02-fixtures)
	is := newIs(t)                 // assertions
	f := mkAccountCompany(t, s)    // provisioned tenant (03-test-data)

	rent := mkExpenseCategory(t, f, "Rent")
	fuel := mkExpenseCategory(t, f, "Fuel")
	uuid := mkExpense(t, f, rent, 100, day(-1))

	// Act through the production path.
	is.NoErr(f.s.updateExpense(f.ctx, uuid, &StoreExpenseForm{
		Category: fuel, Date: day(-1), Amount: 250, Notes: "revised",
	}))

	// Assert via the repo …
	e, err := f.s.findExpenseByUUID(f.ctx, uuid)
	is.NoErr(err)
	is.Equal(e.Category.Name, "Fuel")

	// … and/or against the raw table.
	is.True(scalarString(t, s.db,
		`SELECT updated_at::text FROM expenses WHERE uuid = $1`, uuid) != before,
		"updateExpense should bump updated_at")
}
```

The pattern is always: **arrange** a tenant + data with builders/factories,
**act** by calling the real repository or handler, **assert** on the return value
and on the database.

## Handler-level flow tests

For tests that must exercise the HTTP layer (validation, flash, redirects), use
`newHandlerServer` + `handlerCtx` ([02-fixtures.md](02-fixtures.md)):

```go
s := newHandlerServer(t)
ctx, sess, rec := handlerCtx(t, s, f, "POST", "/customers", body)
s.storeCustomerHandler(ctx)
is.Equal(rec.Code, http.StatusSeeOther)   // or assert sessionErrors(sess)
```

## Naming & organisation

- One file per domain flow: `flow_invoice_create_test.go`,
  `flow_payment_test.go`, `flow_purchase_writes_test.go`,
  `flow_opening_balance_test.go`, `flow_void_test.go`, …
- Files suffixed `_playsql_test.go` were added when a repository was converted to
  playsql and pin the converted read/write behaviour
  (e.g. `flow_catalog_playsql_test.go`).
- Test names describe the scenario and the invariant:
  `TestFindExpenses_IncludeDeleted`, `TestDeleteExpense_StampsBothColumns`.

## What flow tests are especially good at here

- **Tenant isolation** — every read is company-scoped; tests assert a second
  tenant can't see the first's rows.
- **Ledger correctness** — invoice → receivable → payment → balance, and the
  void/reversal paths.
- **Sequencing** — NCF / tax-receipt and document code sequences.
- **Soft-delete semantics** — `deleted_at` scoping via the model's `softdelete`
  tag, and reads that deliberately include trashed rows.
- **Stock movement** — inventory quantity/balance changes on sale/purchase.

## Running

```
go test ./app/                          # all flows
go test ./app/ -run TestFlowInvoice     # a subset by name
go test -race ./app/                    # what CI runs
```

Flow tests are transaction-isolated and safe to run in parallel; the suite takes
~20s without `-race`, ~2–3 min with it.

## Related files

- `app/flow_*_test.go` — the flow suite (~60 files)
- `app/harness_test.go`, `app/factories_test.go` — the scaffolding they build on
