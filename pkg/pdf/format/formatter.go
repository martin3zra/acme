package format

import (
	"fmt"
	"strconv"
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
