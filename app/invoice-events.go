package app

import (
	"context"
	"database/sql"

	"github.com/martin3zra/forge/events"
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

// InvoiceVoided is raised when an invoice has been voided (amounts and lines
// zeroed). Reverse listeners undo its side effects within the same transaction.
type InvoiceVoided struct {
	CompanyID  int
	InvoiceID  int
	CustomerID int
	// MovementRecorded reports whether the invoice had posted stock movements, so
	// the inventory listener only reverses stock that was actually taken.
	MovementRecorded bool
}

func (InvoiceVoided) Name() string { return "invoice.voided" }

// reverseReceivableListener removes the invoice's receivable when it is voided.
type reverseReceivableListener struct{ s *Server }

func (l reverseReceivableListener) Handle(ctx context.Context, tx *sql.Tx, e events.Event) error {
	ev := e.(InvoiceVoided)
	return l.s.deleteInvoiceFromReceivables(tx, ev.CompanyID, ev.InvoiceID, ev.CustomerID)
}

// reverseInventoryListener returns the invoice's stock (a sale_return) when the
// voided invoice had recorded inventory movements.
type reverseInventoryListener struct{ s *Server }

func (l reverseInventoryListener) Handle(ctx context.Context, tx *sql.Tx, e events.Event) error {
	ev := e.(InvoiceVoided)
	if !ev.MovementRecorded {
		return nil
	}
	return l.s.reverseMovements(tx, ev.CompanyID, "invoice", ev.InvoiceID, InventoryMovementKinds.SaleReturn)
}

// dispatcher lazily builds the event dispatcher and registers listeners on first
// use, so every way a *Server is constructed (production, test harness) gets the
// same wiring without touching constructors.
func (s *Server) dispatcher() *events.Dispatcher {
	s.eventsOnce.Do(func() {
		d := events.NewDispatcher()
		d.On(InvoiceCreated{}, receivableListener{s})
		d.On(InvoiceVoided{}, reverseReceivableListener{s}, reverseInventoryListener{s})
		s.events = d
	})
	return s.events
}
