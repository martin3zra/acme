package app

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/martin3zra/acme/pkg/i18n"
)

func ensureUUIDIsValid(str string) bool {
	regex := `^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`
	r := regexp.MustCompile(regex)
	return r.MatchString(str)
}

func filter[T any](s []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(s)) // Pre-allocate for efficiency
	for _, v := range s {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

func trans(namespaces ...string) map[string]string {
	translations, err := i18n.LoadTranslations("es", "en", namespaces...)
	if err != nil {
		panic(err)
	}
	return translations
}

func mapTo[T any](m map[string]any) (T, error) {
	var result T
	data, err := json.Marshal(m)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func round(value float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(value*ratio) / ratio
}

// Helper to convert various types to int
func toInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("unsupported type: %T", value)
	}
}
