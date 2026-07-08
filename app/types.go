package app

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"sync"
	"time"

	"github.com/martin3zra/acme/app/mail"
	"github.com/martin3zra/forge/auth"
	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/forge/mailer"
	"github.com/martin3zra/forge/routing"
	"github.com/martin3zra/forge/support"
	"github.com/martin3zra/forge/validator"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

const PaymentTermsMax = 120 // net120

type LoginFormRequest struct {
	support.FormRequest
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

func (f LoginFormRequest) Rules() map[string]any {
	return map[string]any{
		"email":    "required|email|max:100|lowercase",
		"password": "required",
		"remember": "sometimes|boolean",
	}
}

type CreatePasswordForm struct {
	support.FormRequest
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

func (f CreatePasswordForm) Rules() map[string]any {
	return map[string]any{
		"password": "required|confirmed",
	}
}

type StoreCustomerForm struct {
	support.FormRequest
	Name            string    `json:"name"`
	Contact         string    `json:"contact"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	PaymentMethod   string    `json:"payment_method"`
	PaymentTerms    string    `json:"payment_terms"`
	CreditLimited   bool      `json:"credit_limited"`
	CreditLimit     float64   `json:"credit_limit"`
	CustomerType    string    `json:"customer_type"`
	TaxReceipt      int       `json:"tax_receipt"`
	OpenBalance     float64   `json:"open_balance"`
	OpenBalanceAsOf time.Time `json:"open_balance_as_of"`
}

func (StoreCustomerForm) Rules() map[string]any {
	return map[string]any{
		"name":    "required|min:3|max:120",
		"contact": "sometimes|min:3|max:120",
		"email": []any{
			"required",
			"email",
			"min:8",
			"max:120",
			"lowercase",
			validator.Rule{}.Unique("customers", "email"),
		},
		"phone":              "sometimes|min:3|max:120",
		"payment_method":     "sometimes|in:cash,ck,card,bt",
		"payment_terms":      "sometimes|required",
		"credit_limited":     "required",
		"credit_limit":       "sometimes|required|min:0",
		"customer_type":      "sometimes|required|in:individual,business",
		"tax_receipt":        "sometimes|exists:tax_receipts,id",
		"open_balance":       "sometimes|min:0",
		"open_balance_as_of": "sometimes",
	}
}

func (form StoreCustomerForm) Authorize() bool {
	return Can(form.User(), "create:customer")
}

type StoreVendorForm struct {
	support.FormRequest
	Name            string    `json:"name"`
	Contact         string    `json:"contact"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	PurchaseNote    string    `json:"purchase_note"`
	LeadTimeDays    int       `json:"lead_time_days"`
	PaymentMethod   string    `json:"payment_method"`
	PaymentTerms    string    `json:"payment_terms"`
	CreditLimited   bool      `json:"credit_limited"`
	CreditLimit     float64   `json:"credit_limit"`
	VendorType      string    `json:"vendor_type"`
	TaxReceipt      int       `json:"tax_receipt"`
	OpenBalance     float64   `json:"open_balance"`
	OpenBalanceAsOf time.Time `json:"open_balance_as_of"`
}

func (StoreVendorForm) Rules() map[string]any {
	return map[string]any{
		"name":    "required|min:3|max:120",
		"contact": "sometimes|min:3|max:120",
		"email": []any{
			"required",
			"email",
			"min:8",
			"max:120",
			"lowercase",
			validator.Rule{}.Unique("vendors", "email"),
		},
		"phone":              "sometimes|min:3|max:120",
		"purchase_note":      "sometimes|max:2000",
		"lead_time_days":     "sometimes|integer|min:0",
		"payment_method":     "sometimes|in:cash,ck,card,bt",
		"payment_terms":      "sometimes|required",
		"credit_limited":     "required",
		"credit_limit":       "sometimes|required|min:0",
		"vendor_type":        "sometimes|required|in:individual,business",
		"tax_receipt":        "sometimes|exists:tax_receipts,id",
		"open_balance":       "sometimes|min:0",
		"open_balance_as_of": "sometimes",
	}
}

func (form StoreVendorForm) Authorize() bool {
	return Can(form.User(), "create:vendor")
}

type UpdateCustomerForm struct {
	support.FormRequest
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Contact         string    `json:"contact"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	PaymentMethod   string    `json:"payment_method"`
	PaymentTerms    string    `json:"payment_terms"`
	CreditLimited   bool      `json:"credit_limited"`
	CreditLimit     float64   `json:"credit_limit"`
	CustomerType    string    `json:"customer_type"`
	TaxReceipt      int       `json:"tax_receipt"`
	OpenBalance     float64   `json:"open_balance"`
	OpenBalanceAsOf time.Time `json:"open_balance_as_of"`
}

func (form UpdateCustomerForm) Authorize() bool {
	return Can(form.User(), "update:customer")
}

func (form UpdateCustomerForm) Rules() map[string]any {
	return map[string]any{
		"name":    "required|min:3|max:120",
		"contact": "sometimes|min:3|max:120",
		"email": []any{
			"required",
			"email",
			"min:8",
			"max:120",
			"lowercase",
			validator.Rule{}.Unique("customers", "email").Ignore(form.ID, "id"),
		},
		"phone":              "sometimes|min:3|max:120",
		"payment_method":     "sometimes|in:cash,ck,card,bt",
		"payment_terms":      "sometimes|required",
		"credit_limit":       "sometimes|required|min:0",
		"customer_type":      "sometimes|required|in:individual,business",
		"tax_receipt":        "sometimes|exists:tax_receipts,id",
		"open_balance":       "sometimes|min:0",
		"open_balance_as_of": "sometimes",
	}
}

type UpdateVendorForm struct {
	support.FormRequest
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Contact         string    `json:"contact"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	PurchaseNote    string    `json:"purchase_note"`
	LeadTimeDays    int       `json:"lead_time_days"`
	PaymentMethod   string    `json:"payment_method"`
	PaymentTerms    string    `json:"payment_terms"`
	CreditLimited   bool      `json:"credit_limited"`
	CreditLimit     float64   `json:"credit_limit"`
	VendorType      string    `json:"vendor_type"`
	TaxReceipt      int       `json:"tax_receipt"`
	OpenBalance     float64   `json:"open_balance"`
	OpenBalanceAsOf time.Time `json:"open_balance_as_of"`
}

func (form UpdateVendorForm) Authorize() bool {
	return Can(form.User(), "update:vendor")
}

func (form UpdateVendorForm) Rules() map[string]any {
	return map[string]any{
		"name":    "required|min:3|max:120",
		"contact": "sometimes|min:3|max:120",
		"email": []any{
			"required",
			"email",
			"min:8",
			"max:120",
			"lowercase",
			validator.Rule{}.Unique("vendors", "email").Ignore(form.ID, "id"),
		},
		"phone":              "sometimes|min:3|max:120",
		"purchase_note":      "sometimes|max:2000",
		"lead_time_days":     "sometimes|integer|min:0",
		"payment_method":     "sometimes|in:cash,ck,card,bt",
		"payment_terms":      "sometimes|required",
		"credit_limit":       "sometimes|required|min:0",
		"vendor_type":        "sometimes|required|in:individual,business",
		"tax_receipt":        "sometimes|exists:tax_receipts,id",
		"open_balance":       "sometimes|min:0",
		"open_balance_as_of": "sometimes",
	}
}

type StoreRecurrenceForm struct {
	support.FormRequest
	Recurrence
}

func (form StoreRecurrenceForm) Rules() map[string]any {
	return map[string]any{
		"recurrence":              "bail|required",
		"recurrence.enabled":      "bail|sometimes",
		"recurrence.name":         "required|max:100",
		"recurrence.start_date":   "nullable|date",
		"recurrence.until":        "nullable|date|after_or_equal:start_date",
		"recurrence.frequency":    "required|in:daily,weekly,monthly,quarterly,yearly",
		"recurrence.interval":     "required|integer|min:1",
		"recurrence.weekdays":     "required_if:frequency,weekly|min:1",
		"recurrence.day_of_month": "required_if:frequency,monthly,quarterly,yearly|min:1|max:31",
		"recurrence.month":        "required_if:frequency,yearly|min:1|max:12",
	}
}

func (form StoreRecurrenceForm) AsRecurrence() *Recurrence {
	return &Recurrence{
		Enabled:    form.Enabled,
		Name:       form.Name,
		Type:       form.Type,
		SendEmail:  form.SendEmail,
		Frequency:  form.Frequency,
		Interval:   form.Interval,
		Timezone:   form.Timezone,
		StartDate:  form.StartDate,
		Until:      form.Until,
		DayOfMonth: form.DayOfMonth,
		Weekdays:   form.Weekdays,
		Month:      form.Month,
	}
}

type ConfirmsPasswords struct {
	support.FormRequest
	Password string `json:"current_password"`
}

func (ConfirmsPasswords) Rules() map[string]any {
	return map[string]any{
		"current_password": "required|current_password",
	}
}

type ItemIdentifiers struct {
	Reference       *string `json:"reference,omitempty"`
	Code            *string `json:"code,omitempty"`
	SKU             *string `json:"sku,omitempty"`
	Barcode         *string `json:"barcode,omitempty"`
	VendorReference *string `json:"vendor_reference,omitempty"`
}

func (d *ItemIdentifiers) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *ItemIdentifiers) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type StoreItemForm struct {
	support.FormRequest
	Name        string          `json:"name"`
	Price       float64         `json:"price"`
	Description string          `json:"description"`
	TaxID       int             `json:"tax_id"`
	UnitID      int             `json:"unit_id"`
	ItemType    string          `json:"item_type"` // e.g. "product", "service"
	Identifiers ItemIdentifiers `json:"identifiers,omitempty"`
}

func (form StoreItemForm) Authorize() bool {
	return Can(form.User(), "create:item")
}

func (StoreItemForm) Rules() map[string]any {
	return map[string]any{
		"name": []any{
			"required",
			"min:3",
			"max:120",
			validator.Rule{}.Unique("items", "name"),
		},
		"description":                  "sometimes|min:3|max:120",
		"price":                        "required|min:0",
		"tax_id":                       "bail|required|exists:taxes,id",
		"unit_id":                      "bail|required|exists:units,id",
		"item_type":                    "bail|required|in:product,service",
		"identifiers":                  "sometimes",
		"identifiers.reference":        "sometimes|nullable|max:100",
		"identifiers.code":             "sometimes|nullable|max:50",
		"identifiers.sku":              "sometimes|nullable|max:50",
		"identifiers.barcode":          "sometimes|nullable|max:32",
		"identifiers.vendor_reference": "sometimes|nullable|max:100",
	}
}

type UpdateItemForm struct {
	support.FormRequest
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Price       float64         `json:"price"`
	Description string          `json:"description"`
	TaxID       int             `json:"tax_id"`
	UnitID      int             `json:"unit_id"`
	ItemType    string          `json:"item_type"` // e.g. "product", "service"
	Identifiers ItemIdentifiers `json:"identifiers,omitempty"`
}

type StoreWarehouseForm struct {
	support.FormRequest
	Name     string `json:"name"`
	Location string `json:"location"`
}

func (form StoreWarehouseForm) Authorize() bool {
	return Can(form.User(), "create:inventory")
}

func (StoreWarehouseForm) Rules() map[string]any {
	return map[string]any{
		"name":     []any{"required", "min:3", "max:150", validator.Rule{}.Unique("warehouses", "name")},
		"location": "sometimes|nullable|max:2000",
	}
}

type UpdateWarehouseForm struct {
	support.FormRequest
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

func (form UpdateWarehouseForm) Authorize() bool {
	return Can(form.User(), "update:inventory")
}

func (form UpdateWarehouseForm) Rules() map[string]any {
	return map[string]any{
		"name":     []any{"required", "min:3", "max:150", validator.Rule{}.Unique("warehouses", "name").Ignore(form.ID, "id")},
		"location": "sometimes|nullable|max:2000",
	}
}

func (form UpdateItemForm) Authorize() bool {
	return Can(form.User(), "update:item")
}

func (form UpdateItemForm) Rules() map[string]any {
	return map[string]any{
		"name":                         []any{"required", "min:3", "max:120", validator.Rule{}.Unique("items", "name").Ignore(form.ID, "id")},
		"description":                  "sometimes|min:3|max:120",
		"price":                        "required|min:0",
		"tax_id":                       "required|exists:taxes,id",
		"unit_id":                      "required|exists:units,id",
		"item_type":                    "bail|required|in:product,service",
		"identifiers":                  "sometimes",
		"identifiers.reference":        "sometimes|nullable|max:100",
		"identifiers.code":             "sometimes|nullable|max:50",
		"identifiers.sku":              "sometimes|nullable|max:50",
		"identifiers.barcode":          "sometimes|nullable|max:32",
		"identifiers.vendor_reference": "sometimes|nullable|max:100",
	}
}

type TermType string

const (
	_CASH    TermType = "cash"
	_CREDIT  TermType = "credit"
	_OPENING TermType = "opening"
)

var InvoiceTermType = struct {
	Cash    TermType
	Credit  TermType
	Opening TermType
}{
	Cash:    _CASH,
	Credit:  _CREDIT,
	Opening: _OPENING,
}

type Role string

const (
	_DEVELOPER  Role = "developer"  //  Access to APIs, integrations, and developer tools
	_OWNER      Role = "owner"      // Full control of billing, settings, and organization
	_ADMIN      Role = "admin"      //  Manages users, roles, and global settings
	_SUPERVISOR Role = "supervisor" // Manages team data, limited settings access
	_STANDARD   Role = "standard"   // Regular user with core feature access
)

var Roles = struct {
	Developer  Role
	Owner      Role
	Admin      Role
	Supervisor Role
	Standard   Role
}{
	Developer:  _DEVELOPER,
	Owner:      _OWNER,
	Admin:      _ADMIN,
	Supervisor: _SUPERVISOR,
	Standard:   _STANDARD,
}

var RoleMap = []map[string]any{
	// {"id": string(Roles.Developer), "label": Roles.Developer},
	// {"id": string(Roles.Owner), "label": Roles.Owner},
	{"id": string(Roles.Admin), "label": Roles.Admin, "description": "Manages users, roles, and global settings"},
	{"id": string(Roles.Supervisor), "label": Roles.Supervisor, "description": "Manages team data, limited settings access"},
	{"id": string(Roles.Standard), "label": Roles.Standard, "description": "Regular user with core feature access"},
}

type TransactionKind string

const (
	_TRANSACTION_KIND_INVOICE  TransactionKind = "invoice"
	_TRANSACTION_KIND_ESTIMATE TransactionKind = "estimate"
	_TRANSACTION_KIND_ORDER    TransactionKind = "order"
	_TRANSACTION_KIND_TEMPLATE TransactionKind = "template"
)

var TransactionKinds = struct {
	Invoice  TransactionKind
	Estimate TransactionKind
	Order    TransactionKind
	Template TransactionKind
}{
	Invoice:  _TRANSACTION_KIND_INVOICE,
	Estimate: _TRANSACTION_KIND_ESTIMATE,
	Order:    _TRANSACTION_KIND_ORDER,
	Template: _TRANSACTION_KIND_TEMPLATE,
}

type PurchaseTransactionKind string

const (
	_PURCHASE_ORDER   PurchaseTransactionKind = "purchase_order"
	_PURCHASE_RECEIPT PurchaseTransactionKind = "purchase_receipt"
	_VENDOR_BILL      PurchaseTransactionKind = "vendor_bill"
)

var PurchaseTransactionKinds = struct {
	PurchaseOrder   PurchaseTransactionKind
	PurchaseReceipt PurchaseTransactionKind
	VendorBill      PurchaseTransactionKind
}{
	PurchaseOrder:   _PURCHASE_ORDER,
	PurchaseReceipt: _PURCHASE_RECEIPT,
	VendorBill:      _VENDOR_BILL,
}

// PurchaseStatus represents the lifecycle state of a purchase (purchase_status column).
type PurchaseStatus string

const (
	_PURCHASE_STATUS_DRAFT              PurchaseStatus = "draft"
	_PURCHASE_STATUS_PARTIALLY_RECEIVED PurchaseStatus = "partially_received"
	_PURCHASE_STATUS_RECEIVED           PurchaseStatus = "received"
	_PURCHASE_STATUS_PARTIALLY_PAID     PurchaseStatus = "partially_paid"
	_PURCHASE_STATUS_CLOSED             PurchaseStatus = "closed"
	_PURCHASE_STATUS_POSTED             PurchaseStatus = "posted"
)

var PurchaseStatuses = struct {
	Draft             PurchaseStatus
	PartiallyReceived PurchaseStatus
	Received          PurchaseStatus
	PartiallyPaid     PurchaseStatus
	Closed            PurchaseStatus
	Posted            PurchaseStatus
}{
	Draft:             _PURCHASE_STATUS_DRAFT,
	PartiallyReceived: _PURCHASE_STATUS_PARTIALLY_RECEIVED,
	Received:          _PURCHASE_STATUS_RECEIVED,
	PartiallyPaid:     _PURCHASE_STATUS_PARTIALLY_PAID,
	Closed:            _PURCHASE_STATUS_CLOSED,
	Posted:            _PURCHASE_STATUS_POSTED,
}

// lockedPurchaseStatuses lists purchase statuses where edits and deletes are prohibited.
var lockedPurchaseStatuses = map[PurchaseStatus]bool{
	PurchaseStatuses.Received:          true,
	PurchaseStatuses.PartiallyReceived: true,
	PurchaseStatuses.PartiallyPaid:     true,
	PurchaseStatuses.Closed:            true,
	PurchaseStatuses.Posted:            true,
}

type TransactionSource struct {
	Type TransactionKind `json:"type,omitempty"`
	ID   string          `json:"id,omitempty"`
	Code string          `json:"code,omitempty"`
}

func (d *TransactionSource) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *TransactionSource) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type PurchaseSource struct {
	Type   PurchaseTransactionKind `json:"type,omitempty"`
	ID     string                  `json:"id,omitempty"`
	Code   string                  `json:"code,omitempty"`
	Target *PurchaseSource         `json:"target,omitempty"`
}

func (d *PurchaseSource) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *PurchaseSource) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type FrequencyType string

