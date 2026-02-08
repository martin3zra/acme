package app

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

func mapToStoreItemForm(data map[string]any, taxes []*tax) (*StoreItemForm, error) {
	// Required fields
	name, ok := getString(data, "name")
	if !ok {
		return nil, errors.New("The column `name` is missing or malformed")
	}

	price, ok := getFloat64(data, "price")
	if !ok {
		return nil, errors.New("The column `price` is missing or malformed")
	}

	itemType, ok := getString(data, "item_type")
	if !ok {
		return nil, errors.New("The column `item_type` is missing or malformed")
	}

	if strings.ToUpper(itemType) == "P" {
		itemType = "product"
	} else {
		itemType = "service"
	}

	taxRate, ok := getFloat64(data, "tax_rate")
	if !ok {
		return nil, errors.New("The column `tax_rate` is missing or malformed")
	}
	taxIndex := buildTaxRateIndex(taxes)
	taxID, err := resolveTaxID(taxRate, taxIndex)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &StoreItemForm{
		Name:        name,
		Price:       price,
		Description: derefOrEmpty(getStringPtr(data, "description")),
		TaxID:       taxID,
		UnitID:      1,
		ItemType:    itemType,
		Identifiers: ItemIdentifiers{
			Reference:       getStringPtr(data, "reference"),
			Code:            getStringPtr(data, "code"),
			SKU:             getStringPtr(data, "sku"),
			Barcode:         getStringPtr(data, "barcode"),
			VendorReference: getStringPtr(data, "vendor_reference"),
		},
	}, nil
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

func resolveTaxID(
	rateStr float64,
	taxIndex map[int]int64,
) (int, error) {

	key := normalizeRate(rateStr)

	id, ok := taxIndex[key]
	if !ok {
		return 0, fmt.Errorf("no tax found for rate %f", rateStr)
	}

	return int(id), nil
}

func normalizeRate(r float64) int {
	return int(math.Round(r * 100))
}

func buildTaxRateIndex(taxes []*tax) map[int]int64 {
	index := make(map[int]int64)

	for _, t := range taxes {
		key := normalizeRate(t.Rate)
		index[key] = t.ID
	}

	return index
}
