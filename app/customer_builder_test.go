package app

import "testing"

// customerBuilder is a fluent, faker-backed factory for customers. It builds a
// StoreCustomerForm with realistic defaults and runs it through the real
// storeCustomer path, so tests still exercise production logic (validation,
// sequences, tenant scoping) rather than raw INSERTs.
//
//	id, _ := newCustomer(t, f, g).Credit("net30").CreditLimit(1000).Build()
type customerBuilder struct {
	t    *testing.T
	f    *fixture
	form StoreCustomerForm
}

// newCustomer starts a builder for a cash ("pia"), business customer with
// faker-generated name/contact/phone and a guaranteed-unique email.
func newCustomer(t *testing.T, f *fixture, g fakeGen) *customerBuilder {
	return &customerBuilder{
		t: t,
		f: f,
		form: StoreCustomerForm{
			Name:          g.Company(),
			Contact:       g.Name(),
			Email:         uniq("cust") + "@test.local",
			Phone:         g.Phone(),
			PaymentMethod: "cash",
			PaymentTerms:  "pia",
			CustomerType:  "business",
			TaxReceipt:    f.taxReceiptID,
		},
	}
}

// Credit switches the customer to credit terms (e.g. "net30").
func (b *customerBuilder) Credit(terms string) *customerBuilder {
	b.form.PaymentTerms = terms
	return b
}

// CreditLimit enforces a credit limit (implies a credit-limited customer).
func (b *customerBuilder) CreditLimit(limit float64) *customerBuilder {
	b.form.CreditLimited = true
	b.form.CreditLimit = limit
	return b
}

// Individual makes the customer an individual person rather than a business.
func (b *customerBuilder) Individual(g fakeGen) *customerBuilder {
	b.form.CustomerType = "individual"
	b.form.Name = g.Name()
	return b
}

// Build persists the customer via storeCustomer and returns its id + uuid.
func (b *customerBuilder) Build() (id int, uuid string) {
	b.t.Helper()
	if err := b.f.s.storeCustomer(b.f.ctx, &b.form); err != nil {
		b.t.Fatalf("storeCustomer: %v", err)
	}
	if err := b.f.s.db.QueryRow(
		`SELECT id, uuid FROM customers WHERE company_id = $1 AND email = $2`,
		b.f.company.ID, b.form.Email,
	).Scan(&id, &uuid); err != nil {
		b.t.Fatalf("load customer: %v", err)
	}
	return id, uuid
}
