package app

import (
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

type StoreInvoiceForm struct {
	support.FormRequest
	CustomerID int       `json:"customer_id"`
	Date       time.Time `json:"date"`
	Terms      int       `json:"terms"`
	Notes      string    `json:"notes"`
	Lines      []struct {
		ID    int     `json:"id"`
		Qty   int     `json:"quantity"`
		Price float64 `json:"price"`
		Rate  float64 `json:"rate"`
	} `json:"lines"`

	// considere these fields as protected
	Amount    float64
	AmountDue float64
	Tax       float64
	Total     float64

	// Protected fields
	paidStatus PaidStatus
}

func (StoreInvoiceForm) Rules() map[string]any {
	return map[string]any{
		"customer_id": "required|exists:customers,id",
		"date":        "required",
		"terms":       "required",
		"lines":       "required",
	}
}

func (form *StoreInvoiceForm) PassedValidation() {
	// compute tax for each line
	for _, line := range form.Lines {
		// we need to add the discount here.
		lineAmount := (line.Price * float64(line.Qty))

		form.Tax += lineAmount * (line.Rate / 100)
		form.Amount += lineAmount
		form.Total += lineAmount + form.Tax
	}

	form.paidStatus = PAID
	if form.Terms > 1 {
		form.AmountDue = form.Amount
		form.paidStatus = UNPAID
	}
}
