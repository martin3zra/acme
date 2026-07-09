package app

import (
	"context"
	"database/sql"
	"errors"

	"github.com/martin3zra/forge/foundation"
)

type taxReceipt struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Serie         string `json:"serie"`
	Type          string `json:"type"`
	SequenceStart int    `json:"sequence_start"`
	SequenceEnd   int    `json:"sequence_end"`
	Current       int    `json:"current"`
	// Add timestamps properties
	foundation.Timestamps
}

type taxReceiptSeq struct {
	Seq    int64
	Number string
}

func (s *Server) findTaxReceiptsForSetup(ctx context.Context) ([]*taxReceipt, error) {
	rows, err := s.db.Query(`
		SELECT id, name, serie, type, COALESCE(sequence_start, 0), COALESCE(sequence_end, 0), COALESCE(current, 0), created_at, updated_at, deleted_at
		FROM tax_receipts
		WHERE company_id = $1
		ORDER BY id;
  `, CurrentCompany(ctx).ID)
	if err != nil {
		return nil, err
	}

	data := make([]*taxReceipt, 0)
	for rows.Next() {
		t := new(taxReceipt)
		if err = rows.Scan(
			&t.ID,
			&t.Name,
			&t.Serie,
			&t.Type,
			&t.SequenceStart,
			&t.SequenceEnd,
			&t.Current,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.DeletedAt,
		); err != nil {
			return nil, err
		}

		data = append(data, t)
	}

	return data, nil
}

func (s *Server) findTaxesReceipts(ctx context.Context) ([]*taxReceipt, error) {
	rows, err := s.db.Query("SELECT id, name, serie, type, sequence_start, sequence_end, current, created_at, updated_at, deleted_at "+
		"FROM tax_receipts WHERE company_id = $1 ORDER BY id", CurrentCompany(ctx).ID)
	if err != nil {
		return nil, err
	}

	data := make([]*taxReceipt, 0)
	for rows.Next() {
		t := new(taxReceipt)
		if err = rows.Scan(
			&t.ID,
			&t.Name,
			&t.Serie,
			&t.Type,
			&t.SequenceStart,
			&t.SequenceEnd,
			&t.Current,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.DeletedAt,
		); err != nil {
			return nil, err
		}

		data = append(data, t)
	}

	return data, nil
}

var (
	// ErrTaxReceiptReachEnd means the receipt's sequence range is spent.
	ErrTaxReceiptReachEnd = errors.New("tax receipt reach end")
	// ErrTaxReceiptNotFound means no such receipt exists for this company. The
	// StoreInvoiceForm rule is `exists:tax_receipts,id`, which is not scoped to a
	// company, so a request naming another tenant's receipt id gets here. The
	// company_id predicate below is what actually keeps it from being consumed.
	ErrTaxReceiptNotFound = errors.New("tax receipt not found for this company")
)

// grabTaxReceiptSequence consumes the next fiscal sequence for a tax receipt and
// returns it with its formatted NCF.
//
// This must be one atomic statement. The previous implementation read `current` in
// one query and then wrote `current + 1` from the value it had read, with no row
// lock. Under READ COMMITTED two concurrent invoices both read the same `current`,
// both issued the same NCF, and the counter advanced only once:
//
//	invoice A sequence = 1 -> NCF B9900000001
//	invoice B sequence = 1 -> NCF B9900000001
//	tax_receipts.current after two issues = 2 (should be 3)
//
// Duplicate fiscal numbers are a compliance problem, not just a data one.
// `current = current + 1 ... RETURNING` increments under the row lock the UPDATE
// itself takes, so each caller gets a distinct sequence. GetNextSequence, the
// company-code sequence next door, already guarded itself with FOR UPDATE.
//
// The `current < sequence_end` predicate replaces the old pre-read exhaustion check,
// so the last issuable number is sequence_end - 1, as it always was. Updating no row
// is ambiguous — spent range, unknown id, another company's id, or a soft-deleted
// receipt — so the cause is resolved with a follow-up read rather than reported as
// exhaustion either way.
//
// `deleted_at IS NULL` is defensive: nothing in the codebase soft-deletes a tax
// receipt today, so no live path reaches it. If one is ever added, a retired receipt
// must not keep issuing fiscal numbers.
func (s *Server) grabTaxReceiptSequence(tx *sql.Tx, companyId, taxReceiptID int) (*taxReceiptSeq, error) {
	var sequence int64
	var serie string
	err := tx.QueryRow(
		`UPDATE tax_receipts
		    SET current = current + 1, updated_at = NOW()
		  WHERE company_id = $1 AND id = $2 AND deleted_at IS NULL AND current < sequence_end
		 RETURNING serie, current - 1`,
		companyId, taxReceiptID,
	).Scan(&serie, &sequence)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, s.explainNoSequence(tx, companyId, taxReceiptID)
	}
	if err != nil {
		return nil, err
	}

	taxNumber := foundation.GeneratePrefixedNumber(serie, 8, int(sequence))

	return &taxReceiptSeq{Seq: sequence, Number: taxNumber}, nil
}

// explainNoSequence distinguishes an exhausted receipt from one that does not belong
// to the company. Reading ErrNoRows does not abort the caller's transaction, so this
// second query is safe to run on it.
func (s *Server) explainNoSequence(tx *sql.Tx, companyId, taxReceiptID int) error {
	var exists bool
	err := tx.QueryRow(
		"SELECT true FROM tax_receipts WHERE company_id = $1 AND id = $2 AND deleted_at IS NULL",
		companyId, taxReceiptID,
	).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTaxReceiptNotFound
	}
	if err != nil {
		return err
	}
	return ErrTaxReceiptReachEnd
}
