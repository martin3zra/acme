package app

import (
	"database/sql"
	"log"
	"strconv"
	"time"

	"github.com/martin3zra/acme/pkg/foundation"
)

type invoice struct {
	ID         int        `json:"id"`
	Number     string     `json:"number"`
	Customer   customer   `json:"customer"`
	Date       time.Time  `json:"date"`
	Amount     float64    `json:"amount"`
	Tax        float64    `json:"tax"`
	Status     string     `json:"status"`
	PaidStatus PaidStatus `json:"paid_status"`
}

func (s *Server) findInvoices(companyId int) ([]*invoice, error) {
	rows, err := s.db.Query("SELECT invoices.id, invoices.date, invoices.amount, invoices.tax, invoices.status, invoices.paid_status, "+
		"customers.id as customer, customers.name, customers.email, customers.phone "+
		"FROM invoices "+
		"INNER JOIN companies ON (invoices.company_id = companies.id) "+
		"INNER JOIN customers ON (invoices.company_id = customers.company_id AND invoices.customer_id = customers.id) "+
		"WHERE invoices.company_id = $1", companyId)
	if err != nil {
		return nil, err
	}
	data := make([]*invoice, 0)
	for rows.Next() {
		i := new(invoice)

		if err = rows.Scan(
			&i.ID,
			&i.Date,
			&i.Amount,
			&i.Tax,
			&i.Status,
			&i.PaidStatus,
			&i.Customer.ID,
			&i.Customer.Name,
			&i.Customer.Email,
			&i.Customer.Phone,
		); err != nil {
			return nil, err
		}

		i.Number = s.generatePrefixedInvoiceNumber(i.ID)

		data = append(data, i)
	}
	return data, nil
}

func (s *Server) generatePrefixedInvoiceNumber(value int) string {
	return foundation.GeneratePrefixedNumber("INV-", 10, value)
}

func (s *Server) storeInvoice(companyID int, form StoreInvoiceForm) error {

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO invoices (company_id, date, customer_id, amount, tax, amount_due, note, paid_status) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id")
	if err != nil {
		defer stmt.Close()
		if txErr := tx.Rollback(); txErr != nil {
			log.Fatalf("Error inserting new item: %v", txErr)
			return txErr
		}

		return err
	}

	var invoiceID int
	err = stmt.QueryRow(
		companyID,
		form.Date,
		form.CustomerID,
		form.Amount,
		form.Tax,
		form.AmountDue,
		form.Notes,
		form.paidStatus,
	).Scan(&invoiceID)

	if err != nil {
		return err
	}

	if err = s.attachInvoiceLines(tx, companyID, invoiceID, form); err != nil {
		return err
	}

	// trigger an event for this? Use pipe!!!
	if form.Terms > 1 {
		if err = s.registerReceivable(tx, companyID, invoiceID, form.CustomerID); err != nil {
			return err
		}

		if err = s.updateCustomerAmountDue(tx, companyID, form.CustomerID, form.AmountDue); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Server) attachInvoiceLines(tx *sql.Tx, companyId, invoiceId int, form StoreInvoiceForm) error {
	stmt := "INSERT INTO invoices_items (company_id, invoice_id, item_id, unit_id, qty, price, tax) VALUES "
	vals := []any{}
	for i, line := range form.Lines {
		//
		vals = append(vals, companyId, invoiceId, line.ID, line.Unit, line.Qty, line.Price, line.Rate)

		numFields := 7
		n := i * numFields

		stmt += `(`
		for j := range numFields {
			stmt += `$` + strconv.Itoa(n+j+1) + `,`
		}
		stmt = stmt[:len(stmt)-1] + `),`
	}
	stmt = stmt[:len(stmt)-1]

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

func (s *Server) registerReceivable(tx *sql.Tx, companyId, invoiceId, customerId int) error {
	_, err := tx.Exec("INSERT INTO receivables (company_id, invoice_id, customer_id) VALUES($1, $2, $3)",
		companyId, invoiceId, customerId,
	)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Fatalf("Error registering receivables: %v", txErr)
			return txErr
		}
	}

	return err
}