const (
	_FREQUENCY_WEEKLY   FrequencyType = "weekly"
	_FREQUENCY_MONTHLY  FrequencyType = "monthly"
	_FREQUENCY_QUARTELY FrequencyType = "quarterly"
	_FREQUENCY_YEARLY   FrequencyType = "yearly"
)

var Frequency = struct {
	Weekly    FrequencyType
	Monthly   FrequencyType
	Quarterly FrequencyType
	Yearly    FrequencyType
}{
	Weekly:    _FREQUENCY_WEEKLY,
	Monthly:   _FREQUENCY_MONTHLY,
	Quarterly: _FREQUENCY_QUARTELY,
	Yearly:    _FREQUENCY_YEARLY,
}

type Recurrence struct {
	Enabled   bool       `json:"enabled"`
	Name      string     `json:"name"`
	Type      string     `json:"type"` // schedule, reminder
	SendEmail bool       `json:"send_email"`
	Frequency string     `json:"frequency"` // daily, weekly, monthly, quarterly, yearly
	Interval  int        `json:"interval"`
	Timezone  string     `json:"timezone,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	Until     *time.Time `json:"until"`

	// Optional fields depending on frequency
	DayOfMonth      int        `json:"day_of_month,omitempty"`
	Weekdays        []string   `json:"weekdays,omitempty"`
	Month           int        `json:"month,omitempty"`
	LastGeneratedAt *time.Time `json:"last_generated_at"`
	NextRunAt       *time.Time `json:"next_run_at"`
}

func (d *Recurrence) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *Recurrence) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

// Payable Status (AP Invoice Lifecycle)
// Draft → invoice received, not yet submitted for approval.
// Pending → submitted, awaiting approval.
// Approved → approved, scheduled for payment.
// Partial → partially paid, balance remains.
// Paid → fully settled.
// Cancelled → invoice voided, no payment will be made.
type PayableStatus string

const (
	_PAYABLE_DRAFT    PayableStatus = "draft"
	_PAYABLE_PENDING  PayableStatus = "pending"
	_PAYABLE_APPROVED PayableStatus = "approved"
	_PAYABLE_REJECTED PayableStatus = "rejected"
	_PAYABLE_VOID     PayableStatus = "void"
)

var PayableStatuses = struct {
	Draft    PayableStatus
	Pending  PayableStatus
	Approved PayableStatus
	Rejected PayableStatus
	Void     PayableStatus
}{
	Draft:    _PAYABLE_DRAFT,
	Pending:  _PAYABLE_PENDING,
	Approved: _PAYABLE_APPROVED,
	Rejected: _PAYABLE_REJECTED,
	Void:     _PAYABLE_VOID,
}

// Paid Status (Financial Settlement)
// Unpaid → no payment received.
// Partially Paid → some payment received, balance remains.
// Paid → fully settled.
// Refunded → money returned to customer.
type PaidStatus string

const (
	_PAID_UNPAID   PaidStatus = "unpaid"
	_PAID_PARTIAL  PaidStatus = "partial"
	_PAID_PAID     PaidStatus = "paid"
	_PAID_REFUNDED PaidStatus = "refunded"
)

var PaidStatuses = struct {
	UnPaid   PaidStatus
	Partial  PaidStatus
	Paid     PaidStatus
	Refunded PaidStatus
}{
	UnPaid:   _PAID_UNPAID,
	Partial:  _PAID_PARTIAL,
	Paid:     _PAID_PAID,
	Refunded: _PAID_REFUNDED,
}

// Document Status (Lifecycle)
// Draft → invoice is being prepared, editable.
// Sent/Open → finalized and delivered to customer, awaiting payment.
// Overdue → past due date, still unpaid.
// Void → canceled, no longer valid.
// Uncollectible → written off as bad debt.
// Closed → invoice has reached its end state (usually after payment or cancellation).
type InvoiceStatus string

const (
	_INVOICE_DRAFT         InvoiceStatus = "draft"
	_INVOICE_SENT          InvoiceStatus = "sent"
	_INVOICE_OVERDUE       InvoiceStatus = "overdue"
	_INVOICE_VOID          InvoiceStatus = "void"
	_INVOICE_UNCOLLECTIBLE InvoiceStatus = "uncollectible"
	_INVOICE_CLOSED        InvoiceStatus = "closed"
)

var InvoiceStatuses = struct {
	Draft         InvoiceStatus
	Sent          InvoiceStatus
	Overdue       InvoiceStatus
	Void          InvoiceStatus
	Uncollectible InvoiceStatus
	Closed        InvoiceStatus
}{
	Draft:         _INVOICE_DRAFT,
	Sent:          _INVOICE_SENT,
	Overdue:       _INVOICE_OVERDUE,
	Void:          _INVOICE_VOID,
	Uncollectible: _INVOICE_UNCOLLECTIBLE,
	Closed:        _INVOICE_CLOSED,
}

type PaymentStatus string

const (
	_PAYMENT_VOID      PaymentStatus = "void"
	_PAYMENT_PENDING   PaymentStatus = "pending"
	_PAYMENT_COMPLETED PaymentStatus = "completed"
	_PAYMENT_FAILED    PaymentStatus = "failed"
)

var PaymentStatuses = struct {
	Void      PaymentStatus
	Pending   PaymentStatus
	Completed PaymentStatus
	Failed    PaymentStatus
}{
	Void:      _PAYMENT_VOID,
	Pending:   _PAYMENT_PENDING,
	Completed: _PAYMENT_COMPLETED,
	Failed:    _PAYMENT_FAILED,
}

type RedirectPreferencesValue string

const (
	_STAY   RedirectPreferencesValue = "stay"
	_LIST   RedirectPreferencesValue = "list"
	_DETAIL RedirectPreferencesValue = "detail"
)

type RedirectPreferences struct {
	Invoice  RedirectPreferencesValue `json:"invoice"`
	Estimate RedirectPreferencesValue `json:"estimate"`
	Customer RedirectPreferencesValue `json:"customer"`
	Vendor   RedirectPreferencesValue `json:"vendor"`
	Item     RedirectPreferencesValue `json:"item"`
	Payment  RedirectPreferencesValue `json:"payment"`
	Order    RedirectPreferencesValue `json:"order"`
}

var RedirectPreference = struct {
	Stay   RedirectPreferencesValue
	List   RedirectPreferencesValue
	Detail RedirectPreferencesValue
}{
	Stay:   _STAY,
	List:   _LIST,
	Detail: _DETAIL,
}

type LineAction string

const (
	ADDED     LineAction = "added"
	UPDATED   LineAction = "updated"
	DELETED   LineAction = "deleted"
	UNCHANGED LineAction = "unchanged"
)

var LineActions = struct {
	Added     LineAction
	Updated   LineAction
	Deleted   LineAction
	Unchanged LineAction
}{
	Added:     ADDED,
	Updated:   UPDATED,
	Deleted:   DELETED,
	Unchanged: UNCHANGED,
}

type Discount struct {
	Val  float64 `json:"value"`
	Type string  `json:"type"`
}

func (d *Discount) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *Discount) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type PaymentAmount struct {
	Amount float64 `json:"amount"`
}

type PaymentBase struct {
	PaymentAmount
	Reference string `json:"reference"`
}

type Cash struct {
	PaymentAmount
}

type Check struct {
	PaymentBase
}

type Card struct {
	PaymentBase
	Last4 int    `json:"last4"`
	Brand string `json:"brand"`
}

type Bt struct {
	PaymentBase
}

type Payment struct {
	Cash  Cash  `json:"cash"`
	Check Check `json:"ck"`
	Card  Card  `json:"card"`
	Bt    Bt    `json:"bt"`
}

func (d *Payment) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *Payment) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type Line struct {
	ID          int        `json:"id"`
	VariantID   int        `json:"variant_id"`
	Unit        int        `json:"unit"`
	Qty         int        `json:"qty"`
	Price       float64    `json:"price"`
	Rate        float64    `json:"rate"`
	Action      LineAction `json:"action"`
	WarehouseID int        `json:"warehouse_id"`
	tax         float64
	amount      float64
	discount    float64
	total       float64
}

type StoreInvoiceForm struct {
	support.FormRequest
	CustomerID int                `json:"customer_id"`
	Date       time.Time          `json:"date"`
	Terms      string             `json:"terms"`
	TaxReceipt int                `json:"tax_receipt"`
	Discount   Discount           `json:"discount"`
	Notes      string             `json:"notes"`
	Lines      []*Line            `json:"lines"`
	Payment    Payment            `json:"payment"`
	Kind       TransactionKind    `json:"kind"`
	Source     *TransactionSource `json:"source"`
	Recurrence *Recurrence        `json:"recurrence"`
	// considere these fields as protected
	amount     float64
	amountDue  float64
	tax        float64
	total      float64
	paidStatus PaidStatus
	status     InvoiceStatus
	dueOn      *time.Time
	termType   TermType
}

func (form StoreInvoiceForm) Authorize() bool {
	return Can(form.User(), "create:invoice")
}

func (form StoreInvoiceForm) Rules() map[string]any {
	return map[string]any{
		"customer_id":             "bail|required|exists:customers,id",
		"kind":                    "bail|required|in:invoice,estimate,order,template",
		"recurrence":              "bail|sometimes",
		"recurrence.enabled":      "bail|sometimes",
		"recurrence.name":         "required|max:100",
		"recurrence.start_date":   "nullable|date",
		"recurrence.until":        "nullable|date|after_or_equal:start_date",
		"recurrence.frequency":    "required|in:daily,weekly,monthly,quarterly,yearly",
		"recurrence.interval":     "required|integer|min:1",
		"recurrence.weekdays":     "required_if:frequency,weekly|min:1",
		"recurrence.day_of_month": "required_if:frequency,monthly,quarterly,yearly|min:1|max:31",
		"recurrence.month":        "required_if:frequency,yearly|min:1|max:12",
		"source":                  "sometimes",
		"source.type":             "bail|sometimes|in:estimate,order,template",
		"date":                    "bail|sometimes|required_if:kind,invoice,estimate|date|after:yesterday",
		"terms":                   "bail|required_if:kind,invoice,order,template|min:1",
		"tax_receipt":             "bail|sometimes|required_if:kind,invoice|exists:tax_receipts,id",
		"lines":                   "required|min:1",
		"lines.*.id":              "required|exists:items,id",
		"lines.*.unit":            "required|exists:units,id",
		"lines.*.qty":             "required|min:1",
		"lines.*.price":           "required",
		"lines.*.rate":            "required",
		"lines.*.action":          "required|in:added",
		"lines.*.warehouse_id":    "required|exists:warehouses,id",
		"discount":                "required",
		"discount.value": []any{
			"sometimes",
			validator.Rule{}.When(form.Discount.Type == "percentage", "between:0,100", "min:0"),
		},
		"discount.type": "required|in:percentage,fixed",
	}
}

func (StoreInvoiceForm) Messages() map[string]string {
	return map[string]string{
		"customer_id.required": "You must specify the customer you want to invoice.",
		"lines.min":            "You must specify at least one item to invoice.",
	}
}

func (form *StoreInvoiceForm) Compute() {
	// compute tax for each line
	form.computeTax()

	form.dueOn = nil
	if form.Kind == TransactionKinds.Estimate || form.Kind == TransactionKinds.Template {
		form.status = InvoiceStatuses.Sent
		form.paidStatus = PaidStatuses.UnPaid
		return
	}
	form.status = InvoiceStatuses.Closed
	form.paidStatus = PaidStatuses.Paid
	form.termType = InvoiceTermType.Cash
	termInDays := getNetDays(form.Terms)
	if termInDays >= 0 {
		form.amountDue = form.total
		form.status = InvoiceStatuses.Sent
		form.paidStatus = PaidStatuses.UnPaid
		form.termType = InvoiceTermType.Credit

		dueDate := form.Date.AddDate(0, 0, termInDays)
		form.dueOn = &dueDate
	}
}

func (form *StoreInvoiceForm) PassedValidation() {
	form.Compute()
}

func (form *StoreInvoiceForm) PrepareForValidation() {
	if form.Kind != TransactionKinds.Template {
		form.Recurrence = nil
	}
}

func (form *StoreInvoiceForm) paymentTotalAmount() float64 {
	return form.Payment.Cash.Amount + form.Payment.Card.Amount + form.Payment.Check.Amount + form.Payment.Bt.Amount
}

func (form *StoreInvoiceForm) computeTax() {

	discountPercentage := form.Discount.Val
	if form.Discount.Type == "fixed" {
		totalAmount := float64(0)
		for _, line := range form.Lines {
			totalAmount += (line.Price * float64(line.Qty))
		}

		discountPercentage = float64(discountPercentage/totalAmount) * 100
	}

	for _, line := range form.Lines {
		if line.Action == LineActions.Deleted {
			continue
		}
		// We can store the line discount on the database
		// We can add a discount value amount to the invoice.
		line.amount = round(line.Price*float64(line.Qty), 2)
		line.discount = round(line.amount*(discountPercentage/100), 2)
		line.tax = round((line.amount-line.discount)*(line.Rate/100), 2)
		line.total = round(line.amount-line.discount+line.tax, 2)

		form.tax += line.tax
		form.amount += line.amount
		form.total += line.total
	}

}

type UpdateInvoiceForm struct {
	StoreInvoiceForm
}

func (form UpdateInvoiceForm) Authorize() bool {
	return Can(form.User(), fmt.Sprintf("update:%s", form.Kind))
}

func (form UpdateInvoiceForm) Rules() map[string]any {
	return map[string]any{
		"customer_id":          "bail|required|exists:customers,id",
		"date":                 "bail|required|date",
		"terms":                "bail|sometimes|required_if:kind,invoice,order|min:1",
		"tax_receipt":          "bail|sometimes|required_if:kind,invoice|exists:tax_receipts,id",
		"lines":                "required|min:1",
		"lines.*.id":           "required|exists:items,id",
		"lines.*.unit":         "required|exists:units,id",
		"lines.*.qty":          "required|min:1", // ADD when rule here, only validate when is the action is added or updated
		"lines.*.price":        "required",
		"lines.*.rate":         "required",
		"lines.*.action":       "required|in:added,updated,deleted,unchanged",
		"lines.*.warehouse_id": "required_unless:lines.*.action,deleted,unchanged|exists:warehouses,id",
		"discount":             "required",
		"discount.value": []any{
			"sometimes",
			validator.Rule{}.When(form.Discount.Type == "percentage", "between:0,100", "min:0"),
		},
		"discount.type": "required|in:percentage,fixed",
	}
}

type StorePurchaseForm struct {
	support.FormRequest
	VendorID      int                     `json:"vendor_id"`
	Date          time.Time               `json:"date"`
	Terms         string                  `json:"terms"`
	Discount      Discount                `json:"discount"`
	Notes         string                  `json:"notes"`
	Lines         []*Line                 `json:"lines"`
	Kind          PurchaseTransactionKind `json:"kind"`
	Source        *PurchaseSource         `json:"source"`
	InvoiceNumber string                  `json:"invoice_number"`
	// protected
	amount        float64
	amountDue     float64
	tax           float64
	total         float64
	paymentStatus PaidStatus
	dueOn         *time.Time
}

func (form StorePurchaseForm) Authorize() bool {
	return Can(form.User(), "create:purchase")
}

func (form StorePurchaseForm) Rules() map[string]any {
	return map[string]any{
		"vendor_id":      "bail|required|exists:vendors,id",
		"kind":           "bail|required|in:purchase_order,purchase_receipt,vendor_bill",
		"source":         "sometimes",
		"source.type":    "bail|sometimes|in:purchase_order,purchase_receipt,vendor_bill",
		"date":           "bail|required|date|after:yesterday",
		"terms":          "bail|sometimes|min:1",
		"lines":          "required|min:1",
		"lines.*.id":     "required|exists:items,id",
		"lines.*.unit":   "required|exists:units,id",
		"lines.*.qty":    "required|min:1",
		"lines.*.price":  "required",
		"lines.*.rate":   "required",
		"lines.*.action": "required|in:added",
		"invoice_number": validator.Rule{}.When(string(form.Kind) == "vendor_bill", "required|min:1|max:100", "sometimes"),
		"discount":       "required",
		"discount.value": []any{
			"sometimes",
			validator.Rule{}.When(form.Discount.Type == "percentage", "between:0,100", "min:0"),
		},
		"discount.type": "required|in:percentage,fixed",
	}
}

func (StorePurchaseForm) Messages() map[string]string {
	return map[string]string{
		"vendor_id.required": "You must specify the vendor you want to purchase from.",
		"lines.min":          "You must specify at least one item.",
	}
}

func (form *StorePurchaseForm) Compute() {
	form.computeTax()

	form.amountDue = form.total
	form.paymentStatus = PaidStatuses.UnPaid
	form.dueOn = nil

	termInDays := getNetDays(form.Terms)
	if termInDays >= 0 {
		dueDate := form.Date.AddDate(0, 0, termInDays)
		form.dueOn = &dueDate
	}
}

func (form *StorePurchaseForm) PassedValidation() {
	form.Compute()
}

func (form *StorePurchaseForm) computeTax() {
	discountPercentage := form.Discount.Val
	if form.Discount.Type == "fixed" {
		totalAmount := float64(0)
		for _, line := range form.Lines {
			totalAmount += (line.Price * float64(line.Qty))
		}

		if totalAmount > 0 {
			discountPercentage = float64(discountPercentage/totalAmount) * 100
		}
	}

	for _, line := range form.Lines {
		if line.Action == LineActions.Deleted {
			continue
		}
		line.amount = round(line.Price*float64(line.Qty), 2)
		line.discount = round(line.amount*(discountPercentage/100), 2)
		line.tax = round((line.amount-line.discount)*(line.Rate/100), 2)
		line.total = round(line.amount-line.discount+line.tax, 2)

		form.tax += line.tax
		form.amount += line.amount
		form.total += line.total
	}
}

type UpdatePurchaseForm struct {
	StorePurchaseForm
}

func (form UpdatePurchaseForm) Authorize() bool {
	return Can(form.User(), "update:purchase")
}

func (form UpdatePurchaseForm) Rules() map[string]any {
	return map[string]any{
		"vendor_id":      "bail|required|exists:vendors,id",
		"kind":           "bail|required|in:purchase_order,purchase_receipt,vendor_bill",
		"date":           "bail|required|date",
		"terms":          "bail|sometimes|min:1",
		"lines":          "required|min:1",
		"lines.*.id":     "required|exists:items,id",
		"lines.*.unit":   "required|exists:units,id",
		"lines.*.qty":    "required|min:1",
		"lines.*.price":  "required",
		"lines.*.rate":   "required",
		"lines.*.action": "required|in:added,updated,deleted,unchanged",
		"discount":       "required",
		"discount.value": []any{
			"sometimes",
			validator.Rule{}.When(form.Discount.Type == "percentage", "between:0,100", "min:0"),
		},
		"discount.type": "required|in:percentage,fixed",
	}
}

// VendorPaymentLine is a single AP bill being settled in a vendor payment.
type VendorPaymentLine struct {
	ID        int        `json:"id"`
	UUID      string     `json:"uuid"`
	AmountDue float64    `json:"amount_due"`
	Payment   float64    `json:"payment"`
	Action    LineAction `json:"action"`
}

type StoreVendorPaymentForm struct {
	support.FormRequest
	VendorID string               `json:"vendor_id"`
	Date     time.Time            `json:"date"`
	Notes    string               `json:"notes"`
	Lines    []*VendorPaymentLine `json:"lines"`
	Payment  Payment              `json:"payment"`
	Amount   float64              `json:"amount"`
}

func (form StoreVendorPaymentForm) Authorize() bool {
	return Can(form.User(), "create:payable")
}

func (form StoreVendorPaymentForm) Rules() map[string]any {
	return map[string]any{
		"vendor_id":          "bail|required|exists:vendors,uuid",
		"date":               "bail|required|date|after:yesterday",
		"notes":              "sometime",
		"lines":              "required|min:1",
		"lines.*.uuid":       "required|exists:accounts_payable,uuid",
		"lines.*.amount_due": "required",
		"lines.*.payment":    "required|min:0",
	}
}

type UpdateVendorPaymentForm struct {
	support.FormRequest
	VendorID string               `json:"vendor_id"`
	Date     time.Time            `json:"date"`
	Notes    string               `json:"notes"`
	Lines    []*VendorPaymentLine `json:"lines"`
	Payment  Payment              `json:"payment"`
	Amount   float64              `json:"amount"`
}

func (form UpdateVendorPaymentForm) Authorize() bool {
	return Can(form.User(), "update:payable")
}

func (form UpdateVendorPaymentForm) Rules() map[string]any {
	return map[string]any{
		"vendor_id":          "bail|required|exists:vendors,uuid",
		"date":               "bail|required|date",
		"notes":              "sometime",
		"lines":              "required|min:1",
		"lines.*.id":         "required|exists:vendor_payment_items,id",
		"lines.*.uuid":       "required|exists:accounts_payable,uuid",
		"lines.*.amount_due": "required",
		"lines.*.payment":    "required|min:0",
		"lines.*.action":     "required|in:added,updated,deleted,unchanged",
	}
}

type PaymentLine struct {
	ID        int        `json:"id"`
	Uuid      string     `json:"uuid"`
	AmountDue float64    `json:"amount_due"`
	Payment   float64    `json:"payment"`
	Discount  float64    `json:"discount"`
	Action    LineAction `json:"action"`
}

type StorePaymentForm struct {
	support.FormRequest
	CustomerID string         `json:"customer_id"`
	Date       time.Time      `json:"date"`
	Notes      string         `json:"notes"`
	Lines      []*PaymentLine `json:"lines"`
	Payment    Payment        `json:"payment"`
	Amount     float64        `json:"amount"`
}

func (form StorePaymentForm) Authorize() bool {
	return Can(form.User(), "create:payment")
}

func (form StorePaymentForm) Rules() map[string]any {
	return map[string]any{
		"customer_id":        "bail|required|exists:customers,uuid",
		"date":               "bail|required|date|after:yesterday",
		"notes":              "sometime",
		"lines":              "required|min:1",
		"lines.*.uuid":       "required|exists:invoices,uuid",
		"lines.*.amount_due": "required",
		"lines.*.payment":    "required|min:0",
		"lines.*.discount":   "sometimes",
		// "lines.*.action": "required|in:added",
	}
}

type UpdatePaymentForm struct {
	support.FormRequest
	CustomerID string         `json:"customer_id"`
	Date       time.Time      `json:"date"`
	Notes      string         `json:"notes"`
	Lines      []*PaymentLine `json:"lines"`
	Payment    Payment        `json:"payment"`
	Amount     float64        `json:"amount"`
}

func (form UpdatePaymentForm) Authorize() bool {
	return Can(form.User(), "update:payment")
}

func (form UpdatePaymentForm) Rules() map[string]any {
	return map[string]any{
		"customer_id":        "bail|required|exists:customers,uuid",
		"date":               "bail|required|date",
		"notes":              "sometime",
		"lines":              "required|min:1",
		"lines.*.id":         "required|exists:receivables_income_items,id",
		"lines.*.uuid":       "required|exists:invoices,uuid",
		"lines.*.amount_due": "required",
		"lines.*.payment":    "required|min:0",
		"lines.*.discount":   "sometimes",
		"lines.*.action":     "required|in:added,updated,deleted,unchanged",
	}
}

type MustVerifyAccount interface {
	HasVerifiedAccount() bool
	MarkAccountAsVerified(*sql.DB) bool
	SendAccountVerificationNotification(mailer.Mailer, map[string]string)
	GetEmailAddressForAccountVerification() string
}

type EmailVerificationForm struct {
	support.FormRequest
	Email string `json:"email"`
}

func (EmailVerificationForm) Rules() map[string]any {
	return map[string]any{
		"email": "required|email",
	}
}

type StoreCompanyForm struct {
	support.FormRequest
	Name    string `json:"name"`
	RNC     string `json:"rnc"`
	City    string `json:"city"`
	Address string `json:"address"`
}

func (form StoreCompanyForm) Authorize() bool {
	db := form.Context().Value(database.ConnectionKey{}).(*sql.DB)
	u := UserFromFoundationUser(form.User())
	if u.IsOwner(db) {
		return true
	}
	return Can(form.User(), "create:company") // OR OWNER
}

func (StoreCompanyForm) Rules() map[string]any {
	return map[string]any{
		"name": []any{
			"required",
			"min:3",
			validator.Rule{}.Unique("companies", "name"),
		},
		"rnc":     "required|min:9|max:11:unique:companies,rnc",
		"city":    "required",
		"address": "required",
	}
}

type StoreProfileForm struct {
	support.FormRequest
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (form StoreProfileForm) Rules() map[string]any {

	db := form.Context().Value(database.ConnectionKey{}).(*sql.DB)
	if db == nil {
		panic("database connection need to be set.")
	}

	var userId int
	if err := db.QueryRow("SELECT owner_id FROM accounts WHERE uuid = $1", form.Param("account")).Scan(&userId); err != nil {
		panic("unable to find the account owner.")
	}

	return map[string]any{
		"name": "required|min:3",
		"email": []any{
			"required",
			"email",
			"min:8",
			"max:120",
			"lowercase",
			validator.Rule{}.
				Unique("users", "email").
				Ignore(userId, "id"),
		},
	}
}

type CompanyRole struct {
	Company string `json:"company"`
	Role    string `json:"role"`
}

type StoreUserForm struct {
	support.FormRequest
	Name      string        `json:"name"`
	Email     string        `json:"email"`
	Companies []CompanyRole `json:"companies"`
}

func (form StoreUserForm) Authorize() bool {
	return Can(form.User(), "create:user")
}

func (form StoreUserForm) Rules() map[string]any {
	return map[string]any{
		"name": "required|min:3",
		"email": []any{
			"required",
			"email",
			"min:8",
			"max:120",
			"lowercase",
			validator.Rule{}.Unique("users", "email").Ignore(form.Param("id"), "uuid"),
		},
		"companies":           "required|min:1",
		"companies.*.company": "required|exists:companies,uuid",
		"companies.*.role":    "required|in:admin,supervisor,standard",
	}
}

type User struct {
	AuthUser
	account *account
	Linked  int `json:"linked"`
}

func (u *User) SendEmailVerification(notify mailer.Mailer, attributes map[string]string) {
	url, err := routing.TemporarySignedURL(
		attributes["url"],
		map[string]string{},
		attributes["secret"],
		60*time.Minute,
	)
	if err != nil {
		log.Fatal(err)
		return
	}

	notify.
		To(u.Email, u.Name).
		Send(mail.NewVerification(foundation.AsMap(u), url))
}

func (u *User) SendEmailVerificationChange(notify mailer.Mailer, attributes map[string]string) {
	url, err := routing.TemporarySignedURL(
		attributes["url"],
		map[string]string{},
		attributes["secret"],
		60*time.Minute,
	)
	if err != nil {
		log.Fatal(err)
		return
	}

	notify.
		To(*u.PendingEmail, u.Name).
		Send(mail.NewVerification(foundation.AsMap(u), url))
}

func (u *User) HasVerifiedEmail() bool {
	return u.EmailVerifiedAt != nil
}

func (u *User) MarkEmailAsVerified(db *sql.DB) bool {
	_, err := db.Exec("UPDATE users SET email_verified_at = now(), updated_at = now() WHERE id = $1", u.Id)
	return err == nil
}

func (u *User) Account(db *sql.DB) *account {
	var a = new(account)
	if err := db.QueryRow("SELECT id, uuid, owner_id, status, verified_at, created_at, updated_at, deleted_at FROM accounts WHERE owner_id = $1", u.Id).
		Scan(&a.ID, &a.UUID, &a.Owner.ID, &a.Status, &a.VerifiedAt, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt); err != nil {
		log.Println("An error occurred fetching the account using the ownerID:", err)
		return nil
	}

	u.account = a

	return a
}

func (u *User) OwnedBy(db *sql.DB) (*account, error) {
	if u.account != nil {
		return u.account, nil
	}

	var a = new(account)
	if err := db.QueryRow(`
    SELECT accounts.id, accounts.uuid, accounts.owner_id, accounts.status, accounts.verified_at, accounts.created_at, accounts.updated_at, accounts.deleted_at
    FROM accounts
    INNER JOIN accounts_users on accounts.id = accounts_users.account_id
    WHERE accounts_users.user_id = $1
  `, u.Id).
		Scan(&a.ID, &a.UUID, &a.Owner.ID, &a.Status, &a.VerifiedAt, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt); err != nil {
		return nil, err
	}

	u.account = a

	return a, nil
}

func (u *User) IsOwner(db *sql.DB) bool {
	return u.Account(db) != nil
}

func (u *User) IsNotOwner(db *sql.DB) bool {
	return !u.IsOwner(db)
}

func (u *User) IsOwned(db *sql.DB) bool {
	_, err := u.OwnedBy(db)
	return err == nil
}

func (u *User) IsNotOwned(db *sql.DB) bool {
	return !u.IsOwned(db)
}

func (u *User) IsOrphan(db *sql.DB) bool {
	return u.IsNotOwned(db) && u.IsNotOwner(db)
}

func UserFromContext(ctx context.Context) *User {
	return UserFromFoundationUser(auth.User(ctx))
}

func UserFromFoundationUser(u foundation.Authenticatable) *User {
	au, ok := u.(*AuthUser)
	if !ok || au == nil {
		return &User{}
	}
	return &User{
		AuthUser: *au,
	}
}

func (u *User) currentCompany(db *sql.DB) (*Company, error) {
	result := db.QueryRow(`
    SELECT companies.id, companies.uuid, companies.name, companies.identifier, companies.city,
    companies.address, companies.created_at, companies.updated_at, companies_users.role
    FROM companies
    JOIN companies_users ON companies.id = companies_users.company_id
    WHERE companies_users.user_id = $1 AND companies_users.current = true
  `, u.Id)
	var company Company
	err := result.Scan(
		&company.ID,
		&company.UUID,
		&company.Name,
		&company.Identifier,
		&company.City,
		&company.Address,
		&company.CreatedAt,
		&company.UpdatedAt,
		&company.UserRole,
	)
	if err != nil {
		return nil, err
	}

	return &company, err
}

func CurrentCompany(ctx context.Context) *Company {
	cc := ctx.Value(CompanyKey{})
	if cc == nil {
		return nil
	}

	return cc.(*Company)
}

func CurrentAccount(ctx context.Context) int {
	ac := ctx.Value(AccountKey{})
	if ac == nil {
		return 0
	}

	data, ok := ac.(map[string]any)
	if !ok {
		return 0
	}
	if val, ok := data["id"]; ok {
		if intVal, err := toInt(val); err == nil {
			return intVal
		}
	}
	return 0
}

type SequenceConfig struct {
	Prefix  string `json:"prefix"`
	Next    int    `json:"next"`
	Padding int    `json:"padding"`
}

type InvoiceSequence struct {
	Default SequenceConfig `json:"default"`
	Credit  SequenceConfig `json:"credit"`
	Cash    SequenceConfig `json:"cash"`
}

type CompanySequence struct {
	Invoice  InvoiceSequence `json:"invoice"`
	Template SequenceConfig  `json:"template"`
	Customer SequenceConfig  `json:"customer"`
	Vendor   SequenceConfig  `json:"vendor"`
	Estimate SequenceConfig  `json:"estimate"`
	Payment  SequenceConfig  `json:"payment"`
}

type SequenceForm struct {
	support.FormRequest
	CompanySequence
}

func (SequenceForm) Rules() map[string]any {
	return map[string]any{
		"invoice":                 "required",
		"invoice.default.padding": "required|min:3",
		"invoice.default.next":    "required|min:1",
		"invoice.cash.padding":    "required|min:3",
		"invoice.cash.next":       "required|min:1",
		"invoice.credit.padding":  "required|min:3",
		"invoice.credit.next":     "required|min:1",
		"vendor":                  "required",
		"vendor.padding":          "required|min:3",
		"vendor.next":             "required|min:1",
		"estimate":                "required",
		"estimate.padding":        "required|min:3",
		"estimate.next":           "required|min:1",
		"payment":                 "required",
		"payment.padding":         "required|min:3",
		"payment.next":            "required|min:1",
		"template":                "sometimes",
		"template.padding":        "sometimes|min:3",
		"template.next":           "sometimes|min:1",
	}
}

func (form SequenceForm) Authorize() bool {
	return Can(form.User(), "update:company:sequence")
}

type RedirectPreferencesForm struct {
	support.FormRequest
	Invoice  string `json:"invoice"`
	Estimate string `json:"estimate"`
	Customer string `json:"customer"`
	Vendor   string `json:"vendor"`
	Order    string `json:"order"`
	Item     string `json:"item"`
	Payment  string `json:"payment"`
}

func (RedirectPreferencesForm) Rules() map[string]any {
	return map[string]any{
		"invoice":  "required|in:list,detail,stay",
		"estimate": "required|in:list,detail,stay",
		"customer": "required|in:list,detail,stay",
		"vendor":   "required|in:list,detail,stay",
		"order":    "required|in:list,detail,stay",
		"item":     "required|in:list,detail,stay",
		"payment":  "required|in:list,detail,stay",
	}
}

// HandlesVariantsForm toggles the company-level product-variants feature flag.
type HandlesVariantsForm struct {
	support.FormRequest
	Enabled bool `json:"enabled"`
}

func (HandlesVariantsForm) Rules() map[string]any {
	return map[string]any{
		"enabled": "boolean",
	}
}

type TaxReceiptSequenceForm struct {
	ID    int `json:"id"`
	Start int `json:"start"`
	End   int `json:"end"`
}
type TaxReceiptsForm struct {
	support.FormRequest
	Receipts []TaxReceiptSequenceForm `json:"receipts"`
}

func (TaxReceiptsForm) Rules() map[string]any {
	return map[string]any{
		"receipts":         "required|min:1",
		"receipts.*.id":    "required|exists:shared_tax_receipts,id",
		"receipts.*.start": "required|min:1",
		"receipts.*.end":   "required|min:1|gt:start",
	}
}

type StoreExpenseCategoryForm struct {
	support.FormRequest
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (StoreExpenseCategoryForm) Rules() map[string]any {
	return map[string]any{
		"name":        "required|min:1",
		"description": "required|min:1",
	}
}

type StoreUnitForm struct {
	support.FormRequest
	Name    string `json:"name"`
	BaseQty int    `json:"base_qty"`
}

func (StoreUnitForm) Rules() map[string]any {
	return map[string]any{
		"name":     "required|min:1",
		"base_qty": "required|min:1",
	}
}

type Missing struct {
	Key     string `json:"key"`
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
}

type PrerequisiteResult struct {
	Resource string    `json:"resource"`
	Ok       bool      `json:"ok"`
	Missing  []Missing `json:"missing"`
}

type prereqCacheKeyType struct{}

var prereqCacheKey = prereqCacheKeyType{}

type prereqCache map[string]PrerequisiteResult

var (
	ErrPrerequisitesMissing = errors.New("resource prerequisites missing")
	ErrSettingsNotFound     = errors.New("company settings not found")
	ErrInvalidConfiguration = errors.New("invalid resource configuration")
)

type DateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type PresetRange struct {
	Key  string `json:"key"`
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

type ReportForm struct {
	support.FormRequest
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

func (ReportForm) Rules() map[string]any {
	return map[string]any{
		"from": "required|date",
		"to":   "required|date|before_or_equals:from",
	}
}

type ReportSalesForm struct {
	support.FormRequest
	From         time.Time `json:"from"`
	To           time.Time `json:"to"`
	ReportType   string    `json:"reportType"`
	ShowInvoices bool      `json:"showInvoices"`
}

func (ReportSalesForm) Rules() map[string]any {
	return map[string]any{
		"from":       "required|date",
		"to":         "required|date|before_or_equals:from",
		"reportType": "required|in:sales_by_item,sales_by_customer,sales_by_date",
	}
}

type InvoiceType string

const (
	InvoiceTypeAll    InvoiceType = "all"
	InvoiceTypeCash   InvoiceType = "cash"
	InvoiceTypeCredit InvoiceType = "credit"
)

// Validate ensures the value is one of the allowed constants
func (t InvoiceType) Validate() error {
	switch t {
	case InvoiceTypeAll, InvoiceTypeCash, InvoiceTypeCredit:
		return nil
	default:
		return fmt.Errorf("invalid invoice type: %s", t)
	}
}

type CustomerType string

const (
	CustomerTypeAll        CustomerType = "all"
	CustomerTypeIndividual CustomerType = "individual"
	CustomerTypeBusiness   CustomerType = "business"
)

// Validate ensures the value is one of the allowed constants
func (t CustomerType) Validate() error {
	switch t {
	case CustomerTypeAll, CustomerTypeIndividual, CustomerTypeBusiness:
		return nil
	default:
		return fmt.Errorf("invalid customer type: %s", t)
	}
}

type VendorType string

const (
	VendorTypeAll        VendorType = "all"
	VendorTypeIndividual VendorType = "individual"
	VendorTypeBusiness   VendorType = "business"
)

// Validate ensures the value is one of the allowed constants
func (t VendorType) Validate() error {
	switch t {
	case VendorTypeAll, VendorTypeIndividual, VendorTypeBusiness:
		return nil
	default:
		return fmt.Errorf("invalid vendor type: %s", t)
	}
}

type ItemType string

const (
	ItemTypeAll     ItemType = "all"
	ItemTypeProduct ItemType = "product"
	ItemTypeService ItemType = "service"
)

// Validate ensures the value is one of the allowed constants
func (t ItemType) Validate() error {
	switch t {
	case ItemTypeAll, ItemTypeProduct, ItemTypeService:
		return nil
	default:
		return fmt.Errorf("invalid item type: %s", t)
	}
}

type StoreTaxForm struct {
	support.FormRequest
	Name string  `json:"name"`
	Rate float64 `json:"rate"`
}

func (StoreTaxForm) Rules() map[string]any {
	return map[string]any{
		"name": "required|min:2",
		"rate": "required|min:0|max:99",
	}
}

type UploadSessionForm struct {
	support.FormRequest
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	Mime      string `json:"mime"`
	Delimiter string `json:"delimiter"`
	Encoding  string `json:"encoding"`
}

func (UploadSessionForm) Rules() map[string]any {
	return map[string]any{
		"mime":      "required",
		"filename":  "required",
		"size":      "required",
		"delimiter": "required",
		"encoding":  "required",
	}
}

type UploadChunkForm struct {
	support.FormRequest
	UploadId    string         `json:"upload_id"`
	ChunkIndex  int            `json:"chunk_index"`
	TotalChunks int            `json:"total_chunks"`
	Filename    string         `json:"filename"`
	Chunk       multipart.File `json:"chunk"`
}

func (UploadChunkForm) Rules() map[string]any {
	return map[string]any{
		"upload_id": "required",
		// "chunk_index":  "required|min:0",
		"total_chunks": "required",
		"filename":     "required",
		// "chunk":        "required",
	}
}

type CompleteUploadForm struct {
	support.FormRequest
	UploadID string `json:"upload_id"`
	Filename string `json:"filename"`
}

func (CompleteUploadForm) Rules() map[string]any {
	return map[string]any{
		"upload_id": "required",
		"filename":  "required",
	}
}

type UploadSession struct {
	ID        string `json:"id"`
	UserID    int64  `json:"user_id"`
	Filename  string `json:"filename"`
	FileSize  int64  `json:"file_size"`
	Delimiter string `json:"delimiter"`
	Encoding  string `json:"encoding"`
	// Type           string         `json:"records_type"`
	Status         string         `json:"status"`
	TotalChunks    sql.NullInt64  `json:"total_chunks"`
	UploadedChunks int            `json:"uploaded_chunks"`
	ErrorMessage   sql.NullString `json:"error_message"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type ImportForm struct {
	support.FormRequest
	UploadID string `json:"upload_id"`
	Type     string `json:"type"`
}

func (ImportForm) Rules() map[string]any {
	return map[string]any{
		"upload_id": "required",
		"type":      "required|in:items,customers,vendors",
	}
}

type ImportEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

var importStreams = sync.Map{} // importID → chan ImportEvent

type ImportOptions struct {
	Delimiter rune
}

type UploadEncoding string

const (
	EncodingUTF8    UploadEncoding = "utf-8"
	EncodingLatin1  UploadEncoding = "latin-1"
	EncodingWin1252 UploadEncoding = "windows-1252"
)

var encodingDecoders = map[UploadEncoding]*encoding.Decoder{
	EncodingUTF8:    nil, // no-op
	EncodingLatin1:  charmap.ISO8859_1.NewDecoder(),
	EncodingWin1252: charmap.Windows1252.NewDecoder(),
}

type StoreExpenseForm struct {
	support.FormRequest
	Category string    `json:"category"`
	Date     time.Time `json:"date"`
	Notes    string    `json:"notes"`
	Amount   float64   `json:"amount"`
}

func (form StoreExpenseForm) Authorize() bool {
	return Can(form.User(), "create:expense")
}

func (form StoreExpenseForm) Rules() map[string]any {
	return map[string]any{
		"category": "bail|required|exists:expenses_categories,uuid",
		"date":     "bail|required|date",
		"notes":    "sometime",
		"amount":   "required|min:0",
	}
}

type Date struct {
	time.Time
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}

	formatted := d.Format("2006-01-02")
	return []byte(`"` + formatted + `"`), nil
}

