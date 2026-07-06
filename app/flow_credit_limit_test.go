package app

import (
	"testing"
	"time"
)

// invoiceBody builds the JSON payload for a single-line net30 credit invoice.
func invoiceBody(f *fixture, customerID, itemID int) map[string]any {
	return map[string]any{
		"customer_id": customerID,
		"date":        time.Now().Format(time.RFC3339),
		"terms":       "net30",
		"tax_receipt": f.taxReceiptID,
		"kind":        "invoice",
		"discount":    map[string]any{"value": 0, "type": "percentage"},
		"lines": []map[string]any{{
			"id": itemID, "unit": f.unitID, "qty": 1, "price": 100,
			"rate": 18, "action": "added", "warehouse_id": f.warehouseID,
		}},
	}
}

// TestFlowCreditLimitBlocksInvoice: the handler-level guard rejects a credit
// invoice that would push the customer past their credit limit, and nothing is
// persisted.
func TestFlowCreditLimitBlocksInvoice(t *testing.T) {
	srv := newHandlerServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, srv)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Credit("net30").CreditLimit(50).Build() // limit 50, invoice total 118

	ctx, sess, _ := handlerCtx(t, srv, f, "POST", "/invoices",
		invoiceBody(f, custID, itemID))
	srv.storeInvoiceHandler()(ctx)

	errs := sessionErrors(sess)
	is.True(errs["status"] != nil, "expected a credit-limit error on the session")
	assertNoRow(t, srv.db, "invoices", map[string]any{"customer_id": custID})
}

// TestFlowCreditLimitAllowsWithinLimit: an invoice within the limit goes through.
func TestFlowCreditLimitAllowsWithinLimit(t *testing.T) {
	srv := newHandlerServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, srv)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Credit("net30").CreditLimit(1000).Build() // limit 1000, invoice total 118

	ctx, sess, _ := handlerCtx(t, srv, f, "POST", "/invoices",
		invoiceBody(f, custID, itemID))
	srv.storeInvoiceHandler()(ctx)

	is.True(sessionErrors(sess)["status"] == nil, "no error expected within the credit limit")
	assertRow(t, srv.db, "invoices", map[string]any{"customer_id": custID, "transaction_kind": "invoice"})
}
