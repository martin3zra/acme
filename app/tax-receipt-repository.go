package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/martin3zra/acme/pkg/foundation"
)

type taxReceipt struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Series        string `json:"series"`
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

func (s *Server) findTaxesReceipts(ctx context.Context) ([]*taxReceipt, error) {
	rows, err := s.db.Query("SELECT id, name, series, type, sequence_start, sequence_end, current, created_at, updated_at, deleted_at "+
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
			&t.Series,
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

func (s *Server) grabTaxReceiptSequence(tx *sql.Tx, companyId, taxReceiptID int) (*taxReceiptSeq, error) {
	var sequence, sequenceEnd int64
	var series, taxType string
	row := tx.QueryRow("SELECT series, type, current, sequence_end FROM tax_receipts WHERE company_id = $1 AND id = $2", companyId, taxReceiptID)
	err := row.Scan(&series, &taxType, &sequence, &sequenceEnd)
	if err != nil {
		return nil, err
	}

	// abort the transaction when the current sequence is equals to the end sequence
	if sequence == sequenceEnd {
		return nil, errors.New("tax receipt reach end") //new(ErrTaxReceiptReachEnd)
	}

	_, err = tx.Exec("UPDATE tax_receipts SET current = $3 WHERE company_id = $1 AND id = $2", companyId, taxReceiptID, sequence+1)
	if err != nil {
		return nil, err
	}

	taxNumber := foundation.GeneratePrefixedNumber(fmt.Sprintf("%s%s", series, taxType), 8, int(sequence))

	return &taxReceiptSeq{Seq: sequence, Number: taxNumber}, nil
}
