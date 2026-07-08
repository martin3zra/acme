package app

import (
	"strings"

	"github.com/martin3zra/forge/foundation"
)

// rolePermissionsCache holds the flattened permission set for every known role.
// It is built once at package initialization from groupedPermissions (whose set
// of roles is static) and never written to afterwards, so the concurrent reads
// from request goroutines in Can() need no locking.
var rolePermissionsCache = buildRolePermissions()

var groupedPermissions = map[string]map[string][]string{
	"owner": {"*": {"*"}},
	"admin": {
		"view":    {"dashboard", "invoice", "estimate", "order", "purchase", "customer", "item", "vendor", "inventory", "payment", "payable", "reports", "setting"},
		"viewAny": {"dashboard", "invoice", "estimate", "order", "purchase", "customer", "item", "vendor", "inventory", "movement", "payment", "payable", "reports", "setting", "attribute"},
		"create":  {"dashboard", "invoice", "estimate", "order", "purchase", "customer", "item", "vendor", "inventory", "adjustment", "transfer", "payment", "payable", "reports", "setting", "attribute"},
		"delete":  {"dashboard", "invoice", "estimate", "order", "purchase", "customer", "item", "vendor", "inventory", "payment", "payable", "reports", "setting", "attribute"},
		"update":  {"dashboard", "invoice", "estimate", "order", "purchase", "customer", "item", "vendor", "inventory", "payment", "payable", "reports", "setting", "company:sequence", "attribute"},
		"confirm": {"purchase"},
	},
	"supervisor": {
		"view":    {"dashboard", "customer", "item", "inventory", "payment", "reports"},
		"viewAny": {"dashboard", "invoice", "estimate", "order", "purchase", "customer", "item", "inventory", "movement", "payment", "payable", "reports", "attribute"},
		"create":  {"dashboard", "invoice", "estimate", "order", "purchase", "customer", "item", "inventory", "adjustment", "transfer", "payment", "payable", "reports", "attribute"},
		"delete":  {"dashboard", "invoice", "estimate", "order", "purchase", "customer", "item", "inventory", "payment", "reports", "attribute"},
		"update":  {"dashboard", "invoice", "estimate", "order", "purchase", "customer", "item", "inventory", "payment", "reports", "attribute"},
		"confirm": {"purchase"},
	},
	"standard": {
		"view":   {"invoice", "customer"},
		"create": {"invoice"},
	},
}

// buildRolePermissions flattens every role in groupedPermissions once. Called
// only during package initialization, so it does not race with request reads.
func buildRolePermissions() map[string]map[string]bool {
	cache := make(map[string]map[string]bool, len(groupedPermissions))
	for role, rolePermissions := range groupedPermissions {
		flatPermissions := make(map[string]bool)
		for action, modules := range rolePermissions {
			for _, module := range modules {
				flatPermissions[action+":"+module] = true

				// If role has full module access ("view:*"), create general key
				if module == "*" {
					flatPermissions[action+":*"] = true
				}

				// If role has full action access ("*:invoice"), create wildcard key
				if action == "*" {
					flatPermissions["*:"+module] = true
				}
			}
		}

		// If role has full access ("*:*"), create a general wildcard key
		if _, exists := rolePermissions["*"]; exists {
			flatPermissions["*"] = true
		}

		cache[role] = flatPermissions
	}
	return cache
}

// permissions returns the precomputed permission set for a role. Unknown roles
// return nil; reading a nil map yields false, so Can() denies them. Read-only —
// the cache is never mutated after init.
func permissions(role string) map[string]bool {
	return rolePermissionsCache[role]
}

func Can(user foundation.Authenticatable, actionModule string) bool {
	permissions := permissions(user.GetRole())

	// If the user requests "*" (full access check), return true if full access exists
	if actionModule == "*" {
		return permissions["*"]
	}

	// Standard permission checks
	if permissions[actionModule] {
		return true
	}

	// Wildcard checks need an "action:module" shape. A string without a colon
	// can't be split, so skip straight to the complete-wildcard check rather
	// than slicing on a missing separator (which would panic).
	action, module, ok := strings.Cut(actionModule, ":")
	if !ok {
		return permissions["*"]
	}

	// Action-wide wildcard (e.g., "view:*")
	if permissions[action+":*"] {
		return true
	}

	// Module-wide wildcard (e.g., "*:invoice")
	if permissions["*:"+module] {
		return true
	}

	// Check for complete wildcard "*:*"
	return permissions["*"]
}
