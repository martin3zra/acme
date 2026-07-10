package app

import (
	"testing"
)

// CreateProductWithVariantsService converted to playsql: all seven raw statements.
// generateVariantName also stopped depending on Go's map iteration order.

// mkAttr creates an attribute and returns its id.
func mkAttr(t *testing.T, f *fixture, name string) int {
	t.Helper()
	if err := f.s.storeAttribute(f.ctx, &StoreAttributeForm{
		Name: name, Type: "select", DisplayName: name}); err != nil {
		t.Fatalf("storeAttribute(%s): %v", name, err)
	}
	attrs, err := f.s.findAttributes(f.ctx)
	if err != nil {
		t.Fatalf("findAttributes: %v", err)
	}
	for _, a := range attrs {
		if a.Name == name {
			return a.ID
		}
	}
	t.Fatalf("attribute %s not found", name)
	return 0
}

// mkAttrValue creates a value under an attribute and returns its id.
func mkAttrValue(t *testing.T, f *fixture, attrID int, value string) int {
	t.Helper()
	if err := f.s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{
		AttributeID: attrID, Value: value, DisplayName: value}); err != nil {
		t.Fatalf("storeAttributeValue(%s): %v", value, err)
	}
	return scalarInt(t, f.s.db,
		`SELECT id FROM attribute_values WHERE attribute_id = $1 AND value = $2`, attrID, value)
}

func variantNames(t *testing.T, f *fixture, itemID int) []string {
	t.Helper()
	rows, err := f.s.db.Query(
		`SELECT name FROM items_variants WHERE item_id = $1 ORDER BY id`, itemID)
	if err != nil {
		t.Fatalf("variantNames: %v", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, n)
	}
	return out
}

// itemIDByName resolves the item just created by storeItemWithVariants.
func itemIDByName(t *testing.T, f *fixture, name string) int {
	t.Helper()
	return scalarInt(t, f.s.db,
		`SELECT id FROM items WHERE company_id = $1 AND name = $2`, f.company.ID, name)
}

// ─── variant naming ───────────────────────────────────────────────────────────

// TestGenerateVariantName_OrderFollowsTheForm pins the bug fix.
//
// generateVariantName used to range over the combo's map[int]int directly, and Go
// randomises map iteration order, so a two-attribute combination was named "Red Large"
// on one insert and "Large Red" on the next — nondeterministically, for the same input.
// The order now comes from form.AttributeIDs.
//
// The loop matters: with a two-key map, a single run has a coin-flip chance of landing
// on the right order by accident. Twenty items make that vanishingly unlikely.
func TestGenerateVariantName_OrderFollowsTheForm(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	color := mkAttr(t, f, uniq("Color"))
	size := mkAttr(t, f, uniq("Size"))
	red := mkAttrValue(t, f, color, "Red")
	large := mkAttrValue(t, f, size, "Large")

	for i := 0; i < 20; i++ {
		name := uniq("Shirt")
		form := &StoreItemWithAttributesForm{
			Name: name, Price: 100, TaxID: f.taxID, UnitID: f.unitID, ItemType: "product",
			AttributeIDs: []int{color, size},
			VariantCombos: []VariantCombo{
				{AttributeValueIDs: map[int]int{color: red, size: large}},
			},
		}
		is.NoErr(s.storeItemWithVariants(f.ctx, form))

		names := variantNames(t, f, itemIDByName(t, f, name))
		is.Equal(len(names), 1)
		is.Equal(names[0], "Red Large")
	}
}

// TestGenerateVariantName_ReversedFormOrderReversesTheName: the form's attribute order
// is the name's order, so reversing it reverses the name. This is what proves the name
// tracks the form rather than, say, the attribute ids sorting ascending.
func TestGenerateVariantName_ReversedFormOrderReversesTheName(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	color := mkAttr(t, f, uniq("Color"))
	size := mkAttr(t, f, uniq("Size"))
	red := mkAttrValue(t, f, color, "Red")
	large := mkAttrValue(t, f, size, "Large")

	mk := func(order []int) string {
		name := uniq("Shirt")
		form := &StoreItemWithAttributesForm{
			Name: name, Price: 100, TaxID: f.taxID, UnitID: f.unitID, ItemType: "product",
			AttributeIDs: order,
			VariantCombos: []VariantCombo{
				{AttributeValueIDs: map[int]int{color: red, size: large}},
			},
		}
		is.NoErr(s.storeItemWithVariants(f.ctx, form))
		names := variantNames(t, f, itemIDByName(t, f, name))
		is.Equal(len(names), 1)
		return names[0]
	}

	is.Equal(mk([]int{color, size}), "Red Large")
	is.Equal(mk([]int{size, color}), "Large Red")
}

