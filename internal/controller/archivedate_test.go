package controller

import (
	"testing"

	"github.com/fachschaftinformatik/web/internal/model/archive"
)

func TestArchiveDateValidatesCalendarAndRetainsInput(t *testing.T) {
	for _, test := range []struct{ day, month, year, want string }{
		{"27", "3", "2026", "2026-03-27"},
		{"27", "März", "2026", "2026-03-27"},
		{"27", "Mrz.", "2026", "2026-03-27"},
		{"29", "02", "2024", "2024-02-29"},
		{"29", "02", "2026", ""},
		{"31", "4", "2026", ""},
		{"27", "13", "2026", ""},
		{"27", "3", "26", ""},
		{"27", "", "2026", ""},
		{"", "", "", ""},
	} {
		t.Run(test.day+"/"+test.month+"/"+test.year, func(t *testing.T) {
			d := archive.DateInput{Day: test.day, Month: test.month, Year: test.year}
			got, message := d.Parse()
			if got != test.want || (message == "") != (test.want != "") {
				t.Fatalf("parse() = %q, %q", got, message)
			}
			if d.Day != test.day || d.Month != test.month || d.Year != test.year {
				t.Fatal("validation changed user input")
			}
			if test.month == "" && test.day != "" && (!d.Invalid["month"] || d.Invalid["day"] || d.Invalid["year"]) {
				t.Fatal("incomplete date did not isolate missing month")
			}
		})
	}
}
