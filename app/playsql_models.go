package app

import (
	"database/sql"

	"github.com/martin3zra/playsql"
)

// Phase 2 of the playsql adoption: production writes. This file is the first
// production code to link playsql (and, transitively, its SQL drivers). Write
// paths convert one at a time; everything not yet migrated stays raw database/sql.

// playTx wraps an in-flight *sql.Tx with playsql under the Postgres grammar, so a
// production write can use typed models on the *same* transaction the caller
// already opened (via forge database.WithTransaction). The caller still owns the
// tx and its commit/rollback.
func playTx(tx *sql.Tx) (*playsql.Tx, error) {
	return playsql.UseTx(tx, "postgres")
}

// Receivable is the write model for the receivables table. The pk is DB-assigned
// (serial); playsql omits the zero id on insert and reads it back via RETURNING.
// Columns with database defaults (timestamps) are intentionally not mapped, so
// the INSERT lets the database fill them.
type Receivable struct {
	ID         int64 `db:"id" play:"pk,incrementing"`
	CompanyID  int   `db:"company_id"`
	InvoiceID  int   `db:"invoice_id"`
	CustomerID int   `db:"customer_id"`
}

func (Receivable) TableName() string { return "receivables" }

// InvoiceItem is the write model for invoice line rows (invoices_items). It backs
// the bulk line insert; the pk is DB-assigned, and timestamp columns are left to
// their database defaults (not mapped here).
type InvoiceItem struct {
	ID          int64   `db:"id" play:"pk,incrementing"`
	CompanyID   int     `db:"company_id"`
	InvoiceID   int     `db:"invoice_id"`
	ItemID      int     `db:"item_id"`
	UnitID      int     `db:"unit_id"`
	Qty         int     `db:"qty"`
	Price       float64 `db:"price"`
	Rate        float64 `db:"rate"`
	Amount      float64 `db:"amount"`
	Tax         float64 `db:"tax"`
	Total       float64 `db:"total"`
	WarehouseID int     `db:"warehouse_id"`
}

func (InvoiceItem) TableName() string { return "invoices_items" }
