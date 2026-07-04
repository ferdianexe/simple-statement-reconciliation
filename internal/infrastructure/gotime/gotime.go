package gotime

import (
	// golang package
	"time"
)

//go:generate mockgen -source=./gotime.go -destination=./gotime_mock.go -package=gotime

// Default represents the default instance of goTime.
var Default = goTime{}

// goTime represents an empty struct wrapping time-related helpers. It
// doesn't wrap a swappable external resource the way goCsv wraps
// *csv.Reader, so it needs no inner Reader-style interface of its own -
// the aggregating infrastructure.Service is what gets mocked by callers.
type goTime struct{}

// TruncateToDay strips the time-of-day component. Useful when one side
// of a comparison carries a full timestamp and the other only carries
// a date, since both need to be compared at day granularity.
func (g goTime) TruncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// InRange reports whether d's date falls within [start, end] inclusive.
// start and end are expected to already be day-truncated by the caller.
func (g goTime) InRange(d, start, end time.Time) bool {
	d = g.TruncateToDay(d)
	return !d.Before(start) && !d.After(end)
}

// Now returns the current local time.
func (g goTime) Now() time.Time {
	return time.Now()
}
