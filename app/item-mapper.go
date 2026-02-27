package app

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

type IssueLevelType string

const (
	_Info    IssueLevelType = "info"
	_Warning IssueLevelType = "warning"
	_Error   IssueLevelType = "error"
)

var IssueLevel = struct {
	Info    IssueLevelType
	Warning IssueLevelType
	Error   IssueLevelType
}{
	Info:    _Info,
	Warning: _Warning,
	Error:   _Error,
}

type ImportIssue struct {
	Row     int
	Column  string
	Level   IssueLevelType
	Message string
	Value   string
}

var (
	ErrEmptyValue  = errors.New("empty value")
	ErrMaskedValue = errors.New("masked value")
	ErrInvalidNum  = errors.New("invalid number")
)

func mapToStoreItemForm(rowNum int, data map[string]any, unitID int, taxes []*tax, warnings *[]ImportIssue) (*StoreItemForm, error) {
	// Required fields
	name, ok := getString(data, "name")
	if !ok {
		return nil, errors.New("The column `name` is missing or malformed")
	}

	price, ok, err := getFloat64(data, "price")
	if !ok {
		return nil, errors.New("The column `price` is missing or malformed")
	} else if err != nil {
		switch err {
		case ErrMaskedValue:
			*warnings = append(*warnings, ImportIssue{
				Row:     rowNum,
				Column:  "price",
				Value:   data["price"].(string),
				Message: "Masked value ignored",
			})
		case ErrInvalidNum:
			*warnings = append(*warnings, ImportIssue{
				Row:     rowNum,
				Column:  "price",
				Value:   data["price"].(string),
				Message: "Invalid number",
			})
		}
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

	taxRate, ok, err := getFloat64(data, "tax_rate")
	if !ok {
		return nil, errors.New("The column `tax_rate` is missing or malformed")
	} else if err != nil {
		switch err {
		case ErrMaskedValue:
			*warnings = append(*warnings, ImportIssue{
				Row:     rowNum,
				Column:  "tax_rate",
				Value:   data["tax_rate"].(string),
				Message: "Masked value ignored",
			})
		case ErrInvalidNum:
			*warnings = append(*warnings, ImportIssue{
				Row:     rowNum,
				Column:  "tax_rate",
				Value:   data["tax_rate"].(string),
				Message: "Invalid number",
			})
		}
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
		UnitID:      unitID,
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

func getBoolean(data map[string]any, key string) (bool, bool) {
	v, ok := data[key]
	if !ok || v == nil {
		return false, false
	}

	switch val := v.(type) {

	case bool:
		return val, true

	case int:
		return val != 0, true

	case int64:
		return val != 0, true

	case float64:
		return val != 0, true

	case string:
		s := strings.TrimSpace(strings.ToLower(val))
		if s == "" {
			return false, false
		}

		switch s {
		case "t", "true", "yes", "y", "si", "s", "1":
			return true, true
		case "f", "false", "no", "n", "0":
			return false, true
		default:
			return false, false
		}
	}

	return false, false
}

func getFloat64(data map[string]any, key string) (float64, bool, error) {
	s, ok := getString(data, key)
	if !ok {
		return 0, false, ErrEmptyValue
	}

	// Detect masked / redacted values
	if strings.ContainsAny(s, "*Xx") {
		return 0, false, ErrMaskedValue
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Println("error parsing a float|numeric value", data, err)
		return 0, false, ErrInvalidNum
	}

	return f, true, nil
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
