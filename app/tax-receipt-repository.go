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
// The `current < sequence_end` predicate replaces the old pre-read exhaustion check:
// no row is updated when the range is spent, which surfaces as ErrNoRows.
func (s *Server) grabTaxReceiptSequence(tx *sql.Tx, companyId, taxReceiptID int) (*taxReceiptSeq, error) {
	var sequence int64
	var serie string
	err := tx.QueryRow(
		`UPDATE tax_receipts
		    SET current = current + 1, updated_at = NOW()
		  WHERE company_id = $1 AND id = $2 AND current < sequence_end
		 RETURNING serie, current - 1`,
		companyId, taxReceiptID,
	).Scan(&serie, &sequence)
	if errors.Is(err, sql.ErrNoRows) {
		// Either the receipt does not exist, or its range is exhausted. The caller
		// only ever passes a receipt it just read, so exhaustion is the real case.
		return nil, errors.New("tax receipt reach end") //new(ErrTaxReceiptReachEnd)
	}
	if err != nil {
		return nil, err
	}

	taxNumber := foundation.GeneratePrefixedNumber(serie, 8, int(sequence))

	return &taxReceiptSeq{Seq: sequence, Number: taxNumber}, nil
}
