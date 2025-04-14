// Package luxon provides advanced time handling utilities. The batch subpackage
// contains functions for efficient batch processing of time-related operations,
// optimized for performance in high-throughput scenarios such as bulk time zone
// conversions, recurrence evaluations, and business day calculations.
package luxon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// BatchInZone converts a slice of DateTime instances to a specified time zone in parallel.
// It uses a cached time zone location to minimize disk I/O and supports cancellation
// via context. Returns an error if any conversion fails.
//
// Example:
//
//	ctx := context.Background()
//	dts := []DateTime{dt1, dt2}
//	converted, err := BatchInZone(ctx, dts, "America/New_York")
//	if err != nil {
//	    log.Fatal(err)
//	}
func BatchInZone(ctx context.Context, dts []DateTime, zone string) ([]DateTime, error) {
	if len(dts) == 0 {
		return nil, nil
	}
	result := make([]DateTime, len(dts))
	g, ctx := errgroup.WithContext(ctx)

	// Load or cache time zone once
	loc, err := loadCachedLocation(zone)
	if err != nil {
		return nil, fmt.Errorf("failed to load time zone %s: %w", zone, err)
	}

	for i, dt := range dts {
		i, dt := i, dt // Capture loop variables
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				result[i] = DateTime{time: dt.time.In(loc)}
				return nil
			}
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

// BatchGenerateOccurrences evaluates multiple RecurrenceRule instances in parallel,
// generating occurrences within a specified time range. It supports cancellation
// via context and returns a map of rule indices to their occurrences.
//
// Example:
//
//	ctx := context.Background()
//	rules := []RecurrenceRule{rule1, rule2}
//	start, end := luxon.Now(), luxon.Now().Plus(30*24*time.Hour)
//	results, err := BatchGenerateOccurrences(ctx, rules, start, end)
//	if err != nil {
//	    log.Fatal(err)
//	}
func BatchGenerateOccurrences(ctx context.Context, rules []RecurrenceRule, start, end DateTime) (map[int][]DateTime, error) {
	if len(rules) == 0 {
		return nil, nil
	}
	results := make(map[int][]DateTime, len(rules))
	var mu sync.Mutex
	g, ctx := errgroup.WithContext(ctx)

	for i, rule := range rules {
		i, rule := i, rule // Capture loop variables
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				occs, err := rule.GenerateOccurrences(start, end)
				if err != nil {
					return fmt.Errorf("failed to generate occurrences for rule %d: %w", i, err)
				}
				mu.Lock()
				results[i] = occs
				mu.Unlock()
				return nil
			}
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

// BatchBusinessDaysBetween calculates the number of business days between pairs of DateTime
// instances using a BusinessCalendar, processing in parallel. It supports cancellation via
// context and returns a slice of counts matching the input pairs.
//
// Example:
//
//	ctx := context.Background()
//	cal := &BusinessCalendar{...}
//	pairs := []DateTimePair{{Start: dt1, End: dt2}, {Start: dt3, End: dt4}}
//	counts, err := BatchBusinessDaysBetween(ctx, cal, pairs)
//	if err != nil {
//	    log.Fatal(err)
//	}
func BatchBusinessDaysBetween(ctx context.Context, cal *BusinessCalendar, pairs []DateTimePair) ([]int, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	if cal == nil {
		return nil, errors.New("business calendar cannot be nil")
	}
	results := make([]int, len(pairs))
	g, ctx := errgroup.WithContext(ctx)

	for i, pair := range pairs {
		i, pair := i, pair // Capture loop variables
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if pair.Start.time.After(pair.End.time) {
					return fmt.Errorf("start time after end time for pair %d", i)
				}
				count := cal.BusinessDaysBetween(pair.Start, pair.End)
				results[i] = count
				return nil
			}
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

// BatchFormat formats a slice of DateTime instances using a specified layout in parallel.
// It supports cancellation via context and returns a slice of formatted strings.
//
// Example:
//
//	ctx := context.Background()
//	dts := []DateTime{dt1, dt2}
//	formatted, err := BatchFormat(ctx, dts, time.RFC3339)
//	if err != nil {
//	    log.Fatal(err)
//	}
func BatchFormat(ctx context.Context, dts []DateTime, layout string) ([]string, error) {
	if len(dts) == 0 {
		return nil, nil
	}
	results := make([]string, len(dts))
	g, ctx := errgroup.WithContext(ctx)

	for i, dt := range dts {
		i, dt := i, dt // Capture loop variables
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				results[i] = dt.Format(layout)
				return nil
			}
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

// BatchHumanizeDurations converts a slice of Duration instances to human-readable strings
// in parallel. It supports cancellation via context and returns a slice of formatted strings.
//
// Example:
//
//	ctx := context.Background()
//	durs := []Duration{dur1, dur2}
//	formatted, err := BatchHumanizeDurations(ctx, durs)
//	if err != nil {
//	    log.Fatal(err)
//	}
func BatchHumanizeDurations(ctx context.Context, durs []Duration) ([]string, error) {
	if len(durs) == 0 {
		return nil, nil
	}
	results := make([]string, len(durs))
	g, ctx := errgroup.WithContext(ctx)

	for i, dur := range durs {
		i, dur := i, dur // Capture loop variables
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				results[i] = dur.Humanize()
				return nil
			}
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

// DateTimePair represents a pair of DateTime instances for operations like business day calculations.
type DateTimePair struct {
	Start, End DateTime
}

// loadCachedLocation retrieves a time.Location from the cache or loads it, storing it if necessary.
// It assumes timeZoneCache is a package-level sync.Map.
func loadCachedLocation(zone string) (*time.Location, error) {
	if zone == "" {
		return nil, errors.New("time zone cannot be empty")
	}
	loc, ok := timeZoneCache.Load(zone)
	if !ok {
		var err error
		loc, err = time.LoadLocation(zone)
		if err != nil {
			return nil, err
		}
		timeZoneCache.Store(zone, loc)
	}
	return loc.(*time.Location), nil
}
