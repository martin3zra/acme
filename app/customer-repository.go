package app

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
)

type OpenBalance struct {
	InvoiceID *int       `json:"invoice_id"`
	Date      *time.Time `json:"date"`
	Amount    *float64   `json:"amount"`
}

type customer struct {
	ID            int          `json:"id"`
	UUID          string       `json:"uuid"`
	Code          string       `json:"code"`
	Name          string       `json:"name"`
	ContactName   string       `json:"contact_name,omitempty,"`
	Phone         string       `json:"phone"`
	Email         string       `json:"email"`
	AmountDue     float64      `json:"amount_due,omitempty,"`
	Address       string       `json:"address"`
	CustomerType  string       `json:"customer_type"`
	PaymentMethod string       `json:"payment_method"`
	CreditLimited bool         `json:"credit_limited"`
	CreditLimit   float64      `json:"credit_limit"`
	PaymentTerms  string       `json:"payment_terms"`
	TaxReceipt    *int         `json:"tax_receipt"`
	OpenBalance   *OpenBalance `json:"open_balance"`
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
		NCF        *string    `json:"ncf"`
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

func (s *Server) findCustomeByID(ctx context.Context, customerID int) (*customer, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	// The old query INNER JOINed companies purely to assert the tenant row
	// exists; the company_id predicate already scopes it, so the join is
	// dropped. The softdelete tag on customerRead adds "deleted_at IS NULL".
	var row customerRead
	err = pdb.Model(&customerRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", customerID).
		First(ctx, &row)
	if err != nil {
		return nil, err
	}

	return row.toCustomer(), nil
}

// findCustomeByUUID stays on raw database/sql: it LEFT JOINs a derived
// "opening balance" invoice subquery onto the customer and reads its columns
// into an OpenBalance struct — a correlated join to a filtered subquery that
// playsql's model/relation reads don't express. See playsql-phase2 notes.
func (s *Server) findCustomeByUUID(ctx context.Context, customerID string) (*customer, error) {

	var c customer
	var ob OpenBalance
	err := s.db.QueryRow("SELECT c.id, c.uuid, c.code, c.name, c.contact_name, c.phone, c.email, c.status, c.amount_due, "+
		"invoices.id as invoice_id, invoices.date, invoices.amount, c.customer_type, c.payment_method, c.credit_limited, c.credit_limit, c.payment_terms, c.tax_receipt_id, c.address, "+
		"c.created_at, c.updated_at, c.deleted_at "+
		"FROM customers c "+
		"INNER JOIN companies ON (c.company_id = companies.id) "+
		"LEFT JOIN (SELECT id, customer_id, date, amount FROM invoices WHERE invoices.type = 'opening'::invoice_terms) invoices ON invoices.customer_id = c.id "+
		"WHERE c.company_id = $1 "+
		"AND c.uuid = $2 "+
		"AND c.deleted_at IS NULL", CurrentCompany(ctx).ID, customerID).
		Scan(&c.ID, &c.UUID, &c.Code, &c.Name, &c.ContactName, &c.Phone, &c.Email, &c.Status, &c.AmountDue, &ob.InvoiceID, &ob.Date, &ob.Amount, &c.CustomerType,
			&c.PaymentMethod,
			&c.CreditLimited,
			&c.CreditLimit,
			&c.PaymentTerms,
			&c.TaxReceipt, &c.Address, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
	if err != nil {
		return nil, err
	}

	c.OpenBalance = &ob

	return &c, nil
}

func (s *Server) findCustomers(ctx context.Context, customerType CustomerType) ([]*customer, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []customerRead
	if err := pdb.Model(&customerRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		Unless(customerType == "all", func(q *playsql.Builder) {
			q.WhereEq("customer_type", string(customerType))
		}).
		OrderBy("name", playsql.Asc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*customer, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toCustomer())
	}
	return data, nil
}

func (s *Server) findCustomersBySearchCriteria(ctx context.Context, term string) ([]*customer, error) {
	if len(strings.TrimSpace(term)) == 0 {
		return nil, errors.New("need to specifiy the customer you're looking for")
	}
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []customerRead
	err = pdb.Model(&customerRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		Where("name", "ILIKE", "%"+term+"%").
		WhereEq("status", "enabled").
		OrderBy("name", playsql.Asc).
		Limit(5).
		Get(ctx, &rows)
	if err != nil {
		return nil, err
	}

	data := make([]*customer, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toCustomer())
	}
	return data, nil
}

func (s *Server) storeCustomer(ctx context.Context, form *StoreCustomerForm) error {
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		companyID := CurrentCompany(ctx).ID

		seqInfo, err := GetNextSequence(tx, companyID, "customer")
		if err != nil {
			return err
		}

		return s.storeCustomerInternal(tx, companyID, seqInfo.Code, form)
	})
}

func (s *Server) storeCustomerInternal(tx *sql.Tx, companyID int, code string, form *StoreCustomerForm) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}
	cust := &customerInsert{
		CompanyID:     companyID,
		Name:          form.Name,
		ContactName:   form.Contact,
		Email:         form.Email,
		Phone:         form.Phone,
		PaymentMethod: form.PaymentMethod,
		PaymentTerms:  form.PaymentTerms,
		CreditLimited: form.CreditLimited,
		CreditLimit:   form.CreditLimit,
		AmountDue:     form.OpenBalance,
		CustomerType:  form.CustomerType,
		TaxReceiptID:  form.TaxReceipt,
		Code:          code,
		Address:       form.Address,
	}
	if err = ptx.Insert(context.Background(), cust); err != nil {
		return err
	}
	customerID := int(cust.ID)

	if form.OpenBalance == 0 || form.OpenBalanceAsOf.IsZero() {
		return nil
	}

	// The opening-balance invoice needs a code like any other invoice
	// (invoices.code is NOT NULL); use the credit-invoice sequence.
	openingSeq, err := GetNextSequence(tx, companyID, "invoice.credit")
	if err != nil {
		return err
	}

	opening := &openingInvoiceInsert{
		CompanyID:  companyID,
		Date:       form.OpenBalanceAsOf,
		Type:       InvoiceTermType.Opening,
		DueOn:      form.OpenBalanceAsOf,
		CustomerID: customerID,
		Amount:     form.OpenBalance,
		AmountDue:  form.OpenBalance,
		Total:      form.OpenBalance,
		Note:       "Saldo inicial",
		Status:     InvoiceStatuses.Sent,
		PaidStatus: PaidStatuses.UnPaid,
		Code:       openingSeq.Code,
	}
	if err := ptx.Insert(context.Background(), opening); err != nil {
		return err
	}
	invoiceID := int(opening.ID)

	return s.registerReceivable(tx, companyID, invoiceID, customerID)
}

