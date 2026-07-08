package app

import (
	"testing"
	"time"
)

// invoiceBuilder is a fluent, faker-backed factory for documents
// (invoice/estimate/order). It assembles a StoreInvoiceForm and runs it through
// the real storeInvoice path — so sequences, receivables and inventory
// movements all fire exactly as in production. The document status is not set
// directly: Compute() derives it from Kind + Terms (a cash invoice posts as
// Closed/Paid, a credit invoice as Sent/UnPaid).
//
//	uuid := newInvoice(t, f, g).ForCustomer(id).Credit("net30").
//	    WithItem(itemID).Build()
type invoiceBuilder struct {
	t          *testing.T
	f          *fixture
	g          fakeGen
	customerID int
	kind       TransactionKind
	terms      string
	notes      string
	src        *TransactionSource
	lines      []*Line
}

// newInvoice starts a builder for a cash ("pia") invoice with no lines yet.
func newInvoice(t *testing.T, f *fixture, g fakeGen) *invoiceBuilder {
	return &invoiceBuilder{
		t:     t,
		f:     f,
		g:     g,
		kind:  TransactionKinds.Invoice,
		terms: "pia",
	}
}

// ForCustomer targets an existing customer id.
func (b *invoiceBuilder) ForCustomer(id int) *invoiceBuilder {
	b.customerID = id
	return b
}

// Cash makes the document a cash ("pia") sale (posts as Closed/Paid).
func (b *invoiceBuilder) Cash() *invoiceBuilder {
	b.terms = "pia"
	return b
}

// Credit makes the document a credit sale with the given terms (e.g. "net30";
// posts as Sent/UnPaid and registers a receivable).
func (b *invoiceBuilder) Credit(terms string) *invoiceBuilder {
	b.terms = terms
	return b
}

// Kind overrides the document kind (invoice/estimate/order/template).
func (b *invoiceBuilder) Kind(k TransactionKind) *invoiceBuilder {
	b.kind = k
	return b
}

// FromSource links the document to a source (e.g. converting an estimate).
func (b *invoiceBuilder) FromSource(src *TransactionSource) *invoiceBuilder {
	b.src = src
	return b
}

// Notes attaches a faker-generated note.
func (b *invoiceBuilder) Notes() *invoiceBuilder {
	b.notes = b.g.Sentence()
	return b
}

// WithLine adds an explicit line for an item.
func (b *invoiceBuilder) WithLine(itemID, qty int, price, rate float64) *invoiceBuilder {
	b.lines = append(b.lines, mkLine(itemID, b.f.unitID, b.f.warehouseID, qty, price, rate))
	return b
}

// WithVariantLine adds a line that names an explicit variant of the item.
func (b *invoiceBuilder) WithVariantLine(itemID, variantID, qty int, price, rate float64) *invoiceBuilder {
	l := mkLine(itemID, b.f.unitID, b.f.warehouseID, qty, price, rate)
	l.VariantID = variantID
	b.lines = append(b.lines, l)
	return b
}

// WithItem adds a line for an item with a faker-generated quantity (1-5) and
// price, taxed at the fixture's default rate (18%).
func (b *invoiceBuilder) WithItem(itemID int) *invoiceBuilder {
	return b.WithLine(itemID, b.g.IntRange(1, 5), round(b.g.Price(), 2), 18)
}

// Build persists the document via storeInvoice and returns its uuid.
func (b *invoiceBuilder) Build() string {
	b.t.Helper()
	form := &StoreInvoiceForm{
		CustomerID: b.customerID,
		Date:       time.Now(),
		Terms:      b.terms,
		TaxReceipt: b.f.taxReceiptID,
		Discount:   Discount{Type: "percentage"},
		Notes:      b.notes,
		Lines:      b.lines,
		Kind:       b.kind,
		Source:     b.src,
	}
	form.Compute() // populate protected fields (HTTP layer normally does this)
	if b.kind == TransactionKinds.Invoice && b.terms == "pia" {
		form.Payment.Cash.Amount = form.total
	}
	uuid, err := b.f.s.storeInvoice(b.f.ctx, form)
	if err != nil {
		b.t.Fatalf("storeInvoice(%s,%s): %v", b.kind, b.terms, err)
	}
	return uuid
}
