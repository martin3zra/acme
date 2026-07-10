package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
)

type payment struct {
	ID       int       `json:"id"`
	UUID     string    `json:"uuid"`
	Code     string    `json:"code"`
	Date     time.Time `json:"date"`
	Amount   float64   `json:"amount"`
	Notes    string    `json:"notes"`
	Customer struct {
		ID        int     `json:"_"`
		UUID      string  `json:"uuid"`
		Name      string  `json:"name"`
		Email     string  `json:"email"`
		AmountDue float64 `json:"amount_due"`
		Address   string  `json:"address"`
		Phone     string  `json:"phone"`
	} `json:"customer"`
	Invoices int           `json:"invoices"`
	Payment  Payment       `json:"payment"`
	Status   PaymentStatus `json:"status"`
	foundation.Timestamps
}

type paymentLine struct {
	ID      int     `json:"id"`
	Payment float64 `json:"payment"`
	Invoice struct {
		ID         int        `json:"_"`
		UUID       string     `json:"uuid"`
		Code       string     `json:"code"`
		Date       time.Time  `json:"date"`
		DueOn      *time.Time `json:"due_on"`
		PaidStatus PaidStatus `json:"paid_status"`
		Amount     float64    `json:"amount"`
		AmountDue  float64    `json:"amount_due"`
		NCF        string     `json:"ncf"`
		Notes      string     `json:"notes"`
	} `json:"invoice"`
	Action LineAction `json:"action"`
	foundation.Timestamps
}

func (s *Server) findPayments(ctx context.Context) ([]*payment, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []paymentModel
	if err := pdb.Model(&paymentModel{}).
		WithConstraint("Customer", withTrashedRelation).
		WithCount("Items", playsql.As("invoices")).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		OrderBy("id", playsql.Desc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*payment, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toPayment())
	}
	return data, nil
}

// findPaymentByUUID reads one payment. WithTrashed because the old detail query,
// unlike the list, carried no deleted_at predicate.
func (s *Server) findPaymentByUUID(ctx context.Context, uuid string) (*payment, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row paymentModel
	if err := pdb.Model(&paymentModel{}).
		WithConstraint("Customer", withTrashedRelation).
		WithCount("Items", playsql.As("invoices")).
		WithTrashed().
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("uuid", uuid).
		First(ctx, &row); err != nil {
		return nil, err
	}

	return row.toPayment(), nil
}

// findPaymentLines lists the invoices a payment settled. The old CASE expression is
// computed in Go. The companies and receivables_income joins are dropped: both only
// asserted existence, and the receivable_income_id filter already pins the parent.
//
// The old query also carried `INNER JOIN tax_receipts ON invoices.tax_receipt_id =
// tax_receipts.id`. That was a bug, not an existence assertion: tax_receipt_id is
// nullable, so the join silently dropped every line whose invoice has no tax receipt.
// Callers treat this list as the payment's full set of allocations — voidPayment
// walks it to give the money back — so a payment against a receipt-less invoice
// voided without restoring the invoice or the customer balance. The join is gone;
// invoices.tax_number is nullable too, so such a line simply reports an empty NCF.
func (s *Server) findPaymentLines(ctx context.Context, paymentID int) ([]*paymentLine, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []paymentItemRead
	if err := pdb.Model(&paymentItemRead{}).
		With("Invoice").
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("receivable_income_id", paymentID).
		OrderBy("id", playsql.Asc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*paymentLine, 0, len(rows))
	for _, r := range rows {
		l := &paymentLine{ID: r.ID, Payment: r.PaymentAmount}
		l.CreatedAt = r.CreatedAt
		l.UpdatedAt = r.UpdatedAt
		l.DeletedAt = r.DeletedAt

		l.Invoice.AmountDue = r.AmountDue
		l.Invoice.PaidStatus = PaidStatuses.Partial
		if r.AmountDue-r.PaymentAmount == 0 {
			l.Invoice.PaidStatus = PaidStatuses.Paid
		}
		if inv := r.Invoice; inv != nil {
			l.Invoice.ID = inv.ID
			l.Invoice.UUID = inv.UUID
			l.Invoice.Code = inv.Code
			l.Invoice.Date = inv.Date
			l.Invoice.DueOn = inv.DueOn
			l.Invoice.Amount = inv.Total
			l.Invoice.NCF = inv.TaxNumber
			l.Invoice.Notes = inv.Note
		}
		data = append(data, l)
	}
	return data, nil
}

