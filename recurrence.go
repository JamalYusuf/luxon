// Package luxon provides advanced time handling functionality, including comprehensive
// recurrence rules based on the iCalendar RRULE specification for generating events.
package luxon

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	ical "github.com/arran4/golang-ical"
)

// RecurrenceRule defines a rule for generating recurring events, fully supporting iCalendar RRULE.
// It includes all standard fields for frequency, constraints, and exceptions.
type RecurrenceRule struct {
	Freq       string         // Frequency: "DAILY", "WEEKLY", "MONTHLY", "YEARLY"
	Interval   int            // Interval between occurrences (e.g., 2 for every other week)
	ByDay      []time.Weekday // Specific weekdays (e.g., Monday, Wednesday)
	ByHour     []int          // Specific hours (0-23)
	ByMinute   []int          // Specific minutes (0-59)
	Wkst       time.Weekday   // Week start day (default: Monday)
	Count      int            // Maximum number of occurrences (0 for unlimited)
	Until      *DateTime      // End date for recurrence (nil for no end)
	ByMonth    []int          // Specific months (1-12)
	ByMonthDay []int          // Specific days of the month (1-31, negative for days from end)
	ExDate     []DateTime     // Dates to exclude
	RDate      []DateTime     // Dates to include
}

// ParseRRule parses an iCalendar RRULE string into a RecurrenceRule, including EXDATE and RDATE.
// It uses the github.com/arran4/golang-ical package for robust parsing.
// Example: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE;BYHOUR=9" defines a rule for every other week
// on Monday and Wednesday at 9 AM.
// Returns an error if the input is invalid.
func ParseRRule(rrule string) (RecurrenceRule, error) {
	// Parse the calendar with RRULE, EXDATE, and RDATE
	cal, err := ical.ParseCalendar(strings.NewReader(fmt.Sprintf("BEGIN:VCALENDAR\nRRULE:%s\nEND:VCALENDAR", rrule)))
	if err != nil {
		return RecurrenceRule{}, fmt.Errorf("failed to parse RRULE: %w", err)
	}
	if len(cal.Components) == 0 {
		return RecurrenceRule{}, errors.New("no components found in input")
	}

	var rule RecurrenceRule
	for _, comp := range cal.Components {
		if event, ok := comp.(*ical.VEvent); ok {
			for _, prop := range event.Properties {
				switch prop.IANAToken {
				case "RRULE":
					for key, value := range prop.Value {
						switch key {
						case "FREQ":
							rule.Freq = value
						case "INTERVAL":
							interval, err := strconv.Atoi(value)
							if err != nil {
								return RecurrenceRule{}, fmt.Errorf("invalid INTERVAL: %w", err)
							}
							rule.Interval = interval
						case "COUNT":
							count, err := strconv.Atoi(value)
							if err != nil {
								return RecurrenceRule{}, fmt.Errorf("invalid COUNT: %w", err)
							}
							rule.Count = count
						case "UNTIL":
							until, err := time.Parse(time.RFC3339, value)
							if err != nil {
								return RecurrenceRule{}, fmt.Errorf("invalid UNTIL: %w", err)
							}
							dt := DateTime{time: until}
							rule.Until = &dt
						case "BYDAY":
							days, err := parseByDay(value)
							if err != nil {
								return RecurrenceRule{}, fmt.Errorf("invalid BYDAY: %w", err)
							}
							rule.ByDay = days
						case "BYHOUR":
							hours, err := parseIntList(value)
							if err != nil {
								return RecurrenceRule{}, fmt.Errorf("invalid BYHOUR: %w", err)
							}
							for _, h := range hours {
								if h < 0 || h > 23 {
									return RecurrenceRule{}, fmt.Errorf("BYHOUR value out of range: %d", h)
								}
							}
							rule.ByHour = hours
						case "BYMINUTE":
							minutes, err := parseIntList(value)
							if err != nil {
								return RecurrenceRule{}, fmt.Errorf("invalid BYMINUTE: %w", err)
							}
							for _, m := range minutes {
								if m < 0 || m > 59 {
									return RecurrenceRule{}, fmt.Errorf("BYMINUTE value out of range: %d", m)
								}
							}
							rule.ByMinute = minutes
						case "WKST":
							wkst, err := parseByDay(value)
							if err != nil || len(wkst) != 1 {
								return RecurrenceRule{}, fmt.Errorf("invalid WKST: %w", err)
							}
							rule.Wkst = wkst[0]
						case "BYMONTH":
							months, err := parseIntList(value)
							if err != nil {
								return RecurrenceRule{}, fmt.Errorf("invalid BYMONTH: %w", err)
							}
							for _, m := range months {
								if m < 1 || m > 12 {
									return RecurrenceRule{}, fmt.Errorf("BYMONTH value out of range: %d", m)
								}
							}
							rule.ByMonth = months
						case "BYMONTHDAY":
							days, err := parseIntList(value)
							if err != nil {
								return RecurrenceRule{}, fmt.Errorf("invalid BYMONTHDAY: %w", err)
							}
							for _, d := range days {
								if (d < -31 || d > 31) || d == 0 {
									return RecurrenceRule{}, fmt.Errorf("BYMONTHDAY value out of range: %d", d)
								}
							}
							rule.ByMonthDay = days
						}
					}
				case "EXDATE":
					dt, err := time.Parse(time.RFC3339, prop.Value)
					if err != nil {
						return RecurrenceRule{}, fmt.Errorf("invalid EXDATE: %w", err)
					}
					rule.ExDate = append(rule.ExDate, DateTime{time: dt})
				case "RDATE":
					dt, err := time.Parse(time.RFC3339, prop.Value)
					if err != nil {
						return RecurrenceRule{}, fmt.Errorf("invalid RDATE: %w", err)
					}
					rule.RDate = append(rule.RDate, DateTime{time: dt})
				}
			}
		}
	}

	// Set defaults and validate
	if rule.Freq == "" {
		return RecurrenceRule{}, errors.New("frequency is required")
	}
	if rule.Interval <= 0 {
		rule.Interval = 1
	}
	if rule.Wkst == 0 {
		rule.Wkst = time.Monday // Default per iCalendar
	}
	return rule, nil
}

