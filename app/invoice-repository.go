package app

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
)

type LineAction string

const (
	ADDED     LineAction = "added"
	UPDATED   LineAction = "updated"
	DELETED   LineAction = "deleted"
	UNCHANGED LineAction = "unchanged"
)

type invoice struct {
	ID           int        `json:"id"`
	UUID         string     `json:"uuid"`
	Number       string     `json:"number"`
	NCF          string     `json:"ncf"`
	Customer     customer   `json:"customer"`
	Date         time.Time  `json:"date"`
	DueOn        *time.Time `json:"due_on"`
	Terms        int        `json:"terms"`
	TaxReceiptID int        `json:"tax_receipt_id"`
	Amount       float64    `json:"amount"`
	Discount     Discount   `json:"discount"`
	Tax          float64    `json:"tax"`
	Total        float64    `json:"total"`
	Status       string     `json:"status"`
	PaidStatus   PaidStatus `json:"paid_status"`
	Payment      Payment    `json:"payment"`
	Notes        string     `json:"notes"`
}

type line struct {
	ID          int64   `json:"id"`
	Qty         int64   `json:"qty"`
	Price       float64 `json:"price"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Unit        struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"unit"`
	// Add timestamps properties
	foundation.Timestamps
	Action LineAction `json:"action"`
}

func (s *Server) findInvoices(companyId int) ([]*invoice, error) {
	rows, err := s.db.Query("SELECT invoices.id, invoices.uuid, invoices.date, invoices.due_on, invoices.amount, invoices.discount, invoices.tax, "+
		"invoices.total, invoices.status, invoices.paid_status, invoices.payment, invoices.note, invoices.tax_receipt_id,"+
		"tax_receipts.series || tax_receipts.type || LPAD(invoices.tax_receipt_sequence::varchar,8,'0') as NCF, "+
		"customers.id as customer, customers.name, customers.email, customers.phone "+
		"FROM invoices "+
		"INNER JOIN companies ON (invoices.company_id = companies.id) "+
		"INNER JOIN customers ON (invoices.company_id = customers.company_id AND invoices.customer_id = customers.id) "+
		"INNER JOIN tax_receipts ON (invoices.company_id = tax_receipts.company_id AND invoices.tax_receipt_id = tax_receipts.id) "+
		"WHERE invoices.company_id = $1", companyId)
	if err != nil {
		return nil, err
	}
	data := make([]*invoice, 0)
	for rows.Next() {
		i := new(invoice)

		if err = rows.Scan(
			&i.ID,
			&i.UUID,
			&i.Date,
			&i.DueOn,
			&i.Amount,
			&i.Discount,
			&i.Tax,
			&i.Total,
			&i.Status,
			&i.PaidStatus,
			&i.Payment,
			&i.Notes,
			&i.TaxReceiptID,
			&i.NCF,
			&i.Customer.ID,
			&i.Customer.Name,
			&i.Customer.Email,
			&i.Customer.Phone,
		); err != nil {
			return nil, err
		}
		i.Number = s.generatePrefixedInvoiceNumber(i.ID)
		i.Customer.Address = "LOUISVILLE, Selby 3864 Johnson Street, United States of America"

		data = append(data, i)
	}
	return data, nil
}

