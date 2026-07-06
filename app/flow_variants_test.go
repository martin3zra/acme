package app

import "testing"

// TestFlowCreateProductVariants exercises the ported variant-matrix service
// against main's items_variants schema: creating variants from attribute
// combinations flags the item, links product attributes, inserts non-default
// variants (alongside the item's existing default), maps variant attribute
// values, and coalesces a nil combo cost_price to 0 (NOT NULL column).
func TestFlowCreateProductVariants(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, _ := mkItem(t, f, 100, 60) // item + its default variant

	is.NoErr(f.s.storeAttribute(f.ctx, &StoreAttributeForm{Name: "Color", Type: "select", DisplayName: "Color"}))
	attrs, err := f.s.findAttributes(f.ctx)
	is.NoErr(err)
	colorID := attrs[0].ID
	is.NoErr(f.s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{AttributeID: colorID, Value: "Red", DisplayName: "Red"}))
	is.NoErr(f.s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{AttributeID: colorID, Value: "Blue", DisplayName: "Blue"}))
	vals, err := f.s.findAttributeValuesByAttribute(f.ctx, colorID)
	is.NoErr(err)
	is.Equal(len(vals), 2)

	price := 120.0
	form := &StoreItemWithAttributesForm{
		Name:         "Shirt",
		AttributeIDs: []int{colorID},
		VariantCombos: []VariantCombo{
			{AttributeValueIDs: map[int]int{colorID: vals[0].ID}, Price: &price}, // cost_price nil -> 0
			{AttributeValueIDs: map[int]int{colorID: vals[1].ID}},                // price nil -> 0
		},
	}

	svc := NewCreateProductWithVariantsService(s.db)
	is.NoErr(svc.Execute(f.ctx, form, itemID))

	// Item flagged; attribute linked.
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items WHERE id = $1 AND has_variants = true`, itemID), 1)
	assertRow(t, s.db, "product_attributes", map[string]any{"item_id": itemID, "attribute_id": colorID})

	// Two new non-default variants (the item's default is untouched).
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_variants WHERE item_id = $1 AND is_default = false`, itemID), 2)
	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM items_variants WHERE item_id = $1 AND is_default = true`, itemID), 1)

	// Attribute-value mappings, one per combo.
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM variant_attribute_values WHERE company_id = $1 AND attribute_id = $2`, f.company.ID, colorID), 2)

	// The priced combo kept its price; its nil cost_price coalesced to 0.
	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT price FROM items_variants WHERE item_id = $1 AND price = 120`, itemID), 120)
	is.EqualFloat(scalarFloat(t, s.db,
		`SELECT cost_price FROM items_variants WHERE item_id = $1 AND price = 120`, itemID), 0)
}
