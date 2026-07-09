package app

import (
	"errors"
	"testing"
)

// Every guarded write is scoped by company_id, so a zero-row result means the id or
// uuid names a record that does not exist or belongs to another tenant. Both used to
// return nil, and the handler above flashed success having changed nothing.
//
// These tests drive each guarded path twice: once with an id from another company,
// once with an id that exists nowhere.

func TestWriteGuard_UnknownAndForeignIDs(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	other := mkAccountCompany(t, s)
	g := newFaker(t)

	// Records that exist, but for the other tenant.
	foreignCustomer, _ := newCustomer(t, other, g).Build()
	foreignVendor, _ := newVendor(t, other, g).Build()
	foreignItem, _ := mkItem(t, other, 100, 60)
	foreignWarehouse := other.warehouseID

	const unknownID = 99999999
	const unknownUUID = "00000000-0000-0000-0000-000000000000"

	cases := []struct {
		name string
		call func() error
	}{
		{"deleteCustomer/foreign", func() error { return s.deleteCustomer(f.ctx, foreignCustomer) }},
		{"deleteCustomer/unknown", func() error { return s.deleteCustomer(f.ctx, unknownID) }},
		// Fully-populated form: the enum columns reject an empty string at bind time,
		// which would abort the shared txdb transaction before the guard is reached.
		{"updateCustomer/foreign", func() error {
			return s.updateCustomer(f.ctx, foreignCustomer, &UpdateCustomerForm{
				Name: "x", CustomerType: "business", PaymentMethod: "ck",
				PaymentTerms: "net30", TaxReceipt: f.taxReceiptID,
			})
		}},
		{"deleteVendor/foreign", func() error { return s.deleteVendor(f.ctx, foreignVendor) }},
		{"deleteVendor/unknown", func() error { return s.deleteVendor(f.ctx, unknownID) }},
		{"deleteItem/foreign", func() error { return s.deleteItem(f.ctx, foreignItem) }},
		{"deleteItem/unknown", func() error { return s.deleteItem(f.ctx, unknownID) }},
		{"updateItem/foreign", func() error {
			return s.updateItem(f.ctx, foreignItem, &UpdateItemForm{
				Name: "x", Description: "d", Price: 1, TaxID: f.taxID, ItemType: "product", UnitID: f.unitID,
			})
		}},
		{"deleteWarehouse/foreign", func() error { return s.deleteWarehouse(f.ctx, foreignWarehouse) }},
		{"deleteWarehouse/unknown", func() error { return s.deleteWarehouse(f.ctx, unknownID) }},
		{"updateWarehouse/unknown", func() error {
			return s.updateWarehouse(f.ctx, unknownID, &UpdateWarehouseForm{Name: "x"})
		}},
		{"updateTax/unknown", func() error {
			return s.updateTax(f.ctx, unknownUUID, &StoreTaxForm{Name: "x", Rate: 1})
		}},
		{"updateUnit/unknown", func() error {
			return s.updateUnit(f.ctx, unknownID, &StoreUnitForm{Name: "x", BaseQty: 1})
		}},
		{"updateProfile/unknown-account", func() error {
			return s.updateProfile(unknownUUID, &StoreProfileForm{Name: "x", Email: "x@y.z"})
		}},
	}

	for _, tc := range cases {
		err := tc.call()
		if !errors.Is(err, ErrRecordNotFound) {
			t.Errorf("%s: got %v, want ErrRecordNotFound", tc.name, err)
		}
	}

	// Nothing belonging to the other tenant was touched.
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM customers WHERE id = $1 AND deleted_at IS NULL`, foreignCustomer), 1)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM vendors WHERE id = $1 AND deleted_at IS NULL`, foreignVendor), 1)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM items WHERE id = $1 AND deleted_at IS NULL AND name <> 'x'`, foreignItem), 1)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM warehouses WHERE id = $1 AND deleted_at IS NULL`, foreignWarehouse), 1)
}

// TestWriteGuard_HappyPathsUnaffected: a legitimate id still succeeds.
func TestWriteGuard_HappyPathsUnaffected(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	custID, _ := newCustomer(t, f, g).Build()
	is.NoErr(s.deleteCustomer(f.ctx, custID))
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM customers WHERE id = $1 AND deleted_at IS NOT NULL`, custID), 1)

	vendID, _ := newVendor(t, f, g).Build()
	is.NoErr(s.deleteVendor(f.ctx, vendID))

	itemID, _ := mkItem(t, f, 100, 60)
	is.NoErr(s.updateItem(f.ctx, itemID, &UpdateItemForm{
		Name: "Renamed", Description: "d", Price: 100, TaxID: f.taxID, ItemType: "product", UnitID: f.unitID,
	}))
	is.Equal(scalarString(t, s.db, `SELECT name FROM items WHERE id = $1`, itemID), "Renamed")

	is.NoErr(s.updateUnit(f.ctx, f.unitID, &StoreUnitForm{Name: "Caja", BaseQty: 12}))
	is.Equal(scalarString(t, s.db, `SELECT name FROM units WHERE id = $1`, f.unitID), "Caja")
}

// TestWriteGuard_SoftDeletedAttributeValue: the attribute writes carry
// `AND deleted_at IS NULL`, so editing an already-deleted value matches no row. That
// is a not-found, not a success.
func TestWriteGuard_SoftDeletedAttributeValue(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	is.NoErr(s.storeAttribute(f.ctx, &StoreAttributeForm{
		Name: "Color", Type: "select", DisplayName: "Color",
	}))
	attrs, err := s.findAttributes(f.ctx)
	is.NoErr(err)
	attrID := attrs[0].ID

	is.NoErr(s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{
		AttributeID: attrID, Value: "Blue", DisplayName: "Blue",
	}))
	valueUUID := scalarString(t, s.db,
		`SELECT uuid::text FROM attribute_values WHERE attribute_id = $1`, attrID)

	is.NoErr(s.deleteAttributeValue(f.ctx, valueUUID))

	// The update carries `AND deleted_at IS NULL`, so a soft-deleted value matches
	// no row. That is a not-found, not a success.
	err = s.updateAttributeValue(f.ctx, valueUUID, &StoreAttributeValueForm{
		AttributeID: attrID, Value: "Navy", DisplayName: "Navy",
	})
	is.True(errors.Is(err, ErrRecordNotFound), "editing a soft-deleted value is not-found")

	// An unknown uuid likewise.
	err = s.updateAttributeValue(f.ctx, "00000000-0000-0000-0000-000000000000",
		&StoreAttributeValueForm{AttributeID: attrID, Value: "Teal", DisplayName: "Teal"})
	is.True(errors.Is(err, ErrRecordNotFound), "an unknown value uuid is not-found")
}

// TestWriteGuard_UpdatePendingEmail: matches nothing once pending_email is cleared,
// so the caller cannot report an address as verified that never changed.
func TestWriteGuard_UpdatePendingEmail(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	pending := "new@test.local"
	_, err := s.db.Exec(`UPDATE users SET pending_email = $2 WHERE id = $1`, f.user.Id, pending)
	is.NoErr(err)

	user, err := s.findUserByUUID(f.user.UUID)
	is.NoErr(err)
	is.NoErr(s.updatePendingEmail(user)) // promotes pending_email -> email
	is.Equal(scalarString(t, s.db, `SELECT email FROM users WHERE id = $1`, f.user.Id), pending)

	// Replaying the same promotion now matches nothing.
	err = s.updatePendingEmail(user)
	is.True(errors.Is(err, ErrRecordNotFound), "a replayed promotion must not report success")
}
