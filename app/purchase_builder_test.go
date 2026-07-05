package app

import (
	"testing"
	"time"
)

// purchaseBuilder is a fluent factory for purchase documents (purchase_order/
// purchase_receipt/vendor_bill) over the real storePurchase path. Unlike the
// customer/invoice builders it takes no faker generator: a purchase carries no
// free-text scalar worth faking — the vendor is already faker-built and line
// values are asserted, so they stay explicit.
//
//	uuid := newPurchase(t, f).ForVendor(vendorID).
//	    Kind(PurchaseTransactionKinds.VendorBill).WithLine(itemID, 1, 100, 18).Build()
type purchaseBuilder struct {
	t             *testing.T
	f             *fixture
	vendorID      int
	kind          PurchaseTransactionKind
	terms         string
	invoiceNumber string
	src           *PurchaseSource
	lines         []*Line
}

// newPurchase starts a net30 purchase builder (kind must be set explicitly).
func newPurchase(t *testing.T, f *fixture) *purchaseBuilder {
	return &purchaseBuilder{t: t, f: f, terms: "net30"}
}

// ForVendor targets the vendor id.
func (b *purchaseBuilder) ForVendor(id int) *purchaseBuilder { b.vendorID = id; return b }

// Kind sets the purchase kind (purchase_order/purchase_receipt/vendor_bill).
func (b *purchaseBuilder) Kind(k PurchaseTransactionKind) *purchaseBuilder { b.kind = k; return b }

// Terms overrides the payment terms.
func (b *purchaseBuilder) Terms(terms string) *purchaseBuilder { b.terms = terms; return b }

// InvoiceNumber sets the vendor's invoice/reference number explicitly. When left
// unset, Build generates a unique one for non-order kinds.
func (b *purchaseBuilder) InvoiceNumber(n string) *purchaseBuilder { b.invoiceNumber = n; return b }

// FromSource links the document to a source (e.g. a receipt from an order).
func (b *purchaseBuilder) FromSource(src *PurchaseSource) *purchaseBuilder { b.src = src; return b }

// WithLine adds a line for an item.
func (b *purchaseBuilder) WithLine(itemID, qty int, price, rate float64) *purchaseBuilder {
	b.lines = append(b.lines, mkLine(itemID, b.f.unitID, b.f.warehouseID, qty, price, rate))
	return b
}

// Build persists the purchase via storePurchase and returns its uuid.
func (b *purchaseBuilder) Build() string {
	b.t.Helper()
	invoiceNumber := b.invoiceNumber
	// Purchase orders carry no vendor invoice number; bills/receipts get a unique
	// one unless the caller set it.
	if invoiceNumber == "" && b.kind != PurchaseTransactionKinds.PurchaseOrder {
		invoiceNumber = uniq("PINV")
	}
	form := &StorePurchaseForm{
		VendorID:      b.vendorID,
		Date:          time.Now(),
		Terms:         b.terms,
		Discount:      Discount{Type: "percentage"},
		Lines:         b.lines,
		Kind:          b.kind,
		Source:        b.src,
		InvoiceNumber: invoiceNumber,
	}
	form.SetContext(b.f.ctx) // so form.User() resolves (AP created_by)
	form.Compute()
	uuid, err := b.f.s.storePurchase(b.f.ctx, form)
	if err != nil {
		b.t.Fatalf("storePurchase(%s): %v", b.kind, err)
	}
	return uuid
}
