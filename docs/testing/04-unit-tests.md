# Unit Tests

The small slice of tests that need **no database** — pure functions and
in-memory logic. They are fast and run in isolation from the integration suite.

## Running just these

```
make test-unit
```

which is:

```
go test ./app/ -run 'Test(Map|Get|Parse|Normalize|Detect|NextOccurrence|MapHeaders)'
```

The pattern targets the naming used by the DB-free tests. (Note: it is a name
filter, not a build tag — a new pure-unit test should be named to match, or the
Makefile pattern extended.)

## What lives here

| Area | File | What it pins |
|---|---|---|
| Authorization | `acl_test.go` | `Can(user, "action:module")` — role → permission resolution, wildcards, malformed strings, and a `-race` concurrency test for the permission cache |
| Config | `config_test.go` | Env parsing / config construction |
| CSV import headers | `import-headers_test.go` | Header detection and mapping (`MapHeaders`, delimiter `Detect`) |
| Mappers | `mapper-helpers_test.go`, `vendor-mapper_test.go` | Row/DTO mapping helpers |
| Recurrence | scheduler math (`NextOccurrence`) | Next-run date computation for recurring invoices |
| Logging | `logging_test.go` | `InitLogger` level/format, stdlib bridge (does not need the DB) |

## Style

Table-driven, standard library only, using the `is` helpers where handy:

```go
func TestCan(t *testing.T) {
	tests := []struct {
		name, role, actionModule string
		want                     bool
	}{
		{"owner full wildcard", "owner", "create:invoice", true},
		{"standard denied module", "standard", "create:vendor", false},
		{"malformed no colon, owner has *", "owner", "invoice", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Can(&AuthUser{Role: tt.role}, tt.actionModule); got != tt.want {
				t.Fatalf("Can(%q,%q)=%v want %v", tt.role, tt.actionModule, got, tt.want)
			}
		})
	}
}
```

## When to write a unit test vs a flow test

- **Unit** when the logic is a self-contained function with no persistence:
  a mapper, a date calculation, a permission check, a parser.
- **Flow** ([05-integration-tests.md](05-integration-tests.md)) when the
  behaviour depends on the schema, sequences, tenancy, or multiple repositories
  — which is most of this system.

## Related files

- `app/acl_test.go`, `app/config_test.go`, `app/import-headers_test.go`
- `app/mapper-helpers_test.go`, `app/vendor-mapper_test.go`, `app/logging_test.go`
- `Makefile` — `test-unit`
