# Test Data: fixtures, builders, factories, faker

Test data is created through **production code paths**, not raw INSERTs, so
building a customer or invoice also exercises validation, sequences, and tenant
scoping. Four layers cooperate.

## 1. The tenant fixture — `mkAccountCompany`

`mkAccountCompany(t, s)` (`app/factories_test.go`) provisions a complete tenant
and returns a `fixture`:

```go
type fixture struct {
	s            *Server
	ctx          context.Context   // authCtx wired for repo calls
	company      *Company
	accountID    int
	user         *AuthUser
	unitID, taxID, warehouseID, taxReceiptID int  // catalog prerequisites
}
```

It creates an enabled/verified user + account, then builds the company through
the **real** `storeCompany` path (so `companies_users`, shared-data copy, and
sequences all run), and backfills the catalog rows most flows need (a unit, an
18% tax, a warehouse, a fiscal tax receipt) plus the purchase/vendor-payment
sequence keys. Almost every flow test starts:

```go
s := newTestServer(t)
f := mkAccountCompany(t, s)
```

## 2. Builders — fluent, faker-backed, production-path

One per major aggregate (`customer_builder_test.go`, `invoice_builder_test.go`,
`vendor_builder_test.go`, `purchase_builder_test.go`). They assemble a real
`Store…Form`, run it through the actual repository, and return ids/uuids:

```go
custID, _ := newCustomer(t, f, g).Credit("net30").CreditLimit(1000).Build()

uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
	WithLine(itemID, 2, 100, 18).Build()
```

`Build()` calls `storeCustomer` / `storeInvoice` etc. — so the builder path
covers validation, NCF/sequence assignment, and stock movement, exactly as a
user request would.

## 3. Factories — catalog helpers

Free functions in `app/factories_test.go` for the supporting cast:

- `mkItem(t, f, price, cost)` → item + its default inventory-tracked variant.
- `mkVariantItem(t, f, price)` → a `has_variants` product with a Color attribute
  and two variants (for variant-matrix tests).
- `mkLine(itemID, unitID, warehouseID, qty, price, rate)` → an invoice/order line.
- `mkExpenseCategory`, `mkExpense`, and similar per-domain helpers live beside
  their flow tests.

## 4. Faker — deterministic realism

`newFaker(t)` (`app/faker_test.go`) returns a generator seeded from the **test
name**, so each test gets a reproducible stream (a failure replays with the same
values) while different tests draw different data. Builders take a `fakeGen`
interface (`Name`, `Company`, `Phone`, `Email`, `Price`, …).

**Uniqueness rule:** faker supplies realism where it is free; columns the schema
requires unique (e.g. customer email) are suffixed with `uniq("prefix")` (a
process-global atomic counter, `app/factories_test.go`) so a repeated faker draw
can never trip a `UNIQUE` constraint. Use `uniq()` for anything the DB enforces
unique; use faker for everything else.

## Putting it together

```go
func TestSomething(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Build()
	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithLine(itemID, 2, 100, 18).Build()

	// assert against repos and/or the raw tables …
}
```

## Related files

- `app/factories_test.go` — `fixture`, `mkAccountCompany`, `mkItem`,
  `mkVariantItem`, `mkLine`, `uniq`
- `app/customer_builder_test.go`, `app/invoice_builder_test.go`,
  `app/vendor_builder_test.go`, `app/purchase_builder_test.go`
- `app/faker_test.go` — `newFaker`, `fakeGen`
