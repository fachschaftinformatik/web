package view

import (
	"testing"
	"time"
)

func TestRelativeTime(t *testing.T) {
	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		age  time.Duration
		want string
	}{
		{-time.Hour, "gerade eben"},
		{59 * time.Second, "gerade eben"},
		{time.Minute, "vor einer Minute"},
		{2 * time.Minute, "vor 2 Minuten"},
		{time.Hour, "vor einer Stunde"},
		{2 * time.Hour, "vor 2 Stunden"},
		{24 * time.Hour, "vor einem Tag"},
		{2 * 24 * time.Hour, "vor 2 Tagen"},
		{30 * 24 * time.Hour, "vor einem Monat"},
		{90 * 24 * time.Hour, "vor 3 Monaten"},
		{365 * 24 * time.Hour, "vor einem Jahr"},
		{730 * 24 * time.Hour, "vor 2 Jahren"},
	} {
		if got := relativeTime(now.Add(-tc.age).Format(time.RFC3339), now); got != tc.want {
			t.Fatalf("%s: %s, want %s", tc.age, got, tc.want)
		}
	}
}
