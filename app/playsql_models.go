package app

import (
	"database/sql"
	"time"

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
	DueOn              *time.Time       `db:"due_on"`
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
