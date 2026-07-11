package app

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/forge/support"
	"github.com/martin3zra/forge/validator"
)

type ItemIdentifiers struct {
	Reference       *string `json:"reference,omitempty"`
	Code            *string `json:"code,omitempty"`
	SKU             *string `json:"sku,omitempty"`
	Barcode         *string `json:"barcode,omitempty"`
	VendorReference *string `json:"vendor_reference,omitempty"`
}

func (d *ItemIdentifiers) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *ItemIdentifiers) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}

type StoreItemForm struct {
	support.FormRequest
	Name        string          `json:"name"`
	Price       float64         `json:"price"`
	Description string          `json:"description"`
	TaxID       int             `json:"tax_id"`
	UnitID      int             `json:"unit_id"`
	ItemType    string          `json:"item_type"` // e.g. "product", "service"
	Identifiers ItemIdentifiers `json:"identifiers,omitempty"`
}

func (form StoreItemForm) Authorize() bool {
	return Can(form.User(), "create:item")
}

func (form StoreItemForm) Rules() map[string]any {
	return map[string]any{
		"name": []any{
			"required",
			"min:3",
			"max:120",
			validator.Rule{}.Unique("items", "name"),
		},
		"description":                  "sometimes|min:3|max:120",
		"price":                        "required|min:0",
		"tax_id":                       []any{"bail", "required", tenantExists(form.Context(), "taxes", "id")},
		"unit_id":                      []any{"bail", "required", tenantExists(form.Context(), "units", "id")},
		"item_type":                    "bail|required|in:product,service",
		"identifiers":                  "sometimes",
		"identifiers.reference":        "sometimes|nullable|max:100",
		"identifiers.code":             "sometimes|nullable|max:50",
		"identifiers.sku":              "sometimes|nullable|max:50",
		"identifiers.barcode":          "sometimes|nullable|max:32",
		"identifiers.vendor_reference": "sometimes|nullable|max:100",
	}
}

type UpdateItemForm struct {
	support.FormRequest
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Price       float64         `json:"price"`
	Description string          `json:"description"`
	TaxID       int             `json:"tax_id"`
	UnitID      int             `json:"unit_id"`
	ItemType    string          `json:"item_type"` // e.g. "product", "service"
	Identifiers ItemIdentifiers `json:"identifiers,omitempty"`
}

type StoreWarehouseForm struct {
	support.FormRequest
	Name     string `json:"name"`
	Location string `json:"location"`
}

func (form StoreWarehouseForm) Authorize() bool {
	return Can(form.User(), "create:inventory")
}

func (StoreWarehouseForm) Rules() map[string]any {
	return map[string]any{
		"name":     []any{"required", "min:3", "max:150", validator.Rule{}.Unique("warehouses", "name")},
		"location": "sometimes|nullable|max:2000",
	}
}

