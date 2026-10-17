package content

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrForbidden = errors.New("content access denied")

type FieldError struct{ Field, Message string }
type ValidationError struct{ Fields []FieldError }

func (e *ValidationError) Error() string { return "invalid content input" }
func ValidateText(title, body string) []FieldError {
	var errors []FieldError
	for _, field := range []struct {
		key, value, empty, short, long string
		min, max                       int
	}{
		{"title", title, "Bitte gib einen Titel an.", "Der Titel muss zwischen 3 und 255 Zeichen lang sein.", "Der Titel muss zwischen 3 und 255 Zeichen lang sein.", 3, 255},
		{"body", body, "Gib einen Beitragstext ein.", "Der Beitragstext muss mindestens 10 Zeichen lang sein.", "Der Beitragstext darf höchstens 10.000 Zeichen lang sein.", 10, 10000},
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
	return errors
}
func ValidateSchedule(date, clock string) []FieldError {
	var errors []FieldError
	if date != "" {
		if _, err := time.Parse("2.1.2006", strings.TrimSpace(date)); err != nil {
			errors = append(errors, FieldError{"date", "Gib ein gültiges Datum ein, zum Beispiel 30.01.2027."})
		}
	}
	if clock != "" {
		if _, err := time.Parse("15:04", strings.TrimSpace(clock)); err != nil {
			errors = append(errors, FieldError{"time", "Gib eine gültige Uhrzeit ein, zum Beispiel 17:00."})
		}
		if date == "" {
			errors = append(errors, FieldError{"date", "Gib auch das Datum der Veranstaltung ein."})
		}
	}
	return errors
}
