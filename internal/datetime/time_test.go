package datetime

import (
	"testing"
	"time"
)

func TestFormatUTCAlwaysUsesZ(t *testing.T) {
	for _, offset := range []int{0, 2 * 60 * 60, -8 * 60 * 60, 5*60*60 + 30*60} {
		input := time.Date(2026, 10, 17, 12, 34, 56, 0, time.FixedZone("input", offset))
		stored := FormatUTC(input)
		if stored[len(stored)-1] != 'Z' {
			t.Fatalf("stored offset instead of Z: %s", stored)
		}
		parsed, err := time.Parse(time.RFC3339, stored)
		if err != nil || !parsed.Equal(input) || parsed.Location() != time.UTC {
			t.Fatalf("did not preserve UTC instant: %s %v", stored, err)
		}
	}
}

func TestLocalTimestampStoresUTCAndRejectsDSTAmbiguity(t *testing.T) {
	for _, test := range []struct{ input, want string }{
		{"2026-01-17T17:00", "2026-01-17T16:00:00Z"},
		{"2026-07-17T17:00", "2026-07-17T15:00:00Z"},
		{"2026-07-18T00:30", "2026-07-17T22:30:00Z"},
		{"2026-03-29T02:30", ""},
		{"2026-10-25T02:30", ""},
		{"2026-02-30T12:00", ""},
	} {
		t.Run(test.input, func(t *testing.T) {
			got, err := LocalTimestamp(test.input)
			if got != test.want || (err == nil) != (test.want != "") {
				t.Fatalf("got %q, %v; want %q", got, err, test.want)
			}
		})
	}
}
