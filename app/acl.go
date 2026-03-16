package app

import (
	"strings"

	"github.com/martin3zra/acme/pkg/foundation"
)

var rolePermissionsCache = map[string]map[string]bool{}

var groupedPermissions = map[string]map[string][]string{
	"owner": {"*": {"*"}},
	"admin": {
		"view":    {"dashboard", "invoice", "estimate", "order", "customer", "item", "payment", "reports", "setting", "vendor"},
		"viewAny": {"dashboard", "invoice", "estimate", "order", "customer", "item", "payment", "reports", "setting", "vendor"},
		"create":  {"dashboard", "invoice", "estimate", "order", "customer", "item", "payment", "reports", "setting", "vendor"},
		"delete":  {"dashboard", "invoice", "estimate", "order", "customer", "item", "payment", "reports", "setting", "vendor"},
		"update":  {"dashboard", "invoice", "estimate", "order", "customer", "item", "payment", "reports", "setting", "company:sequence", "vendor"},
	},
	"supervisor": {
		"view":    {"dashboard", "customer", "item", "payment", "reports"},
		"viewAny": {"dashboard", "invoice", "estimate", "order", "customer", "item", "payment", "reports"},
		"create":  {"dashboard", "invoice", "estimate", "order", "customer", "item", "payment", "reports"},
		"delete":  {"dashboard", "invoice", "estimate", "order", "customer", "item", "payment", "reports"},
		"update":  {"dashboard", "invoice", "estimate", "order", "customer", "item", "payment", "reports"},
	},
	"standard": {
		"view":   {"invoice", "customer"},
		"create": {"invoice"},
	},
}

func permissions(role string) map[string]bool {
	if cached, exists := rolePermissionsCache[role]; exists {
		return cached
	}

	flatPermissions := make(map[string]bool)

	if rolePermissions, exists := groupedPermissions[role]; exists {
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
	}

	rolePermissionsCache[role] = flatPermissions
	return flatPermissions
}

func Can(user *foundation.User, actionModule string) bool {
	permissions := permissions(user.Role)

	// If the user requests "*" (full access check), return true if full access exists
	if actionModule == "*" {
		return permissions["*"]
	}

	// Standard permission checks
	if permissions[actionModule] {
		return true
	}

	// Action-wide wildcard (e.g., "view:*")
	action := actionModule[:strings.Index(actionModule, ":")]
	if permissions[action+":*"] {
		return true
	}

	// Module-wide wildcard (e.g., "*:invoice")
	module := actionModule[strings.Index(actionModule, ":")+1:]
	if permissions["*:"+module] {
		return true
	}

	// Check for complete wildcard "*:*"
	return permissions["*"]
}
