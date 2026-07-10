package app

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

// Attribute + attribute-value reads and writes converted to playsql.

// TestFindAttributesWithValues_EagerLoads: this used to run one query per attribute.
// With("Values") loads them in a single second query, still ordered by sort_order
// then value, still excluding soft-deleted values.
func TestFindAttributesWithValues_EagerLoads(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	is.NoErr(s.storeAttribute(f.ctx, &StoreAttributeForm{Name: "Color", Type: "select", DisplayName: "Color"}))
	is.NoErr(s.storeAttribute(f.ctx, &StoreAttributeForm{Name: "Size", Type: "select", DisplayName: "Size"}))
	is.NoErr(s.storeAttribute(f.ctx, &StoreAttributeForm{Name: "Empty", Type: "text", DisplayName: "Empty"}))

	attrs, err := s.findAttributes(f.ctx)
	is.NoErr(err)
	is.Equal(len(attrs), 3)

	byName := map[string]int{}
	for _, a := range attrs {
		byName[a.Name] = a.ID
	}

	// Values deliberately inserted out of order; sort_order then value decides.
	is.NoErr(s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{
		AttributeID: byName["Color"], Value: "Red", DisplayName: "Red", SortOrder: 2}))
	is.NoErr(s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{
		AttributeID: byName["Color"], Value: "Blue", DisplayName: "Blue", SortOrder: 1}))
	is.NoErr(s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{
		AttributeID: byName["Size"], Value: "M", DisplayName: "M", SortOrder: 1}))

	// A soft-deleted value must not come back.
	is.NoErr(s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{
		AttributeID: byName["Size"], Value: "XL", DisplayName: "XL", SortOrder: 9}))
	xlUUID := scalarString(t, s.db,
		`SELECT uuid::text FROM attribute_values WHERE attribute_id = $1 AND value = 'XL'`, byName["Size"])
	is.NoErr(s.deleteAttributeValue(f.ctx, xlUUID))

	withValues, err := s.findAttributesWithValues(f.ctx)
	is.NoErr(err)
	is.Equal(len(withValues), 3)

	// Attributes come back ordered by name: Color, Empty, Size.
	is.Equal(withValues[0].Name, "Color")
	is.Equal(withValues[1].Name, "Empty")
	is.Equal(withValues[2].Name, "Size")

	// Color's values are ordered by sort_order, not insertion order.
	is.Equal(len(withValues[0].Values), 2)
	is.Equal(withValues[0].Values[0].Value, "Blue")
	is.Equal(withValues[0].Values[1].Value, "Red")

	// The attribute with no values carries none.
	is.Equal(len(withValues[1].Values), 0)

	// Size has only its live value; XL was soft-deleted.
	is.Equal(len(withValues[2].Values), 1)
	is.Equal(withValues[2].Values[0].Value, "M")
}

// TestStoreAttributeValue_RevivesSoftDeleted: the ON CONFLICT branch exists to bring
// a soft-deleted value back. deleted_at is written explicitly as nil so
// EXCLUDED.deleted_at clears it.
func TestStoreAttributeValue_RevivesSoftDeleted(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	is.NoErr(s.storeAttribute(f.ctx, &StoreAttributeForm{Name: "Color", Type: "select", DisplayName: "Color"}))
	attrs, err := s.findAttributes(f.ctx)
	is.NoErr(err)
	attrID := attrs[0].ID

	is.NoErr(s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{
		AttributeID: attrID, Value: "Blue", DisplayName: "Blue", SortOrder: 1}))
	uuid := scalarString(t, s.db, `SELECT uuid::text FROM attribute_values WHERE attribute_id = $1`, attrID)

	is.NoErr(s.deleteAttributeValue(f.ctx, uuid))
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM attribute_values WHERE uuid = $1 AND deleted_at IS NOT NULL`, uuid), 1)

	// Storing the same value again revives the row and refreshes its columns.
	is.NoErr(s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{
		AttributeID: attrID, Value: "Blue", DisplayName: "Navy Blue", SortOrder: 7}))

	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM attribute_values WHERE attribute_id = $1`, attrID), 1)
	is.Equal(scalarInt(t, s.db,
		`SELECT count(*) FROM attribute_values WHERE uuid = $1 AND deleted_at IS NULL`, uuid), 1)
	is.Equal(scalarString(t, s.db, `SELECT display_name FROM attribute_values WHERE uuid = $1`, uuid), "Navy Blue")
	is.Equal(scalarInt(t, s.db, `SELECT sort_order FROM attribute_values WHERE uuid = $1`, uuid), 7)

	values, err := s.findAttributeValuesByAttribute(f.ctx, attrID)
	is.NoErr(err)
	is.Equal(len(values), 1)
}