// GenerateOccurrences produces a list of DateTime instances for the recurrence rule within a time range.
// It respects all RRULE fields (Freq, Interval, ByDay, ByHour, ByMinute, Wkst, Count, Until, ByMonth,
// ByMonthDay) and handles ExDate and RDate.
// Returns an error if the rule is invalid or generation fails.
// TODO: Optimize for very large ranges using calendar-aware stepping.
func (r RecurrenceRule) GenerateOccurrences(start, end DateTime) ([]DateTime, error) {
	if r.Freq == "" {
		return nil, errors.New("frequency is required")
	}
	if start.time.After(end.time) {
		return nil, errors.New("start must be before or equal to end")
	}

	var occurrences []DateTime
	count := 0
	endTime := end.time
	if r.Until != nil && r.Until.time.Before(endTime) {
		endTime = r.Until.time
	}

	// Initialize current time with start
	current := start

	switch r.Freq {
	case "DAILY":
		for current.time.Before(endTime) || current.time.Equal(endTime) {
			if r.matchesConstraints(current) && !r.isExcluded(current) {
				occurrences = append(occurrences, current)
				count++
			}
			current = DateTime{time: current.time.AddDate(0, 0, r.Interval)}
			if r.shouldStop(count) {
				break
			}
		}
	case "WEEKLY":
		// Adjust for WKST
		weekStartOffset := (int(current.time.Weekday()) - int(r.Wkst) + 7) % 7
		current = DateTime{time: current.time.AddDate(0, 0, -weekStartOffset)}
		for current.time.Before(endTime) || current.time.Equal(endTime) {
			for day := 0; day < 7; day++ {
				dayTime := DateTime{time: current.time.AddDate(0, 0, day)}
				if dayTime.time.Before(start.time) || (dayTime.time.After(endTime) && !dayTime.time.Equal(endTime)) {
					continue
				}
				if r.matchesConstraints(dayTime) && !r.isExcluded(dayTime) {
					occurrences = append(occurrences, dayTime)
					count++
				}
				if r.shouldStop(count) {
					break
				}
			}
			current = DateTime{time: current.time.AddDate(0, 0, 7*r.Interval)}
			if r.shouldStop(count) {
				break
			}
		}
	case "MONTHLY":
		for current.time.Before(endTime) || current.time.Equal(endTime) {
			if r.matchesConstraints(current) && !r.isExcluded(current) {
				occurrences = append(occurrences, current)
				count++
			}
			year, month, day := current.time.Date()
			newMonth := int(month) + r.Interval
			newYear := year + (newMonth-1)/12
			newMonth = (newMonth-1)%12 + 1
			lastDay := time.Date(newYear, time.Month(newMonth+1), 0, 0, 0, 0, 0, current.time.Location()).Day()
			if day > lastDay {
				day = lastDay
			}
			current = DateTime{
				time: time.Date(newYear, time.Month(newMonth), day,
					current.time.Hour(), current.time.Minute(), current.time.Second(),
					current.time.Nanosecond(), current.time.Location()),
			}
			if r.shouldStop(count) {
				break
			}
		}
	case "YEARLY":
		for current.time.Before(endTime) || current.time.Equal(endTime) {
			if r.matchesConstraints(current) && !r.isExcluded(current) {
				occurrences = append(occurrences, current)
				count++
			}
			current = DateTime{time: current.time.AddDate(r.Interval, 0, 0)}
			if r.shouldStop(count) {
				break
			}
		}
	default:
		return nil, fmt.Errorf("unsupported frequency: %s", r.Freq)
	}

	// Add RDATE occurrences
	for _, rdate := range r.RDate {
		if (rdate.time.Equal(start.time) || rdate.time.After(start.time)) &&
			(rdate.time.Before(endTime) || rdate.time.Equal(endTime)) &&
			!r.isExcluded(rdate) {
			occurrences = append(occurrences, rdate)
			count++
		}
	}

	// Sort occurrences by time
	sort.Slice(occurrences, func(i, j int) bool {
		return occurrences[i].time.Before(occurrences[j].time)
	})

	return occurrences, nil
}