// ── Inventory ─────────────────────────────────────────────────────────────────

// InventoryMovementKind maps to the inv_transaction_kind enum on inventory_movements.
type InventoryMovementKind string

const (
	_INV_MOVEMENT_SALE             InventoryMovementKind = "sale"
	_INV_MOVEMENT_SALE_RETURN      InventoryMovementKind = "sale_return"
	_INV_MOVEMENT_PURCHASE_ORDER   InventoryMovementKind = "purchase_order"
	_INV_MOVEMENT_PURCHASE_RECEIPT InventoryMovementKind = "purchase_receipt"
	_INV_MOVEMENT_PURCHASE_RETURN  InventoryMovementKind = "purchase_return"
	_INV_MOVEMENT_VENDOR_BILL      InventoryMovementKind = "vendor_bill"
	_INV_MOVEMENT_ADJUSTMENT       InventoryMovementKind = "adjustment"
	_INV_MOVEMENT_TRANSFER         InventoryMovementKind = "transfer"
)

var InventoryMovementKinds = struct {
	Sale            InventoryMovementKind
	SaleReturn      InventoryMovementKind
	PurchaseOrder   InventoryMovementKind
	PurchaseReceipt InventoryMovementKind
	PurchaseReturn  InventoryMovementKind
	VendorBill      InventoryMovementKind
	Adjustment      InventoryMovementKind
	Transfer        InventoryMovementKind
}{
	Sale:            _INV_MOVEMENT_SALE,
	SaleReturn:      _INV_MOVEMENT_SALE_RETURN,
	PurchaseOrder:   _INV_MOVEMENT_PURCHASE_ORDER,
	PurchaseReceipt: _INV_MOVEMENT_PURCHASE_RECEIPT,
	PurchaseReturn:  _INV_MOVEMENT_PURCHASE_RETURN,
	VendorBill:      _INV_MOVEMENT_VENDOR_BILL,
	Adjustment:      _INV_MOVEMENT_ADJUSTMENT,
	Transfer:        _INV_MOVEMENT_TRANSFER,
}

