package app

import "embed"

// localesFS holds the application's locale files. The i18n framework package is
// locale-agnostic; the app owns its locales and supplies them here.
//
//go:embed locales/*.json
var localesFS embed.FS

const (
	defaultLang  = "es"
	fallbackLang = "en"
)
