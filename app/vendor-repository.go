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
	err := s.db.QueryRow("SELECT v.id, v.uuid, v.code, v.name, v.contact_name, v.phone, v.email, v.status, v.amount_payable, v.purchase_note, v.lead_time_days, v.address, "+
		"v.created_at, v.updated_at, v.deleted_at "+
		"FROM vendors v "+
		"INNER JOIN companies ON (v.company_id = companies.id) "+
		"WHERE v.company_id = $1 "+
		"AND v.id = $2 "+
		"AND v.deleted_at IS NULL", CurrentCompany(ctx).ID, vendorID).
		Scan(&v.ID, &v.UUID, &v.Code, &v.Name, &v.ContactName, &v.Phone, &v.Email, &v.Status, &v.AmountPayable, &v.PurchaseNote, &v.LeadTimeDays, &v.Address, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func (s *Server) findVendorByUUID(ctx context.Context, vendorID string) (*vendor, error) {

	var v vendor
	var ob OpenBalance
	err := s.db.QueryRow(
		"SELECT v.id, v.uuid, v.code, v.name, v.contact_name, v.phone, v.email, v.status, v.amount_payable, v.purchase_note, v.lead_time_days, "+
			"ap.id as invoice_id, ap.invoice_date, ap.amount_total, v.vendor_type, v.payment_method, v.payment_terms, v.address, "+
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
			&v.VendorType, &v.PaymentMethod, &v.PaymentTerms, &v.Address,
			&v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
		)
	if err != nil {
		return nil, err
	}

	v.OpenBalance = &ob

	return &v, nil
}

func (s *Server) findVendors(ctx context.Context, vendorType VendorType) ([]*vendor, error) {

	rows, err := s.db.Query("SELECT v.id, v.uuid, v.code, v.name, v.contact_name, v.phone, v.email, v.status, v.amount_payable, v.purchase_note, v.lead_time_days, "+
		"v.vendor_type, v.payment_method, v.payment_terms, v.address, v.created_at, v.updated_at, v.deleted_at "+
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
			&row.Address,
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

func (s *Server) findVendorsBySearchCriteria(ctx context.Context, term string) ([]*vendor, error) {
	if len(strings.TrimSpace(term)) == 0 {
		return nil, errors.New("need to specifiy the vendor you're looking for")
	}
	rows, err := s.db.Query("SELECT v.id, v.uuid, v.code, v.name, v.contact_name, v.phone, v.email, v.amount_payable, v.purchase_note, v.lead_time_days, "+
		"v.vendor_type, v.payment_method, v.payment_terms, v.address, v.created_at, v.updated_at, v.deleted_at "+
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
			&row.Address,
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

func (s *Server) storeVendor(ctx context.Context, form *StoreVendorForm) error {
	return database.WithTransaction(s.db, func(tx *sql.Tx) error {
		companyID := CurrentCompany(ctx).ID

		seqInfo, err := GetNextSequence(tx, companyID, "vendor")
		if err != nil {
			return err
		}

		createdBy := AuthUserFromContext(ctx).GetAuthIdentifier()
		return s.storeVendorInternal(tx, companyID, seqInfo.Code, createdBy, form)
	})
}

func (s *Server) storeVendorInternal(tx *sql.Tx, companyID int, code string, createdBy int, form *StoreVendorForm) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}
	vend := &vendorInsert{
		CompanyID:     companyID,
		Name:          form.Name,
		ContactName:   form.Contact,
		Email:         form.Email,
		Phone:         form.Phone,
		PaymentMethod: form.PaymentMethod,
		PaymentTerms:  form.PaymentTerms,
		PurchaseNote:  form.PurchaseNote,
		LeadTimeDays:  form.LeadTimeDays,
		AmountPayable: form.OpenBalance,
		VendorType:    form.VendorType,
		Code:          code,
		Address:       form.Address,
	}
	if err = ptx.Insert(context.Background(), vend); err != nil {
		return err
	}
	vendorID := int(vend.ID)

	if form.OpenBalance == 0 || form.OpenBalanceAsOf.IsZero() {
		return nil
	}

	return s.storeVendorOpenBalance(tx, companyID, vendorID, createdBy, form)
}

// storeVendorOpenBalance inserts an opening balance entry into accounts_payable.
// This represents a pre-existing liability the company owes the vendor
// before they started using the system.
func (s *Server) storeVendorOpenBalance(tx *sql.Tx, companyID int, vendorID int, createdBy int, form *StoreVendorForm) error {
	// Opening balance invoice number: OB-{vendorID}-{YYYYMMDD}
	// Guaranteed unique per vendor since a vendor can only have one opening balance.
	invoiceNumber := fmt.Sprintf("OB-%d-%s", vendorID, form.OpenBalanceAsOf.Format("20060102"))

	ptx, err := playTx(tx)
	if err != nil {
		return err
	}
	// due_date == invoice_date: an opening balance is overdue by definition.
	// amount_payable is a generated column and is not written here.
	ap := &openingPayableInsert{
		CompanyID:      companyID,
		VendorID:       vendorID,
		InvoiceNumber:  invoiceNumber,
		InvoiceDate:    form.OpenBalanceAsOf,
		DueDate:        form.OpenBalanceAsOf,
		AmountTotal:    form.OpenBalance,
		TaxAmount:      0,
		DiscountAmount: 0,
		AmountPaid:     0,
		Currency:       "USD",
		PaymentTerms:   form.PaymentTerms,
		PaymentMethod:  form.PaymentMethod,
		Status:         PayableStatuses.Pending,
		PaidStatus:     PaidStatuses.UnPaid,
		Notes:          "Saldo inicial",
		CreatedBy:      createdBy,
	}
	if err = ptx.Insert(context.Background(), ap); err != nil {
		return err
	}
	apID := int(ap.ID)

	return s.registerPayable(tx, companyID, apID, vendorID)
}

func (s *Server) registerPayable(tx *sql.Tx, companyID int, apID, vendorID int) error {
	ptx, err := playTx(tx)
	if err != nil {
		return err
	}
	return ptx.Insert(context.Background(), &payableRegister{
		CompanyID:         companyID,
		AccountsPayableID: apID,
		VendorID:          vendorID,
	})
}

func (s *Server) updateVendor(ctx context.Context, vendorID int, form *UpdateVendorForm) error {

	_, err := s.db.Exec(
		"UPDATE vendors SET name = $1, contact_name = $2, email = $3, phone = $4, payment_method = $5, payment_terms = $6, vendor_type = $7, purchase_note = $8, lead_time_days = $9, address = $10 WHERE company_id = $11 AND id = $12",
		form.Name, form.Contact, form.Email, form.Phone, form.PaymentMethod, form.PaymentTerms, form.VendorType, form.PurchaseNote, form.LeadTimeDays, form.Address, CurrentCompany(ctx).ID, vendorID,
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