// TestGenerateVariantName_MismatchedAttributeContributesNothing: the old per-value
// query matched on (id, attribute_id, company_id), so a value id paired with the wrong
// attribute yielded no row and no name fragment. The batched read keys on the same
// pair.
func TestGenerateVariantName_MismatchedAttributeContributesNothing(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	color := mkAttr(t, f, uniq("Color"))
	size := mkAttr(t, f, uniq("Size"))
	red := mkAttrValue(t, f, color, "Red")
	large := mkAttrValue(t, f, size, "Large")

	// `size` is mapped to `red`, a value belonging to `color`.
	name := uniq("Shirt")
	form := &StoreItemWithAttributesForm{
		Name: name, Price: 100, TaxID: f.taxID, UnitID: f.unitID, ItemType: "product",
		AttributeIDs: []int{color, size},
		VariantCombos: []VariantCombo{
			{AttributeValueIDs: map[int]int{color: red, size: red}},
		},
	}
	is.NoErr(s.storeItemWithVariants(f.ctx, form))

	names := variantNames(t, f, itemIDByName(t, f, name))
	is.Equal(len(names), 1)
	is.Equal(names[0], "Red") // the mismatched pair drops out
	_ = large
}

// TestGenerateVariantName_SoftDeletedValueStillNames: the per-value lookup never
// filtered deleted_at, so WithTrashed preserves it. A combo referencing a retired value
// still produces its name rather than silently dropping a word.
func TestGenerateVariantName_SoftDeletedValueStillNames(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	color := mkAttr(t, f, uniq("Color"))
	red := mkAttrValue(t, f, color, "Red")

	_, err := s.db.Exec(`UPDATE attribute_values SET deleted_at = now() WHERE id = $1`, red)
	is.NoErr(err)

	name := uniq("Shirt")
	form := &StoreItemWithAttributesForm{
		Name: name, Price: 100, TaxID: f.taxID, UnitID: f.unitID, ItemType: "product",
		AttributeIDs: []int{color},
		VariantCombos: []VariantCombo{
			{AttributeValueIDs: map[int]int{color: red}},
		},
	}
	is.NoErr(s.storeItemWithVariants(f.ctx, form))

	names := variantNames(t, f, itemIDByName(t, f, name))
	is.Equal(len(names), 1)
	is.Equal(names[0], "Red")
}

// TestGenerateVariantName_ScopedToCompany: another tenant's attribute value contributes
// no name fragment, as the company_id predicate always ensured.
func TestGenerateVariantName_ScopedToCompany(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	first := mkAccountCompany(t, s)
	second := mkAccountCompany(t, s)

	theirColor := mkAttr(t, second, uniq("Color"))
	theirRed := mkAttrValue(t, second, theirColor, "Red")

	name := uniq("Shirt")
	form := &StoreItemWithAttributesForm{
		Name: name, Price: 100, TaxID: first.taxID, UnitID: first.unitID, ItemType: "product",
		AttributeIDs: []int{theirColor},
		VariantCombos: []VariantCombo{
			{AttributeValueIDs: map[int]int{theirColor: theirRed}},
		},
	}
	is.NoErr(s.storeItemWithVariants(first.ctx, form))

	names := variantNames(t, first, itemIDByName(t, first, name))
	is.Equal(len(names), 1)
	is.Equal(names[0], "") // no fragment survived the company scope
}

// ─── writes ───────────────────────────────────────────────────────────────────

// TestStoreItemWithVariants_SetsHasVariantsAndLinks covers the has_variants update, the
// product_attributes link, the variant insert and the variant_attribute_values links.
func TestStoreItemWithVariants_SetsHasVariantsAndLinks(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	color := mkAttr(t, f, uniq("Color"))
	size := mkAttr(t, f, uniq("Size"))
	red := mkAttrValue(t, f, color, "Red")
	blue := mkAttrValue(t, f, color, "Blue")
	large := mkAttrValue(t, f, size, "Large")

	name := uniq("Shirt")
	form := &StoreItemWithAttributesForm{
		Name: name, Price: 100, TaxID: f.taxID, UnitID: f.unitID, ItemType: "product",
		AttributeIDs: []int{color, size},
		VariantCombos: []VariantCombo{
			{AttributeValueIDs: map[int]int{color: red, size: large}},
			{AttributeValueIDs: map[int]int{color: blue, size: large}},
		},
	}
	is.NoErr(s.storeItemWithVariants(f.ctx, form))
	itemID := itemIDByName(t, f, name)

	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM items WHERE id = $1 AND has_variants = true`, itemID), 1)

	// Both attributes linked, with the form's index as sort_order.
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM product_attributes WHERE item_id = $1`, itemID), 2)
	is.Equal(scalarInt(t, s.db,
		`SELECT sort_order FROM product_attributes WHERE item_id = $1 AND attribute_id = $2`,
		itemID, color), 0)
	is.Equal(scalarInt(t, s.db,
		`SELECT sort_order FROM product_attributes WHERE item_id = $1 AND attribute_id = $2`,
		itemID, size), 1)

	// Two variants, no default, each with two attribute-value links.
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM items_variants WHERE item_id = $1`, itemID), 2)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM items_variants WHERE item_id = $1 AND is_default = true`, itemID), 0)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM variant_attribute_values vav
		   JOIN items_variants iv ON iv.id = vav.variant_id
		  WHERE iv.item_id = $1`, itemID), 4)

	// uuid comes from the column default now that the insert no longer names it.
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM items_variants WHERE item_id = $1 AND uuid IS NOT NULL`, itemID), 2)

	names := variantNames(t, f, itemID)
	is.Equal(names[0], "Red Large")
	is.Equal(names[1], "Blue Large")
}

