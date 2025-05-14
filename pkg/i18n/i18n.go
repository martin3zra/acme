package i18n

import (
	"encoding/json"
	"fmt"
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
