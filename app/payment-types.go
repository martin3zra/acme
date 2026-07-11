package app

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/martin3zra/forge/support"
)

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
		"vendor_id":          []any{"bail", "required", tenantExists(form.Context(), "vendors", "uuid")},
		"date":               "bail|required|date|after:yesterday",
		"notes":              "sometime",
		"lines":              "required|min:1",
		"lines.*.uuid":       []any{"required", tenantExists(form.Context(), "accounts_payable", "uuid")},
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
		"vendor_id":          []any{"bail", "required", tenantExists(form.Context(), "vendors", "uuid")},
		"date":               "bail|required|date",
		"notes":              "sometime",
		"lines":              "required|min:1",
		"lines.*.id":         []any{"required", tenantExists(form.Context(), "vendor_payment_items", "id")},
		"lines.*.uuid":       []any{"required", tenantExists(form.Context(), "accounts_payable", "uuid")},
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
		"customer_id":        []any{"bail", "required", tenantExists(form.Context(), "customers", "uuid")},
		"date":               "bail|required|date|after:yesterday",
		"notes":              "sometime",
		"lines":              "required|min:1",
		"lines.*.uuid":       []any{"required", tenantExists(form.Context(), "invoices", "uuid")},
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
		"customer_id":        []any{"bail", "required", tenantExists(form.Context(), "customers", "uuid")},
		"date":               "bail|required|date",
		"notes":              "sometime",
		"lines":              "required|min:1",
		"lines.*.id":         []any{"required", tenantExists(form.Context(), "receivables_income_items", "id")},
		"lines.*.uuid":       []any{"required", tenantExists(form.Context(), "invoices", "uuid")},
		"lines.*.amount_due": "required",
		"lines.*.payment":    "required|min:0",
		"lines.*.discount":   "sometimes",
		"lines.*.action":     "required|in:added,updated,deleted,unchanged",
	}
}
