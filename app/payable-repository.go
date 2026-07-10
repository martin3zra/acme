package app

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
)

type vendorPayment struct {
	ID        int       `json:"id"`
	UUID      string    `json:"uuid"`
	CompanyID int       `json:"company_id"`
	Vendor    vendor    `json:"vendor"`
	Date      time.Time `json:"date"`
	Amount    float64   `json:"amount"`
	Notes     string    `json:"notes"`
	Payment   *Payment  `json:"payment"`
	Status    string    `json:"status"`
	Code      string    `json:"code"`
	foundation.Timestamps
}

type vendorPaymentLine struct {
	ID         int       `json:"id"`
	APID       int64     `json:"ap_id"`
	APUUID     string    `json:"ap_uuid"`
	BillNumber string    `json:"bill_number"`
	BillDate   time.Time `json:"bill_date"`
	DueDate    time.Time `json:"due_date"`
	AmountDue  float64   `json:"amount_due"`
	Payment    float64   `json:"payment"`
	PaidStatus string    `json:"paid_status"`
}

// findPayables lists the company's outstanding AP entries with their payables
// register row. The old query was rooted on payables and joined up to
// accounts_payable; this reads accounts_payable as the root because every filter
// and the ordering belong to it, and pulls the register row through Has/With.
//
// The dropped joins are all redundant: INNER JOIN companies and INNER JOIN vendors
// only asserted existence (both are NOT NULL FKs, and company_id already scopes
// the query), and payables.company_id / payables.vendor_id are copies of the AP
// row's own columns, so matching on accounts_payable_id alone selects the same row.
func (s *Server) findPayables(ctx context.Context) ([]*Payable, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []accountsPayableModel
	if err := pdb.Model(&accountsPayableModel{}).
		With("Register").
		Has("Register").
		WhereEq("company_id", CurrentCompany(ctx).ID).
		Where("status", "!=", "void").
		Where("paid_status", "!=", "paid").
		OrderBy("due_date", playsql.Asc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*Payable, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toPayable())
	}
	return data, nil
}

func (s *Server) storeVendorPayment(ctx context.Context, form *StoreVendorPaymentForm) error {
	companyID := CurrentCompany(ctx).ID
	v, err := s.findVendorByUUID(ctx, form.VendorID)
	if err != nil {
		return err
	}

	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		seqInfo, err := GetNextSequence(tx, companyID, "vendor_payment")
		if err != nil {
			return err
		}

		ptx, err := playTx(tx)
		if err != nil {
			return err
		}
		// Map insert so uuid stays unset for the DB default; the merged
		// vendorPaymentModel maps uuid, which a struct insert would write as empty.
		paymentID64, err := ptx.Model(&vendorPaymentModel{}).Insert(context.Background(), map[string]any{
			"company_id": companyID,
			"vendor_id":  int(v.ID),
			"date":       form.Date,
			"amount":     form.Amount,
			"notes":      form.Notes,
			"payment":    foundation.ToJSON(form.Payment),
			"status":     "completed",
			"code":       seqInfo.Code,
		})
		if err != nil {
			return err
		}
		paymentID := int(paymentID64)

		rows := make([]map[string]any, 0, len(form.Lines))
		for _, line := range form.Lines {
			ap, err := s.findAPByUUID(tx, companyID, line.UUID)
			if err != nil {
				return err
			}
			rows = append(rows, map[string]any{
				"company_id":          companyID,
				"vendor_payment_id":   paymentID,
				"accounts_payable_id": ap.ID,
				"date":                form.Date,
				"amount_due":          line.AmountDue,
				"payment_amount":      line.Payment,
			})
			if err = s.updateAPBalance(tx, companyID, ap.ID, line.Payment); err != nil {
				return err
			}
		}
		if _, err = ptx.Model(&vendorPaymentItem{}).InsertMany(context.Background(), rows); err != nil {
			return err
		}

		return s.updateVendorAmountPayable(tx, companyID, v.ID, -form.Amount)
	})
}

// findAPByUUID resolves an AP entry inside the caller's transaction. Unlike
// toPayable, ID/UUID here are the accounts_payable row's own identifiers — the
// caller wants the AP entry, not its payables register row.
func (s *Server) findAPByUUID(tx *sql.Tx, companyID int, uuid string) (*Payable, error) {
	ptx, err := playTx(tx)
	if err != nil {
		return nil, err
	}

	var row accountsPayableModel
	if err := ptx.Model(&accountsPayableModel{}).
		WhereEq("company_id", companyID).
		WhereEq("uuid", uuid).
		First(context.Background(), &row); err != nil {
		return nil, err
	}

	return &Payable{
		ID:            row.ID,
		UUID:          row.UUID,
		AmountPayable: row.AmountPayable,
		AmountPaid:    row.AmountPaid,
	}, nil
}

