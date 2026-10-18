package view

import (
	"fmt"
	"time"
)

func postedLabel(value string) string { return formatDate(value, displayLocale, dateStylePosted) }
func dateLabel(value string) string   { return formatDate(value, displayLocale, dateStyleDate) }
func eventLabel(value string) string  { return formatDate(value, displayLocale, dateStyleEvent) }
func parseDate(value string) (time.Time, error) {
	var err error
	for _, layout := range []string{"2006-01-02", "02.01.2006", time.RFC3339} {
		var date time.Time
		date, err = time.Parse(layout, value)
		if err == nil {
			return date, nil
		}
	}
	return time.Time{}, err
}
func relativeTime(value string, now time.Time) string {
	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return dateLabel(value)
	}
	age := now.Sub(date)
	if age < time.Minute {
		return "gerade eben"
	}
	for _, unit := range []struct {
		duration         time.Duration
		singular, plural string
	}{
		{365 * 24 * time.Hour, "einem Jahr", "Jahren"},
		{30 * 24 * time.Hour, "einem Monat", "Monaten"},
		{24 * time.Hour, "einem Tag", "Tagen"},
		{time.Hour, "einer Stunde", "Stunden"},
		{time.Minute, "einer Minute", "Minuten"},
	} {
		if count := int(age / unit.duration); count > 0 {
			if count == 1 {
				return "vor " + unit.singular
			}
			return fmt.Sprintf("vor %d %s", count, unit.plural)
		}
	}
	return "gerade eben"
}
