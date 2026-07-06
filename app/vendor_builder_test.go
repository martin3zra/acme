package app

import "testing"

// vendorBuilder is a fluent, faker-backed factory for vendors, wrapping the real
// storeVendor path.
//
//	vendorID, vendorUUID := newVendor(t, f, g).Terms("net30").Build()
type vendorBuilder struct {
	t    *testing.T
	f    *fixture
	form StoreVendorForm
}

// newVendor starts a builder for a net30, business vendor with a faker name and
// a guaranteed-unique email.
func newVendor(t *testing.T, f *fixture, g fakeGen) *vendorBuilder {
	return &vendorBuilder{
		t: t,
		f: f,
		form: StoreVendorForm{
			Name:          g.Company(),
			Email:         uniq("vend") + "@test.local",
			PaymentMethod: "cash",
			PaymentTerms:  "net30",
			VendorType:    "business",
			TaxReceipt:    f.taxReceiptID,
		},
	}
}

// Terms overrides the vendor payment terms.
func (b *vendorBuilder) Terms(terms string) *vendorBuilder {
	b.form.PaymentTerms = terms
	return b
}

// Build persists the vendor via storeVendor and returns its id + uuid.
func (b *vendorBuilder) Build() (id int, uuid string) {
	b.t.Helper()
	if err := b.f.s.storeVendor(b.f.ctx, &b.form); err != nil {
		b.t.Fatalf("storeVendor: %v", err)
	}
	if err := b.f.s.db.QueryRow(
		`SELECT id, uuid FROM vendors WHERE company_id = $1 AND email = $2`,
		b.f.company.ID, b.form.Email,
	).Scan(&id, &uuid); err != nil {
		b.t.Fatalf("load vendor: %v", err)
	}
	return id, uuid
}
