package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
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
	rows, err := s.db.Query(`
    SELECT
    receivables_income.id, receivables_income.uuid, receivables_income.code, receivables_income.date, receivables_income.amount,
    receivables_income.payment, receivables_income.status, receivables_income.notes, receivables_income.created_at, receivables_income.updated_at,
    (select count(*) from receivables_income_items
    where receivables_income.company_id = receivables_income_items.company_id
    and receivables_income.id = receivables_income_items.receivable_income_id
    ) as invoices,
    customers.uuid, customers.name, customers.amount_due
    FROM receivables_income
    INNER JOIN customers ON (receivables_income.company_id = customers.company_id AND receivables_income.customer_id = customers.id)
    WHERE receivables_income.company_id = $1
    AND receivables_income.deleted_at IS NULL
    ORDER BY receivables_income.id DESC
  `, CurrentCompany(ctx).ID)
	if err != nil {
		return nil, err
	}
	data := make([]*payment, 0)
	for rows.Next() {
		i := new(payment)

		if err = rows.Scan(
			&i.ID,
			&i.UUID,
			&i.Code,
			&i.Date,
			&i.Amount,
			&i.Payment,
			&i.Status,
			&i.Notes,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Invoices,
			&i.Customer.UUID,
			&i.Customer.Name,
			&i.Customer.AmountDue,
		); err != nil {
			return nil, err
		}

		data = append(data, i)
	}
	return data, nil
}