func (s *Server) findInvoicesByUUID(companyId int, uuid string) (*invoice, error) {
	i := new(invoice)
	err := s.db.QueryRow("SELECT invoices.id, invoices.uuid, invoices.date, invoices.due_on, invoices.amount, invoices.discount, invoices.tax, "+
		"invoices.total, invoices.status, invoices.paid_status, invoices.payment, invoices.note, invoices.tax_receipt_id, "+
		"tax_receipts.series || tax_receipts.type || LPAD(invoices.tax_receipt_sequence::varchar,8,'0') as NCF, invoices.note, "+
		"customers.id as customer, customers.name, customers.email, customers.phone "+
		"FROM invoices "+
		"INNER JOIN companies ON (invoices.company_id = companies.id) "+
		"INNER JOIN customers ON (invoices.company_id = customers.company_id AND invoices.customer_id = customers.id) "+
		"INNER JOIN tax_receipts ON (invoices.company_id = tax_receipts.company_id AND invoices.tax_receipt_id = tax_receipts.id) "+
		"WHERE invoices.company_id = $1 AND invoices.uuid = $2", companyId, uuid).
		Scan(
			&i.ID,
			&i.UUID,
			&i.Date,
			&i.DueOn,
			&i.Amount,
			&i.Discount,
			&i.Tax,
			&i.Total,
			&i.Status,
			&i.PaidStatus,
			&i.Payment,
			&i.Notes,
			&i.TaxReceiptID,
			&i.NCF,
			&i.Notes,
			&i.Customer.ID,
			&i.Customer.Name,
			&i.Customer.Email,
			&i.Customer.Phone)
	if err != nil {
		return nil, err
	}
	i.Terms = 1
	if i.DueOn != nil {
		difference := i.DueOn.Sub(i.Date)
		// Difference in days
		i.Terms = int(difference.Hours()) / 24
	}

	i.Number = s.generatePrefixedInvoiceNumber(i.ID)
	i.Customer.Address = "LOUISVILLE, Selby 3864 Johnson Street, United States of America"

	return i, nil
}

func (s *Server) generatePrefixedInvoiceNumber(value int) string {
	return foundation.GeneratePrefixedNumber("INV-", 10, value)
}

func (s *Server) storeInvoice(companyID int, form StoreInvoiceForm) error {

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	taxReceiptSequence, err := s.grabTaxReceiptSequence(tx, companyID, form.TaxReceipt)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO invoices (company_id, tax_receipt_id, tax_receipt_sequence, date, type, due_on, customer_id, amount, discount, tax, amount_due, total, note, paid_status, payment) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id")
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
		form.TaxReceipt,
		taxReceiptSequence,
		form.Date,
		form.termType,
		&form.dueOn,
		form.CustomerID,
		form.amount,
		foundation.ToJSON(form.Discount),
		form.tax,
		form.amountDue,
		form.total,
		form.Notes,
		form.paidStatus,
		foundation.ToJSON(form.Payment),
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

		if err = s.updateCustomerAmountDue(tx, companyID, form.CustomerID, form.amountDue); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Server) updateInvoice(companyID int, uuid string, form UpdateInvoiceForm) error {
	invoice, err := s.findInvoicesByUUID(companyID, uuid)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	// defer tx.Rollback()

	_, err = tx.Exec(`
    UPDATE invoices
    SET customer_id = $3, date = $4, due_on = $5, amount = $6, discount = $7, tax = $8, total = $9, amount_due = $10, note = $11
    WHERE company_id = $1 AND id = $2
  `,
		companyID, invoice.ID, form.CustomerID, form.Date, form.dueOn, form.amount, foundation.ToJSON(form.Discount), form.tax, form.total, form.amountDue, form.Notes,
	)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Fatalf("Error updating invoice: %v", txErr)
			return txErr
		}

		return err
	}

	if err = s.processInvoiceLines(tx, companyID, invoice.ID, form); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Fatalf("Error updating invoice: %v", txErr)
			return txErr
		}

		return err
	}

	if invoice.Customer.ID != form.CustomerID {
		// Update customer balance. Logs this operations to keep track of it.
		err = s.updateCustomerAmountDue(tx, companyID, invoice.Customer.ID, -invoice.Amount)
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				log.Fatalf("Error updating invoice: %v", txErr)
				return txErr
			}

			return err
		}

		err = s.updateCustomerAmountDue(tx, companyID, form.CustomerID, form.total)
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				log.Fatalf("Error updating invoice: %v", txErr)
				return txErr
			}

			return err
		}
	}

	return tx.Commit()
}

