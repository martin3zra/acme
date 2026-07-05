package app

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/martin3zra/forge/database"
)

// TestFlowRecurringGeneratesInvoice exercises the recurrence generator
// (ProcessRecurrenceAt -> generateInvoice). Writing it surfaced that recurring
// invoice generation had never worked end-to-end — three distinct breakages,
// all now fixed:
//
//  1. findInvoicesByID INNER JOINed tax_receipts, so a template (no tax receipt)
//     was unfindable.                                    [LEFT JOIN]
//  2. mapInvoiceToStoreForm dereferenced *invoice.TaxReceiptID, nil for
//     templates -> panic.                                [nil guard]
//  3. the line copy dropped warehouse_id, so generated lines defaulted to
//     warehouse_id=1 -> FK violation.   [warehouse_id plumbed through
//     findInvoiceLines + the line struct + the mapper]
func TestFlowRecurringGeneratesInvoice(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Credit("net30").Build()

	// Create the recurring template through the real storeInvoice path: it now
	// persists both the recurrence object and the tax_receipt_id, so no manual
	// patching is needed for generation to work.
	now := time.Now()
	r := &Recurrence{
		Enabled: true, Name: "Monthly", Type: "schedule",
		Frequency: "monthly", Interval: 1, Timezone: "America/Santo_Domingo",
		DayOfMonth: now.Day(), StartDate: &now,
	}
	tmplForm := &StoreInvoiceForm{
		CustomerID: custID, Date: now, Terms: "net30", TaxReceipt: f.taxReceiptID,
		Discount: Discount{Type: "percentage"}, Kind: TransactionKinds.Template,
		Recurrence: r, Lines: []*Line{mkLine(itemID, f.unitID, f.warehouseID, 1, 100, 18)},
	}
	tmplForm.Compute()
	tmplUUID, err := f.s.storeInvoice(f.ctx, tmplForm)
	is.NoErr(err)

	var tmplID int
	is.NoErr(s.db.QueryRow(`SELECT id FROM invoices WHERE uuid = $1`, tmplUUID).Scan(&tmplID))

	// Sanity: the template carries its tax receipt (the fix under test).
	is.Equal(scalarInt(t, s.db, `SELECT COALESCE(tax_receipt_id,0) FROM invoices WHERE id = $1`, tmplID), f.taxReceiptID)

	// storeInvoice set r.NextRunAt = StartDate; that is the run we process now.

	var gen uuid.UUID
	is.NoErr(database.WithTransaction(s.db, func(tx *sql.Tx) error {
		var e error
		gen, e = s.ProcessRecurrenceAt(tx, now, f.company.ID, tmplID, r)
		return e
	}))
	is.True(gen != uuid.Nil, "a recurring invoice should have been generated")

	assertRow(t, s.db, "invoices", map[string]any{"uuid": gen.String(), "transaction_kind": "invoice"})
	src := scalarString(t, s.db, `SELECT COALESCE(source::text, '') FROM invoices WHERE uuid = $1`, gen.String())
	is.True(strings.Contains(src, tmplUUID), "generated invoice should reference the template")
}
