package app

import (
	"strings"
	"testing"
)

// TestFlowEstimateToInvoice: converting an estimate into an invoice closes the
// estimate and links the two documents bidirectionally via invoices.source.
func TestFlowEstimateToInvoice(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := mkCustomer(t, f, "pia")

	estUUID := mkInvoice(t, f, custID, TransactionKinds.Estimate, "", nil,
		mkLine(itemID, f.unitID, f.warehouseID, 1, 100, 18))

	invUUID := mkInvoice(t, f, custID, TransactionKinds.Invoice, "pia",
		&TransactionSource{Type: TransactionKinds.Estimate, ID: estUUID},
		mkLine(itemID, f.unitID, f.warehouseID, 1, 100, 18))

	// Source estimate is closed and points forward to the new invoice.
	is.Equal(scalarString(t, s.db, `SELECT status FROM invoices WHERE uuid = $1`, estUUID), "closed")
	estSrc := scalarString(t, s.db, `SELECT COALESCE(source::text, '') FROM invoices WHERE uuid = $1`, estUUID)
	is.True(strings.Contains(estSrc, invUUID), "estimate.source should link to the invoice")

	// New invoice records the inbound estimate as its source.
	invSrc := scalarString(t, s.db, `SELECT COALESCE(source::text, '') FROM invoices WHERE uuid = $1`, invUUID)
	is.True(strings.Contains(invSrc, estUUID), "invoice.source should reference the estimate")
}

// TestFlowOrderToInvoice: same conversion linkage from an order.
func TestFlowOrderToInvoice(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	itemID, _ := mkItem(t, f, 50, 30)
	custID, _ := mkCustomer(t, f, "pia")

	ordUUID := mkInvoice(t, f, custID, TransactionKinds.Order, "pia", nil,
		mkLine(itemID, f.unitID, f.warehouseID, 1, 50, 18))
	invUUID := mkInvoice(t, f, custID, TransactionKinds.Invoice, "pia",
		&TransactionSource{Type: TransactionKinds.Order, ID: ordUUID},
		mkLine(itemID, f.unitID, f.warehouseID, 1, 50, 18))

	is.Equal(scalarString(t, s.db, `SELECT status FROM invoices WHERE uuid = $1`, ordUUID), "closed")
	ordSrc := scalarString(t, s.db, `SELECT COALESCE(source::text, '') FROM invoices WHERE uuid = $1`, ordUUID)
	is.True(strings.Contains(ordSrc, invUUID), "order.source should link to the invoice")
}
