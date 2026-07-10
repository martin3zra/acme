package app

import (
	"context"
	"database/sql"
	"errors"

	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
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

// findTaxesReceipts lists the company's tax receipts, newest id last.
//
// Neither list read filtered deleted_at, so a retired receipt still appears here.
// grabTaxReceiptSequence filters it explicitly; that asymmetry is existing behaviour.
//
// The old query wrapped sequence_start, sequence_end and current in COALESCE(..., 0).
// All three are NOT NULL, so the wrapper never had anything to coalesce.
func (s *Server) findTaxesReceipts(ctx context.Context) ([]*taxReceipt, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []taxReceiptRead
	if err := pdb.Model(&taxReceiptRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		OrderBy("id", playsql.Asc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*taxReceipt, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toTaxReceipt())
	}
	return data, nil
}

// findTaxReceiptsForSetup was a byte-for-byte duplicate of findTaxesReceipts once the
// dead COALESCEs are removed: same columns, same WHERE, same ORDER BY. It stays as a
// named entry point for the account-setup screen.
func (s *Server) findTaxReceiptsForSetup(ctx context.Context) ([]*taxReceipt, error) {
	return s.findTaxesReceipts(ctx)
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
// to the company. A not-found read does not abort the caller's transaction, so this
// second query is safe to run on it.
//
// taxReceiptRead carries no softdelete tag — the list reads never filtered deleted_at
// — so the predicate is written out here, matching grabTaxReceiptSequence above.
func (s *Server) explainNoSequence(tx *sql.Tx, companyId, taxReceiptID int) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	var row taxReceiptRead
	err = ptx.Model(&taxReceiptRead{}).
		Select("id").
		WhereEq("company_id", companyId).
		WhereEq("id", taxReceiptID).
		WhereNull("deleted_at").
		First(context.Background(), &row)
	if errors.Is(err, playsql.ErrNotFound) {
		return ErrTaxReceiptNotFound
	}
	if err != nil {
		return err
	}
	return ErrTaxReceiptReachEnd
}
