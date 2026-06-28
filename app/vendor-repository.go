package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
)

type vendor struct {
	ID            int     `json:"id"`
	UUID          string  `json:"uuid"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	ContactName   string  `json:"contact_name,omitempty"`
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	PurchaseNote  string  `json:"purchase_note"`
	LeadTimeDays  int     `json:"lead_time_days"`
	AmountPayable float64 `json:"amount_payable"`
	Address       string  `json:"address"`
	VendorType    string  `json:"vendor_type"`
	PaymentMethod string  `json:"payment_method"`
	// CreditLimited bool         `json:"credit_limited"`
	// CreditLimit   float64      `json:"credit_limit"`
	PaymentTerms string `json:"payment_terms"`
	// TaxReceipt    *int         `json:"tax_receipt"`
	OpenBalance *OpenBalance `json:"open_balance"`
	foundation.Timestamps
	Status foundation.Status `json:"status"`
}

type Payable struct {
	ID        int64  `json:"id"`
	UUID      string `json:"uuid"`
	CompanyID int64  `json:"company_id"`

	// Payables join fields
	InvoiceID   int64  `json:"invoice_id"`
	InvoiceUUID string `json:"invoice_uuid"`

	// Invoice identifiers
	InvoiceNumber   string `json:"invoice_number"`
	PurchaseOrderID *int64 `json:"purchase_order_id,omitempty"`

	// Related entities
	VendorID   int64  `json:"vendor_id"`
	ApprovedBy *int64 `json:"approved_by,omitempty"`
	CreatedBy  int64  `json:"created_by"`

	// Dates
	InvoiceDate time.Time  `json:"invoice_date"`
	DueDate     time.Time  `json:"due_date"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`

	// Amounts
	AmountTotal    float64 `json:"amount_total"`
	TaxAmount      float64 `json:"tax_amount"`
	DiscountAmount float64 `json:"discount_amount"`
	AmountPayable  float64 `json:"amount_payable"`
	AmountPaid     float64 `json:"amount_paid"`

	// Payment details
	CurrencyCode  string  `json:"currency_code"`
	PaymentTerms  *string `json:"payment_terms,omitempty"`
	PaymentMethod *string `json:"payment_method,omitempty"`

	// Status & notes
	Status     PayableStatus `json:"status"`
	PaidStatus PaidStatus    `json:"paid_status"`
	Notes      *string       `json:"notes,omitempty"`

	foundation.Timestamps
}

func (s *Server) findVendorByID(ctx context.Context, vendorID int) (*vendor, error) {

	var v vendor
	err := s.db.QueryRow("SELECT v.id, v.uuid, v.code, v.name, v.contact_name, v.phone, v.email, v.status, v.amount_payable, v.purchase_note, v.lead_time_days, "+
		"v.created_at, v.updated_at, v.deleted_at "+
		"FROM vendors v "+
		"INNER JOIN companies ON (v.company_id = companies.id) "+
		"WHERE v.company_id = $1 "+
		"AND v.id = $2 "+
		"AND v.deleted_at IS NULL", CurrentCompany(ctx).ID, vendorID).
		Scan(&v.ID, &v.UUID, &v.Code, &v.Name, &v.ContactName, &v.Phone, &v.Email, &v.Status, &v.AmountPayable, &v.PurchaseNote, &v.LeadTimeDays, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt)
	if err != nil {
		return nil, err
	}

	v.Address = "LOUISVILLE, Selby 3864 Johnson Street, United States of America"
	return &v, nil
}

func (s *Server) findVendorByUUID(ctx context.Context, vendorID string) (*vendor, error) {

	var v vendor
	var ob OpenBalance
	err := s.db.QueryRow(
		"SELECT v.id, v.uuid, v.code, v.name, v.contact_name, v.phone, v.email, v.status, v.amount_payable, v.purchase_note, v.lead_time_days, "+
			"ap.id as invoice_id, ap.invoice_date, ap.amount_total, v.vendor_type, v.payment_method, v.payment_terms, "+
			"v.created_at, v.updated_at, v.deleted_at "+
			"FROM vendors v "+
			"INNER JOIN companies ON (v.company_id = companies.id) "+
			"LEFT JOIN (SELECT id, vendor_id, invoice_date, amount_total FROM accounts_payable WHERE accounts_payable.status = 'draft') ap ON ap.vendor_id = v.id "+
			"WHERE v.company_id = $1 "+
			"AND v.uuid = $2 "+
			"AND v.deleted_at IS NULL",
		CurrentCompany(ctx).ID, vendorID).
		Scan(
			&v.ID, &v.UUID, &v.Code, &v.Name, &v.ContactName, &v.Phone, &v.Email,
			&v.Status, &v.AmountPayable, &v.PurchaseNote, &v.LeadTimeDays,
			&ob.InvoiceID, &ob.Date, &ob.Amount,
			&v.VendorType, &v.PaymentMethod, &v.PaymentTerms,
			&v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
		)
	if err != nil {
		return nil, err
	}

	v.OpenBalance = &ob

	v.Address = "LOUISVILLE, Selby 3864 Johnson Street, United States of America"
	return &v, nil
}

func (s *Server) findVendors(ctx context.Context, vendorType VendorType) ([]*vendor, error) {

	rows, err := s.db.Query("SELECT v.id, v.uuid, v.code, v.name, v.contact_name, v.phone, v.email, v.status, v.amount_payable, v.purchase_note, v.lead_time_days, "+
		"v.vendor_type, v.payment_method, v.payment_terms, v.created_at, v.updated_at, v.deleted_at "+
		"FROM vendors v "+
		"INNER JOIN companies ON (v.company_id = companies.id) "+
		"WHERE v.company_id = $1 "+
		"AND v.deleted_at IS NULL AND ($2 = 'all' OR v.vendor_type = $2::vendor_types) ORDER BY v.name", CurrentCompany(ctx).ID, vendorType)
	if err != nil {
		return nil, err
	}
	data := make([]*vendor, 0)
	for rows.Next() {
		row := new(vendor)
		if err = rows.Scan(
			&row.ID,
			&row.UUID,
			&row.Code,
			&row.Name,
			&row.ContactName,
			&row.Phone,
			&row.Email,
			&row.Status,
			&row.AmountPayable,
			&row.PurchaseNote,
			&row.LeadTimeDays,
			&row.VendorType,
			&row.PaymentMethod,
			&row.PaymentTerms,
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

func (s *Server) findVendorsBySearchCriteria(ctx context.Context, term string) ([]*vendor, error) {
	if len(strings.TrimSpace(term)) == 0 {
		return nil, errors.New("need to specifiy the vendor you're looking for")
	}
	rows, err := s.db.Query("SELECT v.id, v.uuid, v.code, v.name, v.contact_name, v.phone, v.email, v.amount_payable, v.purchase_note, v.lead_time_days, "+
		"v.vendor_type, v.payment_method, v.payment_terms, v.created_at, v.updated_at, v.deleted_at "+
		"FROM vendors v "+
		"INNER JOIN companies ON (v.company_id = companies.id) "+
		"WHERE v.company_id = $1 "+
		"AND v.name ILIKE $2 "+
		"AND v.deleted_at IS NULL AND v.status = 'enabled' ORDER BY v.name LIMIT 5 ", CurrentCompany(ctx).ID, "%"+term+"%")
	if err != nil {
		return nil, err
	}
	data := make([]*vendor, 0)
	for rows.Next() {
		row := new(vendor)
		if err = rows.Scan(
			&row.ID,
			&row.UUID,
			&row.Code,
			&row.Name,
			&row.ContactName,
			&row.Phone,
			&row.Email,
			&row.AmountPayable,
			&row.PurchaseNote,
			&row.LeadTimeDays,
			&row.VendorType,
			&row.PaymentMethod,
			&row.PaymentTerms,
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

func (s *Server) storeVendor(ctx context.Context, form *StoreVendorForm) error {
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		companyID := CurrentCompany(ctx).ID

		seqInfo, err := GetNextSequence(tx, companyID, "vendor")
		if err != nil {
			return err
		}

		return s.storeVendorInternal(tx, companyID, seqInfo.Code, form)
	})
}

func (s *Server) storeVendorInternal(tx *sql.Tx, companyID int, code string, form *StoreVendorForm) error {
	stmt, err := tx.Prepare("INSERT INTO vendors (company_id, name, contact_name, email, phone, payment_method, payment_terms, purchase_note, lead_time_days, amount_payable, vendor_type, code) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id")
	if err != nil {
		return err
	}

	var vendorID int
	err = stmt.QueryRow(companyID, form.Name, form.Contact, form.Email, form.Phone, form.PaymentMethod, form.PaymentTerms, form.PurchaseNote, form.LeadTimeDays, form.OpenBalance, form.VendorType, code).Scan(&vendorID)
	if err != nil {
		return err
	}

	if form.OpenBalance == 0 || form.OpenBalanceAsOf.IsZero() {
		return nil
	}

	return s.storeVendorOpenBalance(tx, companyID, vendorID, 0, form)
}

// storeVendorOpenBalance inserts an opening balance entry into accounts_payable.
// This represents a pre-existing liability the company owes the vendor
// before they started using the system.
func (s *Server) storeVendorOpenBalance(tx *sql.Tx, companyID int, vendorID int, createdBy int, form *StoreVendorForm) error {
	stmt, err := tx.Prepare(
		"INSERT INTO accounts_payable " +
			"(company_id, vendor_id, invoice_number, invoice_date, due_date, " +
			"amount_total, tax_amount, discount_amount, amount_paid, " +
			"currency, payment_terms, payment_method, status, paid_status, notes, created_by) " +
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) " +
			"RETURNING id")
	if err != nil {
		return err
	}

	// Opening balance invoice number: OB-{vendorID}-{YYYYMMDD}
	// Guaranteed unique per vendor since a vendor can only have one opening balance.
	invoiceNumber := fmt.Sprintf("OB-%d-%s", vendorID, form.OpenBalanceAsOf.Format("20060102"))

	var apID int
	err = stmt.QueryRow(
		companyID,
		vendorID,
		invoiceNumber,
		form.OpenBalanceAsOf, // invoice_date
		form.OpenBalanceAsOf, // due_date — already overdue by definition
		form.OpenBalance,     // amount_total
		0,                    // tax_amount
		0,                    // discount_amount
		0,                    // amount_paid
		"USD",                // currency_code — adjust to your default
		form.PaymentTerms,
		form.PaymentMethod,
		PayableStatuses.Pending, // status: lifecycle
		PaidStatuses.UnPaid,     // paid_status: payment state
		"Saldo inicial",
		createdBy,
	).Scan(&apID)
	if err != nil {
		return err
	}

	return s.registerPayable(tx, companyID, apID, vendorID)
}

func (s *Server) registerPayable(tx *sql.Tx, companyID int, apID, vendorID int) error {
	_, err := tx.Exec(
		"INSERT INTO payables (company_id, accounts_payable_id, vendor_id) VALUES ($1, $2, $3)",
		companyID, apID, vendorID,
	)

	return err
}

func (s *Server) updateVendor(ctx context.Context, vendorID int, form *UpdateVendorForm) error {

	_, err := s.db.Exec(
		"UPDATE vendors SET name = $1, contact_name = $2, email = $3, phone = $4, payment_method = $5, payment_terms = $6, vendor_type = $7, purchase_note = $8, lead_time_days = $9 WHERE company_id = $10 AND id = $11",
		form.Name, form.Contact, form.Email, form.Phone, form.PaymentMethod, form.PaymentTerms, form.VendorType, form.PurchaseNote, form.LeadTimeDays, CurrentCompany(ctx).ID, vendorID,
	)

	return err
}

func (s *Server) deleteVendor(ctx context.Context, vendorID int) error {
	_, err := s.db.Exec(
		"UPDATE vendors SET deleted_at = now(), updated_at = now() WHERE company_id = $1 AND id = $2",
		CurrentCompany(ctx).ID, vendorID,
	)

	return err
}

func (s *Server) toggleVendorStatus(ctx context.Context, vendor *vendor) error {
	status := vendor.Status
	if status == "enabled" {
		status = "disabled"
	} else {
		status = "enabled"
	}
	_, err := s.db.Exec(
		"UPDATE vendors SET updated_at = now(), status = $3 WHERE company_id = $1 AND id = $2",
		CurrentCompany(ctx).ID, vendor.ID, status,
	)
	return err
}

func (s *Server) updateVendorAmountPayable(tx *sql.Tx, companyId, vendorId int, amountPayable float64) error {

	result, err := tx.Exec("UPDATE vendors SET amount_payable = amount_payable + $3 WHERE company_id = $1 AND id = $2",
		companyId, vendorId, amountPayable,
	)
	if err != nil {
		return err
	}

	if affected, err := result.RowsAffected(); err == nil {
		if affected != 1 {
			return errors.New("unable to update vendor balance") //new(ErrUnprocessableEntity)
		}
	}

	return err
}

func (s *Server) findVendorPayables(ctx context.Context, vendorID string) ([]*Payable, error) {
	rows, err := s.db.Query(`
		SELECT
			p.id, p.uuid,
			ap.uuid, ap.id, ap.invoice_number,
			ap.invoice_date, ap.due_date,
			ap.amount_total, ap.amount_payable, ap.amount_paid,
			ap.status, ap.paid_status, ap.notes
		FROM payables p
		INNER JOIN companies        ON (p.company_id = companies.id)
		INNER JOIN accounts_payable ap ON (p.company_id = ap.company_id AND p.vendor_id = ap.vendor_id AND p.accounts_payable_id = ap.id)
		INNER JOIN vendors          ON (p.company_id = vendors.company_id AND p.vendor_id = vendors.id)
		WHERE p.company_id = $1
		AND ap.paid_status != 'paid'
		AND vendors.uuid = $2
	`, CurrentCompany(ctx).ID, vendorID)
	if err != nil {
		return nil, err
	}

	data := make([]*Payable, 0)
	for rows.Next() {
		row := new(Payable)
		if err = rows.Scan(
			&row.ID,
			&row.UUID,
			&row.InvoiceUUID,
			&row.InvoiceID,
			&row.InvoiceNumber,
			&row.InvoiceDate,
			&row.DueDate,
			&row.AmountTotal,
			&row.AmountPayable,
			&row.AmountPaid,
			&row.Status,
			&row.PaidStatus,
			&row.Notes,
		); err != nil {
			return data, err
		}

		data = append(data, row)
	}

	return data, nil
}
