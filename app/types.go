package app

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/martin3zra/acme/pkg/support"
	"github.com/martin3zra/acme/pkg/validator"
)

type LoginFormRequest struct {
	support.FormRequest
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (f LoginFormRequest) Rules() map[string]any {
	return map[string]any{
		"email":    "required|email|max:100|lowercase",
		"password": "required",
	}
}

type StoreCustomerForm struct {
	support.FormRequest
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
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
		"phone": "sometimes|min:3|max:120",
	}
}

type UpdateCustomerForm struct {
	support.FormRequest
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
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
		"phone": "sometimes|min:3|max:120",
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

type StoreItemForm struct {
	support.FormRequest
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	TaxID       int     `json:"tax_id"`
	UnitID      int     `json:"unit_id"`
}

func (StoreItemForm) Rules() map[string]any {
	return map[string]any{
		"name": []any{
			"required",
			"min:3",
			"max:120",
			validator.Rule{}.Unique("items", "name"),
		},
		"description": "sometimes|min:3|max:120",
		"price":       "required|min:0",
		"tax_id":      "bail|required|exists:taxes,id",
		"unit_id":     "bail|required|exists:units,id",
	}
}

type UpdateItemForm struct {
	support.FormRequest
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	TaxID       int     `json:"tax_id"`
	UnitID      int     `json:"unit_id"`
}

func (form UpdateItemForm) Rules() map[string]any {
	return map[string]any{
		"name":        []any{"required", "min:3", "max:120", validator.Rule{}.Unique("items", "name").Ignore(form.ID, "id")},
		"description": "sometimes|min:3|max:120",
		"price":       "required|min:0",
		"tax_id":      "required|exists:taxes,id",
		"unit_id":     "required|exists:units,id",
	}
}

type TermType string

const (
	_CASH   TermType = "cash"
	_CREDIT TermType = "credit"
)

var InvoiceTermType = struct {
	Cash   TermType
	Credit TermType
}{
	Cash:   _CASH,
	Credit: _CREDIT,
}

type PaidStatus string

const (
	_UNPAID  PaidStatus = "unpaid"
	_PARTIAL PaidStatus = "partial"
	_PAID    PaidStatus = "paid"
	_REMOVED PaidStatus = "removed"
)

var PaidStatuses = struct {
	UnPaid  PaidStatus
	Partial PaidStatus
	Paid    PaidStatus
	Removed PaidStatus
}{
	UnPaid:  _UNPAID,
	Partial: _PARTIAL,
	Paid:    _PAID,
	Removed: _REMOVED,
}

type InvoiceStatus string

const (
	DRAFT     InvoiceStatus = "draft"
	SENT      InvoiceStatus = "sent"
	VIEWED    InvoiceStatus = "viewed"
	OVERDUE   InvoiceStatus = "overdue"
	COMPLETED InvoiceStatus = "completed"
	VOID      InvoiceStatus = "void"
)

var InvoiceStatuses = struct {
	Draft     InvoiceStatus
	Sent      InvoiceStatus
	Viewed    InvoiceStatus
	Overdue   InvoiceStatus
	Completed InvoiceStatus
	Void      InvoiceStatus
}{
	Draft:     DRAFT,
	Sent:      SENT,
	Viewed:    VIEWED,
	Overdue:   OVERDUE,
	Completed: COMPLETED,
	Void:      VOID,
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
	ID       int        `json:"id"`
	Unit     int        `json:"unit"`
	Qty      int        `json:"qty"`
	Price    float64    `json:"price"`
	Rate     float64    `json:"rate"`
	Action   LineAction `json:"action"`
	tax      float64
	amount   float64
	discount float64
	total    float64
}

type StoreInvoiceForm struct {
	support.FormRequest
	CustomerID int       `json:"customer_id"`
	Date       time.Time `json:"date"`
	Terms      int       `json:"terms"`
	TaxReceipt int       `json:"tax_receipt"`
	Discount   Discount  `json:"discount"`
	Notes      string    `json:"notes"`
	Lines      []*Line   `json:"lines"`
	Payment    Payment   `json:"payment"`
	// considere these fields as protected
	amount     float64
	amountDue  float64
	tax        float64
	total      float64
	paidStatus PaidStatus
	dueOn      *time.Time
	termType   TermType
}

func (form StoreInvoiceForm) Rules() map[string]any {
	return map[string]any{
		"customer_id":    "bail|required|exists:customers,id",
		"date":           "bail|required|date|after:yesterday",
		"terms":          "bail|required|min:1",
		"tax_receipt":    "bail|required|exists:tax_receipts,id",
		"lines":          "required|min:1",
		"lines.*.id":     "required|exists:items,id",
		"lines.*.unit":   "required|exists:units,id",
		"lines.*.qty":    "required|min:1",
		"lines.*.price":  "required",
		"lines.*.rate":   "required",
		"lines.*.action": "required|in:added",
		"discount":       "required",
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

func (form *StoreInvoiceForm) PassedValidation() {
	// compute tax for each line
	form.computeTax()

	form.dueOn = nil
	form.paidStatus = PaidStatuses.Paid
	form.termType = InvoiceTermType.Cash
	if form.Terms > 1 {
		form.amountDue = form.total
		form.paidStatus = PaidStatuses.UnPaid
		form.termType = InvoiceTermType.Credit

		dueDate := form.Date.AddDate(0, 0, form.Terms)
		form.dueOn = &dueDate
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
		// We can store the line discoun on the database
		// We can add a discount value amount to the invoice.
		line.amount = (line.Price * float64(line.Qty))
		line.discount = line.amount * (discountPercentage / 100)
		line.tax = (line.amount - line.discount) * (line.Rate / 100)
		line.total = line.amount - line.discount + line.tax

		form.tax += line.tax
		form.amount += line.amount
		form.total += line.total
	}

}

type UpdateInvoiceForm struct {
	StoreInvoiceForm
}

func (form UpdateInvoiceForm) Rules() map[string]any {
	return map[string]any{
		"customer_id":    "bail|required|exists:customers,id",
		"date":           "bail|required|date",
		"terms":          "bail|required|min:1",
		"tax_receipt":    "bail|required|exists:tax_receipts,id",
		"lines":          "required|min:1",
		"lines.*.id":     "required|exists:items,id",
		"lines.*.unit":   "required|exists:units,id",
		"lines.*.qty":    "required|min:1", // ADD when rule here, only validate when is the action is added or updated
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

type PaymentLine struct {
	Uuid      string  `json:"uuid"`
	AmountDue float64 `json:"amount_due"`
	Payment   float64 `json:"payment"`
	Discount  float64 `json:"discount"`
}

type StorePaymentForm struct {
	support.FormRequest
	CustomerID string         `json:"customer_id"`
	Date       time.Time      `json:"date"`
	Lines      []*PaymentLine `json:"lines"`
	Payment    Payment        `json:"payment"`
	Amount     float64        `json:"amount"`
}

func (form StorePaymentForm) Rules() map[string]any {
	return map[string]any{
		"customer_id":        "bail|required|exists:customers,uuid",
		"date":               "bail|required|date|after:yesterday",
		"lines":              "required|min:1",
		"lines.*.uuid":       "required|exists:invoices,uuid",
		"lines.*.amount_due": "required",
		"lines.*.payment":    "required|min:0",
		"lines.*.discount":   "sometimes",
		// "lines.*.action": "required|in:added",
	}
}
