package app

import "testing"

func TestCan(t *testing.T) {
	tests := []struct {
		name         string
		role         string
		actionModule string
		want         bool
	}{
		{"owner full wildcard", "owner", "create:invoice", true},
		{"admin allowed", "admin", "create:invoice", true},
		{"admin action-wide via viewAny", "admin", "viewAny:attribute", true},
		{"standard denied module", "standard", "create:vendor", false},
		{"unknown role", "ghost", "view:invoice", false},
		// Regression: a malformed permission string (no ":") must not panic.
		{"malformed no colon, non-owner", "admin", "invoice", false},
		{"malformed no colon, owner has *", "owner", "invoice", true},
		{"bare star", "owner", "*", true},
		{"empty string, non-owner", "standard", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Can(&AuthUser{Role: tt.role}, tt.actionModule); got != tt.want {
				t.Fatalf("Can(%q, %q) = %v, want %v", tt.role, tt.actionModule, got, tt.want)
			}
		})
	}
}
