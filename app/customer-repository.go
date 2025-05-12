package app

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/martin3zra/acme/pkg/foundation"
)

type customer struct {
	ID          int     `json:"id"`
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	ContactName string  `json:"contact_name,omitempty,"`
	Phone       string  `json:"phone"`
	Email       string  `json:"email"`
	AmountDue   float64 `json:"amount_due,omitempty,"`
	Address     string  `json:"address"`
	// Add timestamps properties
	foundation.Timestamps
	Status foundation.Status `json:"status,omitempty,"`
}

type receivable struct {
	ID      int    `json:"id"`
	UUID    string `json:"uuid"`
	Invoice struct {
		ID         int        `json:"id"`
		UUID       string     `json:"uuid"`
		Number     string     `json:"number"`
		NCF        string     `json:"ncf"`
		Date       time.Time  `json:"date"`
		DueOn      *time.Time `json:"due_on"`
		Total      float64    `json:"total"`
		AmountDue  float64    `json:"amount_due"`
		PaidStatus PaidStatus `json:"paid_status"`
		Notes      string     `json:"notes"`
		Payment    float64    `json:"payment"`
		Discount   float64    `json:"discount"`
		Balance    float64    `json:"balance"`
	} `json:"invoice"`
	foundation.Timestamps
}

func (s *Server) findCustomeByID(companyID, customerID int) (*customer, error) {

	var c customer
	err := s.db.QueryRow("SELECT c.id, c.uuid, c.name, c.contact_name, c.phone, c.email, c.status, c.amount_due, "+
		"c.created_at, c.updated_at, c.deleted_at "+
		"FROM customers c "+
		"INNER JOIN companies ON (c.company_id = companies.id) "+
		"WHERE c.company_id = $1 "+
		"AND c.id = $2 "+
		"AND c.deleted_at IS NULL", companyID, customerID).
		Scan(&c.ID, &c.UUID, &c.Name, &c.ContactName, &c.Phone, &c.Email, &c.Status, &c.AmountDue, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
	if err != nil {
		return nil, err
	}

	c.Address = "LOUISVILLE, Selby 3864 Johnson Street, United States of America"
	return &c, nil
}

func (s *Server) findCustomeByUUID(companyID int, customerID string) (*customer, error) {

	var c customer
	err := s.db.QueryRow("SELECT c.id, c.uuid, c.name, c.contact_name, c.phone, c.email, c.status, c.amount_due, "+
		"c.created_at, c.updated_at, c.deleted_at "+
		"FROM customers c "+
		"INNER JOIN companies ON (c.company_id = companies.id) "+
		"WHERE c.company_id = $1 "+
		"AND c.uuid = $2 "+
		"AND c.deleted_at IS NULL", companyID, customerID).
		Scan(&c.ID, &c.UUID, &c.Name, &c.ContactName, &c.Phone, &c.Email, &c.Status, &c.AmountDue, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
	if err != nil {
		return nil, err
	}

	c.Address = "LOUISVILLE, Selby 3864 Johnson Street, United States of America"
	return &c, nil
}

func (s *Server) findCustomers(companyID int) ([]*customer, error) {

	rows, err := s.db.Query("SELECT c.id, c.uuid, c.name, c.contact_name, c.phone, c.email, c.status, c.amount_due, "+
		"c.created_at, c.updated_at, c.deleted_at "+
		"FROM customers c "+
		"INNER JOIN companies ON (c.company_id = companies.id) "+
		"WHERE c.company_id = $1 "+
		"AND c.deleted_at IS NULL ORDER BY c.name", companyID)
	if err != nil {
		return nil, err
	}
	data := make([]*customer, 0)
	for rows.Next() {
		row := new(customer)
		if err = rows.Scan(
			&row.ID,
			&row.UUID,
			&row.Name,
			&row.ContactName,
			&row.Phone,
			&row.Email,
			&row.Status,
			&row.AmountDue,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.DeletedAt,
		); err != nil {
			return data, err
		}
		row.Address = "LOUISVILLE, Selby 3864 Johnson Street, United States of America"

		data = append(data, row)
	}

	return data, nil
}

func (s *Server) findCustomersBySearchCriteria(companyID int, term string) ([]*customer, error) {
	if len(strings.TrimSpace(term)) == 0 {
		return nil, errors.New("need to specifiy the customer you're looking for")
	}
	rows, err := s.db.Query("SELECT c.id, c.uuid, c.name, c.contact_name, c.phone, c.email, c.amount_due "+
		"FROM customers c "+
		"INNER JOIN companies ON (c.company_id = companies.id) "+
		"WHERE c.company_id = $1 "+
		"AND c.name LIKE $2 "+
		"AND c.deleted_at IS NULL AND c.status = 'enabled' ORDER BY c.name LIMIT 5 ", companyID, "%"+term+"%")
	if err != nil {
		return nil, err
	}
	data := make([]*customer, 0)
	for rows.Next() {
		row := new(customer)
		if err = rows.Scan(
			&row.ID,
			&row.UUID,
			&row.Name,
			&row.ContactName,
			&row.Phone,
			&row.Email,
			&row.AmountDue,
		); err != nil {
			return data, err
		}

		row.Address = "LOUISVILLE, Selby 3864 Johnson Street, United States of America"
		data = append(data, row)
	}

	return data, nil
}

func (s *Server) storeCustomer(companyId int, form StoreCustomerForm) error {

	_, err := s.db.Exec("INSERT INTO customers (company_id, name, contact_name, email, phone) "+
		"VALUES ($1, $2, $3, $4, $5)",
		companyId, form.Name, form.Contact, form.Email, form.Phone,
	)

	return err
}

func (s *Server) updateCustomer(companyID, customerID int, form UpdateCustomerForm) error {

	_, err := s.db.Exec(
		"UPDATE customers SET name = $1, contact_name = $2,  email = $3, phone = $4 WHERE company_id = $5 AND id = $6",
		form.Name, form.Contact, form.Email, form.Phone, companyID, customerID,
	)

	return err
}

func (s *Server) deleteCustomer(companyID, customerID int) error {

	_, err := s.db.Exec(
		"UPDATE customers SET deleted_at = now(), updated_at = now() WHERE company_id = $1 AND id = $2",
		companyID, customerID,
	)

	return err
}

func (s *Server) toggleCustomerStatus(companyID int, customer *customer) error {
	status := customer.Status
	if status == "enabled" {
		status = "disabled"
	} else {
		status = "enabled"
	}
	_, err := s.db.Exec(
		"UPDATE customers SET updated_at = now(), status = $3 WHERE company_id = $1 AND id = $2",
		companyID, customer.ID, status,
	)
	return err
}

func (s *Server) updateCustomerAmountDue(tx *sql.Tx, companyId, customerId int, amountDue float64) error {

	result, err := tx.Exec("UPDATE customers SET amount_due = amount_due + $3 WHERE company_id = $1 AND id = $2",
		companyId, customerId, amountDue,
	)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Fatalf("Error updating customer amount due: %v", txErr)
			return txErr
		}
	}

	if affected, err := result.RowsAffected(); err == nil {
		if affected != 1 {
			txErr := tx.Rollback()
			if txErr != nil {
				return txErr
			}
			return errors.New("unable to update customer balance") //new(ErrUnprocessableEntity)
		}
	}

	return nil
}

