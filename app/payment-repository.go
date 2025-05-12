package app

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

type payment struct {
	ID       int       `json:"id"`
	UUID     string    `json:"uuid"`
	Number   string    `json:"number"`
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
	} `json:"customer"`
	Invoices int           `json:"invoices"`
	Payment  Payment       `json:"payment"`
	Status   PaymentStatus `json:"status"`
	foundation.Timestamps
}

type paymentLine struct {
	ID      int64   `json:"id"`
	Payment float64 `json:"payment"`
	Invoice struct {
		ID         int        `json:"_"`
		UUID       string     `json:"uuid"`
		Number     string     `json:"number"`
		Date       time.Time  `json:"date"`
		PaidStatus PaidStatus `json:"paid_status"`
		Amount     float64    `json:"amount"`
		AmountDue  float64    `json:"amount_due"`
	} `json:"invoice"`
	foundation.Timestamps
}

func (s *Server) findPayments(companyId int) ([]*payment, error) {
	rows, err := s.db.Query(`
    SELECT
    receivables_income.id, receivables_income.uuid, receivables_income.date, receivables_income.amount,
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
    ORDER BY receivables_income.id
  `, companyId)
	if err != nil {
		return nil, err
	}
	data := make([]*payment, 0)
	for rows.Next() {
		i := new(payment)

		if err = rows.Scan(
			&i.ID,
			&i.UUID,
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

		i.Number = s.generatePrefixedPaymentNumber(i.ID)

		data = append(data, i)
	}
	return data, nil
}

func (s *Server) findPaymentByUUID(companyId int, uuid string) (*payment, error) {

	i := new(payment)
	err := s.db.QueryRow(`
    SELECT
    receivables_income.id, receivables_income.uuid, receivables_income.date, receivables_income.amount,
    receivables_income.payment, receivables_income.status, receivables_income.notes, receivables_income.created_at, receivables_income.updated_at,
    (select count(*) from receivables_income_items
    where receivables_income.company_id = receivables_income_items.company_id
    and receivables_income.id = receivables_income_items.receivable_income_id
    ) as invoices,
    customers.id, customers.uuid, customers.name, customers.email, customers.amount_due
    FROM receivables_income
    INNER JOIN customers ON (receivables_income.company_id = customers.company_id AND receivables_income.customer_id = customers.id)
    WHERE receivables_income.company_id = $1
    AND receivables_income.uuid = $2
  `, companyId, uuid).Scan(
		&i.ID,
		&i.UUID,
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
	)
	if err != nil {
		return nil, err
	}

	i.Customer.Address = "LOUISVILLE, Selby 3864 Johnson Street, United States of America"
	i.Number = s.generatePrefixedPaymentNumber(i.ID)

	return i, nil
}

func (s *Server) findPaymentLines(companyID int, paymentID int) ([]*paymentLine, error) {
	rows, err := s.db.Query(`
    select receivables_income_items.id, receivables_income_items.payment_amount,
    receivables_income_items.created_at, receivables_income_items.updated_at,
    receivables_income_items.deleted_at,
    invoices.id, invoices.uuid, invoices.date, invoices.total, invoices.amount_due, invoices.paid_status
    from receivables_income_items
    inner join companies on receivables_income_items.company_id = companies.id
    inner join receivables_income on receivables_income_items.company_id = receivables_income.company_id and receivables_income_items.receivable_income_id = receivables_income.id
    inner join invoices on receivables_income_items.company_id = invoices.company_id and receivables_income_items.invoice_id = invoices.id
    where receivables_income_items.company_id = $1
    and receivables_income_items.receivable_income_id = $2
    ORDER BY receivables_income_items.id
  `, companyID, paymentID)
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
			&i.Invoice.Date,
			&i.Invoice.Amount,
			&i.Invoice.AmountDue,
			&i.Invoice.PaidStatus,
		); err != nil {
			return nil, err
		}

		i.Invoice.Number = s.generatePrefixedInvoiceNumber(i.Invoice.ID)

		data = append(data, i)
	}
	return data, nil
}

func (s *Server) storePayment(companyId int, form StorePaymentForm) error {

	customer, err := s.findCustomeByUUID(companyId, form.CustomerID)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO receivables_income (company_id, customer_id, date, amount, notes, payment, status) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id")
	if err != nil {
		defer stmt.Close()
		if txErr := tx.Rollback(); txErr != nil {
			log.Fatalf("Error inserting new item: %v", txErr)
			return txErr
		}

		return err
	}

	var paymentID int
	err = stmt.QueryRow(
		companyId,
		customer.ID,
		form.Date,
		form.Amount,
		form.Notes,
		foundation.ToJSON(form.Payment),
		PaymentStatuses.Completed,
	).Scan(&paymentID)

	if err != nil {
		return err
	}

	if err = s.attachPaymentLines(tx, companyId, paymentID, form); err != nil {
		return err
	}

	if err = s.updateCustomerAmountDue(tx, companyId, customer.ID, -form.Amount); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Server) attachPaymentLines(tx *sql.Tx, companyId, paymentId int, form StorePaymentForm) error {
	vals := []any{}
	for _, line := range form.Lines {
		invoice, err := s.findInvoicesByUUID(companyId, line.Uuid)
		if err != nil {
			return err
		}
		vals = append(vals, companyId, paymentId, form.Date, invoice.ID, line.AmountDue, line.Payment)

		// Update invoice balance and payment status.
		if err = s.updateInvoiceBalance(tx, companyId, invoice.ID, -form.Amount); err != nil {
			return err
		}
	}

	stmt := "INSERT INTO receivables_income_items (company_id, receivable_income_id, date, invoice_id, amount_due, payment_amount) VALUES "
	stmt += database.PrepareBulkInsert(6, len(form.Lines))

	_, err := tx.Exec(stmt, vals...)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Fatalf("Error inserting new item: %v", txErr)
			return txErr
		}

		return err
	}

	return nil
}

func (s *Server) generatePrefixedPaymentNumber(value int) string {
	return foundation.GeneratePrefixedNumber("PAY-", 10, value)
}

func (s *Server) voidPayment(companyID int, uuid string) error {
	payment, err := s.findPaymentByUUID(companyID, uuid)
	if err != nil {
		return err
	}

	if payment.Status == PaymentStatuses.Void {
		return errors.New("payment already voided")
	}

	lines, err := s.findPaymentLines(companyID, payment.ID)
	if err != nil {
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, pl := range lines {
		_, err = tx.Exec("UPDATE receivables_income_items SET payment_amount = 0 WHERE company_id = $1 AND receivable_income_id = $2 AND invoice_id = $3", companyID, payment.ID, pl.Invoice.ID)
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				log.Fatalf("Error updating invoice payment amount: %v", txErr)
				return txErr
			}

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
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Fatalf("Error updating customer amount due: %v", txErr)
			return txErr
		}
	}

	return tx.Commit()
}