func (s *Server) storePayment(ctx context.Context, form *StorePaymentForm) error {

	customer, err := s.findCustomeByUUID(ctx, form.CustomerID)
	if err != nil {
		return err
	}

	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		companyID := CurrentCompany(ctx).ID

		seqInfo, err := GetNextSequence(tx, companyID, "payment")
		if err != nil {
			return err
		}

		ptx, err := playTx(tx)
		if err != nil {
			return err
		}
		// Map insert so uuid stays unset for the DB default; the merged
		// paymentModel maps uuid, which a struct insert would write as empty.
		paymentID64, err := ptx.Model(&paymentModel{}).Insert(context.Background(), map[string]any{
			"company_id":  companyID,
			"customer_id": int(customer.ID),
			"date":        form.Date,
			"amount":      form.Amount,
			"notes":       form.Notes,
			"payment":     foundation.ToJSON(form.Payment),
			"status":      PaymentStatuses.Completed,
			"code":        seqInfo.Code,
		})
		if err != nil {
			return err
		}
		paymentID := int(paymentID64)

		if err = s.attachPaymentLines(tx, ctx, paymentID, form); err != nil {
			return err
		}

		if err = s.updateCustomerAmountDue(tx, companyID, customer.ID, -form.Amount); err != nil {
			return err
		}

		return nil
	})
}

func (s *Server) attachPaymentLines(tx *sql.Tx, ctx context.Context, paymentId int, form *StorePaymentForm) error {
	companyId := CurrentCompany(ctx).ID
	rows := make([]map[string]any, 0, len(form.Lines))
	for _, line := range form.Lines {
		invoice, err := s.findInvoicesByUUID(ctx, TransactionKinds.Invoice, companyId, line.Uuid)
		if err != nil {
			return err
		}
		rows = append(rows, map[string]any{
			"company_id":           companyId,
			"receivable_income_id": paymentId,
			"date":                 form.Date,
			"invoice_id":           invoice.ID,
			"amount_due":           line.AmountDue,
			"payment_amount":       line.Payment,
		})

		// Update invoice balance and payment status.
		if err = s.updateInvoiceBalance(tx, companyId, invoice.ID, -line.Payment); err != nil {
			return err
		}
	}

	ptx, err := playTx(tx)
	if err != nil {
		return err
	}
	_, err = ptx.Model(&paymentItem{}).InsertMany(context.Background(), rows)
	return err
}

func (s *Server) voidPayment(ctx context.Context, uuid string) error {
	payment, err := s.findPaymentByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	if payment.Status == PaymentStatuses.Void {
		return errors.New("payment already voided")
	}

	lines, err := s.findPaymentLines(ctx, payment.ID)
	if err != nil {
		return err
	}
	companyID := CurrentCompany(ctx).ID
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		ptx, err := playTx(tx)
		if err != nil {
			return err
		}

		for _, pl := range lines {
			if _, err = ptx.Model(&paymentItem{}).
				WhereEq("company_id", companyID).
				WhereEq("receivable_income_id", payment.ID).
				WhereEq("invoice_id", pl.Invoice.ID).
				Update(context.Background(), map[string]any{
					"payment_amount": 0,
					"amount_due":     0,
				}); err != nil {
				return err
			}

			if err = s.updateInvoiceBalance(tx, companyID, pl.Invoice.ID, pl.Payment); err != nil {
				return err
			}
		}

		// Adjust customer balance
		if err = s.updateCustomerAmountDue(tx, companyID, payment.Customer.ID, payment.Amount); err != nil {
			return err
		}

		// Mark the payment voided and drop its payment blob. A nil map value writes
		// NULL; the enum needs no ::payment_status cast, Postgres resolves it.
		_, err = ptx.Model(&paymentModel{}).
			WhereEq("company_id", companyID).
			WhereEq("id", payment.ID).
			Update(context.Background(), map[string]any{
				"amount":  0,
				"status":  string(PaymentStatuses.Void),
				"payment": nil,
			})
		return err
	})
}

func (s *Server) updatePayment(ctx context.Context, uuid string, form *UpdatePaymentForm) error {
	payment, err := s.findPaymentByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	if payment.Status == PaymentStatuses.Void {
		return errors.New("payment already voided")
	}

	customer, err := s.findCustomeByUUID(ctx, form.CustomerID)
	if err != nil {
		return err
	}

	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		ptx, err := playTx(tx)
		if err != nil {
			return err
		}

		// The statement it replaced was prepared for a single execution.
		if _, err = ptx.Model(&paymentModel{}).
			WhereEq("company_id", CurrentCompany(ctx).ID).
			WhereEq("id", payment.ID).
			Update(context.Background(), map[string]any{
				"customer_id": customer.ID,
				"date":        form.Date,
				"amount":      form.Amount,
				"notes":       form.Notes,
				"payment":     foundation.ToJSON(form.Payment),
			}); err != nil {
			return err
		}

		return s.processPaymentLines(tx, ctx, payment.ID, customer, form)
	})
}

