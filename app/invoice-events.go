package app

import (
	"context"
	"database/sql"

	"github.com/martin3zra/acme/app/events"
)

// InvoiceCreated is raised when an invoice row has been persisted. Listeners run
// within the creating transaction, so their work is atomic with the invoice.
type InvoiceCreated struct {
	CompanyID  int
	InvoiceID  int
	CustomerID int
	AmountDue  float64
}

func (InvoiceCreated) Name() string { return "invoice.created" }

// receivableListener registers the receivable and raises the customer's balance
// for a credit invoice — the work formerly inlined at the end of
// storeInvoiceInternal (the "trigger an event for this?" TODO).
type receivableListener struct{ s *Server }

func (l receivableListener) Handle(ctx context.Context, tx *sql.Tx, e events.Event) error {
	ev := e.(InvoiceCreated)
	if err := l.s.registerReceivable(tx, ev.CompanyID, ev.InvoiceID, ev.CustomerID); err != nil {
		return err
	}
	return l.s.updateCustomerAmountDue(tx, ev.CompanyID, ev.CustomerID, ev.AmountDue)
}

// dispatcher lazily builds the event dispatcher and registers listeners on first
// use, so every way a *Server is constructed (production, test harness) gets the
// same wiring without touching constructors.
func (s *Server) dispatcher() *events.Dispatcher {
	s.eventsOnce.Do(func() {
		d := events.NewDispatcher()
		d.On(InvoiceCreated{}, receivableListener{s})
		s.events = d
	})
	return s.events
}
