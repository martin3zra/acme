package app

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

func (s *Server) runRecurrenceScheduler() error {
	ctx := context.Background()
	var (
		companyID   int
		invoiceUUID uuid.UUID
	)

	err := database.WithTransaction(s.db, func(tx *sql.Tx) error {

		var (
			invoiceID      int
			recurrenceData *Recurrence
		)

		err := tx.QueryRowContext(ctx, `
      SELECT id, company_id, recurrence
      FROM invoices
      WHERE transaction_kind = 'template'
        AND recurrence_next_run_at <= now()
        AND recurrence IS NOT NULL
        AND (recurrence->>'enabled')::boolean = TRUE
        AND (
          NOT (recurrence ? 'until')
          OR (recurrence->>'until') IS NULL
          OR (recurrence->>'until')::timestamptz >= now()
        )
      ORDER BY recurrence_next_run_at
      LIMIT 1
      FOR UPDATE SKIP LOCKED;
  `).Scan(&invoiceID, &companyID, &recurrenceData)
		if err == sql.ErrNoRows {
			return nil
		}

		if err != nil {
			return err
		}

		invoiceUUID, err = s.ProcessRecurrenceAt(tx, time.Now(), companyID, invoiceID, recurrenceData)
		if err != nil {
			return err
		}

		if !recurrenceData.SendEmail {
			invoiceUUID = uuid.Nil
		}

		return nil
	})

	if err != nil {
		return err
	}

	if invoiceUUID != uuid.Nil {
		go func() {
			s.enqueueInvoiceEmail(companyID, invoiceUUID.String())
		}()
	}

	return nil
}

func (s *Server) ProcessRecurrenceAt(tx *sql.Tx, now time.Time, companyID, invoiceID int, r *Recurrence) (uuid.UUID, error) {
	if !r.Enabled {
		return uuid.Nil, nil
	}

	loc, err := time.LoadLocation(r.Timezone)
	if err != nil {
		return uuid.Nil, err
	}

	now = now.In(loc)

	// Initialize NextRunAt if missing (first run or legacy data)
	if r.NextRunAt == nil {
		return uuid.Nil, fmt.Errorf("recurrence missing next_run_at (invoice_id=%d)", invoiceID)
	}

	invoiceUUID, err := s.generateInvoice(tx, companyID, invoiceID, r, *r.NextRunAt)
	if err != nil {
		return uuid.Nil, err
	}

	r.LastGeneratedAt = r.NextRunAt

	// Determine next occurrence
	next := s.NextOccurrence(r, loc)

	if r.Until != nil && !next.Before(*r.Until) {
		return invoiceUUID, s.clearNextRunAt(tx, companyID, invoiceID)
	}

	r.NextRunAt = &next

	return invoiceUUID, s.updateNextRunAt(tx, companyID, invoiceID, r)
}

// NextOccurrence calculates the next valid occurrence for a recurrence
func (s *Server) NextOccurrence(r *Recurrence, loc *time.Location) time.Time {
	var anchor time.Time

	switch {
	case r.NextRunAt != nil:
		anchor = *r.NextRunAt
	case r.LastGeneratedAt != nil:
		anchor = *r.LastGeneratedAt
	case r.StartDate != nil:
		anchor = *r.StartDate
	default:
		panic("recurrence has no anchor")
	}

	return s.NextOccurrenceFrom(r, anchor, loc)
}

