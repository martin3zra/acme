package app

import (
	"strings"
	"testing"
)

// TestFlowSearchExpandsVariantRows: a variant item search returns one row per
// variant, each carrying its own variant_id, composed name, and variant price.
func TestFlowSearchExpandsVariantRows(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, variantIDs := mkVariantItem(t, f, 100)

	var name string
	is.NoErr(s.db.QueryRow(`SELECT name FROM items WHERE id = $1`, itemID).Scan(&name))

	rows, err := f.s.findItemsByCriteria(f.ctx, name)
	is.NoErr(err)
	is.Equal(len(rows), 2)

	seen := map[int]bool{}
	for _, r := range rows {
		is.Equal(r.ID, itemID)
		seen[r.VariantID] = true
		is.EqualFloat(r.Price, 100)
		if !strings.Contains(r.Name, "—") {
			t.Fatalf("expected composed variant name, got %q", r.Name)
		}
	}
	is.True(seen[variantIDs[0]], "first variant present")
	is.True(seen[variantIDs[1]], "second variant present")
}

// TestFlowSearchPlainItemUsesDefaultVariant: a plain item resolves to a single
// row on its default variant, name unchanged.
func TestFlowSearchPlainItemUsesDefaultVariant(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, defaultVariant := mkItem(t, f, 100, 60)

	var name string
	is.NoErr(s.db.QueryRow(`SELECT name FROM items WHERE id = $1`, itemID).Scan(&name))

	rows, err := f.s.findItemsByCriteria(f.ctx, name)
	is.NoErr(err)
	is.Equal(len(rows), 1)
	is.Equal(rows[0].VariantID, defaultVariant)
	is.Equal(rows[0].Name, name)
}