// TestDeleteAttributeValue_SecondDeleteIsNotFound pins a deliberate narrowing: the
// raw statement had no deleted_at predicate, so deleting twice reported success
// both times. attributeValueRead's softdelete tag makes the second a not-found.
func TestDeleteAttributeValue_SecondDeleteIsNotFound(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	is.NoErr(s.storeAttribute(f.ctx, &StoreAttributeForm{Name: "Color", Type: "select", DisplayName: "Color"}))
	attrs, err := s.findAttributes(f.ctx)
	is.NoErr(err)

	is.NoErr(s.storeAttributeValue(f.ctx, &StoreAttributeValueForm{
		AttributeID: attrs[0].ID, Value: "Blue", DisplayName: "Blue"}))
	uuid := scalarString(t, s.db, `SELECT uuid::text FROM attribute_values WHERE attribute_id = $1`, attrs[0].ID)

	is.NoErr(s.deleteAttributeValue(f.ctx, uuid))
	err = s.deleteAttributeValue(f.ctx, uuid)
	is.True(errors.Is(err, ErrRecordNotFound), "deleting an already-deleted value is not-found")
}

// TestFindAttribute_NotFoundMessages: the two by-column reads keep their error text.
func TestFindAttribute_NotFoundMessages(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	_, err := s.findAttributeByID(f.ctx, "00000000-0000-0000-0000-000000000000")
	is.Err(err, "unknown uuid")
	is.Equal(err.Error(), "attribute not found")

	_, err = s.findAttributeByIntID(f.ctx, 99999999)
	is.Err(err, "unknown id")
	is.Equal(err.Error(), "attribute not found")

	_, err = s.findAttributeValueByUUID(f.ctx, "00000000-0000-0000-0000-000000000000")
	is.Err(err, "unknown value uuid")
	is.Equal(err.Error(), "attribute value not found")
}

// TestQueryRowErrLeaksConnection demonstrates the defect storeAttribute carried.
//
// `QueryRowContext(...).Err()` never calls Scan, and only Scan closes the underlying
// *Rows — Row.Err() just returns the stored error. The connection therefore stays
// checked out. With a one-connection pool the next query blocks until the context
// expires. Calling Scan, or not asking for RETURNING at all, releases it.
//
// The converted storeAttribute does not request RETURNING, so it cannot regress this.
func TestQueryRowErrLeaksConnection(t *testing.T) {
	is := newIs(t)

	open := func() *sql.DB {
		db, err := sql.Open("postgres", testDSN)
		is.NoErr(err)
		db.SetMaxOpenConns(1)
		return db
	}

	// Err() without Scan: the row is never closed, so the single connection is held.
	leaky := open()
	defer leaky.Close()
	is.NoErr(leaky.QueryRow(`SELECT 1`).Err())

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	var n int
	err := leaky.QueryRowContext(ctx, `SELECT 2`).Scan(&n)
	is.True(errors.Is(err, context.DeadlineExceeded),
		"a row left unscanned holds its connection; got: "+errString(err))

	// Scan releases it, so the pool keeps working.
	sane := open()
	defer sane.Close()
	is.NoErr(sane.QueryRow(`SELECT 1`).Scan(&n))
	is.NoErr(sane.QueryRow(`SELECT 2`).Scan(&n))
	is.Equal(n, 2)
}

func errString(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}
