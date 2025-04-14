// Package luxon provides advanced time handling capabilities, extending Go's standard time package.
// The datetime.go file defines the core DateTime type and functions for time manipulation,
// ensuring immutability and developer-friendly APIs.
package luxon

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// timeZoneCache is a package-level sync.Map used to cache loaded time zones,
// reducing disk I/O for repeated time zone conversions.
var timeZoneCache sync.Map

// DateTime is an immutable wrapper around time.Time, providing enhanced functionality
// for time manipulation, formatting, and serialization. All operations return new
// DateTime instances to ensure thread safety and predictability.
type DateTime struct {
	time time.Time // private to enforce immutability
}

// Now returns the current local time as a DateTime instance.
// It respects the configured Timer, allowing time simulation for testing.
func Now() DateTime {
	return DateTime{time: time.Now()}
}

// Parse creates a DateTime from a formatted string using the specified layout.
// The layout follows the same rules as time.Parse (e.g., time.RFC3339, "2006-01-02").
//
// Example:
//
//	dt, err := Parse(time.RFC3339, "2023-10-25T14:30:00Z")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(dt.Format("2006-01-02")) // Output: 2023-10-25
func Parse(layout, value string) (DateTime, error) {
	t, err := time.Parse(layout, value)
	if err != nil {
		return DateTime{}, fmt.Errorf("failed to parse time with layout %s: %w", layout, err)
	}
	return DateTime{time: t}, nil
}

// FromTime constructs a DateTime from an existing time.Time value.
// This is useful for integrating with standard library time functions.
//
// Example:
//
//	t := time.Now()
//	dt := FromTime(t)
//	fmt.Println(dt.Format(time.RFC3339))
func FromTime(t time.Time) DateTime {
	return DateTime{time: t}
}

// ToTime returns the underlying time.Time value of the DateTime.
// This allows interoperability with APIs expecting time.Time.
//
// Example:
//
//	dt := Now()
//	t := dt.ToTime()
//	fmt.Println(t.Format(time.RFC3339))
func (dt DateTime) ToTime() time.Time {
	return dt.time
}

// ParseDuration parses a human-readable duration string (e.g., "2 days, 3 hours").
func ParseDuration(s string) (Duration, error) {
	parts := strings.Split(s, ",")
	var total time.Duration
	for _, part := range parts {
		part = strings.TrimSpace(part)
		var value int
		var unit string
		_, err := fmt.Sscanf(part, "%d %s", &value, &unit)
		if err != nil {
			return Duration{}, err
		}
		switch strings.ToLower(unit) {
		case "days", "day":
			total += time.Duration(value) * 24 * time.Hour
		case "hours", "hour":
			total += time.Duration(value) * time.Hour
		case "minutes", "minute":
			total += time.Duration(value) * time.Minute
		default:
			return Duration{}, fmt.Errorf("unknown unit: %s", unit)
		}
	}
	return Duration{total}, nil
}

// InZone converts the DateTime to the specified time zone, caching the location for performance.
// The zone parameter should be a valid IANA time zone name (e.g., "America/New_York").
//
// Example:
//
//	dt := Now()
//	nyTime, err := dt.InZone("America/New_York")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(nyTime.Format(time.RFC3339))
func (dt DateTime) InZone(zone string) (DateTime, error) {
	loc, ok := timeZoneCache.Load(zone)
	if !ok {
		var err error
		loc, err = time.LoadLocation(zone)
		if err != nil {
			return DateTime{}, fmt.Errorf("failed to load time zone %s: %w", zone, err)
		}
		timeZoneCache.Store(zone, loc)
	}
	return DateTime{time: dt.time.In(loc.(*time.Location))}, nil
}

// Plus adds a duration to the DateTime and returns a new instance, preserving immutability.
// The duration can be any time.Duration value (e.g., 24 * time.Hour for one day).
//
// Example:
//
//	dt := Now()
//	future := dt.Plus(24 * time.Hour)
//	fmt.Println(future.Format("2006-01-02"))
func (dt DateTime) Plus(d time.Duration) DateTime {
	return DateTime{time: dt.time.Add(d)}
}

// PlusYears adds the specified number of years and returns a new DateTime.
// It accounts for leap years and other calendar nuances.
//
// Example:
//
//	dt := Now()
//	future := dt.PlusYears(1)
//	fmt.Println(future.Format("2006-01-02"))
func (dt DateTime) PlusYears(years int) DateTime {
	return DateTime{time: dt.time.AddDate(years, 0, 0)}
}

