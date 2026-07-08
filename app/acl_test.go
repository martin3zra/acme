package app

import (
	"sync"
	"testing"
)

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

// TestCanConcurrent exercises Can() from many goroutines at once — including an
// unknown role, which previously triggered a cache write. Run under `-race` it
// fails if the permission cache is mutated concurrently.
func TestCanConcurrent(t *testing.T) {
	roles := []string{"owner", "admin", "supervisor", "standard", "ghost"}
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(role string) {
			defer wg.Done()
			_ = Can(&AuthUser{Role: role}, "view:invoice")
		}(roles[i%len(roles)])
	}
	wg.Wait()
}
