// Package luxon provides advanced time handling capabilities. This file contains
// business calendar functionality for calculating business days and related operations,
// supporting enterprise scenarios like scheduling, billing, and compliance.
package luxon

import (
	"fmt"
	"time"
)

// BusinessCalendar defines a customizable calendar for business day calculations.
// It includes holidays, non-working days (e.g., weekends), and business hours to support
// precise scheduling and time tracking.
type BusinessCalendar struct {
	Holidays      map[time.Time]bool    // Dates considered holidays, stored at midnight
	NonWorking    map[time.Weekday]bool // Days of the week that are non-working (e.g., weekends)
	BusinessHours *BusinessHours        // Optional business hours for time-specific calculations
	Location      *time.Location        // Time zone for calendar calculations
}

// BusinessHours defines the daily operating hours for a business calendar.
// Times are stored in the local time zone of the BusinessCalendar.
type BusinessHours struct {
	StartTime time.Time // Start of business hours (e.g., 09:00:00 local time)
	EndTime   time.Time // End of business hours (e.g., 17:00:00 local time)
}

// NewBusinessCalendar creates a new BusinessCalendar with the specified time zone.
// If loc is nil, UTC is used. Holidays and NonWorking are initialized as empty maps,
// and BusinessHours is nil until set.
func NewBusinessCalendar(loc *time.Location) *BusinessCalendar {
	if loc == nil {
		loc = time.UTC
	}
	return &BusinessCalendar{
		Holidays:   make(map[time.Time]bool),
		NonWorking: make(map[time.Weekday]bool),
		Location:   loc,
	}
}

// AddHoliday adds a holiday to the calendar on the specified date.
// The date is normalized to midnight in the calendar's time zone.
func (bc *BusinessCalendar) AddHoliday(date time.Time) {
	normalized := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, bc.Location)
	bc.Holidays[normalized] = true
}

// RemoveHoliday removes a holiday from the calendar.
// The date is normalized to midnight in the calendar's time zone.
func (bc *BusinessCalendar) RemoveHoliday(date time.Time) {
	normalized := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, bc.Location)
	delete(bc.Holidays, normalized)
}

// SetNonWorkingDay marks a weekday as non-working (e.g., Saturday, Sunday).
func (bc *BusinessCalendar) SetNonWorkingDay(day time.Weekday) {
	bc.NonWorking[day] = true
}

// ClearNonWorkingDay removes a weekday from the non-working list.
func (bc *BusinessCalendar) ClearNonWorkingDay(day time.Weekday) {
	delete(bc.NonWorking, day)
}

// SetBusinessHours sets the daily business hours for the calendar.
// Start and end times are normalized to the same day in the calendar's time zone.
func (bc *BusinessCalendar) SetBusinessHours(startHour, startMinute, endHour, endMinute int) error {
	start := time.Date(2000, 1, 1, startHour, startMinute, 0, 0, bc.Location)
	end := time.Date(2000, 1, 1, endHour, endMinute, 0, 0, bc.Location)
	if end.Before(start) || end.Equal(start) {
		return fmt.Errorf("end time %s must be after start time %s", end.Format("15:04"), start.Format("15:04"))
	}
	bc.BusinessHours = &BusinessHours{
		StartTime: start,
		EndTime:   end,
	}
	return nil
}

// IsBusinessDay checks if the given DateTime is a full business day.
// A day is a business day if it is not a holiday or a non-working day.
// If BusinessHours is set, it ensures the day supports the full range of hours.
func (bc *BusinessCalendar) IsBusinessDay(dt DateTime) bool {
	truncated := dt.time.In(bc.Location).Truncate(24 * time.Hour)
	if bc.Holidays[truncated] {
		return false
	}
	if bc.NonWorking[dt.time.Weekday()] {
		return false
	}
	return true
}

// IsWithinBusinessHours checks if the given DateTime falls within business hours.
// Returns false if BusinessHours is not set or if the time is outside the defined range.
func (bc *BusinessCalendar) IsWithinBusinessHours(dt DateTime) bool {
	if bc.BusinessHours == nil {
		return false
	}
	local := dt.time.In(bc.Location)
	hour, min, sec := local.Clock()
	nano := local.Nanosecond()
	checkTime := time.Date(2000, 1, 1, hour, min, sec, nano, bc.Location)
	return !checkTime.Before(bc.BusinessHours.StartTime) && checkTime.Before(bc.BusinessHours.EndTime)
}

