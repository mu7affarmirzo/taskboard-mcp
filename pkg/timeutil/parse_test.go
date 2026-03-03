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

func TestParseNaturalDate_ShortWeekday_Fri(t *testing.T) {
	result, err := ParseNaturalDate("fri")
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

func TestParseNaturalDate_ShortWeekday_Mon(t *testing.T) {
	result, err := ParseNaturalDate("mon")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %s", result.Weekday())
	}
}

func TestParseNaturalDate_InXDays(t *testing.T) {
	result, err := ParseNaturalDate("in 3 days")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Now().AddDate(0, 0, 3).Truncate(24 * time.Hour)
	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestParseNaturalDate_InXWeeks(t *testing.T) {
	result, err := ParseNaturalDate("in 2 weeks")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Now().AddDate(0, 0, 14).Truncate(24 * time.Hour)
	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestParseNaturalDate_NextWeek(t *testing.T) {
	result, err := ParseNaturalDate("next week")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %s", result.Weekday())
	}
	if !result.After(time.Now().Truncate(24 * time.Hour)) {
		t.Error("expected future date")
	}
}

func TestParseNaturalDate_NextMonth(t *testing.T) {
	result, err := ParseNaturalDate("next month")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	now := time.Now()
	expectedMonth := now.Month() + 1
	if expectedMonth > 12 {
		expectedMonth = 1
	}
	if result.Month() != expectedMonth {
		t.Errorf("expected month %s, got %s", expectedMonth, result.Month())
	}
	if result.Day() != 1 {
		t.Errorf("expected day 1, got %d", result.Day())
	}
}

func TestParseNaturalDate_NextFriday(t *testing.T) {
	result, err := ParseNaturalDate("next friday")
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

func TestParseNaturalDate_Invalid(t *testing.T) {
	_, err := ParseNaturalDate("not-a-date")
	if err == nil {
		t.Error("expected error for invalid date string")
	}
}
