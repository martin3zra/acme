package app

import (
	"errors"
	"testing"
)

func TestGetString(t *testing.T) {
	data := map[string]any{
		"name":  "  Acme  ",
		"empty": "   ",
		"nil":   nil,
		"num":   42,
	}

	if v, ok := getString(data, "name"); !ok || v != "Acme" {
		t.Errorf("name: want (Acme,true), got (%q,%v)", v, ok)
	}
	if _, ok := getString(data, "empty"); ok {
		t.Error("whitespace-only should be treated as missing")
	}
	if _, ok := getString(data, "nil"); ok {
		t.Error("nil should be treated as missing")
	}
	if _, ok := getString(data, "absent"); ok {
		t.Error("absent key should be missing")
	}
	if _, ok := getString(data, "num"); ok {
		t.Error("non-string should be missing")
	}
}

func TestGetBoolean(t *testing.T) {
	cases := []struct {
		in     any
		want   bool
		wantOK bool
	}{
		{true, true, true},
		{false, false, true},
		{1, true, true},
		{0, false, true},
		{"TRUE", true, true},
		{"si", true, true},
		{"n", false, true},
		{"0", false, true},
		{"maybe", false, false},
		{"", false, false},
		{nil, false, false},
	}

	for _, c := range cases {
		got, ok := getBoolean(map[string]any{"k": c.in}, "k")
		if got != c.want || ok != c.wantOK {
			t.Errorf("getBoolean(%v): want (%v,%v), got (%v,%v)", c.in, c.want, c.wantOK, got, ok)
		}
	}

	if _, ok := getBoolean(map[string]any{}, "absent"); ok {
		t.Error("absent key should be (_, false)")
	}
}

func TestGetFloat64(t *testing.T) {
	if f, ok, err := getFloat64(map[string]any{"k": "1234.50"}, "k"); err != nil || !ok || f != 1234.50 {
		t.Errorf("valid number: got (%v,%v,%v)", f, ok, err)
	}

	if _, _, err := getFloat64(map[string]any{"k": "   "}, "k"); !errors.Is(err, ErrEmptyValue) {
		t.Errorf("empty: want ErrEmptyValue, got %v", err)
	}

	if _, _, err := getFloat64(map[string]any{"k": "1*234"}, "k"); !errors.Is(err, ErrMaskedValue) {
		t.Errorf("masked: want ErrMaskedValue, got %v", err)
	}

	if _, _, err := getFloat64(map[string]any{"k": "abc"}, "k"); !errors.Is(err, ErrInvalidNum) {
		t.Errorf("invalid: want ErrInvalidNum, got %v", err)
	}
}

func TestParsePhone(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"8095550123", []string{"8095550123"}},
		{"809-555-0123", []string{"8095550123"}},
		{"8095550123 / 8295559876", []string{"8095550123", "8295559876"}},
		{"809,555;999|111", nil}, // each fragment < 7 digits → dropped
		{"   ", nil},
		{"", nil},
	}

	for _, c := range cases {
		got := parsePhone(c.in)
		if len(got) != len(c.want) {
			t.Errorf("parsePhone(%q): want %v, got %v", c.in, c.want, got)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("parsePhone(%q)[%d]: want %q, got %q", c.in, i, c.want[i], got[i])
			}
		}
	}
}

func TestNormalizePhone(t *testing.T) {
	if got := normalizePhone("(809) 555-0123"); got != "8095550123" {
		t.Errorf("strip non-digits: got %q", got)
	}
	if got := normalizePhone("12345"); got != "" {
		t.Errorf("too short (<7) should be dropped, got %q", got)
	}
	long := normalizePhone("123456789012345678901234567")
	if len(long) != 20 {
		t.Errorf("should cap at 20 digits, got len=%d (%q)", len(long), long)
	}
}
