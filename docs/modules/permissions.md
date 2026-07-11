# Permissions

Who can do what. Access is role-based: every request checks the user's role
against a permission of the form `action:module`.

## For users

### Roles

A user's role (per company) decides their access:

- **Owner** — full access to everything.
- **Admin** — broad access across sales, purchases, inventory, payments,
  reports, and settings; can confirm purchases.
- **Supervisor** — day-to-day operations (create/view/update across most
  modules) but a narrower delete/settings reach than admin.
- **Standard** — limited: view and create invoices, view customers.

An unknown or unassigned role gets **no** access (every check denies).

### What a permission looks like

Permissions read as **`action:module`** — e.g. `create:invoice`,
`viewAny:customer`, `confirm:purchase`, `delete:vendor`. Modules line up with the
areas in this documentation set (invoice, estimate, order, purchase, customer,
vendor, item, inventory, payment, payable, reports, setting, attribute, …).

## For developers

### Where it's defined

`app/acl.go`. `groupedPermissions` maps each role to the actions and modules it's
allowed:

```go
var groupedPermissions = map[string]map[string][]string{
	"owner": {"*": {"*"}},                       // full wildcard
	"admin": {
		"viewAny": {"invoice", "purchase", ...},
		"create":  {"invoice", "adjustment", "transfer", ...},
		"confirm": {"purchase"},
		...
	},
	"supervisor": { ... },
	"standard":   {"view": {"invoice", "customer"}, "create": {"invoice"}},
}
```

At package init, `buildRolePermissions()` flattens this into a per-role
`map[string]bool` of `"action:module"` keys (plus wildcard keys), cached in
`rolePermissionsCache`. The cache is **built once and never mutated**, so
concurrent request reads are race-free (`TestCanConcurrent` pins this under
`-race`).

### The check

```go
func Can(user foundation.Authenticatable, actionModule string) bool
```

`Can` resolves the user's role set and checks, in order: the exact
`action:module`, then `action:*` (action-wide), then `*:module` (module-wide),
then `*` (full). A malformed string with no `:` can't be split and falls through
to the full-access check rather than panicking (`strings.Cut`, guarded — see the
regression cases in `acl_test.go`).

### Where it's enforced

- **Routes** (`app/route.go`) via `.Can("action:module")`, e.g.
  `route.GET("/invoices", …).Can("viewAny:invoice")`.
- **Handlers** call `Can` directly where a route serves multiple kinds or needs
  finer control.

### Adding or changing access

Edit `groupedPermissions` in `app/acl.go` (add the module to a role's action
list). No migration or cache wiring needed — the flatten runs at init. Add a
case to `acl_test.go`.

### Tests

- `app/acl_test.go` — `TestCan` (role/wildcard/malformed resolution) and
  `TestCanConcurrent` (the cache under `-race`). This is a pure unit test — see
  [../testing/04-unit-tests.md](../testing/04-unit-tests.md).

## Related

Every module doc lists the permissions it uses in its "For developers" section.
