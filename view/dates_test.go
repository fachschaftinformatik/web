package view

import "testing"

func TestUTCDatePresentation(t *testing.T) {
	for _, test := range []struct {
		input string
		style dateStyle
		want  string
	}{
		{"2026-01-17T23:30:00Z", dateStyleEvent, "18.01.2026, 00:30 Uhr"},
		{"2026-07-17T22:30:00Z", dateStylePosted, "00:30 Uhr am 18.07.2026"},
		{"2026-07-17T22:30:00Z", dateStyleDate, "18.07.2026"},
		{"2026-07-17", dateStyleEvent, "17.07.2026"},
		{"invalid", dateStyleDate, ""},
	} {
		if got := formatDate(test.input, "de-DE", test.style); got != test.want {
			t.Errorf("%s: got %q, want %q", test.input, got, test.want)
		}
	}
	date, clock := ScheduleFields("2026-07-17T22:30:00Z")
	if date != "18.07.2026" || clock != "00:30" {
		t.Fatal("UTC schedule did not return local editor values", date, clock)
	}
}
