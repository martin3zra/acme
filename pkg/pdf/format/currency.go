package format

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// CurrencyFormatter formats currency amounts with configurable separators and symbols.
type CurrencyFormatter struct {
	Symbol            string
	DecimalSeparator  string
	ThousandSeparator string
	Precision         int
}

// NewCurrencyFormatter creates a new currency formatter with the specified parameters.
func NewCurrencyFormatter(symbol, decimalSep, thousandSep string, precision int) *CurrencyFormatter {
	return &CurrencyFormatter{
		Symbol:            symbol,
		DecimalSeparator:  decimalSep,
		ThousandSeparator: thousandSep,
		Precision:         precision,
	}
}

// FormatAmount formats a float amount as currency string
func (cf *CurrencyFormatter) FormatAmount(amount float64) string {
	multiplier := math.Pow(10, float64(cf.Precision))
	rounded := math.Round(amount*multiplier) / multiplier
	intPart := int64(rounded)
	fracPart := int(math.Round((rounded - float64(intPart)) * multiplier))

	intStr := strconv.FormatInt(intPart, 10)
	if cf.ThousandSeparator != "" {
		intStr = addThousandsSeparator(intStr, cf.ThousandSeparator)
	}

	var result string
	if cf.Precision > 0 {
		fracStr := fmt.Sprintf("%0*d", cf.Precision, fracPart)
		result = intStr + cf.DecimalSeparator + fracStr
	} else {
		result = intStr
	}

	if cf.Symbol != "" {
		result = cf.Symbol + " " + result
	}

	return result
}

// FormatWithCode formats with currency code
func (cf *CurrencyFormatter) FormatWithCode(amount float64, code string) string {
	return cf.FormatAmount(amount) + " " + code
}

// addThousandsSeparator adds thousands separators to an integer string.
func addThousandsSeparator(s string, sep string) string {
	if len(s) <= 3 {
		return s
	}

	isNegative := false
	if len(s) > 0 && s[0] == '-' {
		isNegative = true
		s = s[1:]
	}

	var result strings.Builder
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteString(sep)
		}
		result.WriteRune(ch)
	}

	resStr := result.String()
	if isNegative {
		resStr = "-" + resStr
	}

	return resStr
}

// GetDefaultFormatters returns default currency formatters
func GetDefaultFormatters() map[string]*CurrencyFormatter {
	return map[string]*CurrencyFormatter{
		"USD": NewCurrencyFormatter("$", ".", ",", 2),
		"EUR": NewCurrencyFormatter("€", ",", ".", 2),
		"GBP": NewCurrencyFormatter("£", ".", ",", 2),
		"JPY": NewCurrencyFormatter("¥", ".", ",", 0),
		"DOP": NewCurrencyFormatter("RD$", ".", ",", 2),
	}
}
