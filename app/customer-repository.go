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
	// dropped. The softdelete tag on customerModel adds "deleted_at IS NULL".
	var row customerModel
	err = pdb.Model(&customerModel{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", customerID).
		First(ctx, &row)
	if err != nil {
		return nil, err
	}

	return row.toCustomer(), nil
}

// findCustomeByUUID reads a customer with its opening balance.
//
// The raw query LEFT JOINed a subquery over invoices filtered to type = 'opening'.
// That is a hasOne with a constraint: withOpeningInvoice carries the filter, and the
// eager load reproduces the LEFT JOIN's semantics — a customer with no opening
// invoice still comes back, with an OpenBalance of nil pointers.
//
// INNER JOIN companies is dropped: existence-only, and company_id already scopes.
func (s *Server) findCustomeByUUID(ctx context.Context, customerID string) (*customer, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row customerModel
	if err := pdb.Model(&customerModel{}).
		WithConstraint("OpeningInvoice", withOpeningInvoice).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("uuid", customerID).
		First(ctx, &row); err != nil {
		return nil, err
	}

	c := row.toCustomer()
	c.OpenBalance = openBalanceOfInvoice(row.OpeningInvoice)
	return c, nil
}

func (s *Server) findCustomers(ctx context.Context, customerType CustomerType) ([]*customer, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []customerModel
	if err := pdb.Model(&customerModel{}).
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

	var rows []customerModel
	err = pdb.Model(&customerModel{}).
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
	// Map insert (not the struct form) so uuid stays unset and the DB default
	// fills it; the merged customerModel maps uuid, which a struct insert would
	// write as an empty string.
	id, err := ptx.Model(&customerModel{}).Insert(context.Background(), map[string]any{
		"company_id":     companyID,
		"name":           form.Name,
		"contact_name":   form.Contact,
		"email":          form.Email,
		"phone":          form.Phone,
		"payment_method": form.PaymentMethod,
		"payment_terms":  form.PaymentTerms,
		"credit_limited": form.CreditLimited,
		"credit_limit":   form.CreditLimit,
		"amount_due":     form.OpenBalance,
		"customer_type":  form.CustomerType,
		"tax_receipt_id": form.TaxReceipt,
		"code":           code,
		"address":        form.Address,
	})
	if err != nil {
		return err
	}
	customerID := int(id)

	if form.OpenBalance == 0 || form.OpenBalanceAsOf.IsZero() {
		return nil
	}

	// The opening-balance invoice needs a code like any other invoice
	// (invoices.code is NOT NULL); use the credit-invoice sequence.
	openingSeq, err := GetNextSequence(tx, companyID, "invoice.credit")
	if err != nil {
		return err
	}

	// Map insert so uuid stays unset for the DB default; the merged invoiceModel
	// maps uuid, which a struct insert would write as empty.
	invoiceID64, err := ptx.Model(&invoiceModel{}).Insert(context.Background(), map[string]any{
		"company_id":  companyID,
		"date":        form.OpenBalanceAsOf,
		"type":        InvoiceTermType.Opening,
		"due_on":      form.OpenBalanceAsOf,
		"customer_id": customerID,
		"amount":      form.OpenBalance,
		"amount_due":  form.OpenBalance,
		"total":       form.OpenBalance,
		"note":        "Saldo inicial",
		"status":      InvoiceStatuses.Sent,
		"paid_status": PaidStatuses.UnPaid,
		"code":        openingSeq.Code,
	})
	if err != nil {
		return err
	}
	invoiceID := int(invoiceID64)

	return s.registerReceivable(tx, companyID, invoiceID, customerID)
}

// updateCustomer, deleteCustomer and toggleCustomerStatus all pick up
// `deleted_at IS NULL` from customerModel's softdelete tag, which the raw statements
// lacked. Editing or re-deleting a soft-deleted customer is now a not-found.
func (s *Server) updateCustomer(ctx context.Context, customerID int, form *UpdateCustomerForm) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	affected, err := pdb.Model(&customerModel{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", customerID).
		Update(ctx, map[string]any{
			"name":           form.Name,
			"contact_name":   form.Contact,
			"email":          form.Email,
			"phone":          form.Phone,
			"payment_method": form.PaymentMethod,
			"payment_terms":  form.PaymentTerms,
			"credit_limit":   form.CreditLimit,
			"customer_type":  form.CustomerType,
			"tax_receipt_id": form.TaxReceipt,
			"credit_limited": form.CreditLimited,
			"address":        form.Address,
		})
	return mustAffectRows(affected, err, "customer")
}

func (s *Server) deleteCustomer(ctx context.Context, customerID int) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	affected, err := pdb.Model(&customerModel{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", customerID).
		Update(ctx, map[string]any{"deleted_at": time.Now()})
	return mustAffectRows(affected, err, "customer")
}

func (s *Server) toggleCustomerStatus(ctx context.Context, customer *customer) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	status := customer.Status
	if status == "enabled" {
		status = "disabled"
	} else {
		status = "enabled"
	}

	affected, err := pdb.Model(&customerModel{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", customer.ID).
		Update(ctx, map[string]any{"status": string(status)})
	return mustAffectRows(affected, err, "customer")
}

// Stays raw: `amount_due = amount_due + $3` is a self-referencing increment, which
// playsql's Update cannot express.
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

// findCustomeReceivables lists a customer's unpaid receivables with their invoice.
//
// Four of the old query's joins carried no weight:
//
//   - INNER JOIN companies asserted existence of a NOT NULL FK.
//   - LEFT JOIN tax_receipts contributed no column, no predicate, and being a LEFT
//     join could not filter either. It is what the "TODO check for NULL values"
//     referred to; the NCF comes off invoices.tax_number. Removed, not fixed:
//     nothing downstream reads a tax_receipts column.
//   - INNER JOIN customers only resolved the uuid, done here as its own read.
//   - receivables.customer_id = invoices.customer_id is redundant next to
//     receivables.invoice_id = invoices.id.
//
// The customer lookup keeps WithTrashed: the old join had no deleted_at predicate on
// customers, so a soft-deleted customer's receivables were still listed.
func (s *Server) findCustomeReceivables(ctx context.Context, customerID string) ([]*receivable, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	companyID := CurrentCompany(ctx).ID

	var c customerModel
	if err := pdb.Model(&customerModel{}).
		Select("id").
		WithTrashed().
		WhereEq("company_id", companyID).
		WhereEq("uuid", customerID).
		First(ctx, &c); err != nil {
		return nil, err
	}

	var rows []receivableRead
	if err := pdb.Model(&receivableRead{}).
		With("Invoice").
		WhereEq("company_id", companyID).
		WhereEq("customer_id", c.ID).
		WhereRelation("Invoice", "paid_status", "!=", "paid").
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*receivable, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toReceivable())
	}
	return data, nil
}
