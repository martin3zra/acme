package app

import (
	"fmt"
	"time"

	"github.com/martin3zra/forge/support"
	"github.com/martin3zra/forge/validator"
)

type StoreCustomerForm struct {
	support.FormRequest
	Name            string    `json:"name"`
	Contact         string    `json:"contact"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	Address         string    `json:"address"`
	PaymentMethod   string    `json:"payment_method"`
	PaymentTerms    string    `json:"payment_terms"`
	CreditLimited   bool      `json:"credit_limited"`
	CreditLimit     float64   `json:"credit_limit"`
	CustomerType    string    `json:"customer_type"`
	TaxReceipt      int       `json:"tax_receipt"`
	OpenBalance     float64   `json:"open_balance"`
	OpenBalanceAsOf time.Time `json:"open_balance_as_of"`
}

func (form StoreCustomerForm) Rules() map[string]any {
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
		"address":            "sometimes|max:255",
		"payment_method":     "sometimes|in:cash,ck,card,bt",
		"payment_terms":      "sometimes|required",
		"credit_limited":     "required",
		"credit_limit":       "sometimes|required|min:0",
		"customer_type":      "sometimes|required|in:individual,business",
		"tax_receipt":        []any{"sometimes", tenantExists(form.Context(), "tax_receipts", "id")},
		"open_balance":       "sometimes|min:0",
		"open_balance_as_of": "sometimes",
	}
}

func (form StoreCustomerForm) Authorize() bool {
	return Can(form.User(), "create:customer")
}

type UpdateCustomerForm struct {
	support.FormRequest
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Contact         string    `json:"contact"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	Address         string    `json:"address"`
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
		"address":            "sometimes|max:255",
		"payment_method":     "sometimes|in:cash,ck,card,bt",
		"payment_terms":      "sometimes|required",
		"credit_limit":       "sometimes|required|min:0",
		"customer_type":      "sometimes|required|in:individual,business",
		"tax_receipt":        []any{"sometimes", tenantExists(form.Context(), "tax_receipts", "id")},
		"open_balance":       "sometimes|min:0",
		"open_balance_as_of": "sometimes",
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