func (s *Server) updateAPBalance(tx *sql.Tx, companyID int, apID int64, paymentAmount float64) error {
	// Stays raw: playsql's Update replaces a column, it cannot express the
	// self-referencing increment `amount_paid = amount_paid + $3`.
	var newPaid float64
	var amountPayable float64
	err := tx.QueryRow(
		"UPDATE accounts_payable SET amount_paid = amount_paid + $3, updated_at = NOW() WHERE company_id = $1 AND id = $2 RETURNING amount_paid, amount_payable",
		companyID, apID, paymentAmount,
	).Scan(&newPaid, &amountPayable)
	if err != nil {
		return err
	}

	var newStatus PaidStatus
	switch {
	case newPaid >= amountPayable:
		newStatus = PaidStatuses.Paid
	case newPaid > 0:
		newStatus = PaidStatuses.Partial
	default:
		newStatus = PaidStatuses.UnPaid
	}

	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	// accountsPayableModel leaves updated_at/paid_at unmapped, so both are passed
	// explicitly; mass-assignment is unrestricted on these models. The old code ran
	// two near-identical statements to conditionally add paid_at.
	changes := map[string]any{
		"paid_status": string(newStatus),
		"updated_at":  time.Now(),
	}
	if newStatus == PaidStatuses.Paid {
		changes["paid_at"] = time.Now()
	}
	if _, err = ptx.Model(&accountsPayableModel{}).
		WhereEq("company_id", companyID).
		WhereEq("id", apID).
		Update(context.Background(), changes); err != nil {
		return err
	}

	// Propagate payment status to the source purchase order (if this AP entry
	// was created from a vendor bill that is linked to a PO via source). Stays
	// raw: the join predicate casts a JSON field to uuid, which no builder
	// relation can express.
	var poUUID string
	lookupErr := tx.QueryRow(
		`SELECT p_po.uuid::text
		   FROM accounts_payable ap
		   JOIN purchases p_bill ON ap.purchase_id = p_bill.id
		   JOIN purchases p_po   ON p_po.company_id = p_bill.company_id
		                        AND p_po.uuid = (p_bill.source->>'id')::uuid
		                        AND p_po.transaction_kind = 'purchase_order'
		  WHERE ap.company_id = $1 AND ap.id = $2`,
		companyID, apID,
	).Scan(&poUUID)
	// No linked PO is fine; any other error must not be swallowed (it would
	// otherwise abort the surrounding transaction silently).
	if lookupErr != nil && lookupErr != sql.ErrNoRows {
		return lookupErr
	}
	if poUUID != "" {
		if err = updatePOPaymentStatus(tx, companyID, poUUID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) findVendorPaymentByUUID(ctx context.Context, uuid string) (*vendorPayment, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row vendorPaymentModel
	if err := pdb.Model(&vendorPaymentModel{}).
		WithConstraint("Vendor", withTrashedRelation).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("uuid", uuid).
		First(ctx, &row); err != nil {
		return nil, err
	}

	return row.toVendorPayment(), nil
}

// findVendorPaymentLines lists the AP entries a vendor payment settled. The old
// query's CASE expression is computed in Go instead; the bill columns come from
// the eager-loaded AP entry.
func (s *Server) findVendorPaymentLines(ctx context.Context, paymentID int) ([]*vendorPaymentLine, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []vendorPaymentItemRead
	if err := pdb.Model(&vendorPaymentItemRead{}).
		With("AccountsPayable").
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("vendor_payment_id", paymentID).
		OrderBy("id", playsql.Asc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*vendorPaymentLine, 0, len(rows))
	for _, r := range rows {
		l := &vendorPaymentLine{
			ID:         int(r.ID),
			APID:       r.AccountsPayableID,
			AmountDue:  r.AmountDue,
			Payment:    r.PaymentAmount,
			PaidStatus: "partial",
		}
		if r.AmountDue-r.PaymentAmount <= 0 {
			l.PaidStatus = "paid"
		}
		if ap := r.AccountsPayable; ap != nil {
			l.APUUID = ap.UUID
			l.BillNumber = ap.InvoiceNumber
			l.BillDate = ap.InvoiceDate
			l.DueDate = ap.DueDate
		}
		data = append(data, l)
	}
	return data, nil
}

func (s *Server) voidVendorPayment(ctx context.Context, uuid string) error {
	p, err := s.findVendorPaymentByUUID(ctx, uuid)
	if err != nil {
		return err
	}
	if p.Status == "void" {
		return errors.New("vendor payment already voided")
	}

	lines, err := s.findVendorPaymentLines(ctx, p.ID)
	if err != nil {
		return err
	}

	companyID := CurrentCompany(ctx).ID
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		ptx, err := playTx(tx)
		if err != nil {
			return err
		}

		for _, l := range lines {
			if err = s.updateAPBalance(tx, companyID, l.APID, -l.Payment); err != nil {
				return err
			}
			if _, err = ptx.Model(&vendorPaymentItem{}).
				WhereEq("company_id", companyID).
				WhereEq("vendor_payment_id", p.ID).
				WhereEq("accounts_payable_id", l.APID).
				Update(context.Background(), map[string]any{
					"payment_amount": 0,
					"updated_at":     time.Now(),
				}); err != nil {
				return err
			}
		}

		if err = s.updateVendorAmountPayable(tx, companyID, p.Vendor.ID, p.Amount); err != nil {
			return err
		}

		_, err = ptx.Model(&vendorPaymentModel{}).
			WhereEq("company_id", companyID).
			WhereEq("id", p.ID).
			Update(context.Background(), map[string]any{
				"amount":     0,
				"status":     "void",
				"updated_at": time.Now(),
			})
		return err
	})
}

func (s *Server) findVendorPayments(ctx context.Context) ([]*vendorPayment, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []vendorPaymentModel
	if err := pdb.Model(&vendorPaymentModel{}).
		WithConstraint("Vendor", withTrashedRelation).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		OrderBy("id", playsql.Desc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*vendorPayment, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toVendorPayment())
	}
	return data, nil
}