type UpdateWarehouseForm struct {
	support.FormRequest
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

func (form UpdateWarehouseForm) Authorize() bool {
	return Can(form.User(), "update:inventory")
}

func (form UpdateWarehouseForm) Rules() map[string]any {
	return map[string]any{
		"name":     []any{"required", "min:3", "max:150", validator.Rule{}.Unique("warehouses", "name").Ignore(form.ID, "id")},
		"location": "sometimes|nullable|max:2000",
	}
}

func (form UpdateItemForm) Authorize() bool {
	return Can(form.User(), "update:item")
}

func (form UpdateItemForm) Rules() map[string]any {
	return map[string]any{
		"name":                         []any{"required", "min:3", "max:120", validator.Rule{}.Unique("items", "name").Ignore(form.ID, "id")},
		"description":                  "sometimes|min:3|max:120",
		"price":                        "required|min:0",
		"tax_id":                       []any{"required", tenantExists(form.Context(), "taxes", "id")},
		"unit_id":                      []any{"required", tenantExists(form.Context(), "units", "id")},
		"item_type":                    "bail|required|in:product,service",
		"identifiers":                  "sometimes",
		"identifiers.reference":        "sometimes|nullable|max:100",
		"identifiers.code":             "sometimes|nullable|max:50",
		"identifiers.sku":              "sometimes|nullable|max:50",
		"identifiers.barcode":          "sometimes|nullable|max:32",
		"identifiers.vendor_reference": "sometimes|nullable|max:100",
	}
}

type ItemType string

const (
	ItemTypeAll     ItemType = "all"
	ItemTypeProduct ItemType = "product"
	ItemTypeService ItemType = "service"
)

// Validate ensures the value is one of the allowed constants
func (t ItemType) Validate() error {
	switch t {
	case ItemTypeAll, ItemTypeProduct, ItemTypeService:
		return nil
	default:
		return fmt.Errorf("invalid item type: %s", t)
	}
}

// attribute is a product characteristic (Color, Size, ...) with a set of values.
type attribute struct {
	ID          int               `json:"id"`
	UUID        string            `json:"uuid"`
	Name        string            `json:"name"`
	Type        string            `json:"type"` // "select", "text", "numeric"
	DisplayName string            `json:"display_name"`
	Description *string           `json:"description,omitempty"`
	Values      []*attributeValue `json:"values,omitempty"`
	foundation.Timestamps
}

// attributeValue is a specific value for an attribute (Red, Blue, S, M, L, ...).
type attributeValue struct {
	ID          int    `json:"id"`
	UUID        string `json:"uuid"`
	AttributeID int    `json:"attribute_id"`
	Value       string `json:"value"`
	DisplayName string `json:"display_name"`
	SortOrder   int    `json:"sort_order"`
	foundation.Timestamps
}

// StoreAttributeForm handles attribute creation and update.
type StoreAttributeForm struct {
	support.FormRequest
	Name        string `json:"name"`
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
	Description string `json:"description,omitempty"`
}

func (StoreAttributeForm) Rules() map[string]any {
	return map[string]any{
		"name":         []any{"required", "min:2", "max:120"},
		"type":         "required|in:select,text,numeric",
		"display_name": []any{"required", "min:2", "max:255"},
		"description":  "sometimes|max:1000",
	}
}

func (form StoreAttributeForm) Authorize() bool {
	return Can(form.User(), "create:attribute")
}

// StoreAttributeValueForm handles attribute-value creation and update.
type StoreAttributeValueForm struct {
	support.FormRequest
	AttributeID int    `json:"-"`
	Value       string `json:"value"`
	DisplayName string `json:"display_name"`
	SortOrder   int    `json:"sort_order,omitempty"`
}

func (StoreAttributeValueForm) Rules() map[string]any {
	return map[string]any{
		"value":        []any{"required", "min:1", "max:120"},
		"display_name": []any{"required", "min:1", "max:255"},
		"sort_order":   "sometimes|integer|min:0",
	}
}

func (form StoreAttributeValueForm) Authorize() bool {
	return Can(form.User(), "create:attribute")
}

// itemVariant is a concrete sellable variant of an item (a point in the
// attribute-value matrix, or the lone default variant of a simple item).
type itemVariant struct {
	ID                   int         `json:"id"`
	UUID                 string      `json:"uuid"`
	ItemID               int         `json:"item_id"`
	SKU                  string      `json:"sku"`
	Name                 string      `json:"name"`
	Barcode              *string     `json:"barcode,omitempty"`
	Reference            *string     `json:"reference,omitempty"`
	VendorReference      *string     `json:"vendor_reference,omitempty"`
	CombinationSignature string      `json:"combination_signature"`
	IsDefault            bool        `json:"is_default"`
	Price                *float64    `json:"price,omitempty"`
	CostPrice            *float64    `json:"cost_price,omitempty"`
	TrackInventory       bool        `json:"track_inventory"`
	StockByWarehouse     map[int]int `json:"stock_by_warehouse,omitempty"`
	Active               bool        `json:"active"`
	foundation.Timestamps
}

// VariantCombo is one requested variant: a map of attribute_id -> attribute_value_id
// plus its per-variant pricing/identifiers.
type VariantCombo struct {
	VariantID         int         `json:"variant_id,omitempty"`
	AttributeValueIDs map[int]int `json:"attribute_value_ids"`
	Price             *float64    `json:"price,omitempty"`
	CostPrice         *float64    `json:"cost_price,omitempty"`
	TrackInventory    *bool       `json:"track_inventory,omitempty"`
	StockByWarehouse  map[int]int `json:"stock_by_warehouse,omitempty"`
	SKU               string      `json:"sku,omitempty"`
	Barcode           string      `json:"barcode,omitempty"`
	Reference         string      `json:"reference,omitempty"`
	VendorReference   string      `json:"vendor_reference,omitempty"`
	Active            *bool       `json:"active,omitempty"`
}

// StoreItemWithAttributesForm handles item creation with attributes and variants.
// It carries the same base item fields as StoreItemForm so the item row can be
// created, plus the attribute/variant matrix.
type StoreItemWithAttributesForm struct {
	support.FormRequest
	Name          string          `json:"name"`
	Price         float64         `json:"price"`
	Description   string          `json:"description,omitempty"`
	TaxID         int             `json:"tax_id"`
	UnitID        int             `json:"unit_id"`
	ItemType      string          `json:"item_type"`
	Identifiers   ItemIdentifiers `json:"identifiers,omitempty"`
	AttributeIDs  []int           `json:"attribute_ids,omitempty"`
	VariantCombos []VariantCombo  `json:"variant_combos,omitempty"`
}

func (form StoreItemWithAttributesForm) Rules() map[string]any {
	return map[string]any{
		"name":             []any{"required", "min:3", "max:120"},
		"price":            "required|numeric|min:0",
		"description":      "sometimes|max:1000",
		"tax_id":           []any{"required", tenantExists(form.Context(), "taxes", "id")},
		"unit_id":          []any{"required", tenantExists(form.Context(), "units", "id")},
		"item_type":        "required",
		"attribute_ids.*":  []any{"sometimes", tenantExists(form.Context(), "attributes", "id")},
		"variant_combos.*": "sometimes|array",
	}
}

func (form StoreItemWithAttributesForm) Authorize() bool {
	return Can(form.User(), "create:item")
}
