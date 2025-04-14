// Package luxon provides advanced time handling capabilities.
// This file implements natural language parsing for time expressions.
package luxon

import (
	"fmt"
	"github.com/olebedev/when"
	"time"
)

// ParseNatural parses a natural language time expression into a DateTime.
// It supports a wide range of expressions common in business scenarios, such as:
//   - Relative times: "in 2 days", "3 weeks ago"
//   - Specific weekdays: "next Monday", "last Friday"
//   - Month-specific: "first Tuesday of next month"
//   - Simple references: "today", "tomorrow"
//
// The function uses the current time as a reference point, which can be overridden
// using the Timer interface for testing.
//
// Examples:
//
//	ParseNatural("next Monday") // Returns DateTime for the upcoming Monday
//	ParseNatural("in 2 weeks") // Returns DateTime for 2 weeks from now
//	ParseNatural("last Friday") // Returns DateTime for the previous Friday
//
// Errors are returned for unsupported or ambiguous expressions.
func ParseNatural(expr string) (DateTime, error) {
	w := when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)

	// Use the current time as the reference
	now := Now().time
	result, err := w.Parse(expr, now)
	if err != nil {
		return DateTime{}, fmt.Errorf("failed to parse expression '%s': %w", expr, err)
	}
	if result == nil {
		return DateTime{}, fmt.Errorf("no valid time found in expression '%s'", expr)
	}

	return DateTime{time: result.Time}, nil
}

// ToNatural converts a DateTime to a human-readable relative string.
// It expresses the time relative to the current time, using natural language
// terms suitable for business contexts, such as:
//   - "in 2 days"
//   - "3 hours ago"
//   - "next week"
//   - "yesterday"
//
// The output is concise and avoids overly verbose phrasing.
//
// Examples:
//
//	dt := Now().Plus(2 * 24 * time.Hour)
//	dt.ToNatural() // Returns "in 2 days"
//	dt := Now().Plus(-1 * 24 * time.Hour)
//	dt.ToNatural() // Returns "yesterday"
func (dt DateTime) ToNatural() string {
	now := Now().time
	diff := dt.time.Sub(now)
	absDiff := diff
	if diff < 0 {
		absDiff = -diff
	}

	switch {
	case absDiff < time.Minute:
		return "now"
	case absDiff < time.Hour:
		minutes := int(absDiff / time.Minute)
		return formatNatural(minutes, "minute", diff >= 0)
	case absDiff < 24*time.Hour:
		if absDiff < 2*24*time.Hour {
			if diff >= 0 {
				return "tomorrow"
			}
			return "yesterday"
		}
		days := int(absDiff / (24 * time.Hour))
		return formatNatural(days, "day", diff >= 0)
	case absDiff < 7*24*time.Hour:
		if absDiff < 2*7*24*time.Hour {
			if diff >= 0 {
				return "next week"
			}
			return "last week"
		}
		weeks := int(absDiff / (7 * 24 * time.Hour))
		return formatNatural(weeks, "week", diff >= 0)
	case absDiff < 30*24*time.Hour:
		if absDiff < 2*30*24*time.Hour {
			if diff >= 0 {
				return "next month"
			}
			return "last month"
		}
		months := int(absDiff / (30 * 24 * time.Hour))
		return formatNatural(months, "month", diff >= 0)
	default:
		years := int(absDiff / (365 * 24 * time.Hour))
		return formatNatural(years, "year", diff >= 0)
	}
}

// AdjustToBusinessDay adjusts a DateTime to the nearest business day based on a BusinessCalendar.
// If the input is already a business day, it is returned unchanged.
// Otherwise, it moves forward to the next business day.
//
// Examples:
//
//	cal := &BusinessCalendar{Weekend: map[time.Weekday]bool{time.Saturday: true, time.Sunday: true}}
//	dt := DateTime{time: time.Date(2023, 10, 28, 0, 0, 0, 0, time.UTC)} // Saturday
//	adjusted, _ := dt.AdjustToBusinessDay(cal) // Returns Monday, October 30, 2023
func (dt DateTime) AdjustToBusinessDay(cal *BusinessCalendar) (DateTime, error) {
	if cal == nil {
		return DateTime{}, fmt.Errorf("business calendar cannot be nil")
	}
	current := dt
	for !cal.IsBusinessDay(current) {
		current = current.Plus(24 * time.Hour)
	}
	return current, nil
}

// formatNatural formats a number and unit into a natural language string.
// It handles singular/plural forms and future/past tenses.
func formatNatural(num int, unit string, isFuture bool) string {
	if num == 0 {
		return "now"
	}
	prefix := "in "
	if !isFuture {
		prefix = ""
	}
	suffix := ""
	if num > 1 {
		suffix = "s"
	}
	return fmt.Sprintf("%s%d %s%s", prefix, num, unit, suffix)
}
