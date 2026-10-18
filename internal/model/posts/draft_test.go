package posts

import (
	"strings"
	"testing"
)

func TestDraftValidation(t *testing.T) {
	valid := Draft{Title: "Frage zum Studium", Body: "Wie läuft das Kolloquium ab?", Type: "discussion"}
	if errors := valid.Validate(); len(errors) != 0 {
		t.Fatal(errors)
	}
	for _, tc := range []struct {
		name, field string
		draft       Draft
	}{
		{"missing title", "title", Draft{Body: valid.Body, Type: valid.Type}},
		{"short body", "body", Draft{Title: valid.Title, Body: "kurz", Type: valid.Type}},
		{"long title", "title", Draft{Title: strings.Repeat("ü", 256), Body: valid.Body, Type: valid.Type}},
		{"invalid type", "type", Draft{Title: valid.Title, Body: valid.Body, Type: "other"}},
		{"invalid date", "date", Draft{Title: valid.Title, Body: valid.Body, Type: "event", Date: "31.02.2027"}},
		{"invalid time", "time", Draft{Title: valid.Title, Body: valid.Body, Type: "event", Date: "30.01.2027", Time: "25:00"}},
		{"time without date", "date", Draft{Title: valid.Title, Body: valid.Body, Type: "event", Time: "17:00"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			found := false
			for _, err := range tc.draft.Validate() {
				if err.Field == tc.field {
					found = true
				}
			}
			if !found {
				t.Fatal(tc.draft.Validate())
			}
		})
	}
}
