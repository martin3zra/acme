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
		"email":   "required|email|unique.ignore:customers|min:8|max:120|lowercase",
		"phone":   "sometimes|min:3|max:120",
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
		"name":        "required|min:3|max:120|unique.ignore:items,name",
		"description": "sometimes|min:3|max:120",
		"price":       "required|min:0",
		"tax_id":      "required|exists:taxes,id",
		"unit_id":     "required|exists:units,id",
	}
}

type TermType string

const (
	CASH   TermType = "cash"
	CREDIT TermType = "credit"
)

type PaidStatus string

const (
	UNPAID  PaidStatus = "unpaid"
	PARTIAL PaidStatus = "partial"
	PAID    PaidStatus = "paid"
)

type Discount struct {
	Val  float64 `json:"value"`
	Type string  `json:"type"`
}

func (d *Discount) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *Discount) Scan(value any) error {
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
	Check Check `json:"check"`
	Card  Card  `json:"card"`
	Bt    Bt    `json:"bt"`
}

func (d *Payment) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *Payment) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type Line struct {
	ID    int     `json:"id"`
	Unit  int     `json:"unit"`
	Qty   int     `json:"quantity"`
	Price float64 `json:"price"`
	Rate  float64 `json:"rate"`
}

type StoreInvoiceForm struct {
	support.FormRequest
	CustomerID int       `json:"customer_id"`
	Date       time.Time `json:"date"`
	Terms      int       `json:"terms"`
	Discount   Discount  `json:"discount"`
	Notes      string    `json:"notes"`
	Lines      []Line    `json:"lines"`
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
		"customer_id":   "bail|required|exists:customers,id",
		"date":          "bail|required|date|after:yesterday",
		"terms":         "bail|required|min:1",
		"lines":         "required|min:1",
		"lines.*.id":    "required|exists:items,id",
		"lines.*.unit":  "required|exists:units,id",
		"lines.*.qty":   "required|min:1",
		"lines.*.price": "required",
		"lines.*.rate":  "required",
		"discount":      "required",
		"discount.value": []any{
			"required",
			validator.Rule{}.When(form.Discount.Type == "percentage", "between:0,100", "min:0"),
		},
		"discount.type": "required|in:percentage,fixed",
	}
}

func (form *StoreInvoiceForm) PassedValidation() {
	// compute tax for each line
	form.computeTax()
	form.applyDiscount()

	form.dueOn = nil
	form.paidStatus = PAID
	form.termType = CASH
	if form.Terms > 1 {
		form.amountDue = form.total
		form.paidStatus = UNPAID
		form.termType = CREDIT

		dueDate := form.Date.AddDate(0, 0, form.Terms)
		form.dueOn = &dueDate
	}
}

func (form *StoreInvoiceForm) paymentTotalAmount() float64 {
	return form.Payment.Cash.Amount + form.Payment.Card.Amount + form.Payment.Check.Amount + form.Payment.Bt.Amount
}

func (form *StoreInvoiceForm) computeTax() {
	for _, line := range form.Lines {
		// we need to add the discount here.
		lineAmount := (line.Price * float64(line.Qty))

		form.tax += lineAmount * (line.Rate / 100)
		form.amount += lineAmount
		form.total += lineAmount + form.tax
	}
}

func (form *StoreInvoiceForm) applyDiscount() {
	if form.Discount.Type == "percentage" {
		form.total = form.total - (form.total * (form.Discount.Val / 100))
		return
	}

	form.total = form.total - form.Discount.Val
}