func (s *Server) processPaymentLines(tx *sql.Tx, ctx context.Context, paymentId int, customer *customer, form *UpdatePaymentForm) error {

	paymentLines, err := s.findPaymentLines(ctx, paymentId)
	if err != nil {
		return err
	}

	companyID := CurrentCompany(ctx).ID
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}

	// Only UPDATED and DELETED refer to an already-stored line. Resolving it before
	// the switch made ADDED unreachable: a newly added line carries no id, so the
	// lookup never matched and every add failed with "payment line not found".
	existingLine := func(id int) (*paymentLine, error) {
		pLines := filter(paymentLines, func(pl *paymentLine) bool { return pl.ID == id })
		if len(pLines) == 0 {
			return nil, errors.New("payment line not found")
		}
		return pLines[0], nil
	}

	lines := s.filterPaymentLines(form.Lines, ADDED, UPDATED, DELETED)
	for _, line := range lines {
		switch line.Action {
		case ADDED:
			invoice, err := s.findInvoicesByUUID(ctx, TransactionKinds.Invoice, companyID, line.Uuid)
			if err != nil {
				return err
			}
			if _, err = ptx.Model(&paymentItem{}).Insert(context.Background(), map[string]any{
				"company_id":           companyID,
				"receivable_income_id": paymentId,
				"date":                 form.Date,
				"invoice_id":           invoice.ID,
				"amount_due":           line.AmountDue,
				"payment_amount":       line.Payment,
			}); err != nil {
				return err
			}
			if err = s.updateInvoiceBalance(tx, companyID, invoice.ID, -line.Payment); err != nil {
				return err
			}
			if err = s.updateCustomerAmountDue(tx, companyID, invoice.Customer.ID, -line.Payment); err != nil {
				return err
			}
		case UPDATED:
			stored, err := existingLine(line.ID)
			if err != nil {
				return err
			}
			invoice, err := s.findInvoicesByID(companyID, stored.Invoice.ID)
			if err != nil {
				return err
			}

			diff := stored.Payment - line.Payment
			invoice.AmountDue += diff
			if invoice.AmountDue < 0 {
				invoice.AmountDue = 0
			}

			if _, err = ptx.Model(&paymentItem{}).
				WhereEq("company_id", companyID).
				WhereEq("receivable_income_id", paymentId).
				WhereEq("id", line.ID).
				Update(context.Background(), map[string]any{
					"payment_amount": line.Payment,
					"amount_due":     invoice.AmountDue,
				}); err != nil {
				return err
			}
			if err = s.updateInvoiceBalance(tx, companyID, stored.Invoice.ID, diff); err != nil {
				return err
			}

			if err = s.updateCustomerAmountDue(tx, companyID, customer.ID, diff); err != nil {
				return err
			}
		case DELETED:
			stored, err := existingLine(line.ID)
			if err != nil {
				return err
			}
			// paymentItem carries no softdelete tag, so this is a hard DELETE,
			// matching the statement it replaced.
			if _, err := ptx.Model(&paymentItem{}).
				WhereEq("company_id", companyID).
				WhereEq("receivable_income_id", paymentId).
				WhereEq("id", line.ID).
				Delete(context.Background()); err != nil {
				return err
			}
			if err = s.updateInvoiceBalance(tx, companyID, stored.Invoice.ID, line.Payment); err != nil {
				return err
			}
			if err = s.updateCustomerAmountDue(tx, companyID, customer.ID, line.Payment); err != nil {
				return err
			}
		default:
			fmt.Println("Unknown action:", line.Action)
			return errors.New("unknown action")
		}
	}
	return nil
}

func (s *Server) filterPaymentLines(lines []*PaymentLine, actions ...LineAction) []*PaymentLine {
	if len(actions) == 0 {
		return nil
	}

	actionSet := make(map[string]struct{}, len(actions))
	for _, action := range actions {
		actionSet[string(action)] = struct{}{}
	}
	filteredLines := make([]*PaymentLine, 0, len(lines))
	for _, line := range lines {
		if _, ok := actionSet[string(line.Action)]; ok {
			filteredLines = append(filteredLines, line)
		}
	}
	return filteredLines
}
