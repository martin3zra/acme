package app

import (
	"testing"
	"time"
)

// cashInvoiceBody builds a pia (cash) invoice payload with the given tendered
// payment amount.
func cashInvoiceBody(f *fixture, customerID, itemID int, payment float64) map[string]any {
	return map[string]any{
		"customer_id": customerID,
		"date":        time.Now().Format(time.RFC3339),
		"terms":       "pia",
		"tax_receipt": f.taxReceiptID,
		"kind":        "invoice",
		"discount":    map[string]any{"value": 0, "type": "percentage"},
		"payment":     map[string]any{"cash": map[string]any{"amount": payment}},
		"lines": []map[string]any{{
			"id": itemID, "unit": f.unitID, "qty": 1, "price": 100,
			"rate": 18, "action": "added", "warehouse_id": f.warehouseID,
		}},
	}
}

// TestFlowCashPaymentMismatchBlocked: a cash invoice whose tendered payment does
// not equal the total is rejected, and nothing is persisted.
func TestFlowCashPaymentMismatchBlocked(t *testing.T) {
	srv := newHandlerServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, srv)

	g := newFaker(t)
	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Build()

	// total is 118, but only 50 tendered.
	ctx, sess, _ := handlerCtx(t, srv, f, "POST", "/invoices", cashInvoiceBody(f, custID, itemID, 50))
	srv.storeInvoiceHandler()(ctx)

	is.True(sessionErrors(sess)["status"] != nil, "payment/total mismatch should be blocked")
	assertNoRow(t, srv.db, "invoices", map[string]any{"customer_id": custID})
}

// TestFlowCashPaymentMatchesCreates: a cash invoice with exact payment succeeds.
func TestFlowCashPaymentMatchesCreates(t *testing.T) {
	srv := newHandlerServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, srv)

	g := newFaker(t)
	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Build()

	ctx, sess, _ := handlerCtx(t, srv, f, "POST", "/invoices", cashInvoiceBody(f, custID, itemID, 118))
	srv.storeInvoiceHandler()(ctx)

	is.True(sessionErrors(sess)["status"] == nil, "exact cash payment should be accepted")
	assertRow(t, srv.db, "invoices", map[string]any{
		"customer_id": custID, "transaction_kind": "invoice", "paid_status": "paid",
	})
}
