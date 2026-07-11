package app

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/martin3zra/forge/support"
	"github.com/martin3zra/forge/validator"
)

type PurchaseTransactionKind string

const (
	_PURCHASE_ORDER   PurchaseTransactionKind = "purchase_order"
	_PURCHASE_RECEIPT PurchaseTransactionKind = "purchase_receipt"
	_VENDOR_BILL      PurchaseTransactionKind = "vendor_bill"
)

var PurchaseTransactionKinds = struct {
	PurchaseOrder   PurchaseTransactionKind
	PurchaseReceipt PurchaseTransactionKind
	VendorBill      PurchaseTransactionKind
}{
	PurchaseOrder:   _PURCHASE_ORDER,
	PurchaseReceipt: _PURCHASE_RECEIPT,
	VendorBill:      _VENDOR_BILL,
}

// PurchaseStatus represents the lifecycle state of a purchase (purchase_status column).
type PurchaseStatus string

const (
	_PURCHASE_STATUS_DRAFT              PurchaseStatus = "draft"
	_PURCHASE_STATUS_PARTIALLY_RECEIVED PurchaseStatus = "partially_received"
	_PURCHASE_STATUS_RECEIVED           PurchaseStatus = "received"
	_PURCHASE_STATUS_PARTIALLY_PAID     PurchaseStatus = "partially_paid"
	_PURCHASE_STATUS_CLOSED             PurchaseStatus = "closed"
	_PURCHASE_STATUS_POSTED             PurchaseStatus = "posted"
)

var PurchaseStatuses = struct {
	Draft             PurchaseStatus
	PartiallyReceived PurchaseStatus
	Received          PurchaseStatus
	PartiallyPaid     PurchaseStatus
	Closed            PurchaseStatus
	Posted            PurchaseStatus
}{
	Draft:             _PURCHASE_STATUS_DRAFT,
	PartiallyReceived: _PURCHASE_STATUS_PARTIALLY_RECEIVED,
	Received:          _PURCHASE_STATUS_RECEIVED,
	PartiallyPaid:     _PURCHASE_STATUS_PARTIALLY_PAID,
	Closed:            _PURCHASE_STATUS_CLOSED,
	Posted:            _PURCHASE_STATUS_POSTED,
}

// lockedPurchaseStatuses lists purchase statuses where edits and deletes are prohibited.
var lockedPurchaseStatuses = map[PurchaseStatus]bool{
	PurchaseStatuses.Received:          true,
	PurchaseStatuses.PartiallyReceived: true,
	PurchaseStatuses.PartiallyPaid:     true,
	PurchaseStatuses.Closed:            true,
	PurchaseStatuses.Posted:            true,
}

type PurchaseSource struct {
	Type   PurchaseTransactionKind `json:"type,omitempty"`
	ID     string                  `json:"id,omitempty"`
	Code   string                  `json:"code,omitempty"`
	Target *PurchaseSource         `json:"target,omitempty"`
}

func (d *PurchaseSource) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *PurchaseSource) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type StorePurchaseForm struct {
	support.FormRequest
	VendorID      int                     `json:"vendor_id"`
	Date          time.Time               `json:"date"`
	Terms         string                  `json:"terms"`
	Discount      Discount                `json:"discount"`
	Notes         string                  `json:"notes"`
	Lines         []*Line                 `json:"lines"`
	Kind          PurchaseTransactionKind `json:"kind"`
	Source        *PurchaseSource         `json:"source"`
	InvoiceNumber string                  `json:"invoice_number"`
	// protected
	amount        float64
	amountDue     float64
	tax           float64
	total         float64
	paymentStatus PaidStatus
	dueOn         *time.Time
}

func (form StorePurchaseForm) Authorize() bool {
	return Can(form.User(), "create:purchase")
}

func (form StorePurchaseForm) Rules() map[string]any {
	return map[string]any{
		"vendor_id":      []any{"bail", "required", tenantExists(form.Context(), "vendors", "id")},
		"kind":           "bail|required|in:purchase_order,purchase_receipt,vendor_bill",
		"source":         "sometimes",
		"source.type":    "bail|sometimes|in:purchase_order,purchase_receipt,vendor_bill",
		"date":           "bail|required|date|after:yesterday",
		"terms":          "bail|sometimes|min:1",
		"lines":          "required|min:1",
		"lines.*.id":     []any{"required", tenantExists(form.Context(), "items", "id")},
		"lines.*.unit":   []any{"required", tenantExists(form.Context(), "units", "id")},
		"lines.*.qty":    "required|min:1",
		"lines.*.price":  "required",
		"lines.*.rate":   "required",
		"lines.*.action": "required|in:added",
		"invoice_number": validator.Rule{}.When(string(form.Kind) == "vendor_bill", "required|min:1|max:100", "sometimes"),
		"discount":       "required",
		"discount.value": []any{
			"sometimes",
			validator.Rule{}.When(form.Discount.Type == "percentage", "between:0,100", "min:0"),
		},
		"discount.type": "required|in:percentage,fixed",
	}
}

func (StorePurchaseForm) Messages() map[string]string {
	return map[string]string{
		"vendor_id.required": "You must specify the vendor you want to purchase from.",
		"lines.min":          "You must specify at least one item.",
	}
}

func (form *StorePurchaseForm) Compute() {
	form.computeTax()

	form.amountDue = form.total
	form.paymentStatus = PaidStatuses.UnPaid
	form.dueOn = nil

	termInDays := getNetDays(form.Terms)
	if termInDays >= 0 {
		dueDate := form.Date.AddDate(0, 0, termInDays)
		form.dueOn = &dueDate
	}
}

func (form *StorePurchaseForm) PassedValidation() {
	form.Compute()
}

func (form *StorePurchaseForm) computeTax() {
	discountPercentage := form.Discount.Val
	if form.Discount.Type == "fixed" {
		totalAmount := float64(0)
		for _, line := range form.Lines {
			totalAmount += (line.Price * float64(line.Qty))
		}

		if totalAmount > 0 {
			discountPercentage = float64(discountPercentage/totalAmount) * 100
		}
	}

	for _, line := range form.Lines {
		if line.Action == LineActions.Deleted {
			continue
		}
		line.amount = round(line.Price*float64(line.Qty), 2)
		line.discount = round(line.amount*(discountPercentage/100), 2)
		line.tax = round((line.amount-line.discount)*(line.Rate/100), 2)
		line.total = round(line.amount-line.discount+line.tax, 2)

		form.tax += line.tax
		form.amount += line.amount
		form.total += line.total
	}
}

type UpdatePurchaseForm struct {
	StorePurchaseForm
}

func (form UpdatePurchaseForm) Authorize() bool {
	return Can(form.User(), "update:purchase")
}

func (form UpdatePurchaseForm) Rules() map[string]any {
	return map[string]any{
		"vendor_id":      []any{"bail", "required", tenantExists(form.Context(), "vendors", "id")},
		"kind":           "bail|required|in:purchase_order,purchase_receipt,vendor_bill",
		"date":           "bail|required|date",
		"terms":          "bail|sometimes|min:1",
		"lines":          "required|min:1",
		"lines.*.id":     []any{"required", tenantExists(form.Context(), "items", "id")},
		"lines.*.unit":   []any{"required", tenantExists(form.Context(), "units", "id")},
		"lines.*.qty":    "required|min:1",
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