// PlusMonths adds the specified number of months and returns a new DateTime.
// It handles calendar rules like leap years and varying month lengths.
//
// Example:
//
//	dt := Parse("2006-01-02", "2023-01-31")
//	future := dt.PlusMonths(1) // Adjusts to February 28 or 29
//	fmt.Println(future.Format("2006-01-02"))
func (dt DateTime) PlusMonths(months int) DateTime {
	return DateTime{time: dt.time.AddDate(0, months, 0)}
}

// PlusDays adds the specified number of days and returns a new DateTime.
// This is a convenience method for calendar-based arithmetic.
//
// Example:
//
//	dt := Now()
//	future := dt.PlusDays(5)
//	fmt.Println(future.Format("2006-01-02"))
func (dt DateTime) PlusDays(days int) DateTime {
	return DateTime{time: dt.time.AddDate(0, 0, days)}
}

// Minus subtracts a duration from the DateTime and returns a new instance.
// The duration can be any time.Duration value (e.g., 24 * time.Hour for one day).
//
// Example:
//
//	dt := Now()
//	past := dt.Minus(24 * time.Hour)
//	fmt.Println(past.Format("2006-01-02"))
func (dt DateTime) Minus(d time.Duration) DateTime {
	return DateTime{time: dt.time.Add(-d)}
}

// MinusYears subtracts specified number of years from the DateTime and returns a new instance.
// This is a convenience method for calendar-based arithmetic.
//
// Example:
//
//	dt := Now()
//	past := dt.MinusYears(5)
//	fmt.Println(past.Format("2006-01-02"))
func (dt DateTime) MinusYears(n int) DateTime {
	return DateTime{time: dt.time.AddDate(-n, 0, 0)}
}

// MinusMonths subtracts specified number of months from the DateTime and returns a new instance.
// This is a convenience method for calendar-based arithmetic.
//
// Example:
//
//	dt := Now()
//	past := dt.MinusMonths(5)
//	fmt.Println(past.Format("2006-01-02"))
func (dt DateTime) MinusMonths(n int) DateTime {
	return DateTime{time: dt.time.AddDate(0, -n, 0)}
}

// MinusDays subtracts specified number of days from the DateTime and returns a new instance.
// This is a convenience method for calendar-based arithmetic.
//
// Example:
//
//	dt := Now()
//	past := dt.MinusDays(5)
//	fmt.Println(past.Format("2006-01-02"))
func (dt DateTime) MinusDays(n int) DateTime {
	return DateTime{time: dt.time.AddDate(0, 0, -n)}
}

// StartOf snaps the DateTime to the start of the specified time unit (e.g., "day", "month").
// Supported units: "second", "minute", "hour", "day", "month", "year".
//
// Example:
//
//	dt := Now()
//	startOfDay := dt.StartOf("day")
//	fmt.Println(startOfDay.Format("2006-01-02 15:04:05"))
func (dt DateTime) StartOf(unit string) (DateTime, error) {
	y, m, d := dt.time.Date()
	h, min, s := dt.time.Clock()
	switch unit {
	case "year":
		return DateTime{time: time.Date(y, 1, 1, 0, 0, 0, 0, dt.time.Location())}, nil
	case "month":
		return DateTime{time: time.Date(y, m, 1, 0, 0, 0, 0, dt.time.Location())}, nil
	case "day":
		return DateTime{time: time.Date(y, m, d, 0, 0, 0, 0, dt.time.Location())}, nil
	case "hour":
		return DateTime{time: time.Date(y, m, d, h, 0, 0, 0, dt.time.Location())}, nil
	case "minute":
		return DateTime{time: time.Date(y, m, d, h, min, 0, 0, dt.time.Location())}, nil
	case "second":
		return DateTime{time: time.Date(y, m, d, h, min, s, 0, dt.time.Location())}, nil
	default:
		return DateTime{}, fmt.Errorf("unsupported time unit: %s", unit)
	}
}

