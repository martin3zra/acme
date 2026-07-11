package app

import (
	"time"

	"github.com/martin3zra/forge/support"
)

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
		"category": []any{"bail", "required", tenantExists(form.Context(), "expenses_categories", "uuid")},
		"date":     "bail|required|date",
		"notes":    "sometime",
		"amount":   "required|min:0",
	}
}
