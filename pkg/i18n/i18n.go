package i18n

import (
	"encoding/json"
	"fmt"
	"strings"
)

func LoadTranslations(lang string, fallbackLang string, namespaces ...string) (map[string]string, error) {
	primary, err := loadLangFromFS(lang)
	if err != nil {
		return nil, fmt.Errorf("load primary language: %w", err)
	}

	fallback := map[string]interface{}{}
	if fallbackLang != "" && fallbackLang != lang {
		fallback, _ = loadLangFromFS(fallbackLang)
	}

	result := make(map[string]string)

	for _, ns := range namespaces {
		// Try from primary
		val, ok := getNestedKey(primary, ns)
		if !ok {
			// fallback
			val, ok = getNestedKey(fallback, ns)
		}

		if ok {
			if sub, ok := val.(map[string]interface{}); ok {
				flat := flatten(sub, ns)
				for k, v := range flat {
					result[k] = v
				}
			}
		}
	}

	return result, nil
}

func loadLangFromFS(lang string) (map[string]any, error) {
	filePath := fmt.Sprintf("locales/%s.json", lang)
	data, err := LocalFS.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", filePath, err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}
	return out, nil
}

// Replacements is a map for placeholder replacements
type Replacements map[string]string

// Translator holds the translations map
type Translator struct {
	Translations map[string]string
}

// NewTranslator creates a new Translator with the given translations
func NewTranslator(translations map[string]string) *Translator {
	if translations == nil {
		translations = map[string]string{}
	}
	return &Translator{Translations: translations}
}

// Trans returns the translated string with replacements
func (t *Translator) Trans(key string, replacements ...Replacements) string {
	translation, ok := t.Translations[key]
	if !ok {
		translation = key
	}

	if len(replacements) == 0 {
		return translation
	}

	for k, v := range replacements[0] {

		nounKey := getNounFromKey(v)
		if strings.HasPrefix(v, "@") {
			refKey := strings.TrimPrefix(v, "@")
			if val, exists := t.Translations[refKey]; exists {
				v = val
			}
		}
		translation = strings.ReplaceAll(translation, ":"+k, v)

		gender, ok := t.Translations[fmt.Sprintf("global.nouns.%s.gender", nounKey)]
		if ok {
			actionNoun := "a"
			if gender == "m" {
				actionNoun = "o"
			}
			translation = strings.ReplaceAll(translation, "@action", actionNoun)
		} else {
			translation = strings.ReplaceAll(translation, "@action", "o")
		}

	}

	return translation
}

func getNounFromKey(key string) string {
	// Remove leading "@" if present
	key = strings.TrimPrefix(key, "@")

	// Split by dot
	parts := strings.Split(key, ".")

	if len(parts) > 1 {
		return parts[1] // returns "customer" or "invoice"
	}

	return ""
}
