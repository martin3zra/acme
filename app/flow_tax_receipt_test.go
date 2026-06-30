package app

import (
	"testing"
	"time"
)

// TestFlowTaxReceiptConsumed: issuing an invoice consumes one tax-receipt (NCF)
// sequence number and stamps it on the invoice.
func TestFlowTaxReceiptConsumed(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	before := scalarFloat(t, s.db, `SELECT current FROM tax_receipts WHERE id = $1`, f.taxReceiptID)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := mkCustomer(t, f, "pia")
	uuid := mkInvoice(t, f, custID, TransactionKinds.Invoice, "pia", nil,
		mkLine(itemID, f.unitID, f.warehouseID, 1, 100, 18))

	after := scalarFloat(t, s.db, `SELECT current FROM tax_receipts WHERE id = $1`, f.taxReceiptID)
	is.EqualFloat(after, before+1)

	var taxReceiptID int
	var taxNumber string
	is.NoErr(s.db.QueryRow(`SELECT tax_receipt_id, COALESCE(tax_number,'') FROM invoices WHERE uuid = $1`, uuid).Scan(&taxReceiptID, &taxNumber))
	is.Equal(taxReceiptID, f.taxReceiptID)
	is.True(taxNumber != "", "invoice should carry a tax (NCF) number")
}

// TestFlowTaxReceiptExhausted: an invoice fails when the tax-receipt sequence is
// exhausted (current == sequence_end), and nothing is persisted.
func TestFlowTaxReceiptExhausted(t *testing.T) {
	is := newIs(t)
	srv := newTestServer(t)
	f := mkAccountCompany(t, srv)

	// A tax receipt already at its end.
	var exhaustedID int
	is.NoErr(srv.db.QueryRow(
		`INSERT INTO tax_receipts (company_id, name, serie, type, sequence_start, sequence_end, current)
		 VALUES ($1, 'Exhausted', 'B02', 'fiscal', 1, 5, 5) RETURNING id`, f.company.ID,
	).Scan(&exhaustedID))

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := mkCustomer(t, f, "pia")

	form := &StoreInvoiceForm{
		CustomerID: custID, Date: time.Now(), Terms: "pia", TaxReceipt: exhaustedID,
		Discount: Discount{Type: "percentage"}, Kind: TransactionKinds.Invoice,
		Lines: []*Line{mkLine(itemID, f.unitID, f.warehouseID, 1, 100, 18)},
	}
	form.Compute()
	_, err := f.s.storeInvoice(f.ctx, form)
	is.Err(err, "invoice must fail when the tax receipt sequence is exhausted")

	// Transaction rolled back: no invoice persisted for this customer.
	assertNoRow(t, srv.db, "invoices", map[string]any{"customer_id": custID})
}
