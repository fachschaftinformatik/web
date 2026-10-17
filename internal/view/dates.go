package view

import (
	"github.com/fachschaftinformatik/web/internal/datetime"
	"time"
)

const displayLocale = "de-DE"

type dateStyle int

const (
	dateStyleDate dateStyle = iota
	dateStyleEvent
	dateStylePosted
)

// Locale controls formatting; the display timezone is configured independently.
// Add another locale's layouts here without changing persisted UTC timestamps.
var dateLayouts = map[string][3]string{
	"de-DE": {"02.01.2006", "02.01.2006, 15:04 Uhr", "15:04 Uhr am 02.01.2006"},
}

func formatDate(value, locale string, style dateStyle) string {
	layouts, ok := dateLayouts[locale]
	if !ok {
		return ""
	}
	date, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return date.In(datetime.Berlin).Format(layouts[style])
	}
	// A calendar date has no timezone and must not move to another day.
	date, err = time.Parse(time.DateOnly, value)
	if err != nil {
		return ""
	}
	return date.Format(layouts[dateStyleDate])
}

// ScheduleFields converts persisted UTC into the local event editor fields.
func ScheduleFields(value string) (string, string) {
	if date, err := time.Parse(time.RFC3339, value); err == nil {
		date = date.In(datetime.Berlin)
		return date.Format(dateLayouts[displayLocale][dateStyleDate]), date.Format("15:04")
	}
	return formatDate(value, displayLocale, dateStyleDate), ""
}
