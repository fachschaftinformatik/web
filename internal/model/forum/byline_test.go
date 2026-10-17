package forum

import "testing"

func TestPostedLabelUsesGermanLocalTime(t *testing.T) {
	for _, tc := range []struct{ created, want string }{
		{"2026-07-17T12:30:00Z", "14:30 Uhr am 17. Juli 2026"},
		{"2026-01-17T12:30:00Z", "13:30 Uhr am 17. Januar 2026"},
		{"2026-07-17T23:30:00Z", "01:30 Uhr am 18. Juli 2026"},
	} {
		if got := (Post{Created: tc.created}).PostedLabel(); got != tc.want {
			t.Fatalf("%s: %s", tc.created, got)
		}
	}
}
