package app

import (
	"fmt"
	"time"

	"github.com/martin3zra/forge/support"
	"github.com/martin3zra/forge/validator"
)

type StoreVendorForm struct {
	support.FormRequest
	Name            string    `json:"name"`
	Contact         string    `json:"contact"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	Address         string    `json:"address"`
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

func (form StoreVendorForm) Rules() map[string]any {
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
		"address":            "sometimes|max:255",
		"purchase_note":      "sometimes|max:2000",
		"lead_time_days":     "sometimes|integer|min:0",
		"payment_method":     "sometimes|in:cash,ck,card,bt",
		"payment_terms":      "sometimes|required",
		"credit_limited":     "required",
		"credit_limit":       "sometimes|required|min:0",
		"vendor_type":        "sometimes|required|in:individual,business",
		"tax_receipt":        []any{"sometimes", tenantExists(form.Context(), "tax_receipts", "id")},
		"open_balance":       "sometimes|min:0",
		"open_balance_as_of": "sometimes",
	}
}

func (form StoreVendorForm) Authorize() bool {
	return Can(form.User(), "create:vendor")
}

type UpdateVendorForm struct {
	support.FormRequest
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Contact         string    `json:"contact"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	Address         string    `json:"address"`
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
		"address":            "sometimes|max:255",
		"purchase_note":      "sometimes|max:2000",
		"lead_time_days":     "sometimes|integer|min:0",
		"payment_method":     "sometimes|in:cash,ck,card,bt",
		"payment_terms":      "sometimes|required",
		"credit_limit":       "sometimes|required|min:0",
		"vendor_type":        "sometimes|required|in:individual,business",
		"tax_receipt":        []any{"sometimes", tenantExists(form.Context(), "tax_receipts", "id")},
		"open_balance":       "sometimes|min:0",
		"open_balance_as_of": "sometimes",
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
