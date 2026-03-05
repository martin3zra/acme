package pdf

import (
	"regexp"
	"strings"
)

// VariableBinder handles {{variable}} replacement in template content
type VariableBinder struct {
	data map[string]any
}

// NewVariableBinder creates a new variable binder
func NewVariableBinder(data map[string]any) *VariableBinder {
	return &VariableBinder{data: data}
}

// Bind replaces {{variable}} placeholders with actual values
// Supports dot notation: {{business.name}}, {{items[0].name}}
func (vb *VariableBinder) Bind(content string) string {
	// Regex to match {{...}} placeholders
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		// Extract variable name (remove {{ }})
		varName := strings.TrimPrefix(strings.TrimSuffix(match, "}}"), "{{")
		varName = strings.TrimSpace(varName)

		// Get the value
		value := vb.Get(varName)
		if value == nil {
			return "" // Return empty string if variable not found
		}

		return toString(value)
	})
}

// Get retrieves a value from the data using dot notation and array access
// Examples: "business.name", "items[0].name", "customer"
func (vb *VariableBinder) Get(path string) any {
	if path == "" {
		return nil
	}

	// Parse the path: split by dots and handle array notation
	parts := parseExpression(path)
	if len(parts) == 0 {
		return nil
	}

	// Start from root data
	current := any(vb.data)

	for _, part := range parts {
		current = vb.navigate(current, part)
		if current == nil {
			return nil
		}
	}

	return current
}

// navigate moves to the next level in the data structure
func (vb *VariableBinder) navigate(current any, part string) any {
	// Check for array access: "items[0]"
	if strings.ContainsRune(part, '[') && strings.ContainsRune(part, ']') {
		return vb.navigateArray(current, part)
	}

	// Regular key access
	if m, ok := current.(map[string]any); ok {
		return m[part]
	}

	return nil
}

// navigateArray handles array access notation: "items[0]"
func (vb *VariableBinder) navigateArray(current any, part string) any {
	// Extract key and index: "items[0]" -> key="items", index=0
	bracketIdx := strings.Index(part, "[")
	if bracketIdx == -1 {
		return nil
	}

	key := part[:bracketIdx]
	indexStr := part[bracketIdx+1 : len(part)-1]

	// Get the array
	var arr any
	if m, ok := current.(map[string]any); ok {
		arr = m[key]
	} else {
		return nil
	}

	// Parse index
	var index int
	if _, err := scanInt(indexStr, &index); err != nil {
		return nil
	}

	// Get array element
	switch v := arr.(type) {
	case []any:
		if index >= 0 && index < len(v) {
			return v[index]
		}
	case []map[string]any:
		if index >= 0 && index < len(v) {
			return v[index]
		}
	}

	return nil
}

// parseExpression parses a dot-notation path into parts
// Examples:
//   "business.name" -> ["business", "name"]
//   "items[0].name" -> ["items[0]", "name"]
func parseExpression(expr string) []string {
	var parts []string
	var current strings.Builder

	for i := 0; i < len(expr); i++ {
		ch := expr[i]

		switch ch {
		case '.':
			if i > 0 && expr[i-1] != ']' {
				// Add current part
				if current.Len() > 0 {
					parts = append(parts, current.String())
					current.Reset()
				}
			} else {
				current.WriteByte(ch)
			}
		case '[':
			// Include array brackets in the part
			current.WriteByte(ch)
		case ']':
			current.WriteByte(ch)
			// After closing bracket, next dot separates parts
		default:
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// scanInt parses an integer from a string
func scanInt(s string, v *int) (string, error) {
	*v = 0
	sign := 1

	i := 0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		if s[i] == '-' {
			sign = -1
		}
		i++
	}

	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		*v = *v*10 + int(s[i]-'0')
		i++
	}

	*v *= sign
	return s[i:], nil
}

// toString converts a value to string
func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return formatInt(val)
	case int32:
		return formatInt(int(val))
	case int64:
		return formatInt(int(val))
	case float32:
		return formatFloat(float64(val))
	case float64:
		return formatFloat(val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case nil:
		return ""
	default:
		return ""
	}
}

// formatInt formats an integer
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var result strings.Builder
	for n > 0 {
		result.WriteByte(byte('0' + n%10))
		n /= 10
	}

	s := result.String()
	// Reverse the string
	chars := []rune(s)
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}

	if negative {
		return "-" + string(chars)
	}
	return string(chars)
}

// formatFloat formats a float
func formatFloat(n float64) string {
	// Simple float formatting
	if n == float64(int64(n)) {
		return formatInt(int(n))
	}
	return ""
}