// TransferStatus maps to the transfer_status enum on inventory_transfers.
// Lifecycle: requested -> in_transit -> received, with cancelled reachable
// only from requested (once goods are in transit they must be received).
type TransferStatus string

const (
	_TRANSFER_REQUESTED  TransferStatus = "requested"
	_TRANSFER_IN_TRANSIT TransferStatus = "in_transit"
	_TRANSFER_RECEIVED   TransferStatus = "received"
	_TRANSFER_CANCELLED  TransferStatus = "cancelled"
)

var TransferStatuses = struct {
	Requested TransferStatus
	InTransit TransferStatus
	Received  TransferStatus
	Cancelled TransferStatus
}{
	Requested: _TRANSFER_REQUESTED,
	InTransit: _TRANSFER_IN_TRANSIT,
	Received:  _TRANSFER_RECEIVED,
	Cancelled: _TRANSFER_CANCELLED,
}

// transferTransitions lists the statuses each status may move to.
var transferTransitions = map[TransferStatus][]TransferStatus{
	_TRANSFER_REQUESTED:  {_TRANSFER_IN_TRANSIT, _TRANSFER_CANCELLED},
	_TRANSFER_IN_TRANSIT: {_TRANSFER_RECEIVED},
	_TRANSFER_RECEIVED:   {},
	_TRANSFER_CANCELLED:  {},
}

