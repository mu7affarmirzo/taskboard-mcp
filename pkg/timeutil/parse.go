package timeutil

import (
	"fmt"
	"strings"
	"time"
)

// ParseNaturalDate parses natural language date strings into time.Time.
// Supports: "today", "tomorrow", weekday names (full and short), ISO dates,
// "Month Day" formats, "in X days/weeks", "next week/month/weekday".
func ParseNaturalDate(s string) (time.Time, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	now := time.Now()

	switch s {
	case "today":
		return now.Truncate(24 * time.Hour), nil
	case "tomorrow":
		return now.AddDate(0, 0, 1).Truncate(24 * time.Hour), nil
	case "next week":
		daysUntilMonday := int(time.Monday - now.Weekday())
		if daysUntilMonday <= 0 {
			daysUntilMonday += 7
		}
		return now.AddDate(0, 0, daysUntilMonday).Truncate(24 * time.Hour), nil
	case "next month":
		next := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
		return next, nil
	}

	weekdays := map[string]time.Weekday{
		"monday": time.Monday, "tuesday": time.Tuesday,
		"wednesday": time.Wednesday, "thursday": time.Thursday,
		"friday": time.Friday, "saturday": time.Saturday, "sunday": time.Sunday,
		"mon": time.Monday, "tue": time.Tuesday, "wed": time.Wednesday,
		"thu": time.Thursday, "fri": time.Friday, "sat": time.Saturday, "sun": time.Sunday,
	}

	// "next <weekday>" pattern
	if strings.HasPrefix(s, "next ") {
		dayName := strings.TrimPrefix(s, "next ")
		if wd, ok := weekdays[dayName]; ok {
			daysUntil := int(wd - now.Weekday())
			if daysUntil <= 0 {
				daysUntil += 7
			}
			return now.AddDate(0, 0, daysUntil).Truncate(24 * time.Hour), nil
		}
	}

	// "in X days/weeks" pattern
	if strings.HasPrefix(s, "in ") {
		var n int
		var unit string
		if _, err := fmt.Sscanf(s, "in %d %s", &n, &unit); err == nil && n > 0 {
			unit = strings.TrimSuffix(unit, "s") // normalize "days" -> "day"
			switch unit {
			case "day":
				return now.AddDate(0, 0, n).Truncate(24 * time.Hour), nil
			case "week":
				return now.AddDate(0, 0, n*7).Truncate(24 * time.Hour), nil
			}
		}
	}

	// Weekday lookup (full and short names)
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