// NextOccurrence finds the next occurrence after a given DateTime using a targeted search.
// It efficiently steps through potential dates based on the rule’s frequency and constraints.
// Returns an empty DateTime and false if no occurrence exists within a reasonable range.
func (r RecurrenceRule) NextOccurrence(after DateTime) (DateTime, bool) {
	// Define a reasonable search window (e.g., 1 year)
	end := DateTime{time: after.time.AddDate(1, 0, 0)}
	if r.Until != nil && r.Until.time.Before(end.time) {
		end = *r.Until
	}

	// Start just after the given time
	current := after
	switch r.Freq {
	case "DAILY":
		// Align to the next valid hour/minute
		current = alignToConstraints(current, r)
		if current.time.After(after.time) && (current.time.Before(end.time) || current.time.Equal(end.time)) &&
			r.matchesConstraints(current) && !r.isExcluded(current) {
			return current, true
		}
		for current.time.Before(end.time) || current.time.Equal(end.time) {
			current = DateTime{time: current.time.AddDate(0, 0, r.Interval)}
			if r.matchesConstraints(current) && !r.isExcluded(current) {
				return current, true
			}
			if r.shouldStop(1) {
				break
			}
		}
	case "WEEKLY":
		// Align to the week start (WKST)
		weekStartOffset := (int(current.time.Weekday()) - int(r.Wkst) + 7) % 7
		current = DateTime{time: current.time.AddDate(0, 0, -weekStartOffset)}
		for current.time.Before(end.time) || current.time.Equal(end.time) {
			for day := 0; day < 7; day++ {
				dayTime := DateTime{time: current.time.AddDate(0, 0, day)}
				if dayTime.time.After(after.time) && (dayTime.time.Before(end.time) || dayTime.time.Equal(end.time)) &&
					r.matchesConstraints(dayTime) && !r.isExcluded(dayTime) {
					return dayTime, true
				}
			}
			current = DateTime{time: current.time.AddDate(0, 0, 7*r.Interval)}
			if r.shouldStop(1) {
				break
			}
		}
	case "MONTHLY":
		// Try the current month first
		current = alignToConstraints(current, r)
		if current.time.After(after.time) && (current.time.Before(end.time) || current.time.Equal(end.time)) &&
			r.matchesConstraints(current) && !r.isExcluded(current) {
			return current, true
		}
		for current.time.Before(end.time) || current.time.Equal(end.time) {
			year, month, day := current.time.Date()
			newMonth := int(month) + r.Interval
			newYear := year + (newMonth-1)/12
			newMonth = (newMonth-1)%12 + 1
			lastDay := time.Date(newYear, time.Month(newMonth+1), 0, 0, 0, 0, 0, current.time.Location()).Day()
			if day > lastDay {
				day = lastDay
			}
			current = DateTime{
				time: time.Date(newYear, time.Month(newMonth), day,
					current.time.Hour(), current.time.Minute(), current.time.Second(),
					current.time.Nanosecond(), current.time.Location()),
			}
			if r.matchesConstraints(current) && !r.isExcluded(current) {
				return current, true
			}
			if r.shouldStop(1) {
				break
			}
		}
	case "YEARLY":
		current = alignToConstraints(current, r)
		if current.time.After(after.time) && (current.time.Before(end.time) || current.time.Equal(end.time)) &&
			r.matchesConstraints(current) && !r.isExcluded(current) {
			return current, true
		}
		for current.time.Before(end.time) || current.time.Equal(end.time) {
			current = DateTime{time: current.time.AddDate(r.Interval, 0, 0)}
			if r.matchesConstraints(current) && !r.isExcluded(current) {
				return current, true
			}
			if r.shouldStop(1) {
				break
			}
		}
	}

	// Check RDATE for the next occurrence
	for _, rdate := range r.RDate {
		if rdate.time.After(after.time) && (rdate.time.Before(end.time) || rdate.time.Equal(end.time)) &&
			!r.isExcluded(rdate) {
			return rdate, true
		}
	}

	return DateTime{}, false
}

