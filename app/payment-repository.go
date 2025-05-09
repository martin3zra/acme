package app

import (
	"database/sql"
	"log"
	"time"

	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

type payment struct {
	ID       int       `json:"id"`
	Date     time.Time `json:"date"`
	Amount   float64   `json:"amount"`
	Customer struct {
		UUID      string  `json:"uuid"`
		Name      string  `json:"name"`
		AmountDue float64 `json:"amount_due"`
	} `json:"customer"`
	foundation.Timestamps
}

func (s *Server) findPayments(companyId int) ([]*payment, error) {
	rows, err := s.db.Query(`
    SELECT
    receivables_income.id, receivables_income.date, receivables_income.amount, receivables_income.created_at, receivables_income.updated_at,
    customers.uuid, customers.name, customers.amount_due
    FROM receivables_income
    INNER JOIN customers ON (receivables_income.company_id = customers.company_id AND receivables_income.customer_id = customers.id)
    WHERE receivables_income.company_id = $1
    AND receivables_income.deleted_at IS NULL
  `, companyId)
	if err != nil {
		return nil, err
	}
	data := make([]*payment, 0)
	for rows.Next() {
		i := new(payment)

		if err = rows.Scan(
			&i.ID,
			&i.Date,
			&i.Amount,
			&i.CreatedAt,
			&i.UpdatedAt,
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

func (s *Server) storePayment(companyId int, form StorePaymentForm) error {

	customer, err := s.findCustomeByUUID(companyId, form.CustomerID)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO receivables_income (company_id, customer_id, date, amount) " +
		"VALUES ($1, $2, $3, $4) RETURNING id")
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