func (s *Server) findPaymentByUUID(ctx context.Context, uuid string) (*payment, error) {

	i := new(payment)
	err := s.db.QueryRow(`
    SELECT
    receivables_income.id, receivables_income.uuid, receivables_income.code, receivables_income.date, receivables_income.amount,
    receivables_income.payment, receivables_income.status, receivables_income.notes, receivables_income.created_at, receivables_income.updated_at,
    (select count(*) from receivables_income_items
    where receivables_income.company_id = receivables_income_items.company_id
    and receivables_income.id = receivables_income_items.receivable_income_id
    ) as invoices,
    customers.id, customers.uuid, customers.name, customers.email, customers.amount_due, customers.address
    FROM receivables_income
    INNER JOIN customers ON (receivables_income.company_id = customers.company_id AND receivables_income.customer_id = customers.id)
    WHERE receivables_income.company_id = $1
    AND receivables_income.uuid = $2
  `, CurrentCompany(ctx).ID, uuid).Scan(
		&i.ID,
		&i.UUID,
		&i.Code,
		&i.Date,
		&i.Amount,
		&i.Payment,
		&i.Status,
		&i.Notes,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Invoices,
		&i.Customer.ID,
		&i.Customer.UUID,
		&i.Customer.Name,
		&i.Customer.Email,
		&i.Customer.AmountDue,
		&i.Customer.Address,
	)
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (s *Server) findPaymentLines(ctx context.Context, paymentID int) ([]*paymentLine, error) {
	rows, err := s.db.Query(`
    select receivables_income_items.id, receivables_income_items.payment_amount,
    receivables_income_items.created_at, receivables_income_items.updated_at,
    receivables_income_items.deleted_at,
    invoices.id, invoices.uuid, invoices.code, invoices.date, invoices.due_on, invoices.total, receivables_income_items.amount_due,
    CASE WHEN (receivables_income_items.amount_due - receivables_income_items.payment_amount) = 0 THEN 'paid' ELSE 'partial' END AS paid_status,
	invoices.tax_number, invoices.note
    from receivables_income_items
    inner join companies on receivables_income_items.company_id = companies.id
    inner join receivables_income on receivables_income_items.company_id = receivables_income.company_id and receivables_income_items.receivable_income_id = receivables_income.id
    inner join invoices on receivables_income_items.company_id = invoices.company_id and receivables_income_items.invoice_id = invoices.id
    INNER JOIN tax_receipts ON (invoices.company_id = tax_receipts.company_id AND invoices.tax_receipt_id = tax_receipts.id)
    where receivables_income_items.company_id = $1
    and receivables_income_items.receivable_income_id = $2
    ORDER BY receivables_income_items.id
  `, CurrentCompany(ctx).ID, paymentID)
	if err != nil {
		return nil, err
	}
	data := make([]*paymentLine, 0)
	for rows.Next() {
		i := new(paymentLine)
		if err = rows.Scan(
			&i.ID,
			&i.Payment,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
			&i.Invoice.ID,
			&i.Invoice.UUID,
			&i.Invoice.Code,
			&i.Invoice.Date,
			&i.Invoice.DueOn,
			&i.Invoice.Amount,
			&i.Invoice.AmountDue,
			&i.Invoice.PaidStatus,
			&i.Invoice.NCF,
			&i.Invoice.Notes,
		); err != nil {
			return nil, err
		}

		data = append(data, i)
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
		payment := &paymentInsert{
			CompanyID:  companyID,
			CustomerID: int(customer.ID),
			Date:       form.Date,
			Amount:     form.Amount,
			Notes:      form.Notes,
			Payment:    foundation.ToJSON(form.Payment),
			Status:     PaymentStatuses.Completed,
			Code:       seqInfo.Code,
		}
		if err = ptx.Insert(context.Background(), payment); err != nil {
			return err
		}
		paymentID := int(payment.ID)

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
		for _, pl := range lines {
			_, err = tx.Exec("UPDATE receivables_income_items SET payment_amount = 0, amount_due = 0 WHERE company_id = $1 AND receivable_income_id = $2 AND invoice_id = $3", companyID, payment.ID, pl.Invoice.ID)
			if err != nil {
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

		// Mark payment as voided
		// Reset all amount on the receivable incomes items
		_, err = tx.Exec("UPDATE receivables_income SET amount = 0, status = 'void'::payment_status, payment = NULL WHERE company_id = $1 AND id = $2", companyID, payment.ID)

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

		stmt, err := tx.Prepare("UPDATE receivables_income SET customer_id = $1, date = $2, amount = $3, notes = $4, payment = $5 WHERE company_id = $6 AND id = $7")
		if err != nil {
			return err
		}

		_, err = stmt.Exec(
			customer.ID,
			form.Date,
			form.Amount,
			form.Notes,
			foundation.ToJSON(form.Payment),
			CurrentCompany(ctx).ID,
			payment.ID,
		)

		if err != nil {
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
	lines := s.filterPaymentLines(form.Lines, ADDED, UPDATED, DELETED)
	for _, line := range lines {
		pLines := filter(paymentLines, func(pl *paymentLine) bool { return pl.ID == line.ID })
		if len(pLines) == 0 {
			return errors.New("payment line not found")
		}
		switch line.Action {
		case ADDED:
			invoice, err := s.findInvoicesByUUID(ctx, TransactionKinds.Invoice, companyID, line.Uuid)
			if err != nil {
				return err
			}
			_, err = tx.Exec(
				"INSERT INTO receivables_income_items (company_id, receivable_income_id, date, invoice_id, amount_due, payment_amount) VALUES ($1, $2, $3, $4, $5, $6)",
				companyID, paymentId, form.Date, invoice.ID, line.AmountDue, line.Payment)
			if err != nil {
				return err
			}
			if err = s.updateInvoiceBalance(tx, companyID, invoice.ID, -line.Payment); err != nil {
				return err
			}
			if err = s.updateCustomerAmountDue(tx, companyID, invoice.Customer.ID, -line.Payment); err != nil {
				return err
			}
		case UPDATED:
			invoice, err := s.findInvoicesByID(companyID, pLines[0].Invoice.ID)
			if err != nil {
				return err
			}

			diff := pLines[0].Payment - line.Payment
			invoice.AmountDue += diff
			if invoice.AmountDue < 0 {
				invoice.AmountDue = 0
			}

			_, err = tx.Exec(
				"UPDATE receivables_income_items SET payment_amount = $4, amount_due = $5 WHERE company_id = $1 AND receivable_income_id = $2 AND id = $3",
				companyID, paymentId, line.ID, line.Payment, invoice.AmountDue)
			if err != nil {
				return err
			}
			if err = s.updateInvoiceBalance(tx, companyID, pLines[0].Invoice.ID, diff); err != nil {
				return err
			}

			if err = s.updateCustomerAmountDue(tx, companyID, customer.ID, diff); err != nil {
				return err
			}
		case DELETED:
			_, err := tx.Exec(
				"DELETE FROM receivables_income_items WHERE company_id = $1 AND receivable_income_id = $2 AND id = $3",
				companyID, paymentId, line.ID)
			if err != nil {
				return err
			}
			if err = s.updateInvoiceBalance(tx, companyID, pLines[0].Invoice.ID, line.Payment); err != nil {
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