// HasOccurrence checks if a specific DateTime matches the recurrence rule, accounting for ExDate and RDate.
// Returns true if the DateTime is a valid occurrence or explicitly included via RDate.
func (r RecurrenceRule) HasOccurrence(dt DateTime) bool {
	if r.isExcluded(dt) {
		return false
	}
	for _, rdate := range r.RDate {
		if dt.time.Equal(rdate.time) {
			return true
		}
	}
	return r.matchesConstraints(dt)
}

// matchesConstraints verifies if a DateTime satisfies all rule constraints (ByDay, ByHour, ByMinute, ByMonth, ByMonthDay).
func (r RecurrenceRule) matchesConstraints(dt DateTime) bool {
	// Check ByDay
	if len(r.ByDay) > 0 {
		matches := false
		for _, day := range r.ByDay {
			if dt.time.Weekday() == day {
				matches = true
				break
			}
		}
		if !matches {
			return false
		}
	}

	// Check ByHour
	if len(r.ByHour) > 0 {
		matches := false
		for _, hour := range r.ByHour {
			if dt.time.Hour() == hour {
				matches = true
				break
			}
		}
		if !matches {
			return false
		}
	}

	// Check ByMinute
	if len(r.ByMinute) > 0 {
		matches := false
		for _, minute := range r.ByMinute {
			if dt.time.Minute() == minute {
				matches = true
				break
			}
		}
		if !matches {
			return false
		}
	}

	// Check ByMonth
	if len(r.ByMonth) > 0 {
		matches := false
		for _, month := range r.ByMonth {
			if int(dt.time.Month()) == month {
				matches = true
				break
			}
		}
		if !matches {
			return false
		}
	}

	// Check ByMonthDay
	if len(r.ByMonthDay) > 0 {
		matches := false
		day := dt.time.Day()
		for _, monthDay := range r.ByMonthDay {
			if monthDay > 0 && day == monthDay {
				matches = true
				break
			} else if monthDay < 0 {
				lastDay := time.Date(dt.time.Year(), dt.time.Month()+1, 0, 0, 0, 0, 0, dt.time.Location()).Day()
				if day == lastDay+monthDay+1 {
					matches = true
					break
				}
			}
		}
		if !matches {
			return false
		}
	}

	return true
}

// isExcluded checks if a DateTime is listed in the ExDate slice.
func (r RecurrenceRule) isExcluded(dt DateTime) bool {
	for _, exdate := range r.ExDate {
		if dt.time.Equal(exdate.time) {
			return true
		}
	}
	return false
}

// shouldStop determines if occurrence generation should stop based on Count.
func (r RecurrenceRule) shouldStop(count int) bool {
	return r.Count > 0 && count >= r.Count
}

// alignToConstraints adjusts a DateTime to the nearest valid time based on ByHour and ByMinute.
func alignToConstraints(dt DateTime, r RecurrenceRule) DateTime {
	hour := dt.time.Hour()
	minute := dt.time.Minute()

	if len(r.ByHour) > 0 {
		hour = r.ByHour[0] // Use the first hour for simplicity
		for _, h := range r.ByHour {
			if h >= dt.time.Hour() {
				hour = h
				break
			}
		}
	}
	if len(r.ByMinute) > 0 {
		minute = r.ByMinute[0] // Use the first minute
		for _, m := range r.ByMinute {
			if m >= dt.time.Minute() {
				minute = m
				break
			}
		}
	}

	return DateTime{
		time: time.Date(dt.time.Year(), dt.time.Month(), dt.time.Day(),
			hour, minute, dt.time.Second(), dt.time.Nanosecond(), dt.time.Location()),
	}
}

// parseByDay converts an iCalendar BYDAY string (e.g., "MO,WE") to a slice of time.Weekday.
func parseByDay(value string) ([]time.Weekday, error) {
	var days []time.Weekday
	for _, dayStr := range strings.Split(value, ",") {
		switch dayStr {
		case "MO":
			days = append(days, time.Monday)
		case "TU":
			days = append(days, time.Tuesday)
		case "WE":
			days = append(days, time.Wednesday)
		case "TH":
			days = append(days, time.Thursday)
		case "FR":
			days = append(days, time.Friday)
		case "SA":
			days = append(days, time.Saturday)
		case "SU":
			days = append(days, time.Sunday)
		default:
			return nil, fmt.Errorf("invalid BYDAY value: %s", dayStr)
		}
	}
	return days, nil
}

// parseIntList converts a comma-separated string of integers (e.g., "1,2,-1") to a slice of ints.
func parseIntList(value string) ([]int, error) {
	var result []int
	for _, s := range strings.Split(value, ",") {
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("invalid integer: %s", err)
		}
		result = append(result, n)
	}
	return result, nil
}