func (s *Server) updateCustomer(ctx context.Context, customerID int, form *UpdateCustomerForm) error {

	res, err := s.db.Exec(
		"UPDATE customers SET name = $1, contact_name = $2,  email = $3, phone = $4, payment_method = $5, payment_terms = $6, credit_limit = $7, customer_type = $8, tax_receipt_id = $9, credit_limited = $10, address = $11 WHERE company_id = $12 AND id = $13",
		form.Name, form.Contact, form.Email, form.Phone, form.PaymentMethod, form.PaymentTerms, form.CreditLimit, form.CustomerType, form.TaxReceipt, form.CreditLimited, form.Address, CurrentCompany(ctx).ID, customerID,
	)

	return mustAffectRow(res, err, "customer")
}

func (s *Server) deleteCustomer(ctx context.Context, customerID int) error {
	res, err := s.db.Exec(
		"UPDATE customers SET deleted_at = now(), updated_at = now() WHERE company_id = $1 AND id = $2",
		CurrentCompany(ctx).ID, customerID,
	)

	return mustAffectRow(res, err, "customer")
}

func (s *Server) toggleCustomerStatus(ctx context.Context, customer *customer) error {
	status := customer.Status
	if status == "enabled" {
		status = "disabled"
	} else {
		status = "enabled"
	}
	res, err := s.db.Exec(
		"UPDATE customers SET updated_at = now(), status = $3 WHERE company_id = $1 AND id = $2",
		CurrentCompany(ctx).ID, customer.ID, status,
	)

	return mustAffectRow(res, err, "customer")
}

func (s *Server) updateCustomerAmountDue(tx *sql.Tx, companyId, customerId int, amountDue float64) error {

	result, err := tx.Exec("UPDATE customers SET amount_due = amount_due + $3 WHERE company_id = $1 AND id = $2",
		companyId, customerId, amountDue,
	)
	if err != nil {
		return err
	}

	if affected, err := result.RowsAffected(); err == nil {
		if affected != 1 {
			return errors.New("unable to update customer balance") //new(ErrUnprocessableEntity)
		}
	}

	return err
}

// TODO we set the LEFT JOIN check for NULL values. on Tax receipt.
func (s *Server) findCustomeReceivables(ctx context.Context, customerID string) ([]*receivable, error) {
	rows, err := s.db.Query(`
    SELECT receivables.id, receivables.uuid, invoices.uuid, invoices.id, invoices.code,
    invoices.date, invoices.due_on, invoices.total, invoices.amount_due, invoices.paid_status, invoices.tax_number
		FROM receivables
		INNER JOIN companies ON (receivables.company_id = companies.id)
		INNER JOIN invoices ON (receivables.company_id = invoices.company_id AND receivables.customer_id = invoices.customer_id AND receivables.invoice_id = invoices.id)
    LEFT JOIN tax_receipts ON (invoices.company_id = tax_receipts.company_id AND invoices.tax_receipt_id = tax_receipts.id)
		INNER JOIN customers ON (receivables.company_id = customers.company_id AND receivables.customer_id = customers.id)
		WHERE receivables.company_id = $1
    AND invoices.paid_status != 'paid'::paid_status
    AND customers.uuid = $2
		AND receivables.deleted_at IS NULL
  `, CurrentCompany(ctx).ID, customerID)
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
			&row.Invoice.Number,
			&row.Invoice.Date,
			&row.Invoice.DueOn,
			&row.Invoice.Total,
			&row.Invoice.AmountDue,
			&row.Invoice.PaidStatus,
			&row.Invoice.NCF,
		); err != nil {
			return data, err
		}

		data = append(data, row)
	}

	return data, nil
}
