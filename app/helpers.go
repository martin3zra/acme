package app

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

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

	return -1 // Not a recognized "Net" term
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

type Font struct {
	Name  string
	Style string
	Path  string
}

//go:embed fonts/DejaVuSans.ttf
var dejavuRegular []byte

//go:embed fonts/DejaVu-Oblique.ttf
var dejavuItalic []byte

//go:embed fonts/DejaVu-Bold.ttf
var dejavuBold []byte

func registerFonts(pdf *fpdf.Fpdf) error {
	pdf.AddUTF8FontFromBytes("DejaVu", "", dejavuRegular)
	if pdf.Error() != nil {
		return pdf.Error()
	}

	pdf.AddUTF8FontFromBytes("DejaVu", "I", dejavuItalic)
	if pdf.Error() != nil {
		return pdf.Error()
	}

	pdf.AddUTF8FontFromBytes("DejaVu", "B", dejavuBold)
	if pdf.Error() != nil {
		return pdf.Error()
	}

	return nil
}

func truncate(pdf *fpdf.Fpdf, text string, maxWidth float64) string {
	if pdf.GetStringWidth(text) <= maxWidth {
		return text // no truncation → no ellipsis
	}

	ellipsis := "…"
	ellipsisWidth := pdf.GetStringWidth(ellipsis)

	for pdf.GetStringWidth(text)+ellipsisWidth > maxWidth && len(text) > 0 {
		text = text[:len(text)-1]
	}

	return text + ellipsis
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func dateWithTimeZone(date time.Time) time.Time {
	loc, _ := time.LoadLocation("America/Santo_Domingo")
	return date.In(loc)
}

func nowWithTimeZone() time.Time {
	loc, _ := time.LoadLocation("America/Santo_Domingo")
	return time.Now().In(loc)
}

func Today() PresetRange {
	now := nowWithTimeZone()
	return PresetRange{
		Key:  "today",
		From: formatDate(now),
		To:   formatDate(now),
	}
}

func ThisWeek() PresetRange {
	now := nowWithTimeZone()
	start := startOfWeek(now)
	end := start.AddDate(0, 0, 6)
	return PresetRange{
		Key:  "this_week",
		From: formatDate(start),
		To:   formatDate(end),
	}
}

func ThisMonth() PresetRange {
	now := nowWithTimeZone()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return PresetRange{
		Key:  "this_month",
		From: formatDate(start),
		To:   formatDate(now),
	}
}

func ThisQuarter() PresetRange {
	now := nowWithTimeZone()
	start := startOfQuarter(now)
	return PresetRange{
		Key:  "this_quarter",
		From: formatDate(start),
		To:   formatDate(now),
	}
}

func ThisYear() PresetRange {
	now := nowWithTimeZone()
	start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	return PresetRange{
		Key:  "this_year",
		From: formatDate(start),
		To:   formatDate(now),
	}
}

func PreviousWeek() PresetRange {
	now := nowWithTimeZone()

	weekday := int(now.Weekday())
	offset := (weekday + 6) % 7 // Monday start

	startOfThisWeek := now.AddDate(0, 0, -offset)
	startOfPrevWeek := startOfThisWeek.AddDate(0, 0, -7)
	endOfPrevWeek := startOfPrevWeek.AddDate(0, 0, 6)

	return PresetRange{
		Key:  "previous_week",
		From: formatDate(startOfPrevWeek),
		To:   formatDate(endOfPrevWeek),
	}
}

func PreviousMonth() PresetRange {
	now := nowWithTimeZone()
	firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastOfPrevMonth := firstOfThisMonth.AddDate(0, 0, -1)
	startOfPrevMonth := time.Date(lastOfPrevMonth.Year(), lastOfPrevMonth.Month(), 1, 0, 0, 0, 0, now.Location())
	return PresetRange{
		Key:  "previous_month",
		From: formatDate(startOfPrevMonth),
		To:   formatDate(lastOfPrevMonth),
	}
}

func PreviousQuarter() PresetRange {
	now := nowWithTimeZone()
	startOfThisQuarter := startOfQuarter(now)
	startOfPrevQuarter := startOfThisQuarter.AddDate(0, -3, 0)
	endOfPrevQuarter := startOfThisQuarter.AddDate(0, 0, -1)
	return PresetRange{
		Key:  "previous_quarter",
		From: formatDate(startOfPrevQuarter),
		To:   formatDate(endOfPrevQuarter),
	}
}

func PreviousYear() PresetRange {
	now := nowWithTimeZone()
	startOfPrevYear := time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, now.Location())
	endOfPrevYear := time.Date(now.Year()-1, 12, 31, 0, 0, 0, 0, now.Location())
	return PresetRange{
		Key:  "previous_year",
		From: formatDate(startOfPrevYear),
		To:   formatDate(endOfPrevYear),
	}
}

func DateRangePresets() []PresetRange {
	return []PresetRange{
		Today(),
		ThisWeek(),
		ThisMonth(),
		ThisQuarter(),
		ThisYear(),
		PreviousWeek(),
		PreviousMonth(),
		PreviousQuarter(),
		PreviousYear(),
		{Key: "custom"},
	}
}

func startOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	offset := (weekday + 6) % 7 // Monday start
	return t.AddDate(0, 0, -offset)
}

func startOfQuarter(t time.Time) time.Time {
	month := (int(t.Month())-1)/3*3 + 1
	return time.Date(t.Year(), time.Month(month), 1, 0, 0, 0, 0, t.Location())
}

// redirectAfterCreate builds the redirect URL based on kind, entity ID, and user preference.
func redirectAfterCreate(kind string, id string, pref RedirectPreferencesValue) string {
	switch pref {
	case RedirectPreference.List:
		return fmt.Sprintf("/%ss", kind) // e.g. /invoices
	case RedirectPreference.Detail:
		return fmt.Sprintf("/%ss?id=%s", kind, id) // e.g. /invoices?id=123
	case RedirectPreference.Stay:
		return fmt.Sprintf("/%ss/create", kind) // e.g. /invoices/create
	default:
		// fallback to list
		return fmt.Sprintf("/%ss", kind)
	}
}

func debugSQL(query string, args []any) string {
	for i, arg := range args {
		placeholder := fmt.Sprintf("$%d", i+1)

		var value string
		switch v := arg.(type) {
		case string:
			value = fmt.Sprintf("'%s'", v)
		case time.Time:
			value = fmt.Sprintf("'%s'", v.Format("2006-01-02"))
		default:
			value = fmt.Sprintf("%v", v)
		}

		query = strings.ReplaceAll(query, placeholder, value)
	}
	return query
}
