package app

import (
	"time"

	"github.com/martin3zra/forge/support"
)

type DateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type PresetRange struct {
	Key  string `json:"key"`
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

type ReportForm struct {
	support.FormRequest
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

func (ReportForm) Rules() map[string]any {
	return map[string]any{
		"from": "required|date",
		"to":   "required|date|before_or_equals:from",
	}
}

type ReportSalesForm struct {
	support.FormRequest
	From         time.Time `json:"from"`
	To           time.Time `json:"to"`
	ReportType   string    `json:"reportType"`
	ShowInvoices bool      `json:"showInvoices"`
}

func (ReportSalesForm) Rules() map[string]any {
	return map[string]any{
		"from":       "required|date",
		"to":         "required|date|before_or_equals:from",
		"reportType": "required|in:sales_by_item,sales_by_customer,sales_by_date",
	}
}
