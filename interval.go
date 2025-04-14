// Package luxon provides enhanced time handling capabilities.
// This file defines the Interval type and its operations for managing time ranges.
package luxon

import (
	"errors"
	"fmt"
	"time"
)

// Interval represents a time range with a start and end DateTime.
// It is used for operations like intersection, duration calculation, and containment checks.
// The interval is inclusive of the start time and exclusive of the end time ([start, end)).
type Interval struct {
	Start DateTime // The start of the time range.
	End   DateTime // The end of the time range.
}

// NewInterval creates a new Interval from start and end DateTime instances.
// It returns an error if the start time is after or equal to the end time, as this would create an invalid interval.
func NewInterval(start, end DateTime) (Interval, error) {
	if start.time.After(end.time) || start.time.Equal(end.time) {
		return Interval{}, errors.New("invalid interval: start time must be before end time")
	}
	return Interval{Start: start, End: end}, nil
}

// Intersect returns the intersection of two intervals, if it exists.
// It returns the overlapping Interval and true if there is an overlap, or an empty Interval and false otherwise.
// The resulting interval follows the [start, end) convention.
func (i Interval) Intersect(other Interval) (Interval, bool) {
	start := i.Start.time
	if other.Start.time.After(start) {
		start = other.Start.time
	}
	end := i.End.time
	if other.End.time.Before(end) {
		end = other.End.time
	}
	if start.Before(end) {
		return Interval{Start: DateTime{time: start}, End: DateTime{time: end}}, true
	}
	return Interval{}, false
}

// Contains checks if a DateTime is within the interval (inclusive).
func (i Interval) Contains(dt DateTime) bool {
	return (dt.time.Equal(i.Start.time) || dt.time.After(i.Start.time)) &&
		(dt.time.Equal(i.End.time) || dt.time.Before(i.End.time))
}

// IsEmpty checks whether the interval is empty (start >= end).
func (i Interval) IsEmpty() bool {
	return !i.Start.time.Before(i.End.time)
}

// Duration returns the duration of the interval.
func (i Interval) Duration() time.Duration {
	return i.End.time.Sub(i.Start.time)
}

// Overlaps returns true if the two intervals overlap.
func (i Interval) Overlaps(other Interval) bool {
	_, ok := i.Intersect(other)
	return ok
}

// Union returns the minimal interval that covers both intervals.
// If the intervals do not overlap or touch, returns false.
func (i Interval) Union(other Interval) (Interval, bool) {
	if i.Overlaps(other) || i.Touches(other) {
		start := i.Start.time
		if other.Start.time.Before(start) {
			start = other.Start.time
		}
		end := i.End.time
		if other.End.time.After(end) {
			end = other.End.time
		}
		return Interval{Start: DateTime{time: start}, End: DateTime{time: end}}, true
	}
	return Interval{}, false
}

// Split divides the interval into a specified number of equal sub-intervals.
// Returns a slice of Intervals, or an error if the count is less than 1.
func (i Interval) Split(count int) ([]Interval, error) {
	if count < 1 {
		return nil, errors.New("split count must be at least 1")
	}
	duration := i.End.time.Sub(i.Start.time) / time.Duration(count)
	result := make([]Interval, count)
	current := i.Start
	for j := 0; j < count; j++ {
		end := DateTime{time: current.time.Add(duration)}
		if j == count-1 {
			end = i.End // Ensure the last interval reaches the end exactly
		}
		interval, err := NewInterval(current, end)
		if err != nil {
			return nil, fmt.Errorf("failed to create sub-interval %d: %w", j, err)
		}
		result[j] = interval
		current = end
	}
	return result, nil
}

// Touches checks whether two intervals touch each other (end == start).
func (i Interval) Touches(other Interval) bool {
	return i.End.time.Equal(other.Start.time) || other.End.time.Equal(i.Start.time)
}

// Subtract returns the parts of i that are not overlapped by other.
// It can return 0, 1, or 2 intervals depending on the overlap.
func (i Interval) Subtract(other Interval) []Interval {
	var result []Interval
	if !i.Overlaps(other) {
		return []Interval{i}
	}

	// If there's a left non-overlapping part
	if other.Start.time.After(i.Start.time) {
		result = append(result, Interval{
			Start: i.Start,
			End:   DateTime{time: other.Start.time},
		})
	}

	// If there's a right non-overlapping part
	if other.End.time.Before(i.End.time) {
		result = append(result, Interval{
			Start: DateTime{time: other.End.time},
			End:   i.End,
		})
	}

	return result
}

// Shift moves the interval by a specified duration, returning a new Interval.
// Both start and end times are shifted by the same amount.
func (i Interval) Shift(d time.Duration) Interval {
	return Interval{
		Start: DateTime{time: i.Start.time.Add(d)},
		End:   DateTime{time: i.End.time.Add(d)},
	}
}

// String returns a string representation of the interval in the format "[start, end)".
func (i Interval) String() string {
	return fmt.Sprintf("[%s, %s)", i.Start.Format(time.RFC3339), i.End.Format(time.RFC3339))
}
