package app

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"codeberg.org/go-pdf/fpdf"
	"github.com/martin3zra/acme/pkg/i18n"
)

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

func getNetDays(term string) int {
	term = strings.TrimSpace(strings.ToLower(term))

	re := regexp.MustCompile(`^net\s*(\d+)$`)
	matches := re.FindStringSubmatch(term)

	if len(matches) == 2 {
		if days, err := strconv.Atoi(matches[1]); err == nil {
			return days
		}
	}

	return 0 // Not a recognized "Net" term
}

type WatermarkOpt struct {
	FontFamily string
	FontStyle  string  // style ("", "B", "I", "BI")
	FontSize   float64 // fontSize in points
	AngleDeg   float64 // angleDeg rotation
	Opacity    float64 // alpha opacity (0.0–1.0)
	Spacing    float64 // letterSpacing (negative tightens)
	Color      struct {
		Red   int
		Green int
		Blue  int
	} // RGB
}

// drawTextWatermarkKerned draws centered, rotated text with custom letter spacing.
func drawTextWatermarkKerned(
	pdf *fpdf.Fpdf,
	text string,
	opts WatermarkOpt,
	// txt, family, style string,
	// fontSize, angleDeg, alpha, letterSpacing float64,
	// r, g, b int,
) {
	// page center
	pw, ph := pdf.GetPageSize()
	cx, cy := pw/2, ph/2

	// font setup
	pdf.SetFont(opts.FontFamily, opts.FontStyle, opts.FontSize)
	pdf.SetTextColor(opts.Color.Red, opts.Color.Green, opts.Color.Blue)

	// measure total width (raw glyph widths + spacing)
	runes := []rune(text)
	total := 0.0
	for i, r := range runes {
		total += pdf.GetStringWidth(string(r))
		if i < len(runes)-1 {
			total += opts.Spacing
		}
	}

	// approximate text height
	_, h := pdf.GetFontSize()

	// start graphic state
	pdf.TransformBegin()
	// rotate around center
	pdf.TransformRotate(opts.AngleDeg, cx, cy)
	// set opacity
	pdf.SetAlpha(opts.Opacity, "Normal")

	// draw each glyph, centered
	x := cx - total/2
	y := cy + h*0.35
	for _, r := range runes {
		s := string(r)
		pdf.Text(x, y, s)
		x += pdf.GetStringWidth(s) + opts.Spacing
	}

	// restore
	pdf.SetAlpha(1.0, "Normal")
	pdf.TransformEnd()
}
