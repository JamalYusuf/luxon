// Package luxon provides enhanced time handling utilities.
// This file defines the Duration type and related functions for manipulating time durations.
package luxon

import (
	"fmt"
	"strings"
	"time"
)

const (
	Day   = 24 * time.Hour //Approximate day (not exact)
	Month = 30 * Day       //Approximate months as 30 days each (not exact)
	Year  = 12 * Month     //Approximate months as 30 days each, 12 months in year (not exact)
)

// Duration wraps time.Duration to provide additional functionality such as human-readable
// formatting, arithmetic operations, and conversion utilities.
// It is immutable; operations return new Duration instances.
type Duration struct {
	d time.Duration
}

// NewDuration creates a new Duration from a time.Duration.
// It is useful for converting standard durations into the luxon Duration type.
//
// Example:
//
//	d := luxon.NewDuration(2 * time.Hour)
//	fmt.Println(d.Humanize()) // Output: "2 hours"
func NewDuration(d time.Duration) Duration {
	return Duration{d: d}
}

// Years creates a Duration representing the specified number of Years.
// Approximate months as 30 days each (not exact)
//
// Example:
//
//	d := luxon.Months(5)
//	fmt.Println(d.Humanize()) // Output: "5 Months"
func Years(n int) Duration {
	return Duration{d: time.Duration(n) * Year}
}

// Months creates a Duration representing the specified number of Months.
// Approximate months as 30 days each (not exact)
//
// Example:
//
//	d := luxon.Months(5)
//	fmt.Println(d.Humanize()) // Output: "5 Months"
func Months(n int) Duration {
	return Duration{d: time.Duration(n) * Month}
}

// Days creates a Duration representing the specified number of days.
//
// Example:
//
//	d := luxon.Days(5)
//	fmt.Println(d.Humanize()) // Output: "5 Days"
func Days(n int) Duration {
	return Duration{d: time.Duration(n) * Day}
}

// Hours creates a Duration representing the specified number of hours.
//
// Example:
//
//	d := luxon.Hours(5)
//	fmt.Println(d.Humanize()) // Output: "5 hours"
func Hours(n int) Duration {
	return Duration{d: time.Duration(n) * time.Hour}
}

// Minutes creates a Duration representing the specified number of minutes.
//
// Example:
//
//	d := luxon.Minutes(45)
//	fmt.Println(d.Humanize()) // Output: "45 minutes"
func Minutes(n int) Duration {
	return Duration{d: time.Duration(n) * time.Minute}
}

// Seconds creates a Duration representing the specified number of seconds.
//
// Example:
//
//	d := luxon.Seconds(30)
//	fmt.Println(d.Humanize()) // Output: "30 seconds"
func Seconds(n int) Duration {
	return Duration{d: time.Duration(n) * time.Second}
}

// Humanize returns a human-readable string representation of the duration.
// It breaks down the duration into days, hours, minutes, and seconds, omitting zero-valued units.
// If the duration is zero, it returns "0 seconds".
//
// Example:
//
//	d := luxon.Hours(2).Plus(luxon.Minutes(30))
//	fmt.Println(d.Humanize()) // Output: "2 hours, 30 minutes"
func (d Duration) Humanize() string {
	if d.d == 0 {
		return "0 seconds"
	}

	var parts []string
	days := d.d / (24 * time.Hour)
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", days, pluralize("day", int64(days))))
	}
	remainder := d.d % (24 * time.Hour)
	hours := remainder / time.Hour
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", hours, pluralize("hour", int64(hours))))
	}
	remainder = remainder % time.Hour
	minutes := remainder / time.Minute
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", minutes, pluralize("minute", int64(minutes))))
	}
	remainder = remainder % time.Minute
	seconds := remainder / time.Second
	if seconds > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", seconds, pluralize("second", int64(seconds))))
	}

	return strings.Join(parts, ", ")
}

// Plus adds another Duration and returns a new Duration instance.
// It supports combining durations for cumulative time calculations.
//
// Example:
//
//	d1 := luxon.Days(1)
//	d2 := luxon.Hours(12)
//	sum := d1.Plus(d2)
//	fmt.Println(sum.Humanize()) // Output: "1 day, 12 hours"
func (d Duration) Plus(other Duration) Duration {
	return Duration{d: d.d + other.d}
}

// Minus subtracts another Duration and returns a new Duration instance.
// If the result would be negative, it returns a zero Duration.
//
// Example:
//
//	d1 := luxon.Hours(5)
//	d2 := luxon.Hours(2)
//	diff := d1.Minus(d2)
//	fmt.Println(diff.Humanize()) // Output: "3 hours"
func (d Duration) Minus(other Duration) Duration {
	if d.d < other.d {
		return Duration{d: 0}
	}
	return Duration{d: d.d - other.d}
}

// Scale multiplies the Duration by a factor and returns a new instance.
// It is useful for proportional adjustments of time spans.
//
// Example:
//
//	d := luxon.Minutes(10)
//	scaled := d.Scale(3)
//	fmt.Println(scaled.Humanize()) // Output: "30 minutes"
func (d Duration) Scale(factor int) Duration {
	return Duration{d: d.d * time.Duration(factor)}
}

// ToStandard returns the underlying time.Duration for interoperability with standard library functions.
//
// Example:
//
//	d := luxon.Days(2)
//	std := d.ToStandard()
//	fmt.Println(std) // Output: 48h0m0s
func (d Duration) ToStandard() time.Duration {
	return d.d
}

// IsZero checks if the Duration is zero.
// It is useful for conditional logic involving durations.
//
// Example:
//
//	d := luxon.Duration{d: 0}
//	fmt.Println(d.IsZero()) // Output: true
func (d Duration) IsZero() bool {
	return d.d == 0
}

// AsDays returns the Duration as a number of days (fractional).
// It is useful for reporting or conversions.
//
// Example:
//
//	d := luxon.Hours(36)
//	fmt.Println(d.AsDays()) // Output: 1.5
func (d Duration) AsDays() float64 {
	return float64(d.d) / float64(24*time.Hour)
}

// AsHours returns the Duration as a number of hours (fractional).
// It provides a convenient way to express durations in hours.
//
// Example:
//
//	d := luxon.Minutes(90)
//	fmt.Println(d.AsHours()) // Output: 1.5
func (d Duration) AsHours() float64 {
	return float64(d.d) / float64(time.Hour)
}

// AsMinutes returns the Duration as a number of minutes (fractional).
// It simplifies conversions to minute-based units.
//
// Example:
//
//	d := luxon.Seconds(90)
//	fmt.Println(d.AsMinutes()) // Output: 1.5
func (d Duration) AsMinutes() float64 {
	return float64(d.d) / float64(time.Minute)
}

// AsSeconds returns the Duration as a number of seconds (fractional).
// It is useful for precise calculations.
//
// Example:
//
//	d := luxon.Seconds(90)
//	fmt.Println(d.AsSeconds()) // Output: 90
func (d Duration) AsSeconds() float64 {
	return float64(d.d) / float64(time.Second)
}

// String returns a string representation of the Duration in a concise format.
// It uses the standard time.Duration string format for consistency.
//
// Example:
//
//	d := luxon.Hours(2)
//	fmt.Println(d.String()) // Output: "2h0m0s"
func (d Duration) String() string {
	return d.d.String()
}

// pluralize returns the singular or plural form of a word based on the count.
func pluralize(word string, count int64) string {
	if count == 1 {
		return word
	}
	return word + "s"
}
