package app

import "testing"

// TestFlowStoreItemWithVariants: the wired path creates the item and its variant
// matrix atomically — flagged has_variants, non-default combo variants, product
// attributes and value mappings, and crucially NO default variant.
func TestFlowStoreItemWithVariants(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	is.NoErr(f.s.storeAttribute(f.ctx, &StoreAttributeForm{Name: "Color", Type: "select", DisplayName: "Color"}))
	attrs, err := f.s.findAttributes(f.ctx)
	is.NoErr(err)
	colorID := attrs[0].ID
	is.NoErr(f.s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{AttributeID: colorID, Value: "Red", DisplayName: "Red"}))
	is.NoErr(f.s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{AttributeID: colorID, Value: "Blue", DisplayName: "Blue"}))
	vals, err := f.s.findAttributeValuesByAttribute(f.ctx, colorID)
	is.NoErr(err)

	form := &StoreItemWithAttributesForm{
		Name: "Shirt", Price: 100, TaxID: f.taxID, UnitID: f.unitID, ItemType: "product",
		AttributeIDs: []int{colorID},
		VariantCombos: []VariantCombo{
			{AttributeValueIDs: map[int]int{colorID: vals[0].ID}},
			{AttributeValueIDs: map[int]int{colorID: vals[1].ID}},
		},
	}
	is.NoErr(f.s.storeItemWithVariants(f.ctx, form))

	var itemID int
	is.NoErr(s.db.QueryRow(`SELECT id FROM items WHERE company_id = $1 AND name = 'Shirt'`, f.company.ID).Scan(&itemID))

	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items WHERE id = $1 AND has_variants = true`, itemID), 1)
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_variants WHERE item_id = $1 AND is_default = false`, itemID), 2)
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_variants WHERE item_id = $1 AND is_default = true`, itemID), 0)
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_units WHERE item_id = $1`, itemID), 1)
	assertRow(t, s.db, "product_attributes", map[string]any{"item_id": itemID, "attribute_id": colorID})
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM variant_attribute_values WHERE company_id = $1 AND attribute_id = $2`, f.company.ID, colorID), 2)
}

// TestFlowStoreItemWithVariantsSimple: with no combos the wired path behaves like
// a plain item — a single default variant, has_variants left false.
func TestFlowStoreItemWithVariantsSimple(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	form := &StoreItemWithAttributesForm{
		Name: "Mug", Price: 10, TaxID: f.taxID, UnitID: f.unitID, ItemType: "product",
	}
	is.NoErr(f.s.storeItemWithVariants(f.ctx, form))

	var itemID int
	is.NoErr(s.db.QueryRow(`SELECT id FROM items WHERE company_id = $1 AND name = 'Mug'`, f.company.ID).Scan(&itemID))

	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_variants WHERE item_id = $1 AND is_default = true`, itemID), 1)
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items WHERE id = $1 AND has_variants = false`, itemID), 1)
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_units WHERE item_id = $1`, itemID), 1)
}
