package posts

import (
	"strings"
	"time"
	"unicode/utf8"
)

type Draft struct {
	Title, Body, Type, Tags, Date, Time, Location string
}

type FieldError struct{ Field, Message string }

// Validate uses the existing application's title and body length policy.
func (d Draft) Validate() []FieldError {
	var errors []FieldError
	for _, field := range []struct {
		key, value, empty, short, long string
		min, max                       int
	}{
		{"title", d.Title, "Bitte gib einen Titel an.", "Der Titel muss zwischen 3 und 255 Zeichen lang sein.", "Der Titel muss zwischen 3 und 255 Zeichen lang sein.", 3, 255},
		{"body", d.Body, "Gib einen Beitragstext ein.", "Der Beitragstext muss mindestens 10 Zeichen lang sein.", "Der Beitragstext darf höchstens 10.000 Zeichen lang sein.", 10, 10000},
	} {
		length := utf8.RuneCountInString(strings.TrimSpace(field.value))
		message := ""
		switch {
		case length == 0:
			message = field.empty
		case length < field.min:
			message = field.short
		case length > field.max:
			message = field.long
		}
		if message != "" {
			errors = append(errors, FieldError{field.key, message})
		}
	}
	switch d.Type {
	case "discussion", "news", "event":
	default:
		errors = append(errors, FieldError{"type", "Wähle eine Beitragsart aus."})
	}
	if d.Type == "event" {
		if d.Date != "" {
			if _, err := time.Parse("2.1.2006", strings.TrimSpace(d.Date)); err != nil {
				errors = append(errors, FieldError{"date", "Gib ein gültiges Datum ein, zum Beispiel 30.01.2027."})
			}
		}
		if d.Time != "" {
			if _, err := time.Parse("15:04", strings.TrimSpace(d.Time)); err != nil {
				errors = append(errors, FieldError{"time", "Gib eine gültige Uhrzeit ein, zum Beispiel 17:00."})
			}
			if d.Date == "" {
				errors = append(errors, FieldError{"date", "Gib auch das Datum der Veranstaltung ein."})
			}
		}
	}
	return errors
}
