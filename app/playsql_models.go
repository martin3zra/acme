package app

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
)

// Phase 2 of the playsql adoption: production writes. This file is the first
// production code to link playsql (and, transitively, its SQL drivers). Write
// paths convert one at a time; everything not yet migrated stays raw database/sql.

// playTx wraps an in-flight *sql.Tx with playsql under the Postgres grammar, so a
// production write can use typed models on the *same* transaction the caller
// already opened (via forge database.WithTransaction). The caller still owns the
// tx and its commit/rollback.
func playTx(tx *sql.Tx) (*playsql.Tx, error) {
	return playsql.UseTx(tx, "postgres")
}

// play wraps the server's *sql.DB with playsql under the Postgres grammar for
// read paths that are not inside a transaction (list/detail queries). Reads use
// dedicated *Read models below (the JSON response structs can't double as playsql
// models: they embed foundation.Timestamps, which the parser skips, and hold
// pointer-to-struct fields it would misread as relations).
func (s *Server) play() (*playsql.DB, error) {
	return playsql.Use(s.db, "postgres")
}

// customerRead is the playsql read model for the customers table. Only real
// columns are mapped (db tags); deleted_at carries play:"softdelete" so queries
// exclude soft-deleted rows automatically, matching the old "deleted_at IS NULL".
type customerRead struct {
	ID            int        `db:"id" play:"pk,incrementing"`
	UUID          string     `db:"uuid"`
	Code          string     `db:"code"`
	Name          string     `db:"name"`
	ContactName   string     `db:"contact_name"`
	Phone         string     `db:"phone"`
	Email         string     `db:"email"`
	Status        string     `db:"status"`
	AmountDue     float64    `db:"amount_due"`
	Address       string     `db:"address"`
	CustomerType  string     `db:"customer_type"`
	PaymentMethod string     `db:"payment_method"`
	CreditLimited bool       `db:"credit_limited"`
	CreditLimit   float64    `db:"credit_limit"`
	PaymentTerms  string     `db:"payment_terms"`
	TaxReceiptID  *int       `db:"tax_receipt_id"`
	CreatedAt     *time.Time `db:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at" play:"softdelete"`
}

func (customerRead) TableName() string { return "customers" }

// toCustomer maps the read model onto the JSON response struct.
func (r customerRead) toCustomer() *customer {
	c := &customer{
		ID:            r.ID,
		UUID:          r.UUID,
		Code:          r.Code,
		Name:          r.Name,
		ContactName:   r.ContactName,
		Phone:         r.Phone,
		Email:         r.Email,
		AmountDue:     r.AmountDue,
		Address:       r.Address,
		CustomerType:  r.CustomerType,
		PaymentMethod: r.PaymentMethod,
		CreditLimited: r.CreditLimited,
		CreditLimit:   r.CreditLimit,
		PaymentTerms:  r.PaymentTerms,
		TaxReceipt:    r.TaxReceiptID,
		Status:        foundation.Status(r.Status),
	}
	c.CreatedAt = r.CreatedAt
	c.UpdatedAt = r.UpdatedAt
	c.DeletedAt = r.DeletedAt
	return c
}

// vendorRead is the playsql read model for the vendors table — same pattern as
// customerRead (real columns only, softdelete on deleted_at).
type vendorRead struct {
	ID            int        `db:"id" play:"pk,incrementing"`
	UUID          string     `db:"uuid"`
	Code          string     `db:"code"`
	Name          string     `db:"name"`
	ContactName   string     `db:"contact_name"`
	Phone         string     `db:"phone"`
	Email         string     `db:"email"`
	Status        string     `db:"status"`
	AmountPayable float64    `db:"amount_payable"`
	PurchaseNote  string     `db:"purchase_note"`
	LeadTimeDays  int        `db:"lead_time_days"`
	Address       string     `db:"address"`
	VendorType    string     `db:"vendor_type"`
	PaymentMethod string     `db:"payment_method"`
	PaymentTerms  string     `db:"payment_terms"`
	CreatedAt     *time.Time `db:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at" play:"softdelete"`
}

