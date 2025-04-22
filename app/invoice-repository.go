package app

import (
	"database/sql"
	"log"
	"strconv"
)

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
		vals = append(vals, companyId, invoiceId, line.ID, 1, line.Qty, line.Price, line.Rate)

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