func (s *Server) NextOccurrenceFrom(r *Recurrence, anchor time.Time, loc *time.Location) time.Time {
	anchor = anchor.In(loc)

	switch r.Frequency {
	case "daily":
		return anchor.AddDate(0, 0, r.Interval)
	case "weekly":
		// Step 1: move forward at least one day
		searchStart := anchor.AddDate(0, 0, 1)

		// Step 2: compute the start of the interval window
		// interval=1 → same week
		// interval=2 → skip one full week
		if r.Interval > 1 {
			searchStart = searchStart.AddDate(0, 0, 7*(r.Interval-1))
		}

		// Start searching from the NEXT week boundary
		searchStart = schedulerStartOfWeek(searchStart).AddDate(0, 0, 7)

		// No weekdays → same weekday as anchor
		if len(r.Weekdays) == 0 {
			return searchStart
		}

		// Step 3: Find earliest allowed weekday
		for i := 0; i < 7; i++ {
			candidate := searchStart.AddDate(0, 0, i)
			wd := strings.ToLower(candidate.Weekday().String())

			for _, allowed := range r.Weekdays {
				if wd == strings.ToLower(allowed) {
					return candidate
				}
			}
		}

		panic("weekly recurrence: no valid weekday")

	case "monthly":
		year, month, _ := anchor.Date()
		nextMonth := month + time.Month(r.Interval)
		daysInMonth := daysIn(nextMonth, year)
		day := min(r.DayOfMonth, daysInMonth)
		return time.Date(year, nextMonth, day,
			anchor.Hour(), anchor.Minute(), anchor.Second(), 0, loc)
	case "yearly":
		return anchor.AddDate(r.Interval, 0, 0)
	default:
		return anchor // fallback
	}
}

func (s *Server) ComputeNextRunAt(r *Recurrence, loc *time.Location) *time.Time {
	next := s.NextOccurrence(r, loc)

	// UNTIL (exclusive)
	if r.Until != nil && !next.Before(*r.Until) {
		return nil
	}

	return &next
}

// schedulerStartOfWeek: start of week (Monday-based)
func schedulerStartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	return time.Date(
		t.Year(), t.Month(), t.Day()-weekday+1,
		t.Hour(), t.Minute(), t.Second(), 0, t.Location(),
	)
}

func daysIn(month time.Month, year int) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func (s *Server) generateInvoice(tx *sql.Tx, companyID, invoiceID int, r *Recurrence, now time.Time) (uuid.UUID, error) {

	invoice, err := s.findInvoicesByID(companyID, invoiceID)
	if err != nil {
		return uuid.Nil, err
	}

	lines, err := s.findInvoiceLines(context.Background(), companyID, invoiceID)
	if err != nil {
		return uuid.Nil, err
	}

	invoiceForm := mapInvoiceToStoreForm(invoice, lines)
	invoiceForm.Date = now
	invoiceForm.Terms = "net30" // do we need to take this from the invoice or customer?
	invoiceForm.Kind = TransactionKinds.Invoice
	invoiceForm.Compute()
	invoiceForm.Source = &TransactionSource{
		ID:   invoice.UUID,
		Type: TransactionKinds.Template,
	}
	invoiceForm.Recurrence = nil
	// Merge notes: recurrence can append a tag
	if r.Name != "" {
		invoiceForm.Notes = fmt.Sprintf("%s (recurrence: %s)", invoiceForm.Notes, r.Name)
	}

	invoiceUUID, err := s.storeInvoiceBackground(tx, companyID, invoiceForm)
	if err != nil {
		return uuid.Nil, err
	}

	return foundation.AsUUID(invoiceUUID)
}

func (s *Server) updateNextRunAt(tx *sql.Tx, companyID, invoiceID int, r *Recurrence) error {

	_, err := tx.Exec(`
    UPDATE invoices
    SET recurrence = 
      jsonb_set(
        jsonb_set(
          COALESCE(recurrence, '{}'::jsonb),
          '{last_generated_at}', to_jsonb($3::timestamptz), true
        ),
        '{next_run_at}', to_jsonb($4::timestamptz), true
      )
    WHERE company_id = $1
      AND id = $2 
      AND transaction_kind = 'template';
  `, companyID, invoiceID, r.LastGeneratedAt, r.NextRunAt)

	return err
}

func (s *Server) clearNextRunAt(tx *sql.Tx, companyID, invoiceID int) error {
	_, err := tx.Exec(`
		UPDATE invoices
		SET recurrence = recurrence - 'next_run_at'
		WHERE company_id = $1
		  AND id = $2
		  AND transaction_kind = 'template'
	`, companyID, invoiceID)

	return err
}