func (s *Server) findCustomeReceivables(companyID int, customerID string) ([]*receivable, error) {
	rows, err := s.db.Query(`
    SELECT receivables.id, receivables.uuid, invoices.uuid, invoices.id,
    invoices.date, invoices.due_on, invoices.total, invoices.amount_due, invoices.paid_status,
    tax_receipts.series || tax_receipts.type || LPAD(invoices.tax_receipt_sequence::varchar,8,'0') as NCF
		FROM receivables
		INNER JOIN companies ON (receivables.company_id = companies.id)
		INNER JOIN invoices ON (receivables.company_id = invoices.company_id AND receivables.customer_id = invoices.customer_id AND receivables.invoice_id = invoices.id)
    INNER JOIN tax_receipts ON (invoices.company_id = tax_receipts.company_id AND invoices.tax_receipt_id = tax_receipts.id)
		INNER JOIN customers ON (receivables.company_id = customers.company_id AND receivables.customer_id = customers.id)
		WHERE receivables.company_id = $1
    AND invoices.paid_status != 'paid'::paid_status
    AND customers.uuid = $2
		AND receivables.deleted_at IS NULL
  `, companyID, customerID)
	if err != nil {
		return nil, err
	}
	data := make([]*receivable, 0)
	for rows.Next() {
		row := new(receivable)
		if err = rows.Scan(
			&row.ID,
			&row.UUID,
			&row.Invoice.UUID,
			&row.Invoice.ID,
			&row.Invoice.Date,
			&row.Invoice.DueOn,
			&row.Invoice.Total,
			&row.Invoice.AmountDue,
			&row.Invoice.PaidStatus,
			&row.Invoice.NCF,
		); err != nil {
			return data, err
		}

		row.Invoice.Number = s.generatePrefixedInvoiceNumber(row.Invoice.ID)

		data = append(data, row)
	}

	return data, nil
}
