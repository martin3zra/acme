package app_test

import (
	"testing"
	"time"

	"github.com/martin3zra/acme/app"
)

func TestNextOccurrence(t *testing.T) {
	loc, _ := time.LoadLocation("America/Santo_Domingo")

	tests := []struct {
		name     string
		r        app.Recurrence
		anchor   time.Time
		expected time.Time
	}{
		{
			name:     "daily interval=1",
			r:        app.Recurrence{Frequency: "daily", Interval: 1},
			anchor:   time.Date(2026, 1, 1, 0, 0, 0, 0, loc),
			expected: time.Date(2026, 1, 2, 0, 0, 0, 0, loc),
		},
		{
			name:     "weekly monday after anchor",
			r:        app.Recurrence{Frequency: "weekly", Interval: 1, Weekdays: []string{"monday"}},
			anchor:   time.Date(2026, 1, 1, 0, 0, 0, 0, loc), // Thursday
			expected: time.Date(2026, 1, 5, 0, 0, 0, 0, loc), // Monday
		},
		{name: "weekly interval=2 monday skips one full week",
			r:        app.Recurrence{Frequency: "weekly", Interval: 2, Weekdays: []string{"monday"}},
			anchor:   time.Date(2026, 1, 1, 0, 0, 0, 0, loc),  // Thursday
			expected: time.Date(2026, 1, 12, 0, 0, 0, 0, loc), // Monday after skipping 1 week
		},
		{
			name: "multiple weekdays chooses earliest",
			r: app.Recurrence{
				Frequency: "weekly",
				Interval:  2,
				Weekdays:  []string{"friday"},
			},
			anchor:   time.Date(2026, 1, 1, 0, 0, 0, 0, loc),
			expected: time.Date(2026, 1, 16, 0, 0, 0, 0, loc), // Friday only
		},
		{
			name: "lastGeneratedAt overrides startDate",
			r: app.Recurrence{
				Frequency: "weekly",
				Interval:  1,
				Weekdays:  []string{"monday"},
				StartDate: func() *time.Time {
					t := time.Date(2026, 1, 1, 0, 0, 0, 0, loc)
					return &t
				}(),
				LastGeneratedAt: func() *time.Time {
					t := time.Date(2026, 1, 5, 0, 0, 0, 0, loc) // Monday
					return &t
				}(),
			},
			expected: time.Date(2026, 1, 12, 0, 0, 0, 0, loc),
		},
		{
			name: "anchor already on weekday still advances",
			r: app.Recurrence{
				Frequency: "weekly",
				Interval:  1,
				Weekdays:  []string{"monday"},

				StartDate: func() *time.Time {
					t := time.Date(2026, 1, 5, 0, 0, 0, 0, loc)
					return &t
				}(), // Monday
			},
			expected: time.Date(2026, 1, 12, 0, 0, 0, 0, loc),
		},
		{
			name: "weekly interval=2 monday+friday",
			r: app.Recurrence{
				Frequency: "weekly",
				Interval:  2,
				Weekdays:  []string{"monday", "friday"},
			},
			anchor:   time.Date(2026, 1, 1, 0, 0, 0, 0, loc),  // Thursday
			expected: time.Date(2026, 1, 12, 0, 0, 0, 0, loc), // Monday after skipping 1 week
		},
		{
			name:     "monthly clamp",
			r:        app.Recurrence{Frequency: "monthly", Interval: 1, DayOfMonth: 31},
			anchor:   time.Date(2026, 1, 31, 0, 0, 0, 0, loc),
			expected: time.Date(2026, 2, 28, 0, 0, 0, 0, loc), // Feb clamp
		},
	}

	s := &app.Server{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.anchor.IsZero() {
				tt.r.StartDate = &tt.anchor
			}

			got := s.NextOccurrence(&tt.r, loc)

			if !got.Equal(tt.expected) {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, got)
			}
		})
	}
}
