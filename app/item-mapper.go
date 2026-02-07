package app

import (
	"strconv"
	"strings"
)

func mapToStoreItemForm(data map[string]any) *StoreItemForm {
	// Required fields
	name, ok := getString(data, "name")
	if !ok {
		return nil
	}

	price, ok := getFloat64(data, "price")
	if !ok {
		return nil
	}

	itemType, ok := getString(data, "item_type")
	if !ok {
		return nil
	}

	if strings.ToUpper(itemType) == "P" {
		itemType = "product"
	} else {
		itemType = "service"
	}

	return &StoreItemForm{
		Name:        name,
		Price:       price,
		Description: derefOrEmpty(getStringPtr(data, "description")),
		TaxID:       1, // 18%
		UnitID:      1,
		ItemType:    itemType,
		Identifiers: ItemIdentifiers{
			Reference:       getStringPtr(data, "reference"),
			Code:            getStringPtr(data, "code"),
			SKU:             getStringPtr(data, "sku"),
			Barcode:         getStringPtr(data, "barcode"),
			VendorReference: getStringPtr(data, "vendor_reference"),
		},
	}
}

func getString(data map[string]any, key string) (string, bool) {
	v, ok := data[key]
	if !ok || v == nil {
		return "", false
	}

	s, ok := v.(string)
	if !ok {
		return "", false
	}

	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}

	return s, true
}

func getStringPtr(data map[string]any, key string) *string {
	s, ok := getString(data, key)
	if !ok {
		return nil
	}
	return &s
}

func getFloat64(data map[string]any, key string) (float64, bool) {
	s, ok := getString(data, key)
	if !ok {
		return 0, false
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}

	return f, true
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
