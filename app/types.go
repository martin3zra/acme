package app

import (
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