// CanTransitionTo reports whether the transfer may move from its current
// status to the target status.
func (s TransferStatus) CanTransitionTo(target TransferStatus) bool {
	for _, allowed := range transferTransitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}

type StoreAdjustmentForm struct {
	support.FormRequest
	VariantID   int     `json:"variant_id"`
	WarehouseID int     `json:"warehouse_id"`
	Qty         float64 `json:"qty"`
	Reason      string  `json:"reason"`
	Notes       string  `json:"notes"`
}

func (form StoreAdjustmentForm) Authorize() bool {
	return Can(form.User(), "create:adjustment")
}

func (form StoreAdjustmentForm) Rules() map[string]any {
	return map[string]any{
		"variant_id":   "bail|required|exists:items_variants,id",
		"warehouse_id": "bail|required|exists:warehouses,id",
		"qty":          "bail|required",
		"reason":       "bail|required|min:3|max:255",
		"notes":        "sometimes|max:1000",
	}
}

// TransferLineInput is a single product line on a transfer document. ID is the
// item id (resolved to its default variant on store, like purchases/invoices).
type TransferLineInput struct {
	ID          int     `json:"id"`
	Qty         float64 `json:"qty"`
	Unit        int     `json:"unit"`
	Cost        float64 `json:"cost"`
	Description string  `json:"description"`
}

