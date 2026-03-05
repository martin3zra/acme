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



































































}	return cf.FormatAmount(amount) + " " + codefunc (cf *CurrencyFormatter) FormatWithCode(amount float64, code string) string {// FormatWithCode formats currency with a code (e.g., "1234.56" -> "$1,234.56 USD")}	}		"DOP": NewCurrencyFormatter("RD$", ".", ",", 2), // Dominican Peso		"JPY": NewCurrencyFormatter("¥", ".", ",", 0),		"GBP": NewCurrencyFormatter("£", ".", ",", 2),		"EUR": NewCurrencyFormatter("€", ",", ".", 2),		"USD": NewCurrencyFormatter("$", ".", ",", 2),	return map[string]*CurrencyFormatter{func GetDefaultFormatters() map[string]*CurrencyFormatter {// GetDefaultFormatters returns common currency formatters}	return result.String()	}		result.WriteRune(ch)		}			result.WriteString(sep)		if i > 0 && (len(s)-i)%3 == 0 {	for i, ch := range s {	var result strings.Builder	}		return s	if len(s) <= 3 {func addThousandsSeparator(s string, sep string) string {// addThousandsSeparator adds thousands separator to integer string}	return result	}		result = cf.Symbol + " " + result	if cf.Symbol != "" {	// Add symbol	}		result = intStr	} else {		result = intStr + cf.DecimalSeparator + fracStr	if cf.Precision > 0 {	var result string	// Combine	}		fracStr = fmt.Sprintf("%0*d", cf.Precision, fracPart)	if cf.Precision > 0 {	var fracStr string	// Format fractional part	}		intStr = addThousandsSeparator(intStr, cf.ThousandSeparator)	if cf.ThousandSeparator != "" {	intStr := strconv.FormatInt(intPart, 10)	// Format integer part with thousands separator	fracPart := int(math.Round((rounded - float64(intPart)) * multiplier))	intPart := int64(rounded)	// Split into integer and fractional parts	rounded := math.Round(amount*multiplier) / multiplier	multiplier := math.Pow(10, float64(cf.Precision))	// Round to precisionfunc (cf *CurrencyFormatter) FormatAmount(amount float64) string {// Example: 1234.56 -> "$1,234.56" or "€ 1.234,56"// FormatAmount formats a number as currency}	}		Precision:         precision,		ThousandSeparator: thousandSeparator,		DecimalSeparator:  decimalSeparator,		Symbol:            symbol,	return &CurrencyFormatter{func NewCurrencyFormatter(symbol string, decimalSeparator string, thousandSeparator string, precision int) *CurrencyFormatter {// NewCurrencyFormatter creates a new currency formatter}	Precision         int	ThousandSeparator string	DecimalSeparator  string	Symbol            stringtype CurrencyFormatter struct {// CurrencyFormatter formats numbers as currency)	"strings"	"strconv"	"math"	"fmt"import (