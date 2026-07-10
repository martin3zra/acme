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
	"github.com/martin3zra/playsql"
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
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	// INNER JOIN companies dropped — company_id already scopes the row; the
	// softdelete tag on vendorRead adds "deleted_at IS NULL".
	var row vendorRead
	err = pdb.Model(&vendorRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", vendorID).
		First(ctx, &row)
	if err != nil {
		return nil, err
	}

	return row.toVendor(), nil
}

// findVendorByUUID stays on raw database/sql: it LEFT JOINs a filtered
// accounts_payable ("draft") subquery to derive the vendor's opening balance —
// a correlated join playsql's model/relation reads don't express. See
// playsql-phase2 notes.
// findVendorByUUID reads a vendor with its opening balance.
//
// The raw query LEFT JOINed a subquery over accounts_payable filtered to
// status = 'draft'. That is a hasOne with a constraint: withOpeningPayable carries the
// filter, and the eager load reproduces the LEFT JOIN's semantics — a vendor with no
// draft payable still comes back, with an OpenBalance of nil pointers.
//
// INNER JOIN companies is dropped: existence-only, and company_id already scopes.
func (s *Server) findVendorByUUID(ctx context.Context, vendorID string) (*vendor, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var row vendorRead
	if err := pdb.Model(&vendorRead{}).
		WithConstraint("OpeningPayable", withOpeningPayable).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("uuid", vendorID).
		First(ctx, &row); err != nil {
		return nil, err
	}

	v := row.toVendor()
	v.OpenBalance = openBalanceOfPayable(row.OpeningPayable)
	return v, nil
}

func (s *Server) findVendors(ctx context.Context, vendorType VendorType) ([]*vendor, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []vendorRead
	if err := pdb.Model(&vendorRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		Unless(vendorType == "all", func(q *playsql.Builder) {
			q.WhereEq("vendor_type", string(vendorType))
		}).
		OrderBy("name", playsql.Asc).
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*vendor, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toVendor())
	}
	return data, nil
}

func (s *Server) findVendorsBySearchCriteria(ctx context.Context, term string) ([]*vendor, error) {
	if len(strings.TrimSpace(term)) == 0 {
		return nil, errors.New("need to specifiy the vendor you're looking for")
	}
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	var rows []vendorRead
	err = pdb.Model(&vendorRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		Where("name", "ILIKE", "%"+term+"%").
		WhereEq("status", "enabled").
		OrderBy("name", playsql.Asc).
		Limit(5).
		Get(ctx, &rows)
	if err != nil {
		return nil, err
	}

	data := make([]*vendor, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toVendor())
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

// updateVendor, deleteVendor and toggleVendorStatus all pick up `deleted_at IS NULL`
// from vendorRead's softdelete tag, which the raw statements lacked. Editing or
// re-deleting a soft-deleted vendor is now a not-found.
func (s *Server) updateVendor(ctx context.Context, vendorID int, form *UpdateVendorForm) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	affected, err := pdb.Model(&vendorRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", vendorID).
		Update(ctx, map[string]any{
			"name":           form.Name,
			"contact_name":   form.Contact,
			"email":          form.Email,
			"phone":          form.Phone,
			"payment_method": form.PaymentMethod,
			"payment_terms":  form.PaymentTerms,
			"vendor_type":    form.VendorType,
			"purchase_note":  form.PurchaseNote,
			"lead_time_days": form.LeadTimeDays,
			"address":        form.Address,
		})
	return mustAffectRows(affected, err, "vendor")
}

func (s *Server) deleteVendor(ctx context.Context, vendorID int) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	affected, err := pdb.Model(&vendorRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", vendorID).
		Update(ctx, map[string]any{"deleted_at": time.Now()})
	return mustAffectRows(affected, err, "vendor")
}

func (s *Server) toggleVendorStatus(ctx context.Context, vendor *vendor) error {
	pdb, err := s.play()
	if err != nil {
		return err
	}

	status := vendor.Status
	if status == "enabled" {
		status = "disabled"
	} else {
		status = "enabled"
	}

	affected, err := pdb.Model(&vendorRead{}).
		WhereEq("company_id", CurrentCompany(ctx).ID).
		WhereEq("id", vendor.ID).
		Update(ctx, map[string]any{"status": string(status)})
	return mustAffectRows(affected, err, "vendor")
}

// Stays raw: `amount_payable = amount_payable + $3` is a self-referencing increment,
// which playsql's Update cannot express.
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

// findVendorPayables lists a vendor's unpaid AP entries with their payables register
// row. Same shape as findPayables: rooted on accounts_payable, with the register row
// pulled through Has/With, since every column the response needs lives on one of the
// two. The dropped joins are redundant for the reasons given there, and payables.id /
// payables.uuid still come from the register row.
//
// The vendor lookup keeps WithTrashed: the old join had no deleted_at predicate on
// vendors, so a soft-deleted vendor's payables were still listed.
func (s *Server) findVendorPayables(ctx context.Context, vendorID string) ([]*Payable, error) {
	pdb, err := s.play()
	if err != nil {
		return nil, err
	}

	companyID := CurrentCompany(ctx).ID

	var v vendorRead
	if err := pdb.Model(&vendorRead{}).
		Select("id").
		WithTrashed().
		WhereEq("company_id", companyID).
		WhereEq("uuid", vendorID).
		First(ctx, &v); err != nil {
		return nil, err
	}

	var rows []accountsPayableRead
	if err := pdb.Model(&accountsPayableRead{}).
		With("Register").
		Has("Register").
		WhereEq("company_id", companyID).
		WhereEq("vendor_id", v.ID).
		Where("paid_status", "!=", "paid").
		Get(ctx, &rows); err != nil {
		return nil, err
	}

	data := make([]*Payable, 0, len(rows))
	for _, r := range rows {
		data = append(data, r.toPayable())
	}
	return data, nil
}