type StoreTransferForm struct {
	support.FormRequest
	FromWarehouseID int                 `json:"from_warehouse_id"`
	ToWarehouseID   int                 `json:"to_warehouse_id"`
	Date            time.Time           `json:"date"`
	Notes           string              `json:"notes"`
	Lines           []TransferLineInput `json:"lines"`
}

func (form StoreTransferForm) Authorize() bool {
	return Can(form.User(), "create:transfer")
}

func (form StoreTransferForm) Rules() map[string]any {
	return map[string]any{
		"from_warehouse_id": "bail|required|exists:warehouses,id",
		"to_warehouse_id":   "bail|required|exists:warehouses,id",
		"notes":             "sometimes|max:1000",
		"lines":             "bail|required",
	}
}

// attribute is a product characteristic (Color, Size, ...) with a set of values.
type attribute struct {
	ID          int               `json:"id"`
	UUID        string            `json:"uuid"`
	Name        string            `json:"name"`
	Type        string            `json:"type"` // "select", "text", "numeric"
	DisplayName string            `json:"display_name"`
	Description *string           `json:"description,omitempty"`
	Values      []*attributeValue `json:"values,omitempty"`
	foundation.Timestamps
}

// attributeValue is a specific value for an attribute (Red, Blue, S, M, L, ...).
type attributeValue struct {
	ID          int    `json:"id"`
	UUID        string `json:"uuid"`
	AttributeID int    `json:"attribute_id"`
	Value       string `json:"value"`
	DisplayName string `json:"display_name"`
	SortOrder   int    `json:"sort_order"`
	foundation.Timestamps
}

