package app

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/martin3zra/forge/support"
	"github.com/martin3zra/forge/validator"
)

const PaymentTermsMax = 120 // net120

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

// Validate ensures the value is one of the allowed constants. The value reaches
// the transaction_kind enum column, so an unchecked string is a query-time error.
func (t TransactionKind) Validate() error {
	switch t {
	case _TRANSACTION_KIND_INVOICE, _TRANSACTION_KIND_ESTIMATE, _TRANSACTION_KIND_ORDER, _TRANSACTION_KIND_TEMPLATE:
		return nil
	default:
		return fmt.Errorf("invalid transaction kind: %s", t)
	}
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
		"customer_id":             []any{"bail", "required", tenantExists(form.Context(), "customers", "id")},
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
		"tax_receipt":             []any{"bail", "sometimes", "required_if:kind,invoice", tenantExists(form.Context(), "tax_receipts", "id")},
		"lines":                   "required|min:1",
		"lines.*.id":              []any{"required", tenantExists(form.Context(), "items", "id")},
		"lines.*.unit":            []any{"required", tenantExists(form.Context(), "units", "id")},
		"lines.*.qty":             "required|min:1",
		"lines.*.price":           "required",
		"lines.*.rate":            "required",
		"lines.*.action":          "required|in:added",
		"lines.*.warehouse_id":    []any{"required", tenantExists(form.Context(), "warehouses", "id")},
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
		"customer_id":          []any{"bail", "required", tenantExists(form.Context(), "customers", "id")},
		"date":                 "bail|required|date",
		"terms":                "bail|sometimes|required_if:kind,invoice,order|min:1",
		"tax_receipt":          []any{"bail", "sometimes", "required_if:kind,invoice", tenantExists(form.Context(), "tax_receipts", "id")},
		"lines":                "required|min:1",
		"lines.*.id":           []any{"required", tenantExists(form.Context(), "items", "id")},
		"lines.*.unit":         []any{"required", tenantExists(form.Context(), "units", "id")},
		"lines.*.qty":          "required|min:1", // ADD when rule here, only validate when is the action is added or updated
		"lines.*.price":        "required",
		"lines.*.rate":         "required",
		"lines.*.action":       "required|in:added,updated,deleted,unchanged",
		"lines.*.warehouse_id": []any{"required_unless:lines.*.action,deleted,unchanged", tenantExists(form.Context(), "warehouses", "id")},
		"discount":             "required",
		"discount.value": []any{
			"sometimes",
			validator.Rule{}.When(form.Discount.Type == "percentage", "between:0,100", "min:0"),
		},
		"discount.type": "required|in:percentage,fixed",
	}
}
