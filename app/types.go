package app

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/martin3zra/acme/pkg/support"
)

type LoginFormRequest struct {
	support.FormRequest
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (f LoginFormRequest) Rules() map[string]any {
	return map[string]any{
		"email":    "required|email|max:100",
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
		"email":   "required|email|unique.ignore:customers|min:8|max:120",
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

type StoreInvoiceForm struct {
	support.FormRequest
	CustomerID int       `json:"customer_id"`
	Date       time.Time `json:"date"`
	Terms      int       `json:"terms"`
	Discount   Discount  `json:"discount"`
	Notes      string    `json:"notes"`
	Lines      []struct {
		ID    int     `json:"id"`
		Unit  int     `json:"unit"`
		Qty   int     `json:"quantity"`
		Price float64 `json:"price"`
		Rate  float64 `json:"rate"`
	} `json:"lines"`

	// considere these fields as protected
	amount     float64
	amountDue  float64
	tax        float64
	total      float64
	paidStatus PaidStatus
}

func (StoreInvoiceForm) Rules() map[string]any {
	return map[string]any{
		"customer_id": "required|exists:customers,id",
		"date":        "required",
		"terms":       "required",
		"lines":       "required",
		"discount":    "required",
	}
}

func (form *StoreInvoiceForm) PassedValidation() {
	// compute tax for each line
	form.computeTax()
	form.applyDiscount()

	form.paidStatus = PAID
	if form.Terms > 1 {
		form.amountDue = form.total
		form.paidStatus = UNPAID
	}
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