// StoreAttributeForm handles attribute creation and update.
type StoreAttributeForm struct {
	support.FormRequest
	Name        string `json:"name"`
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
	Description string `json:"description,omitempty"`
}

func (StoreAttributeForm) Rules() map[string]any {
	return map[string]any{
		"name":         []any{"required", "min:2", "max:120"},
		"type":         "required|in:select,text,numeric",
		"display_name": []any{"required", "min:2", "max:255"},
		"description":  "sometimes|max:1000",
	}
}

func (form StoreAttributeForm) Authorize() bool {
	return Can(form.User(), "create:attribute")
}

// StoreAttributeValueForm handles attribute-value creation and update.
type StoreAttributeValueForm struct {
	support.FormRequest
	AttributeID int    `json:"-"`
	Value       string `json:"value"`
	DisplayName string `json:"display_name"`
	SortOrder   int    `json:"sort_order,omitempty"`
}

func (StoreAttributeValueForm) Rules() map[string]any {
	return map[string]any{
		"value":        []any{"required", "min:1", "max:120"},
		"display_name": []any{"required", "min:1", "max:255"},
		"sort_order":   "sometimes|integer|min:0",
	}
}

func (form StoreAttributeValueForm) Authorize() bool {
	return Can(form.User(), "create:attribute")
}