// TestStoreItemWithVariants_NoCombosCreatesDefaultVariant.
func TestStoreItemWithVariants_NoCombosCreatesDefaultVariant(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	name := uniq("Simple")
	form := &StoreItemWithAttributesForm{
		Name: name, Price: 100, TaxID: f.taxID, UnitID: f.unitID, ItemType: "product",
	}
	is.NoErr(s.storeItemWithVariants(f.ctx, form))
	itemID := itemIDByName(t, f, name)

	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM items_variants WHERE item_id = $1 AND is_default = true`, itemID), 1)
	is.Equal(scalarString(t, s.db,
		`SELECT name FROM items_variants WHERE item_id = $1`, itemID), "Default")
	// A simple product is not flagged as having variants.
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM items WHERE id = $1 AND has_variants = false`, itemID), 1)
}

// TestStoreItemWithVariants_AttributesWithoutCombosRejected.
func TestStoreItemWithVariants_AttributesWithoutCombosRejected(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	color := mkAttr(t, f, uniq("Color"))
	form := &StoreItemWithAttributesForm{
		Name: uniq("Shirt"), Price: 100, TaxID: f.taxID, UnitID: f.unitID, ItemType: "product",
		AttributeIDs: []int{color},
	}
	err := s.storeItemWithVariants(f.ctx, form)
	is.Err(err, "attributes without combinations are meaningless")
}

// TestExecute_ForeignItemIsNotFound pins the has_variants update's company scope and
// its mustAffectRows guard.
//
// The raw statement was `UPDATE items SET has_variants = true WHERE id = $1` — no
// company_id, and the affected-row count discarded. Handed another tenant's item id it
// would have flipped that item's flag; handed an unknown id it would have reported
// success. Now both are ErrRecordNotFound, and the transaction rolls back, so no
// variants are left behind.
func TestExecute_ForeignItemIsNotFound(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	first := mkAccountCompany(t, s)
	second := mkAccountCompany(t, s)

	theirItem, _ := mkItem(t, first, 100, 60)

	color := mkAttr(t, second, uniq("Color"))
	red := mkAttrValue(t, second, color, "Red")

	svc := NewCreateProductWithVariantsService(s.db)
	form := &StoreItemWithAttributesForm{
		Name: uniq("Shirt"), Price: 100, TaxID: second.taxID, UnitID: second.unitID,
		ItemType:     "product",
		AttributeIDs: []int{color},
		VariantCombos: []VariantCombo{
			{AttributeValueIDs: map[int]int{color: red}},
		},
	}

	// second's context, first's item id.
	err := svc.Execute(second.ctx, form, theirItem)
	is.True(err != nil, "another tenant's item must not be flagged as having variants")

	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM items WHERE id = $1 AND has_variants = false`, theirItem), 1)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM product_attributes WHERE item_id = $1`, theirItem), 0)
	// The default variant mkItem created is the only one; the matrix rolled back.
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM items_variants WHERE item_id = $1`, theirItem), 1)
}

// TestAttachProductAttribute_ReattachKeepsSortOrder pins the DO NOTHING arm.
//
// Upsert fills in a default DO UPDATE list only when the update-column slice is nil; an
// empty non-nil slice compiles to DO NOTHING. With a DO UPDATE the second attach would
// overwrite sort_order.
func TestAttachProductAttribute_ReattachKeepsSortOrder(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	color := mkAttr(t, f, uniq("Color"))
	itemID, _ := mkItem(t, f, 100, 60)

	svc := NewCreateProductWithVariantsService(s.db)

	tx, err := s.db.Begin()
	is.NoErr(err)
	is.NoErr(svc.attachProductAttribute(tx, f.company.ID, itemID, color, 3))
	// Re-attach with a different sort_order: the conflict does nothing.
	is.NoErr(svc.attachProductAttribute(tx, f.company.ID, itemID, color, 9))
	is.NoErr(tx.Commit())

	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM product_attributes WHERE item_id = $1 AND attribute_id = $2`,
		itemID, color), 1)
	is.Equal(scalarInt(t, s.db,
		`SELECT sort_order FROM product_attributes WHERE item_id = $1 AND attribute_id = $2`,
		itemID, color), 3)
}
