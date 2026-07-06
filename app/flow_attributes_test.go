package app

import "testing"

// TestFlowAttributeAndValues exercises the ported attribute backend against the
// new attribute/attribute_values schema: create an attribute, reject a duplicate
// name (case-insensitive), add values, and load attributes-with-values.
func TestFlowAttributeAndValues(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	is.NoErr(f.s.storeAttribute(f.ctx, &StoreAttributeForm{
		Name: "Color", Type: "select", DisplayName: "Color",
	}))

	attrs, err := f.s.findAttributes(f.ctx)
	is.NoErr(err)
	is.Equal(len(attrs), 1)
	is.Equal(attrs[0].Name, "Color")

	// Duplicate name (case-insensitive) is rejected.
	is.Err(f.s.storeAttribute(f.ctx, &StoreAttributeForm{
		Name: "  color ", Type: "select", DisplayName: "Color",
	}), "duplicate attribute name must be rejected")

	attrID := attrs[0].ID
	is.NoErr(f.s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{AttributeID: attrID, Value: "Red", DisplayName: "Red"}))
	is.NoErr(f.s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{AttributeID: attrID, Value: "Blue", DisplayName: "Blue"}))

	vals, err := f.s.findAttributeValuesByAttribute(f.ctx, attrID)
	is.NoErr(err)
	is.Equal(len(vals), 2)

	// Duplicate value under the same attribute is rejected.
	is.Err(f.s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{AttributeID: attrID, Value: "red", DisplayName: "Red"}),
		"duplicate attribute value must be rejected")

	withVals, err := f.s.findAttributesWithValues(f.ctx)
	is.NoErr(err)
	is.Equal(len(withVals), 1)
	is.Equal(len(withVals[0].Values), 2)
}
