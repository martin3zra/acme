package format

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Formatter provides formatting utilities for dates, times, and numbers.
type Formatter struct {
	DateFormat     string
	TimeFormat     string
	DateTimeFormat string
}

// NewFormatter creates a new formatter with default settings.
func NewFormatter() *Formatter {
	return &Formatter{
		DateFormat:     "2006-01-02",
		TimeFormat:     "15:04:05",
		DateTimeFormat: "2006-01-02 15:04:05",
	}
}

// FormatDate formats a time.Time as a date string.
func (f *Formatter) FormatDate(t time.Time) string {
	return t.Format(f.DateFormat)
}

// FormatTime formats a time.Time as a time string.
func (f *Formatter) FormatTime(t time.Time) string {
	return t.Format(f.TimeFormat)
}

// FormatDateTime formats a time.Time as a datetime string.
func (f *Formatter) FormatDateTime(t time.Time) string {
	return t.Format(f.DateTimeFormat)
}

// FormatNumber formats an int with separator
func (f *Formatter) FormatNumber(n int64, sep string) string {
	return addThousandsSeparator(strconv.FormatInt(n, 10), sep)
}

// FormatFloat formats a float64 with a specified precision.
func (f *Formatter) FormatFloat(n float64, precision int) string {
	formatStr := "%." + strconv.Itoa(precision) + "f"
	return fmt.Sprintf(formatStr, n)
}

// ParseDate parses a date string
func (f *Formatter) ParseDate(dateStr string) (time.Time, error) {
	return time.Parse(f.DateFormat, dateStr)
}

// ParseDateTime parses a datetime string
func (f *Formatter) ParseDateTime(dateTimeStr string) (time.Time, error) {
	return time.Parse(f.DateTimeFormat, dateTimeStr)
}

// addThousandsSeparator adds separator to number
func addThousandsSeparator(s string, sep string) string {
	if len(s) <= 3 {
		return s
	}

	var result strings.Builder
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteString(sep)
		}
		result.WriteRune(ch)
	}

	return result.String()
}













































}// ParseDateTime parses a datetime string into a time.Time.
func (f *Formatter) ParseDateTime(dateTimeStr string) (time.Time, error) {
	return time.Parse(f.DateTimeFormat, dateTimeStr)
}	return time.Parse(f.DateFormat, dateStr)func (f *Formatter) ParseDate(dateStr string) (time.Time, error) {// ParseDate parses a date string}	return format	format := "%." + strconv.Itoa(precision) + "f"func (f *Formatter) FormatFloat(n float64, precision int) string {// FormatFloat formats a float with precision}	return addThousandsSeparator(strconv.FormatInt(n, 10), sep)func (f *Formatter) FormatNumber(n int64, sep string) string {// FormatNumber formats a number with thousand separators}	return t.Format(f.DateTimeFormat)func (f *Formatter) FormatDateTime(t time.Time) string {// FormatDateTime formats a time as a datetime string}	return t.Format(f.TimeFormat)func (f *Formatter) FormatTime(t time.Time) string {// FormatTime formats a time as a time string}	return t.Format(f.DateFormat)func (f *Formatter) FormatDate(t time.Time) string {// FormatDate formats a time as a date string}	}		DateTimeFormat: "2006-01-02 15:04:05",		TimeFormat:     "15:04:05",		DateFormat:     "2006-01-02",	return &Formatter{func NewFormatter() *Formatter {// NewFormatter creates a new formatter with default formats}	DateTimeFormat string	TimeFormat     string	DateFormat     stringtype Formatter struct {// Formatter provides general formatting utilities)	"time"	"strconv"