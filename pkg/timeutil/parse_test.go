package timeutil

import (
	"testing"
	"time"
)

func TestParseNaturalDate_Today(t *testing.T) {
	result, err := ParseNaturalDate("today")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Now().Truncate(24 * time.Hour)
	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestParseNaturalDate_Tomorrow(t *testing.T) {
	result, err := ParseNaturalDate("tomorrow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestParseNaturalDate_Weekday(t *testing.T) {
	result, err := ParseNaturalDate("Friday")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Weekday() != time.Friday {
		t.Errorf("expected Friday, got %s", result.Weekday())
	}
	if !result.After(time.Now().Truncate(24 * time.Hour)) {
		t.Error("expected future date")
	}
}

func TestParseNaturalDate_ISO(t *testing.T) {
	result, err := ParseNaturalDate("2025-03-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestParseNaturalDate_MonthDay(t *testing.T) {
	result, err := ParseNaturalDate("March 15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Month() != time.March || result.Day() != 15 {
		t.Errorf("expected March 15, got %s %d", result.Month(), result.Day())
	}
}

func TestParseNaturalDate_MonthDayYear(t *testing.T) {
	result, err := ParseNaturalDate("Jan 2, 2026")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestParseNaturalDate_CaseInsensitive(t *testing.T) {
	result, err := ParseNaturalDate("TODAY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Now().Truncate(24 * time.Hour)
	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestParseNaturalDate_Invalid(t *testing.T) {
	_, err := ParseNaturalDate("not-a-date")
	if err == nil {
		t.Error("expected error for invalid date string")
	}
}
