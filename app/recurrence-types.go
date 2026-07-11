package app

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/martin3zra/forge/support"
)

type StoreRecurrenceForm struct {
	support.FormRequest
	Recurrence
}

func (form StoreRecurrenceForm) Rules() map[string]any {
	return map[string]any{
		"recurrence":              "bail|required",
		"recurrence.enabled":      "bail|sometimes",
		"recurrence.name":         "required|max:100",
		"recurrence.start_date":   "nullable|date",
		"recurrence.until":        "nullable|date|after_or_equal:start_date",
		"recurrence.frequency":    "required|in:daily,weekly,monthly,quarterly,yearly",
		"recurrence.interval":     "required|integer|min:1",
		"recurrence.weekdays":     "required_if:frequency,weekly|min:1",
		"recurrence.day_of_month": "required_if:frequency,monthly,quarterly,yearly|min:1|max:31",
		"recurrence.month":        "required_if:frequency,yearly|min:1|max:12",
	}
}

func (form StoreRecurrenceForm) AsRecurrence() *Recurrence {
	return &Recurrence{
		Enabled:    form.Enabled,
		Name:       form.Name,
		Type:       form.Type,
		SendEmail:  form.SendEmail,
		Frequency:  form.Frequency,
		Interval:   form.Interval,
		Timezone:   form.Timezone,
		StartDate:  form.StartDate,
		Until:      form.Until,
		DayOfMonth: form.DayOfMonth,
		Weekdays:   form.Weekdays,
		Month:      form.Month,
	}
}

type FrequencyType string

const (
	_FREQUENCY_WEEKLY   FrequencyType = "weekly"
	_FREQUENCY_MONTHLY  FrequencyType = "monthly"
	_FREQUENCY_QUARTELY FrequencyType = "quarterly"
	_FREQUENCY_YEARLY   FrequencyType = "yearly"
)

var Frequency = struct {
	Weekly    FrequencyType
	Monthly   FrequencyType
	Quarterly FrequencyType
	Yearly    FrequencyType
}{
	Weekly:    _FREQUENCY_WEEKLY,
	Monthly:   _FREQUENCY_MONTHLY,
	Quarterly: _FREQUENCY_QUARTELY,
	Yearly:    _FREQUENCY_YEARLY,
}

type Recurrence struct {
	Enabled   bool       `json:"enabled"`
	Name      string     `json:"name"`
	Type      string     `json:"type"` // schedule, reminder
	SendEmail bool       `json:"send_email"`
	Frequency string     `json:"frequency"` // daily, weekly, monthly, quarterly, yearly
	Interval  int        `json:"interval"`
	Timezone  string     `json:"timezone,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	Until     *time.Time `json:"until"`

	// Optional fields depending on frequency
	DayOfMonth      int        `json:"day_of_month,omitempty"`
	Weekdays        []string   `json:"weekdays,omitempty"`
	Month           int        `json:"month,omitempty"`
	LastGeneratedAt *time.Time `json:"last_generated_at"`
	NextRunAt       *time.Time `json:"next_run_at"`
}

func (d *Recurrence) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *Recurrence) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &d)
}
