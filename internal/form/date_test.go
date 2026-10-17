package form

import (
	"testing"
	"time"
)

func TestDateInputCalendarBoundaries(t *testing.T) {
	for _, test := range []struct{ day, month, year, want string }{
		{"29", "2", "2000", "2000-02-29"},
		{"29", "2", "1900", ""},
		{"29", "Februar", "2024", "2024-02-29"},
		{"29", "Feb.", "2023", ""},
		{"31", "4", "2026", ""},
		{"0", "3", "2026", ""},
		{"1", "13", "2026", ""},
		{"1", "1", "26", ""},
		{"1", "1", "0000", ""},
		{"1", "1", "0001", "0001-01-01"},
		{"31", "12", "9999", "9999-12-31"},
	} {
		t.Run(test.day+"/"+test.month+"/"+test.year, func(t *testing.T) {
			input := DateInput{Day: test.day, Month: test.month, Year: test.year}
			got, message := input.Parse("Datum")
			if got != test.want || (message == "") != (test.want != "") {
				t.Fatalf("Parse() = %q, %q; want %q", got, message, test.want)
			}
			if got != "" {
				date, err := time.Parse(time.DateOnly, got)
				if err != nil || date.Format(time.DateOnly) != got {
					t.Fatalf("non-canonical stored date %q: %v", got, err)
				}
			}
		})
	}
}
