
![Go](https://img.shields.io/badge/Go-1.18+-00ADD8.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

## This package is not yet stable
**Still under active developer, no tagged version released  yet.** 

[View documentation - https://jamalyusuf.github.io/luxon/](https://jamalyusuf.github.io/luxon/)
## Overview
![luxon Logo](docs/assets/logo.png)
The `luxon` package is a powerful, feature-rich extension of Go’s standard `time` package, designed to simplify and enhance time-related operations in Go applications. Inspired by the JavaScript library Luxon, it provides an immutable `DateTime` type, advanced time zone handling, business day calculations, recurrence rules, natural language parsing, and more. The package aims to address common pain points in time manipulation, making it easier to handle scheduling, reporting, and user-facing time displays in business applications.

### Goals
- **Simplify Time Handling**: Offer an intuitive API for parsing, formatting, and manipulating dates and times.
- **Ensure Safety**: Use immutability to prevent concurrency issues and improve reliability.
- **Enhance Functionality**: Provide features like business calendars, recurrence rules, and natural language support not available in the standard library.
- **Optimize Performance**: Include caching and batch processing for high-throughput scenarios.
- **Support Testing**: Enable time simulation for robust testing of time-sensitive logic.

### What It Solves
- **Complex Scheduling**: Simplifies recurring events and business day calculations for enterprise applications.
- **User-Friendly Displays**: Converts times to human-readable formats, improving UX in dashboards and reports.
- **Global Applications**: Handles time zones efficiently with caching, ideal for distributed systems.
- **Testing Challenges**: Offers a mockable time source to simulate different time scenarios.
- **Data Integration**: Seamlessly integrates with JSON and SQL, reducing boilerplate for persistence.

### Use Cases
- **Enterprise Scheduling**: Calculate business days for payroll, billing, or delivery schedules.
- **Event Management**: Generate recurring events for calendars or booking systems.
- **User Interfaces**: Display relative times like "2 days ago" or parse inputs like "next Monday".
- **Analytics**: Process large datasets with efficient batch time zone conversions.
- **Testing**: Simulate time for testing expiration logic or scheduled tasks.

## Installation

To use `luxon`, install it with:

```bash
go get github.com/JamalYusuf/luxon
```

The package requires Go 1.18 or later.

## Components

The `luxon` package is organized into logical modules, each addressing specific time-related needs. Below is a detailed overview of each component, its functionality, and how to use it.

### Core Types
- **`DateTime`**: An immutable wrapper around `time.Time`, used for all time operations. It ensures thread safety and supports serialization.
- **`Duration`**: Extends `time.Duration` with human-readable formatting and arithmetic methods.
- **`Interval`**: Represents a time range with start and end `DateTime` instances, supporting operations like intersection and containment checks.
- **`BusinessCalendar`**: Defines holidays, non-working days, and business hours for business day calculations.
- **`BusinessHours`**: Specifies daily operating hours within a `BusinessCalendar`.
- **`RecurrenceRule`**: Configures recurring events, supporting iCalendar RRULE standards.
- **`Timer`**: An interface for abstracting time operations, with `DefaultTimer` for production and `MockTimer` for testing.
- **`DateTimePair`**: A utility struct for pairing `DateTime` instances in batch operations.

### Functionality by Module

#### Time Creation and Manipulation (`datetime.go`)
Handles creation, parsing, and modification of `DateTime` instances.

- **Functions**:
    - `Now() DateTime`: Returns the current time.
    - `Parse(layout, value string) (DateTime, error)`: Parses a string into a `DateTime`.
    - `FromTime(t time.Time) DateTime`: Converts a `time.Time` to a `DateTime`.
    - `ParseDuration(s string) (Duration, error)`: Parses a human-readable duration string.

- **Methods**:
    - `(dt DateTime) ToTime() time.Time`: Returns the underlying `time.Time`.
    - `(dt DateTime) InZone(zone string) (DateTime, error)`: Converts to a specified time zone.
    - `(dt DateTime) Plus(d time.Duration) DateTime`: Adds a duration.
    - `(dt DateTime) PlusYears(years int) DateTime`: Adds years.
    - `(dt DateTime) PlusMonths(months int) DateTime`: Adds months.
    - `(dt DateTime) PlusDays(days int) DateTime`: Adds days.
    - `(dt DateTime) Minus(d time.Duration) DateTime`: Subtracts a duration.
    - `(dt DateTime) MinusYears(n int) DateTime`: Subtracts years.
    - `(dt DateTime) MinusMonths(n int) DateTime`: Subtracts months.
    - `(dt DateTime) MinusDays(n int) DateTime`: Subtracts days.
    - `(dt DateTime) StartOf(unit string) (DateTime, error)`: Snaps to the start of a time unit.
    - `(dt DateTime) EndOf(unit string) (DateTime, error)`: Snaps to the end of a time unit.
    - `(dt DateTime) Format(layout string) string`: Formats using Go layout strings.
    - `(dt DateTime) FormatLuxon(tokens string) string`: Formats using Luxon-style tokens.
    - `(dt DateTime) Weekday() time.Weekday`: Returns the day of the week.
    - `(dt DateTime) ISOWeek() (year int, week int)`: Returns the ISO 8601 week and year.
    - `(dt DateTime) MarshalJSON() ([]byte, error)`: Serializes to JSON.
    - `(dt *DateTime) UnmarshalJSON(data []byte) error`: Deserializes from JSON.
    - `(dt DateTime) Value() (driver.Value, error)`: Supports SQL value conversion.
    - `(dt *DateTime) Scan(value interface{}) error`: Supports SQL scanning.

**Usage**:
```go
package main

import (
	"fmt"
	"time"
	"github.com/JamalYusuf/luxon"
)

func main() {
	// Create and manipulate a DateTime
	dt := luxon.Now()
	future := dt.PlusDays(5).InZone("America/New_York")
	fmt.Println(future.FormatLuxon("yyyy-MM-dd HH:mm"))

	// Parse a string
	parsed, err := luxon.Parse(time.RFC3339, "2023-10-25T14:30:00Z")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Parsed:", parsed.Format("2006-01-02"))
	}
}
```

**Business Value**: Simplifies date calculations for reporting (e.g., quarterly summaries) and scheduling (e.g., delivery dates).

#### Duration Handling (`duration.go`)
Enhances `time.Duration` with creation, arithmetic, and formatting utilities.

- **Functions**:
    - `NewDuration(d time.Duration) Duration`: Creates a `Duration` from a `time.Duration`.
    - `Years(n int) Duration`: Creates a duration of years (approximated as 365 days).
    - `Months(n int) Duration`: Creates a duration of months (approximated as 30 days).
    - `Days(n int) Duration`: Creates a duration of days.
    - `Hours(n int) Duration`: Creates a duration of hours.
    - `Minutes(n int) Duration`: Creates a duration of minutes.
    - `Seconds(n int) Duration`: Creates a duration of seconds.

- **Methods**:
    - `(d Duration) Humanize() string`: Returns a human-readable string.
    - `(d Duration) Plus(other Duration) Duration`: Adds another duration.
    - `(d Duration) Minus(other Duration) Duration`: Subtracts another duration.
    - `(d Duration) Scale(factor int) Duration`: Multiplies by a factor.
    - `(d Duration) ToStandard() time.Duration`: Returns the underlying `time.Duration`.
    - `(d Duration) IsZero() bool`: Checks if the duration is zero.
    - `(d Duration) AsDays() float64`: Converts to fractional days.
    - `(d Duration) AsHours() float64`: Converts to fractional hours.
    - `(d Duration) AsMinutes() float64`: Converts to fractional minutes.
    - `(d Duration) AsSeconds() float64`: Converts to fractional seconds.
    - `(d Duration) String() string`: Returns a standard duration string.

**Usage**:
```go
package main

import (
	"fmt"
	"github.com/JamalYusuf/luxon"
)

func main() {
	d := luxon.Days(2).Plus(luxon.Hours(3))
	fmt.Println(d.Humanize()) // Output: "2 days, 3 hours"
	fmt.Println(d.AsHours())  // Output: 51
}
```

**Business Value**: Useful for billing systems (e.g., calculating subscription periods) and time tracking (e.g., employee hours).

#### Interval Operations (`interval.go`)
Manages time ranges with start and end `DateTime` instances.

- **Functions**:
    - `NewInterval(start, end DateTime) (Interval, error)`: Creates a new interval.

- **Methods**:
    - `(i Interval) Intersect(other Interval) (Interval, bool)`: Finds the intersection.
    - `(i Interval) Contains(dt DateTime) bool`: Checks if a `DateTime` is within the interval.
    - `(i Interval) IsEmpty() bool`: Checks if the interval is empty.
    - `(i Interval) Duration() time.Duration`: Returns the interval’s duration.
    - `(i Interval) Overlaps(other Interval) bool`: Checks for overlap.
    - `(i Interval) Union(other Interval) (Interval, bool)`: Combines overlapping intervals.
    - `(i Interval) Split(count int) ([]Interval, error)`: Divides into equal sub-intervals.
    - `(i Interval) Touches(other Interval) bool`: Checks if intervals touch.
    - `(i Interval) Subtract(other Interval) []Interval`: Removes overlap, returning remaining parts.
    - `(i Interval) Shift(d time.Duration) Interval`: Shifts the interval.
    - `(i Interval) String() string`: Returns a string representation.

**Usage**:
```go
package main

import (
	"fmt"
	"github.com/JamalYusuf/luxon"
)

func main() {
	start := luxon.Now()
	end := start.PlusDays(10)
	interval, _ := luxon.NewInterval(start, end)
	fmt.Println(interval.Duration().Hours()) // Output: 240
}
```

**Business Value**: Ideal for availability checks in booking systems or analyzing time ranges in analytics.

#### Business Day Calculations (`business.go`)
Handles scheduling with custom business calendars.

- **Types**:
    - `BusinessCalendar`: Defines holidays, non-working days, and business hours.
    - `BusinessHours`: Specifies daily operating hours.

- **Functions**:
    - `NewBusinessCalendar(loc *time.Location) *BusinessCalendar`: Creates a new calendar.

- **Methods**:
    - `(bc *BusinessCalendar) AddHoliday(date time.Time)`: Adds a holiday.
    - `(bc *BusinessCalendar) RemoveHoliday(date time.Time)`: Removes a holiday.
    - `(bc *BusinessCalendar) SetNonWorkingDay(day time.Weekday)`: Marks a day as non-working.
    - `(bc *BusinessCalendar) ClearNonWorkingDay(day time.Weekday)`: Clears a non-working day.
    - `(bc *BusinessCalendar) SetBusinessHours(startHour, startMinute, endHour, endMinute int) error`: Sets business hours.
    - `(bc *BusinessCalendar) IsBusinessDay(dt DateTime) bool`: Checks if a day is a business day.
    - `(bc *BusinessCalendar) IsWithinBusinessHours(dt DateTime) bool`: Checks if a time is within business hours.
    - `(bc *BusinessCalendar) BusinessDaysBetween(start, end DateTime) int`: Counts business days.
    - `(bc *BusinessCalendar) NextBusinessDay(dt DateTime) DateTime`: Finds the next business day.
    - `(bc *BusinessCalendar) PreviousBusinessDay(dt DateTime) DateTime`: Finds the previous business day.
    - `(bc *BusinessCalendar) AddBusinessDays(dt DateTime, n int) DateTime`: Adds business days.
    - `(bc *BusinessCalendar) BusinessHoursBetween(start, end DateTime) time.Duration`: Calculates business hours duration.

**Usage**:
```go
package main

import (
	"fmt"
	"time"
	"github.com/JamalYusuf/luxon"
)

func main() {
	cal := luxon.NewBusinessCalendar(time.UTC)
	cal.SetNonWorkingDay(time.Saturday)
	cal.SetNonWorkingDay(time.Sunday)
	start := luxon.Now()
	end := start.PlusDays(7)
	count := cal.BusinessDaysBetween(start, end)
	fmt.Println("Business days:", count) // e.g., 5
}
```

**Business Value**: Critical for payroll (e.g., calculating workdays) and logistics (e.g., delivery schedules).

#### Natural Language Parsing (`natural.go`)
Parses and formats human-readable time expressions.

- **Functions**:
    - `ParseNatural(expr string) (DateTime, error)`: Parses expressions like "next Monday".

- **Methods**:
    - `(dt DateTime) ToNatural() string`: Converts to a relative string like "in 2 days".
    - `(dt DateTime) AdjustToBusinessDay(cal *BusinessCalendar) (DateTime, error)`: Adjusts to the nearest business day.

**Usage**:
```go
package main

import (
	"fmt"
	"github.com/JamalYusuf/luxon"
)

func main() {
	dt, err := luxon.ParseNatural("next Monday")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Next Monday:", dt.Format("2006-01-02"))
	}
}
```

**Business Value**: Enhances UX in customer-facing apps (e.g., booking forms, reminders).

#### Recurrence Rules (`recurrence.go`)
Supports complex event scheduling with iCalendar RRULE standards.

- **Functions**:
    - `ParseRRule(rrule string) (RecurrenceRule, error)`: Parses an RRule string.

- **Methods**:
    - `(r RecurrenceRule) GenerateOccurrences(start, end DateTime) ([]DateTime, error)`: Generates event occurrences.
    - `(r RecurrenceRule) NextOccurrence(after DateTime) (DateTime, bool)`: Finds the next occurrence.
    - `(r RecurrenceRule) HasOccurrence(dt DateTime) bool`: Checks if a `DateTime` is an occurrence.

**Usage**:
```go
package main

import (
	"fmt"
	"github.com/JamalYusuf/luxon"
)

func main() {
	rule, _ := luxon.ParseRRule("FREQ=WEEKLY;INTERVAL=2;BYDAY=MO")
	start := luxon.Now()
	end := start.PlusDays(30)
	occurrences, _ := rule.GenerateOccurrences(start, end)
	for _, occ := range occurrences {
		fmt.Println(occ.Format("2006-01-02"))
	}
}
```

**Business Value**: Powers calendar apps and automated reminders.

#### Testing Support (`timer.go`)
Facilitates testing with time simulation.

- **Types**:
    - `Timer`: Interface for time operations.
    - `DefaultTimer`: Uses real system time.
    - `MockTimer`: Simulates time for testing.

- **Functions**:
    - `NewDefaultTimer() *DefaultTimer`: Creates a default timer.
    - `NewMockTimer(t time.Time, loc *time.Location) *MockTimer`: Creates a mock timer.
    - `SetTimer(t Timer)`: Sets the global timer.
    - `ResetTimer()`: Resets to the default timer.

- **Methods** (for `DefaultTimer` and `MockTimer`):
    - `Now() DateTime`
    - `Sleep(d time.Duration)`
    - `Advance(d time.Duration) error`
    - `InZone(zone string) (DateTime, error)`
    - `ToTime() time.Time`

**Usage**:
```go
package main

import (
	"fmt"
	"time"
	"github.com/JamalYusuf/luxon"
)

func main() {
	// Simulate time for testing
	mock := luxon.NewMockTimer(time.Now(), time.UTC)
	luxon.SetTimer(mock)
	fmt.Println(luxon.Now().Format("2006-01-02"))
	mock.Advance(24 * time.Hour)
	fmt.Println(luxon.Now().Format("2006-01-02"))
	luxon.ResetTimer()
}
```

**Business Value**: Ensures reliable testing for time-sensitive features like expirations.

#### Batch Processing (`batch.go`)
Optimizes performance for large datasets.

- **Types**:
    - `DateTimePair`: Pairs `DateTime` instances for batch calculations.

- **Functions**:
    - `BatchInZone(ctx context.Context, dts []DateTime, zone string) ([]DateTime, error)`: Converts multiple `DateTime` instances to a time zone.
    - `BatchGenerateOccurrences(ctx context.Context, rules []RecurrenceRule, start, end DateTime) (map[int][]DateTime, error)`: Generates occurrences for multiple rules.
    - `BatchBusinessDaysBetween(ctx context.Context, cal *BusinessCalendar, pairs []DateTimePair) ([]int, error)`: Calculates business days for multiple pairs.
    - `BatchFormat(ctx context.Context, dts []DateTime, layout string) ([]string, error)`: Formats multiple `DateTime` instances.
    - `BatchHumanizeDurations(ctx context.Context, durs []Duration) ([]string, error)`: Humanizes multiple durations.

**Usage**:
```go
package main

import (
	"context"
	"fmt"
	"github.com/JamalYusuf/luxon"
)

func main() {
	ctx := context.Background()
	dts := []luxon.DateTime{luxon.Now(), luxon.Now().PlusDays(1)}
	formatted, err := luxon.BatchFormat(ctx, dts, "2006-01-02")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Formatted:", formatted)
	}
}
```

**Business Value**: Speeds up data processing in analytics and reporting systems.

## Contributing

Contributions are welcome! Please:
1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/YourFeature`).
3. Commit your changes (`git commit -m 'Add YourFeature'`).
4. Push to the branch (`git push origin feature/YourFeature`).
5. Open a pull request.

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## Testing

Run tests with:

```bash
go test ./... -v
```

The package includes unit tests for all components, leveraging `MockTimer` for time-sensitive tests.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