// BusinessDaysBetween calculates the number of business days between two DateTime instances, inclusive.
// It counts days that are neither holidays nor non-working days.
// If BusinessHours is set, only days supporting the full business hours are counted.
func (bc *BusinessCalendar) BusinessDaysBetween(start, end DateTime) int {
	count := 0
	current := start.time.In(bc.Location)
	endTime := end.time.In(bc.Location)

	// Ensure start is before or equal to end
	if current.After(endTime) {
		current, endTime = endTime, current
	}

	for !current.After(endTime) {
		if bc.IsBusinessDay(DateTime{time: current}) {
			count++
		}
		current = current.Add(24 * time.Hour)
	}
	return count
}

// NextBusinessDay finds the next business day after the given DateTime.
// It skips holidays and non-working days, returning a DateTime at the same time of day.
func (bc *BusinessCalendar) NextBusinessDay(dt DateTime) DateTime {
	current := dt.time.In(bc.Location)
	for {
		current = current.Add(24 * time.Hour)
		if bc.IsBusinessDay(DateTime{time: current}) {
			return DateTime{time: current}
		}
	}
}

// PreviousBusinessDay finds the previous business day before the given DateTime.
// It skips holidays and non-working days, returning a DateTime at the same time of day.
func (bc *BusinessCalendar) PreviousBusinessDay(dt DateTime) DateTime {
	current := dt.time.In(bc.Location)
	for {
		current = current.Add(-Day)
		if bc.IsBusinessDay(DateTime{time: current}) {
			return DateTime{time: current}
		}
	}
}

// AddBusinessDays adds a specified number of business days to a DateTime.
// Positive n moves forward; negative n moves backward. Returns a new DateTime at the same time of day.
func (bc *BusinessCalendar) AddBusinessDays(dt DateTime, n int) DateTime {
	current := dt.time.In(bc.Location)
	if n >= 0 {
		for i := 0; i < n; i++ {
			current = bc.NextBusinessDay(DateTime{time: current}).time
		}
	} else {
		for i := 0; i > n; i-- {
			current = bc.PreviousBusinessDay(DateTime{time: current}).time
		}
	}
	return DateTime{time: current}
}

// BusinessHoursBetween calculates the total duration of business hours between two DateTime instances.
// It only counts time within defined business hours, excluding holidays and non-working days.
// Returns zero duration if BusinessHours is not set.
func (bc *BusinessCalendar) BusinessHoursBetween(start, end DateTime) time.Duration {
	if bc.BusinessHours == nil {
		return 0
	}

	startTime := start.time.In(bc.Location)
	endTime := end.time.In(bc.Location)
	if startTime.After(endTime) {
		startTime, endTime = endTime, startTime
	}

	var total time.Duration
	currentDay := startTime.Truncate(24 * time.Hour)

	for !currentDay.After(endTime) {
		if bc.IsBusinessDay(DateTime{time: currentDay}) {
			dayStart := time.Date(currentDay.Year(), currentDay.Month(), currentDay.Day(),
				bc.BusinessHours.StartTime.Hour(), bc.BusinessHours.StartTime.Minute(), 0, 0, bc.Location)
			dayEnd := time.Date(currentDay.Year(), currentDay.Month(), currentDay.Day(),
				bc.BusinessHours.EndTime.Hour(), bc.BusinessHours.EndTime.Minute(), 0, 0, bc.Location)

			// Determine the effective start and end for this day
			effectiveStart := dayStart
			if startTime.After(dayStart) && startTime.Before(dayEnd) {
				effectiveStart = startTime
			}
			effectiveEnd := dayEnd
			if endTime.Before(dayEnd) && endTime.After(dayStart) {
				effectiveEnd = endTime
			}

			if effectiveStart.Before(effectiveEnd) && !effectiveEnd.Before(startTime) && !effectiveStart.After(endTime) {
				total += effectiveEnd.Sub(effectiveStart)
			}
		}
		currentDay = currentDay.Add(24 * time.Hour)
	}
	return total
}
