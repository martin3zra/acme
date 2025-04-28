package app

import (
	"database/sql"
	"log"

	"github.com/martin3zra/acme/pkg/foundation"
)

type customer struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	ContactName string  `json:"contact_name,omitempty"`
	Phone       string  `json:"phone"`
	Email       string  `json:"email"`
	AmountDue   float64 `json:"amount_due,omitempty"`
	// Add timestamps properties
	foundation.Timestamps
	Status foundation.Status `json:"status,omitempty"`
}

func (s *Server) findCustomeByID(companyID, customerID int) (*customer, error) {

	var c customer
	err := s.db.QueryRow("SELECT c.id, c.name, c.contact_name, c.phone, c.email, c.status, c.amount_due, "+
		"c.created_at, c.updated_at, c.deleted_at "+
		"FROM customers c "+
		"INNER JOIN companies ON (c.company_id = companies.id) "+
		"WHERE c.company_id = $1 "+
		"AND c.id = $2 "+
		"AND c.deleted_at IS NULL", companyID, customerID).
		Scan(&c.ID, &c.Name, &c.ContactName, &c.Phone, &c.Email, &c.Status, &c.AmountDue, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (s *Server) findCustomers(companyID int) ([]*customer, error) {

	rows, err := s.db.Query("SELECT c.id, c.name, c.contact_name, c.phone, c.email, c.status, c.amount_due, "+
		"c.created_at, c.updated_at, c.deleted_at "+
		"FROM customers c "+
		"INNER JOIN companies ON (c.company_id = companies.id) "+
		"WHERE c.company_id = $1 "+
		"AND c.deleted_at IS NULL", companyID)
	if err != nil {
		return nil, err
	}
	data := make([]*customer, 0)
	for rows.Next() {
		row := new(customer)
		if err = rows.Scan(
			&row.ID,
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

		data = append(data, row)
	}

	return data, nil
}

func (s *Server) findCustomersBySearchCriteria(companyID int, term string) ([]*customer, error) {

	rows, err := s.db.Query("SELECT c.id, c.name, c.contact_name, c.phone, c.email, c.amount_due "+
		"FROM customers c "+
		"INNER JOIN companies ON (c.company_id = companies.id) "+
		"WHERE c.company_id = $1 "+
		"AND c.name LIKE $2 "+
		"AND c.deleted_at IS NULL AND c.status = 'enabled' LIMIT 5", companyID, "%"+term+"%")
	if err != nil {
		return nil, err
	}
	data := make([]*customer, 0)
	for rows.Next() {
		row := new(customer)
		if err = rows.Scan(
			&row.ID,
			&row.Name,
			&row.ContactName,
			&row.Phone,
			&row.Email,
			&row.AmountDue,
		); err != nil {
			return data, err
		}

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

	_, err := tx.Exec("UPDATE customers SET amount_due = amount_due + $3 WHERE company_id = $1 AND id = $2",
		companyId, customerId, amountDue,
	)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Fatalf("Error updating customer amount due: %v", txErr)
			return txErr
		}
	}

	return nil
}
