// Package datetime defines timestamp storage and the website's input/display zone.
package datetime

import (
	"errors"
	"time"
	_ "time/tzdata"
)

var Berlin = func() *time.Location {
	zone, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}
	return zone
}()

// FormatUTC serializes database timestamps at second precision. Converting to
// UTC before formatting guarantees RFC 3339's explicit Z suffix, never an offset.
func FormatUTC(value time.Time) string { return value.UTC().Format(time.RFC3339) }

// LocalTimestamp converts an entered wall-clock time to canonical UTC. Reject
// skipped and repeated DST times rather than silently choosing another instant.
func LocalTimestamp(value string) (string, error) {
	const layout = "2006-01-02T15:04"
	date, err := time.ParseInLocation(layout, value, Berlin)
	if err != nil {
		return "", err
	}
	if date.Format(layout) != value {
		return "", errors.New("local time does not exist")
	}
	if date.Add(time.Hour).Format(layout) == value || date.Add(-time.Hour).Format(layout) == value {
		return "", errors.New("local time is ambiguous")
	}
	return FormatUTC(date), nil
}