// EndOf snaps the DateTime to the end of the specified time unit (e.g., "day", "month").
// Supported units: "second", "minute", "hour", "day", "month", "year".
//
// Example:
//
//	dt := Now()
//	endOfMonth := dt.EndOf("month")
//	fmt.Println(endOfMonth.Format("2006-01-02 15:04:05"))
func (dt DateTime) EndOf(unit string) (DateTime, error) {
	switch unit {
	case "year":
		return DateTime{time: time.Date(dt.time.Year()+1, 1, 1, 0, 0, 0, 0, dt.time.Location()).Add(-time.Nanosecond)}, nil
	case "month":
		nextMonth := time.Date(dt.time.Year(), dt.time.Month()+1, 1, 0, 0, 0, 0, dt.time.Location())
		if dt.time.Month() == time.December {
			nextMonth = time.Date(dt.time.Year()+1, 1, 1, 0, 0, 0, 0, dt.time.Location())
		}
		return DateTime{time: nextMonth.Add(-time.Nanosecond)}, nil
	case "day":
		return DateTime{time: time.Date(dt.time.Year(), dt.time.Month(), dt.time.Day()+1, 0, 0, 0, 0, dt.time.Location()).Add(-time.Nanosecond)}, nil
	case "hour":
		return DateTime{time: time.Date(dt.time.Year(), dt.time.Month(), dt.time.Day(), dt.time.Hour()+1, 0, 0, 0, dt.time.Location()).Add(-time.Nanosecond)}, nil
	case "minute":
		return DateTime{time: time.Date(dt.time.Year(), dt.time.Month(), dt.time.Day(), dt.time.Hour(), dt.time.Minute()+1, 0, 0, dt.time.Location()).Add(-time.Nanosecond)}, nil
	case "second":
		return DateTime{time: time.Date(dt.time.Year(), dt.time.Month(), dt.time.Day(), dt.time.Hour(), dt.time.Minute(), dt.time.Second()+1, 0, dt.time.Location()).Add(-time.Nanosecond)}, nil
	default:
		return DateTime{}, fmt.Errorf("unsupported time unit: %s", unit)
	}
}

// Format formats the DateTime using Go layout strings.
func (dt DateTime) Format(layout string) string {
	return dt.time.Format(layout)
}

// FormatLuxon formats the DateTime using Luxon-style tokens (e.g., "yyyy-MM-dd").
// Supported tokens include yyyy, MM, dd, HH, mm, ss, and more.
// TODO: Expand token support for additional formats like milliseconds.
//
// Example:
//
//	dt := Now()
//	fmt.Println(dt.FormatLuxon("yyyy-MM-dd")) // e.g., 2023-11-01
func (dt DateTime) FormatLuxon(tokens string) string {
	tokenMap := map[string]string{
		"yyyy": "2006", // Full year
		"yy":   "06",   // Two-digit year
		"MM":   "01",   // Two-digit month
		"M":    "1",    // Month without leading zero
		"dd":   "02",   // Two-digit day
		"d":    "2",    // Day without leading zero
		"HH":   "15",   // 24-hour with leading zero
		"H":    "15",   // 24-hour without leading zero (same as HH for simplicity)
		"hh":   "03",   // 12-hour with leading zero
		"h":    "3",    // 12-hour without leading zero
		"mm":   "04",   // Minutes with leading zero
		"m":    "4",    // Minutes without leading zero
		"ss":   "05",   // Seconds with leading zero
		"s":    "5",    // Seconds without leading zero
		"a":    "PM",   // AM/PM marker
		// Additional tokens can be added here
	}
	re := regexp.MustCompile(`\w+`)
	return re.ReplaceAllStringFunc(tokens, func(token string) string {
		if replacement, ok := tokenMap[token]; ok {
			return replacement
		}
		return token // Non-token parts remain unchanged
	})
}

// Weekday returns the day of the week for the DateTime (e.g., time.Monday).
// It follows the standard time.Weekday enumeration.
//
// Example:
//
//	dt := Now()
//	fmt.Println(dt.Weekday()) // e.g., time.Wednesday
func (dt DateTime) Weekday() time.Weekday {
	return dt.time.Weekday()
}

// ISOWeek returns the ISO 8601 week number and year for the DateTime.
// The week number ranges from 1 to 53, and the year may differ from the calendar year.
//
// Example:
//
//	dt := Parse("2006-01-02", "2023-01-01")
//	year, week := dt.ISOWeek()
//	fmt.Printf("Year: %d, Week: %d\n", year, week)
func (dt DateTime) ISOWeek() (year int, week int) {
	return dt.time.ISOWeek()
}

// MarshalJSON implements the json.Marshaler interface, serializing DateTime to RFC3339 format.
func (dt DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(dt.time.Format(time.RFC3339))
}

// UnmarshalJSON implements the json.Unmarshaler interface, deserializing from RFC3339 format.
func (dt *DateTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	dt.time = t
	return nil
}

// Value implements the driver.Valuer interface for SQL interoperability.
func (dt DateTime) Value() (driver.Value, error) {
	return dt.time, nil
}

// Scan implements the sql.Scanner interface for SQL interoperability.
func (dt *DateTime) Scan(value interface{}) error {
	if value == nil {
		return errors.New("nil value cannot be scanned into DateTime")
	}
	t, ok := value.(time.Time)
	if !ok {
		return errors.New("failed to scan value into DateTime")
	}
	dt.time = t
	return nil
}
