package app

import (
	"context"
	"net/http"
	"regexp"

	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/romsar/gonertia/v2"
)

func flash(w http.ResponseWriter, name string, value any) {
	foundation.SetFlash(w, name, value)
}

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

func loadTranslations(namespaces ...string) map[string]string {
	translations, err := i18n.LoadTranslations("es", "en", namespaces...)
	if err != nil {
		panic(err)
	}
	return translations
}

// mergeTranslations merges shared "translations" with page-specific ones.
func mergeTranslations(ctx context.Context, pageTranslations map[string]string) map[string]string {
	merged := map[string]string{}

	// ✅ Get existing props from context
	sharedProps := gonertia.PropsFromContext(ctx)

	// Get shared translations if available
	if shared, ok := sharedProps["translations"].(map[string]string); ok {
		for k, v := range shared {
			merged[k] = v
		}
	}

	// Merge page-specific
	for k, v := range pageTranslations {
		merged[k] = v
	}

	return merged
}
