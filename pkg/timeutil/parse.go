package timeutil

import (
	"fmt"
	"strings"
	"time"
)

// ParseNaturalDate parses natural language date strings into time.Time.
// Supports: "today", "tomorrow", weekday names, ISO dates, and "Month Day" formats.
func ParseNaturalDate(s string) (time.Time, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	now := time.Now()

	switch s {
	case "today":
		return now.Truncate(24 * time.Hour), nil
	case "tomorrow":
		return now.AddDate(0, 0, 1).Truncate(24 * time.Hour), nil
	}

	weekdays := map[string]time.Weekday{
		"monday": time.Monday, "tuesday": time.Tuesday,
		"wednesday": time.Wednesday, "thursday": time.Thursday,
		"friday": time.Friday, "saturday": time.Saturday, "sunday": time.Sunday,
	}
	if wd, ok := weekdays[s]; ok {
		daysUntil := int(wd - now.Weekday())
		if daysUntil <= 0 {
			daysUntil += 7
		}
		return now.AddDate(0, 0, daysUntil).Truncate(24 * time.Hour), nil
	}

	formats := []string{"2006-01-02", "January 2", "Jan 2", "January 2, 2006", "Jan 2, 2006"}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			if t.Year() == 0 {
				t = t.AddDate(now.Year(), 0, 0)
			}
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse date: %s", s)
}