func (vendorRead) TableName() string { return "vendors" }

// toVendor maps the read model onto the JSON response struct.
func (r vendorRead) toVendor() *vendor {
	v := &vendor{
		ID:            r.ID,
		UUID:          r.UUID,
		Code:          r.Code,
		Name:          r.Name,
		ContactName:   r.ContactName,
		Phone:         r.Phone,
		Email:         r.Email,
		PurchaseNote:  r.PurchaseNote,
		LeadTimeDays:  r.LeadTimeDays,
		AmountPayable: r.AmountPayable,
		Address:       r.Address,
		VendorType:    r.VendorType,
		PaymentMethod: r.PaymentMethod,
		PaymentTerms:  r.PaymentTerms,
		Status:        foundation.Status(r.Status),
	}
	v.CreatedAt = r.CreatedAt
	v.UpdatedAt = r.UpdatedAt
	v.DeletedAt = r.DeletedAt
	return v
}

// unitRead is the playsql read model for the units table. Unlike customerRead/
// vendorRead it carries NO play:"softdelete" tag: the original unit reads never
// filtered deleted_at (units are not soft-deleted — the repo has no delete path),
// so the model maps deleted_at as a plain column to keep "return all rows".
type unitRead struct {
	ID        int64      `db:"id" play:"pk,incrementing"`
	Name      string     `db:"name"`
	BaseQty   int        `db:"base_qty"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (unitRead) TableName() string { return "units" }

// toUnit maps the read model onto the JSON response struct.
func (r unitRead) toUnit() *unit {
	u := &unit{
		ID:      r.ID,
		Name:    r.Name,
		BaseQty: r.BaseQty,
	}
	u.CreatedAt = r.CreatedAt
	u.UpdatedAt = r.UpdatedAt
	u.DeletedAt = r.DeletedAt
	return u
}

// expenseRead is the playsql read model for the expenses table. receipt_url is
// deliberately unmapped: it is nullable, no read ever selected it, and mapping it
// would pull a NULL into the default projection.
type expenseRead struct {
	ID         int        `db:"id" play:"pk,incrementing"`
	UUID       string     `db:"uuid"`
	CategoryID int        `db:"category_id"`
	Date       time.Time  `db:"date"`
	Amount     float64    `db:"amount"`
	Notes      string     `db:"notes"`
	CreatedAt  *time.Time `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at" play:"softdelete"`

	Category *expenseCategoryRead `play:"belongsTo,fk=category_id"`
}

func (expenseRead) TableName() string { return "expenses" }

// toExpense maps the read model onto the JSON response struct. Only the three
// category columns the old INNER JOIN selected are copied across.
func (r expenseRead) toExpense() *expense {
	e := &expense{
		ID:     r.ID,
		UUID:   r.UUID,
		Date:   Date{Time: r.Date},
		Amount: r.Amount,
		Notes:  r.Notes,
	}
	e.CreatedAt = r.CreatedAt
	e.UpdatedAt = r.UpdatedAt
	e.DeletedAt = r.DeletedAt
	if r.Category != nil {
		e.Category = expenseCategory{
			ID:   r.Category.ID,
			UUID: r.Category.UUID,
			Name: r.Category.Name,
		}
	}
	return e
}

// expenseCategoryRead is the playsql read model for the expenses_categories table.
// It carries no play:"softdelete" tag even though the column exists: only
// findExpensesCategories filtered deleted_at, and the other three reads must keep
// resolving a soft-deleted category (storeExpense/updateExpense look one up by
// uuid). That one read filters explicitly with WhereNull.
//
// TotalAmount is play:"readonly" — it has no backing column and is excluded from
// the default projection, appearing only as the WithSum aggregate alias in
// findExpensesByCategories.
type expenseCategoryRead struct {
	ID          int        `db:"id" play:"pk,incrementing"`
	UUID        string     `db:"uuid"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	CreatedAt   *time.Time `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
	TotalAmount float64    `db:"total_amount" play:"readonly"`

	Expenses []*expenseRead `play:"hasMany,fk=category_id"`
}