// itemVariant is a concrete sellable variant of an item (a point in the
// attribute-value matrix, or the lone default variant of a simple item).
type itemVariant struct {
	ID                   int         `json:"id"`
	UUID                 string      `json:"uuid"`
	ItemID               int         `json:"item_id"`
	SKU                  string      `json:"sku"`
	Name                 string      `json:"name"`
	Barcode              *string     `json:"barcode,omitempty"`
	Reference            *string     `json:"reference,omitempty"`
	VendorReference      *string     `json:"vendor_reference,omitempty"`
	CombinationSignature string      `json:"combination_signature"`
	IsDefault            bool        `json:"is_default"`
	Price                *float64    `json:"price,omitempty"`
	CostPrice            *float64    `json:"cost_price,omitempty"`
	TrackInventory       bool        `json:"track_inventory"`
	StockByWarehouse     map[int]int `json:"stock_by_warehouse,omitempty"`
	Active               bool        `json:"active"`
	foundation.Timestamps
}

// VariantCombo is one requested variant: a map of attribute_id -> attribute_value_id
// plus its per-variant pricing/identifiers.
type VariantCombo struct {
	VariantID         int         `json:"variant_id,omitempty"`
	AttributeValueIDs map[int]int `json:"attribute_value_ids"`
	Price             *float64    `json:"price,omitempty"`
	CostPrice         *float64    `json:"cost_price,omitempty"`
	TrackInventory    *bool       `json:"track_inventory,omitempty"`
	StockByWarehouse  map[int]int `json:"stock_by_warehouse,omitempty"`
	SKU               string      `json:"sku,omitempty"`
	Barcode           string      `json:"barcode,omitempty"`
	Reference         string      `json:"reference,omitempty"`
	VendorReference   string      `json:"vendor_reference,omitempty"`
	Active            *bool       `json:"active,omitempty"`
}

// StoreItemWithAttributesForm handles item creation with attributes and variants.
// It carries the same base item fields as StoreItemForm so the item row can be
// created, plus the attribute/variant matrix.
type StoreItemWithAttributesForm struct {
	support.FormRequest
	Name          string          `json:"name"`
	Price         float64         `json:"price"`
	Description   string          `json:"description,omitempty"`
	TaxID         int             `json:"tax_id"`
	UnitID        int             `json:"unit_id"`
	ItemType      string          `json:"item_type"`
	Identifiers   ItemIdentifiers `json:"identifiers,omitempty"`
	AttributeIDs  []int           `json:"attribute_ids,omitempty"`
	VariantCombos []VariantCombo  `json:"variant_combos,omitempty"`
}

func (StoreItemWithAttributesForm) Rules() map[string]any {
	return map[string]any{
		"name":             []any{"required", "min:3", "max:120"},
		"price":            "required|numeric|min:0",
		"description":      "sometimes|max:1000",
		"tax_id":           "required|exists:taxes,id",
		"unit_id":          "required|exists:units,id",
		"item_type":        "required",
		"attribute_ids.*":  "sometimes|exists:attributes,id",
		"variant_combos.*": "sometimes|array",
	}
}

func (form StoreItemWithAttributesForm) Authorize() bool {
	return Can(form.User(), "create:item")
}
