package posts

import (
	"fmt"
	"time"
)

func (p Post) RelativeTime() string    { return relativeTime(p.Created, time.Now()) }
func (c Comment) RelativeTime() string { return relativeTime(c.Created, time.Now()) }
func (c Comment) PostedLabel() string  { return postedLabel(c.Created) }
func (t Thread) EditedLabel() string   { return postedLabel(t.Updated) }
func (c Comment) EditedLabel() string  { return postedLabel(c.Updated) }

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