func (expenseCategoryRead) TableName() string { return "expenses_categories" }

// toExpenseCategory maps the read model onto the JSON response struct.
func (r expenseCategoryRead) toExpenseCategory() *expenseCategory {
	c := &expenseCategory{
		ID:          r.ID,
		UUID:        r.UUID,
		Name:        r.Name,
		Description: r.Description,
		TotalAmount: r.TotalAmount,
	}
	c.CreatedAt = r.CreatedAt
	c.UpdatedAt = r.UpdatedAt
	c.DeletedAt = r.DeletedAt
	return c
}

// expenseInsert is the write model for the expenses table. uuid, receipt_url and
// the timestamps stay unmapped so the database fills them. It also backs
// updateExpense, which — like the statement it replaced — must not bump
// updated_at: playsql only stamps it when the column is mapped.
type expenseInsert struct {
	ID         int64     `db:"id" play:"pk,incrementing"`
	CompanyID  int       `db:"company_id"`
	CategoryID int       `db:"category_id"`
	Date       time.Time `db:"date"`
	Amount     float64   `db:"amount"`
	Notes      string    `db:"notes"`
}

func (expenseInsert) TableName() string { return "expenses" }

// expenseCategoryInsert is the write model for expenses_categories.
type expenseCategoryInsert struct {
	ID          int64  `db:"id" play:"pk,incrementing"`
	CompanyID   int    `db:"company_id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

func (expenseCategoryInsert) TableName() string { return "expenses_categories" }

// accountsPayableRead is the playsql read model for the accounts_payable table.
// The table has no deleted_at, so there is no softdelete tag. amount_payable is a
// generated column: it is read here but must never be written (see
// accountsPayableInsert, which leaves it unmapped).
//
// Register is the one-to-one payables cross-reference row. findPayables reads the
// AP entry as the root (it owns the filters and the due_date ordering) and pulls
// p.id / p.uuid through this relation, which is the inverse of the old query's
// payables-rooted INNER JOIN.
type accountsPayableRead struct {
	ID            int64     `db:"id" play:"pk,incrementing"`
	UUID          string    `db:"uuid"`
	VendorID      int64     `db:"vendor_id"`
	InvoiceNumber string    `db:"invoice_number"`
	InvoiceDate   time.Time `db:"invoice_date"`
	DueDate       time.Time `db:"due_date"`
	AmountTotal   float64   `db:"amount_total"`
	AmountPayable float64   `db:"amount_payable"`
	AmountPaid    float64   `db:"amount_paid"`
	Status        string    `db:"status"`
	PaidStatus    string    `db:"paid_status"`
	Notes         *string   `db:"notes"`

	Register *payableRegisterRead `play:"hasOne,fk=accounts_payable_id"`
}

func (accountsPayableRead) TableName() string { return "accounts_payable" }

// toPayable maps an AP entry plus its payables register row onto the response
// struct. Payable.ID/UUID identify the payables row; Payable.InvoiceID/InvoiceUUID
// identify the accounts_payable row.
func (r accountsPayableRead) toPayable() *Payable {
	notes := ""
	if r.Notes != nil {
		notes = *r.Notes
	}
	p := &Payable{
		InvoiceID:     r.ID,
		InvoiceUUID:   r.UUID,
		InvoiceNumber: r.InvoiceNumber,
		InvoiceDate:   r.InvoiceDate,
		DueDate:       r.DueDate,
		AmountTotal:   r.AmountTotal,
		AmountPayable: r.AmountPayable,
		AmountPaid:    r.AmountPaid,
		Status:        PayableStatus(r.Status),
		PaidStatus:    PaidStatus(r.PaidStatus),
		Notes:         &notes,
	}
	if r.Register != nil {
		p.ID = r.Register.ID
		p.UUID = r.Register.UUID
	}
	return p
}

// payableRegisterRead is the read side of the payables table. It is separate from
// the payableRegister write model because it maps uuid, which is DB-generated and
// must stay unmapped on insert.
type payableRegisterRead struct {
	ID                int64  `db:"id" play:"pk,incrementing"`
	UUID              string `db:"uuid"`
	AccountsPayableID int64  `db:"accounts_payable_id"`
	VendorID          int64  `db:"vendor_id"`
}

func (payableRegisterRead) TableName() string { return "payables" }

// vendorPaymentRead is the playsql read model for the vendor_payments table.
// company_id and the timestamps are deliberately unmapped: the old reads never
// selected them and vendorPayment leaves them zero.
//
// payment is jsonb; it is scanned raw and unmarshalled in toVendorPayment so a
// malformed blob leaves Payment nil instead of failing the whole read, matching
// the previous behaviour.
type vendorPaymentRead struct {
	ID   int64  `db:"id" play:"pk,incrementing"`
	UUID string `db:"uuid"`
	// int, not int64: eager loading matches this against vendorRead.ID by Go value,
	// so a widened type would silently never match and leave Vendor nil.
	VendorID int       `db:"vendor_id"`
	Date     time.Time `db:"date"`
	Amount   float64   `db:"amount"`
	Notes    *string   `db:"notes"`
	Payment  []byte    `db:"payment"`
	Status   string    `db:"status"`
	Code     string    `db:"code"`

	Vendor *vendorRead `play:"belongsTo,fk=vendor_id"`
}

func (vendorPaymentRead) TableName() string { return "vendor_payments" }

// toVendorPayment maps the read model onto the JSON response struct. Only the
// five vendor columns the old INNER JOIN selected are copied across.
func (r vendorPaymentRead) toVendorPayment() *vendorPayment {
	p := &vendorPayment{
		ID:     int(r.ID),
		UUID:   r.UUID,
		Date:   r.Date,
		Amount: r.Amount,
		Status: r.Status,
		Code:   r.Code,
	}
	if r.Notes != nil {
		p.Notes = *r.Notes
	}
	if len(r.Payment) > 0 {
		pm := new(Payment)
		if err := json.Unmarshal(r.Payment, pm); err == nil {
			p.Payment = pm
		}
	}
	if r.Vendor != nil {
		p.Vendor = vendor{
			ID:    r.Vendor.ID,
			UUID:  r.Vendor.UUID,
			Name:  r.Vendor.Name,
			Email: r.Vendor.Email,
			Phone: r.Vendor.Phone,
		}
	}
	return p
}

// vendorPaymentItemRead is the read side of vendor_payment_items, carrying the
// settled AP entry as a belongsTo so the line can report the bill's number and
// dates. The write model (vendorPaymentItem) stays separate.
type vendorPaymentItemRead struct {
	ID                int64   `db:"id" play:"pk,incrementing"`
	VendorPaymentID   int64   `db:"vendor_payment_id"`
	AccountsPayableID int64   `db:"accounts_payable_id"`
	AmountDue         float64 `db:"amount_due"`
	PaymentAmount     float64 `db:"payment_amount"`

	AccountsPayable *accountsPayableRead `play:"belongsTo,fk=accounts_payable_id"`
}

func (vendorPaymentItemRead) TableName() string { return "vendor_payment_items" }

// Receivable is the write model for the receivables table. The pk is DB-assigned
// (serial); playsql omits the zero id on insert and reads it back via RETURNING.
// Columns with database defaults (timestamps) are intentionally not mapped, so
// the INSERT lets the database fill them.
type Receivable struct {
	ID         int64 `db:"id" play:"pk,incrementing"`
	CompanyID  int   `db:"company_id"`
	InvoiceID  int   `db:"invoice_id"`
	CustomerID int   `db:"customer_id"`
}

func (Receivable) TableName() string { return "receivables" }

// InvoiceItem is the write model for invoice line rows (invoices_items). It backs
// the bulk line insert; the pk is DB-assigned, and timestamp columns are left to
// their database defaults (not mapped here).
type InvoiceItem struct {
	ID          int64   `db:"id" play:"pk,incrementing"`
	CompanyID   int     `db:"company_id"`
	InvoiceID   int     `db:"invoice_id"`
	ItemID      int     `db:"item_id"`
	VariantID   int     `db:"variant_id"`
	UnitID      int     `db:"unit_id"`
	Qty         int     `db:"qty"`
	Price       float64 `db:"price"`
	Rate        float64 `db:"rate"`
	Amount      float64 `db:"amount"`
	Tax         float64 `db:"tax"`
	Total       float64 `db:"total"`
	WarehouseID int     `db:"warehouse_id"`
}

func (InvoiceItem) TableName() string { return "invoices_items" }

// invoiceInsert is the write model for creating an invoice row. uuid is
// deliberately NOT mapped: it is DB-generated, and mapping it would make playsql
// insert an empty string over the default. After Insert sets ID, the caller reads
// the generated uuid back. JSON columns hold pre-encoded values (foundation
// ToJSON/AsJSON) exactly as the prior raw INSERT passed them, so encoding is
// unchanged. Timestamp columns are left to their database defaults.
type invoiceInsert struct {
	ID                 int64           `db:"id" play:"pk,incrementing"`
	CompanyID          int             `db:"company_id"`
	TaxReceiptID       *int            `db:"tax_receipt_id"`
	TaxReceiptSequence *int64          `db:"tax_receipt_sequence"`
	TaxNumber          *string         `db:"tax_number"`
	Date               time.Time       `db:"date"`
	Type               *string         `db:"type"`
	DueOn              *time.Time      `db:"due_on"`
	CustomerID         int             `db:"customer_id"`
	Amount             float64         `db:"amount"`
	Discount           string          `db:"discount"`
	Tax                float64         `db:"tax"`
	AmountDue          float64         `db:"amount_due"`
	Total              float64         `db:"total"`
	Note               string          `db:"note"`
	Status             InvoiceStatus   `db:"status"`
	PaidStatus         PaidStatus      `db:"paid_status"`
	Payment            string          `db:"payment"`
	Code               string          `db:"code"`
	TransactionKind    TransactionKind `db:"transaction_kind"`
	Source             *[]byte         `db:"source"`
	Recurrence         *[]byte         `db:"recurrence"`
}

func (invoiceInsert) TableName() string { return "invoices" }

// InventoryMovement is the write model for a stock movement row. created_at is
// set explicitly by the caller (no DB default is relied on). The balance upsert
// that follows a movement stays raw SQL: it increments (quantity +=
// EXCLUDED.quantity), which playsql's replace-style Upsert cannot express.
type InventoryMovement struct {
	ID              int64                 `db:"id" play:"pk,incrementing"`
	CompanyID       int                   `db:"company_id"`
	VariantID       int                   `db:"variant_id"`
	WarehouseID     int                   `db:"warehouse_id"`
	TransactionKind InventoryMovementKind `db:"transaction_kind"`
	Qty             float64               `db:"qty"`
	UnitCost        float64               `db:"unit_cost"`
	ReferenceType   string                `db:"reference_type"`
	ReferenceID     int                   `db:"reference_id"`
	CreatedAt       time.Time             `db:"created_at"`
}

func (InventoryMovement) TableName() string { return "inventory_movements" }

// customerInsert is the write model for creating a customer row. amount_due seeds
// from the opening balance; the pk is DB-assigned.
type customerInsert struct {
	ID            int64   `db:"id" play:"pk,incrementing"`
	CompanyID     int     `db:"company_id"`
	Name          string  `db:"name"`
	ContactName   string  `db:"contact_name"`
	Email         string  `db:"email"`
	Phone         string  `db:"phone"`
	PaymentMethod string  `db:"payment_method"`
	PaymentTerms  string  `db:"payment_terms"`
	CreditLimited bool    `db:"credit_limited"`
	CreditLimit   float64 `db:"credit_limit"`
	AmountDue     float64 `db:"amount_due"`
	CustomerType  string  `db:"customer_type"`
	TaxReceiptID  int     `db:"tax_receipt_id"`
	Code          string  `db:"code"`
	Address       string  `db:"address"`
}

func (customerInsert) TableName() string { return "customers" }

// vendorInsert is the write model for creating a vendor row. amount_payable seeds
// from the opening balance; the pk is DB-assigned.
type vendorInsert struct {
	ID            int64   `db:"id" play:"pk,incrementing"`
	CompanyID     int     `db:"company_id"`
	Name          string  `db:"name"`
	ContactName   string  `db:"contact_name"`
	Email         string  `db:"email"`
	Phone         string  `db:"phone"`
	PaymentMethod string  `db:"payment_method"`
	PaymentTerms  string  `db:"payment_terms"`
	PurchaseNote  string  `db:"purchase_note"`
	LeadTimeDays  int     `db:"lead_time_days"`
	AmountPayable float64 `db:"amount_payable"`
	VendorType    string  `db:"vendor_type"`
	Code          string  `db:"code"`
	Address       string  `db:"address"`
}

func (vendorInsert) TableName() string { return "vendors" }

// PurchaseItem is the write model for purchase line rows (purchase_items). Backs
// the bulk line insert; the pk is DB-assigned.
type PurchaseItem struct {
	ID         int64   `db:"id" play:"pk,incrementing"`
	CompanyID  int     `db:"company_id"`
	PurchaseID int     `db:"purchase_id"`
	VariantID  int     `db:"variant_id"`
	Qty        int     `db:"qty"`
	UnitPrice  float64 `db:"unit_price"`
	LineTotal  float64 `db:"line_total"`
	UnitID     int     `db:"unit_id"`
	Discount   float64 `db:"discount"`
	TaxID      int     `db:"tax_id"`
	TaxAmount  float64 `db:"tax_amount"`
}

func (PurchaseItem) TableName() string { return "purchase_items" }

// paymentInsert is the write model for a customer payment (receivables_income).
type paymentInsert struct {
	ID         int64         `db:"id" play:"pk,incrementing"`
	CompanyID  int           `db:"company_id"`
	CustomerID int           `db:"customer_id"`
	Date       time.Time     `db:"date"`
	Amount     float64       `db:"amount"`
	Notes      string        `db:"notes"`
	Payment    string        `db:"payment"`
	Status     PaymentStatus `db:"status"`
	Code       string        `db:"code"`
}

func (paymentInsert) TableName() string { return "receivables_income" }

// vendorPaymentInsert is the write model for a vendor payment (vendor_payments).
type vendorPaymentInsert struct {
	ID        int64     `db:"id" play:"pk,incrementing"`
	CompanyID int       `db:"company_id"`
	VendorID  int       `db:"vendor_id"`
	Date      time.Time `db:"date"`
	Amount    float64   `db:"amount"`
	Notes     string    `db:"notes"`
	Payment   string    `db:"payment"`
	Status    string    `db:"status"`
	Code      string    `db:"code"`
}

func (vendorPaymentInsert) TableName() string { return "vendor_payments" }

// paymentItem is the write model for customer payment allocation rows
// (receivables_income_items). Backs the bulk allocation insert.
type paymentItem struct {
	ID                 int64     `db:"id" play:"pk,incrementing"`
	CompanyID          int       `db:"company_id"`
	ReceivableIncomeID int       `db:"receivable_income_id"`
	Date               time.Time `db:"date"`
	InvoiceID          int       `db:"invoice_id"`
	AmountDue          float64   `db:"amount_due"`
	PaymentAmount      float64   `db:"payment_amount"`
}

func (paymentItem) TableName() string { return "receivables_income_items" }

// vendorPaymentItem is the write model for vendor payment allocation rows
// (vendor_payment_items). Backs the bulk allocation insert.
type vendorPaymentItem struct {
	ID                int64     `db:"id" play:"pk,incrementing"`
	CompanyID         int       `db:"company_id"`
	VendorPaymentID   int       `db:"vendor_payment_id"`
	AccountsPayableID int64     `db:"accounts_payable_id"`
	Date              time.Time `db:"date"`
	AmountDue         float64   `db:"amount_due"`
	PaymentAmount     float64   `db:"payment_amount"`
}

func (vendorPaymentItem) TableName() string { return "vendor_payment_items" }

// accountsPayableInsert is the write model for an AP entry (accounts_payable).
// amount_payable is a generated column and is intentionally not mapped.
type accountsPayableInsert struct {
	ID             int64         `db:"id" play:"pk,incrementing"`
	CompanyID      int           `db:"company_id"`
	VendorID       int           `db:"vendor_id"`
	PurchaseID     int           `db:"purchase_id"`
	InvoiceNumber  string        `db:"invoice_number"`
	InvoiceDate    time.Time     `db:"invoice_date"`
	DueDate        *time.Time    `db:"due_date"`
	AmountTotal    float64       `db:"amount_total"`
	TaxAmount      float64       `db:"tax_amount"`
	DiscountAmount float64       `db:"discount_amount"`
	AmountPaid     float64       `db:"amount_paid"`
	Currency       string        `db:"currency"`
	PaymentTerms   string        `db:"payment_terms"`
	Status         PayableStatus `db:"status"`
	PaidStatus     PaidStatus    `db:"paid_status"`
	CreatedBy      int           `db:"created_by"`
}

func (accountsPayableInsert) TableName() string { return "accounts_payable" }

// payableRegister is the write model for the payables cross-reference row.
type payableRegister struct {
	ID                int64 `db:"id" play:"pk,incrementing"`
	CompanyID         int   `db:"company_id"`
	AccountsPayableID int   `db:"accounts_payable_id"`
	VendorID          int   `db:"vendor_id"`
}

func (payableRegister) TableName() string { return "payables" }

// openingInvoiceInsert is the write model for a customer's opening-balance invoice
// — a partial-column insert (the rest of invoices' columns take DB defaults), so
// it is a dedicated model rather than the full invoiceInsert.
type openingInvoiceInsert struct {
	ID         int64         `db:"id" play:"pk,incrementing"`
	CompanyID  int           `db:"company_id"`
	Date       time.Time     `db:"date"`
	Type       TermType      `db:"type"`
	DueOn      time.Time     `db:"due_on"`
	CustomerID int           `db:"customer_id"`
	Amount     float64       `db:"amount"`
	AmountDue  float64       `db:"amount_due"`
	Total      float64       `db:"total"`
	Note       string        `db:"note"`
	Status     InvoiceStatus `db:"status"`
	PaidStatus PaidStatus    `db:"paid_status"`
	Code       string        `db:"code"`
}

func (openingInvoiceInsert) TableName() string { return "invoices" }

// openingPayableInsert is the write model for a vendor's opening-balance AP entry.
// Its column set differs from accountsPayableInsert (no purchase_id; carries
// payment_method + notes), so it is its own model. amount_payable is generated.
type openingPayableInsert struct {
	ID             int64         `db:"id" play:"pk,incrementing"`
	CompanyID      int           `db:"company_id"`
	VendorID       int           `db:"vendor_id"`
	InvoiceNumber  string        `db:"invoice_number"`
	InvoiceDate    time.Time     `db:"invoice_date"`
	DueDate        time.Time     `db:"due_date"`
	AmountTotal    float64       `db:"amount_total"`
	TaxAmount      float64       `db:"tax_amount"`
	DiscountAmount float64       `db:"discount_amount"`
	AmountPaid     float64       `db:"amount_paid"`
	Currency       string        `db:"currency"`
	PaymentTerms   string        `db:"payment_terms"`
	PaymentMethod  string        `db:"payment_method"`
	Status         PayableStatus `db:"status"`
	PaidStatus     PaidStatus    `db:"paid_status"`
	Notes          string        `db:"notes"`
	CreatedBy      int           `db:"created_by"`
}

func (openingPayableInsert) TableName() string { return "accounts_payable" }
