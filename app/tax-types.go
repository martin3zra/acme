package app

import (
	"github.com/martin3zra/forge/support"
)

type TaxReceiptSequenceForm struct {
	ID    int `json:"id"`
	Start int `json:"start"`
	End   int `json:"end"`
}
type TaxReceiptsForm struct {
	support.FormRequest
	Receipts []TaxReceiptSequenceForm `json:"receipts"`
}

func (TaxReceiptsForm) Rules() map[string]any {
	return map[string]any{
		"receipts":         "required|min:1",
		"receipts.*.id":    "required|exists:shared_tax_receipts,id",
		"receipts.*.start": "required|min:1",
		"receipts.*.end":   "required|min:1|gt:start",
	}
}

type StoreTaxForm struct {
	support.FormRequest
	Name string  `json:"name"`
	Rate float64 `json:"rate"`
}

func (StoreTaxForm) Rules() map[string]any {
	return map[string]any{
		"name": "required|min:2",
		"rate": "required|min:0|max:99",
	}
}