func (s *Server) attachInvoiceLines(tx *sql.Tx, companyId, invoiceId int, form StoreInvoiceForm) error {
	vals := []any{}
	for _, line := range form.Lines {
		vals = append(vals, companyId, invoiceId, line.ID, line.Unit, line.Qty, line.Price, line.Rate)
	}

	stmt := "INSERT INTO invoices_items (company_id, invoice_id, item_id, unit_id, qty, price, tax) VALUES "
	stmt += database.PrepareBulkInsert(7, len(form.Lines))

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

func (s *Server) processInvoiceLines(tx *sql.Tx, companyId, invoiceId int, form UpdateInvoiceForm) error {

	lines := s.filterInvoiceLines(form.Lines, ADDED, UPDATED, DELETED)
	for _, line := range lines {
		switch line.Action {
		case ADDED:
			stmt := "INSERT INTO invoices_items (company_id, invoice_id, item_id, unit_id, qty, price, tax) VALUES($1,$2,$3,$4,$5,$6,$7) "
			if _, err := tx.Exec(stmt, companyId, invoiceId, line.ID, line.Unit, line.Qty, line.Price, line.Rate); err != nil {
				return err
			}
		case UPDATED:
			stmt := "UPDATE invoices_items SET qty = $4, unit_id = $5 WHERE company_id = $1 AND invoice_id = $2 AND item_id = $3 "
			if _, err := tx.Exec(stmt, companyId, invoiceId, line.ID, line.Qty, line.Unit); err != nil {
				return err
			}
		case DELETED:
			stmt := "DELETE FROM invoices_items WHERE company_id = $1 AND invoice_id = $2 AND item_id = $3"
			if _, err := tx.Exec(stmt, companyId, invoiceId, line.ID); err != nil {
				return err
			}
		default:
			fmt.Println("Nothing to happen here.")
		}
	}
	return nil
}

func (s *Server) findInvoiceLines(companyId int, invoiceId int) ([]*line, error) {
	rows, err := s.db.Query(`
    SELECT ii.item_id, ii.qty, ii.price, iu.unit_id, it.name, it.description, u.name,
    ii.created_at, ii.updated_at, ii.deleted_at, 'unchanged' as action
    FROM invoices_items AS ii
    INNER JOIN companies AS com ON (ii.company_id = com.id)
    INNER JOIN invoices AS i ON (ii.invoice_id = i.id AND ii.company_id = i.company_id)
    INNER JOIN items AS it ON(ii.item_id = it.id AND ii.company_id = it.company_id)
    INNER JOIN items_units AS iu ON(ii.unit_id = iu.unit_id AND ii.item_id = iu.item_id AND ii.company_id = iu.company_id)
    INNER JOIN units AS u ON (iu.unit_id = u.id AND iu.company_id = u.company_id)
    WHERE ii.company_id = $1
    AND ii.invoice_id = $2`, companyId, invoiceId)
	if err != nil {
		return nil, err
	}
	data := make([]*line, 0)
	for rows.Next() {
		i := new(line)

		if err = rows.Scan(
			&i.ID,
			&i.Qty,
			&i.Price,
			&i.Unit.ID,
			&i.Name,
			&i.Description,
			&i.Unit.Name,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
			&i.Action,
		); err != nil {
			return nil, err
		}

		data = append(data, i)
	}
	return data, nil
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

func (s *Server) filterInvoiceLines(lines []Line, actions ...LineAction) []Line {
	if len(actions) == 0 {
		return nil
	}

	// Convert actions to a lookup map for O(1) checks
	actionSet := make(map[string]struct{}, len(actions))
	for _, a := range actions {
		actionSet[string(a)] = struct{}{}
	}

	// Filter lines
	filtered := make([]Line, 0, len(lines))
	for _, line := range lines {
		if _, ok := actionSet[string(line.Action)]; ok {
			filtered = append(filtered, line)
		}
	}

	return filtered
}
