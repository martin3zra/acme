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

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := mkCustomer(t, f, "net30")
	tmplUUID := mkInvoice(t, f, custID, TransactionKinds.Template, "net30", nil,
		mkLine(itemID, f.unitID, f.warehouseID, 1, 100, 18))

	var tmplID int
	is.NoErr(s.db.QueryRow(`SELECT id FROM invoices WHERE uuid = $1`, tmplUUID).Scan(&tmplID))

	now := time.Now()
	r := &Recurrence{
		Enabled: true, Name: "Monthly", Type: "schedule",
		Frequency: "monthly", Interval: 1, Timezone: "America/Santo_Domingo",
		DayOfMonth: now.Day(), StartDate: &now, NextRunAt: &now,
	}

	// A recurring template stores its full recurrence object and the tax receipt
	// to stamp on each generated invoice.
	_, err := s.db.Exec(`UPDATE invoices SET recurrence = $2, tax_receipt_id = $3 WHERE id = $1`, tmplID, r, f.taxReceiptID)
	is.NoErr(err)

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
