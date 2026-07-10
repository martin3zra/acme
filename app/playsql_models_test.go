package app

import (
	"context"
	"testing"

	"github.com/martin3zra/playsql"
)

// Phase 1 of the playsql adoption: typed reads in tests only. The models and the
// wiring live in _test.go, so the production binary does not link playsql (and
// its MySQL/SQL-Server/SQLite drivers). Promote to a production file when phase 2
// (writes) begins.

// psInvoice is a playsql read model for the invoices table. Plain struct + db
// tags; the pk carries play:"pk,incrementing". Only the columns the tests assert
// on are mapped — unmapped columns are ignored on hydration.
type psInvoice struct {
	ID               int64   `db:"id" play:"pk,incrementing"`
	UUID             string  `db:"uuid"`
	CompanyID        int64   `db:"company_id"`
	CustomerID       int64   `db:"customer_id"`
	Status           string  `db:"status"`
	PaidStatus       string  `db:"paid_status"`
	Type             string  `db:"type"`
	Total            float64 `db:"total"`
	AmountDue        float64 `db:"amount_due"`
	MovementRecorded bool    `db:"movement_recorded"`
}

func (psInvoice) TableName() string { return "invoices" }

// TestConfigurePlan covers the production Boot wiring directly: newTestServer sets
// s.plan inline, but Boot builds it via configurePlan. Assert that method leaves a
// live executor that s.play() hands back and that reads through it work.
func TestConfigurePlan(t *testing.T) {
	is := newIs(t)
	s := newTestServer(t)
	s.plan = nil // undo the harness so configurePlan is what wires it

	s.configurePlan()
	is.True(s.plan != nil, "configurePlan should wire s.plan")

	pdb, err := s.play()
	is.NoErr(err)
	is.True(pdb == s.plan, "play() should return the cached plan")

	// The executor is usable: a trivial read compiles and runs.
	_, err = pdb.Model(&customerModel{}).Count(context.Background())
	is.NoErr(err)
}

// TestPlaysqlReadInvoice proves phase-1 reads: create a cash invoice through the
// real store path, then hydrate it back with a typed playsql query and assert
// the same values the raw-SQL flow test checks.
func TestPlaysqlReadInvoice(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)
	g := newFaker(t)

	itemID, _ := mkItem(t, f, 100, 60)
	custID, _ := newCustomer(t, f, g).Build()
	uuid := newInvoice(t, f, g).ForCustomer(custID).Cash().
		WithLine(itemID, 2, 100, 18).Build()

	inv, err := playsql.Query[psInvoice](s.plan).
		WhereEq("company_id", int64(f.company.ID)).
		WhereEq("uuid", uuid).
		First(context.Background())
	is.NoErr(err)

	is.Equal(inv.UUID, uuid)
	is.Equal(int(inv.CustomerID), custID)
	is.Equal(inv.Status, "closed")
	is.Equal(inv.PaidStatus, "paid")
	is.Equal(inv.Type, "cash")
	is.EqualFloat(inv.Total, 236) // 2*100 + 18% tax(36)
	is.EqualFloat(inv.AmountDue, 0)
	is.True(inv.MovementRecorded, "cash invoice should record stock movement")
}
