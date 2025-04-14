// Package luxon provides advanced time handling utilities. The timer subpackage defines
// a Timer interface and implementations for controlling time in tests and simulations,
// supporting common business scenarios like fixed time points, time freezing, and stepping.
package luxon

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Timer defines an interface for abstracting time operations, enabling testing and simulation
// of time-sensitive logic. It supports retrieving the current time, sleeping, advancing time,
// managing time zones, and converting to time.Time, making it suitable for scenarios like
// scheduling, deadlines, and time zone conversions.
//
// Implementations must be safe for concurrent use unless explicitly documented otherwise.
type Timer interface {
	// Now returns the current DateTime according to the timer's state.
	Now() DateTime

	// Sleep pauses execution for the specified duration, respecting the timer's context.
	// In simulated environments, it may advance the timer's internal clock instead.
	Sleep(d time.Duration)

	// Advance moves the timer's clock forward by the given duration.
	// Returns an error if advancing is not supported (e.g., in a real-time timer).
	Advance(d time.Duration) error

	// InZone returns a DateTime in the specified time zone based on the timer's state.
	// Returns an error if the zone is invalid or unsupported.
	InZone(zone string) (DateTime, error)

	// ToTime returns the timer's current time as a time.Time.
	ToTime() time.Time
}

// DefaultTimer is the standard implementation of the Timer interface, using real system time.
// It relies on the time package for accurate timekeeping and is suitable for production use.
type DefaultTimer struct {
	// loc holds the current time zone, defaulting to time.Local.
	loc *time.Location
	// mu protects loc from concurrent modifications.
	mu sync.RWMutex
}

// NewDefaultTimer creates a new DefaultTimer with the local time zone.
func NewDefaultTimer() *DefaultTimer {
	return &DefaultTimer{
		loc: time.Local,
	}
}

// Now returns the current system time as a DateTime in the timer's time zone.
func (t *DefaultTimer) Now() DateTime {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return DateTime{time: time.Now().In(t.loc)}
}

// Sleep pauses execution for the specified duration using the system clock.
func (t *DefaultTimer) Sleep(d time.Duration) {
	time.Sleep(d)
}

// Advance returns an error, as DefaultTimer uses real time and cannot be advanced manually.
func (t *DefaultTimer) Advance(d time.Duration) error {
	return errors.New("cannot advance real-time clock")
}

// InZone returns a DateTime in the specified time zone based on the current system time.
func (t *DefaultTimer) InZone(zone string) (DateTime, error) {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return DateTime{}, fmt.Errorf("failed to load time zone %s: %w", zone, err)
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return DateTime{time: time.Now().In(loc)}, nil
}

// ToTime returns the current system time as a time.Time in the timer's time zone.
func (t *DefaultTimer) ToTime() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return time.Now().In(t.loc)
}

// MockTimer is a Timer implementation for testing, allowing full control over time.
// It supports freezing time, advancing it manually, and simulating time zone changes,
// covering scenarios like expiration checks, scheduling, and deadline testing.
//
// MockTimer is safe for concurrent use.
type MockTimer struct {
	// current holds the simulated time.
	current time.Time
	// loc holds the current time zone.
	loc *time.Location
	// mu protects current and loc from concurrent access.
	mu sync.RWMutex
}

// NewMockTimer creates a new MockTimer set to the specified time and time zone.
// If loc is nil, it defaults to UTC.
func NewMockTimer(t time.Time, loc *time.Location) *MockTimer {
	if loc == nil {
		loc = time.UTC
	}
	return &MockTimer{
		current: t,
		loc:     loc,
	}
}

// Now returns the simulated current time as a DateTime in the timer's time zone.
func (t *MockTimer) Now() DateTime {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return DateTime{time: t.current.In(t.loc)}
}

// Sleep advances the timer's internal clock by the specified duration, simulating a pause.
func (t *MockTimer) Sleep(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.current = t.current.Add(d)
}

// Advance moves the timer's clock forward by the given duration.
func (t *MockTimer) Advance(d time.Duration) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.current = t.current.Add(d)
	return nil
}

// InZone returns a DateTime in the specified time zone based on the simulated time.
func (t *MockTimer) InZone(zone string) (DateTime, error) {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return DateTime{}, fmt.Errorf("failed to load time zone %s: %w", zone, err)
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	return DateTime{time: t.current.In(loc)}, nil
}

// ToTime returns the simulated current time as a time.Time in the timer's time zone.
func (t *MockTimer) ToTime() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.current.In(t.loc)
}

// SetTimer sets the global timer for the luxon package, used by functions like Now.
// It is typically called during test setup to install a MockTimer or custom implementation.
// Reset to DefaultTimer after tests to avoid affecting production code.
func SetTimer(t Timer) {
	currentTimer = t
}

// ResetTimer resets the global timer to the default implementation.
func ResetTimer() {
	currentTimer = NewDefaultTimer()
}

// currentTimer is the package-level timer, defaulting to DefaultTimer.
var currentTimer Timer = NewDefaultTimer()
