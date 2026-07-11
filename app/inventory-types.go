package app

import (
	"time"

	"github.com/martin3zra/forge/support"
)

// InventoryMovementKind maps to the inv_transaction_kind enum on inventory_movements.
type InventoryMovementKind string

const (
	_INV_MOVEMENT_SALE             InventoryMovementKind = "sale"
	_INV_MOVEMENT_SALE_RETURN      InventoryMovementKind = "sale_return"
	_INV_MOVEMENT_PURCHASE_ORDER   InventoryMovementKind = "purchase_order"
	_INV_MOVEMENT_PURCHASE_RECEIPT InventoryMovementKind = "purchase_receipt"
	_INV_MOVEMENT_PURCHASE_RETURN  InventoryMovementKind = "purchase_return"
	_INV_MOVEMENT_VENDOR_BILL      InventoryMovementKind = "vendor_bill"
	_INV_MOVEMENT_ADJUSTMENT       InventoryMovementKind = "adjustment"
	_INV_MOVEMENT_TRANSFER         InventoryMovementKind = "transfer"
)

var InventoryMovementKinds = struct {
	Sale            InventoryMovementKind
	SaleReturn      InventoryMovementKind
	PurchaseOrder   InventoryMovementKind
	PurchaseReceipt InventoryMovementKind
	PurchaseReturn  InventoryMovementKind
	VendorBill      InventoryMovementKind
	Adjustment      InventoryMovementKind
	Transfer        InventoryMovementKind
}{
	Sale:            _INV_MOVEMENT_SALE,
	SaleReturn:      _INV_MOVEMENT_SALE_RETURN,
	PurchaseOrder:   _INV_MOVEMENT_PURCHASE_ORDER,
	PurchaseReceipt: _INV_MOVEMENT_PURCHASE_RECEIPT,
	PurchaseReturn:  _INV_MOVEMENT_PURCHASE_RETURN,
	VendorBill:      _INV_MOVEMENT_VENDOR_BILL,
	Adjustment:      _INV_MOVEMENT_ADJUSTMENT,
	Transfer:        _INV_MOVEMENT_TRANSFER,
}

// TransferStatus maps to the transfer_status enum on inventory_transfers.
// Lifecycle: requested -> in_transit -> received, with cancelled reachable
// only from requested (once goods are in transit they must be received).
type TransferStatus string

const (
	_TRANSFER_REQUESTED  TransferStatus = "requested"
	_TRANSFER_IN_TRANSIT TransferStatus = "in_transit"
	_TRANSFER_RECEIVED   TransferStatus = "received"
	_TRANSFER_CANCELLED  TransferStatus = "cancelled"
)

var TransferStatuses = struct {
	Requested TransferStatus
	InTransit TransferStatus
	Received  TransferStatus
	Cancelled TransferStatus
}{
	Requested: _TRANSFER_REQUESTED,
	InTransit: _TRANSFER_IN_TRANSIT,
	Received:  _TRANSFER_RECEIVED,
	Cancelled: _TRANSFER_CANCELLED,
}

// transferTransitions lists the statuses each status may move to.
var transferTransitions = map[TransferStatus][]TransferStatus{
	_TRANSFER_REQUESTED:  {_TRANSFER_IN_TRANSIT, _TRANSFER_CANCELLED},
	_TRANSFER_IN_TRANSIT: {_TRANSFER_RECEIVED},
	_TRANSFER_RECEIVED:   {},
	_TRANSFER_CANCELLED:  {},
}

// CanTransitionTo reports whether the transfer may move from its current
// status to the target status.
func (s TransferStatus) CanTransitionTo(target TransferStatus) bool {
	for _, allowed := range transferTransitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}

type StoreAdjustmentForm struct {
	support.FormRequest
	VariantID   int     `json:"variant_id"`
	WarehouseID int     `json:"warehouse_id"`
	Qty         float64 `json:"qty"`
	Reason      string  `json:"reason"`
	Notes       string  `json:"notes"`
}

func (form StoreAdjustmentForm) Authorize() bool {
	return Can(form.User(), "create:adjustment")
}

func (form StoreAdjustmentForm) Rules() map[string]any {
	return map[string]any{
		"variant_id":   []any{"bail", "required", tenantExists(form.Context(), "items_variants", "id")},
		"warehouse_id": []any{"bail", "required", tenantExists(form.Context(), "warehouses", "id")},
		"qty":          "bail|required",
		"reason":       "bail|required|min:3|max:255",
		"notes":        "sometimes|max:1000",
	}
}

// TransferLineInput is a single product line on a transfer document. ID is the
// item id plus the specific variant to move. VariantID may be 0 for a plain
// item (resolved to its default variant on store, like purchases/invoices); a
// has_variants item must name one explicitly.
type TransferLineInput struct {
	ID          int     `json:"id"`
	VariantID   int     `json:"variant_id"`
	Qty         float64 `json:"qty"`
	Unit        int     `json:"unit"`
	Cost        float64 `json:"cost"`
	Description string  `json:"description"`
}

type StoreTransferForm struct {
	support.FormRequest
	FromWarehouseID int                 `json:"from_warehouse_id"`
	ToWarehouseID   int                 `json:"to_warehouse_id"`
	Date            time.Time           `json:"date"`
	Notes           string              `json:"notes"`
	Lines           []TransferLineInput `json:"lines"`
}

func (form StoreTransferForm) Authorize() bool {
	return Can(form.User(), "create:transfer")
}

func (form StoreTransferForm) Rules() map[string]any {
	return map[string]any{
		"from_warehouse_id": []any{"bail", "required", tenantExists(form.Context(), "warehouses", "id")},
		"to_warehouse_id":   []any{"bail", "required", tenantExists(form.Context(), "warehouses", "id")},
		"notes":             "sometimes|max:1000",
		"lines":             "bail|required",
	}
}
