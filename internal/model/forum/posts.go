package forum

import (
	_ "embed"
	"fmt"
	"time"
)

type Post struct {
	AuthorID                                        int64
	AuthorPublic                                    bool
	ID, Title, Body, Summary, Author, Created, Type string
	BodyFormat                                      string
	Comments                                        int
	Votes, UserVote                                 int
	Tags                                            []string
}

func (p Post) URL() string { return "/d/" + p.ID }

func (p Post) AuthorURL() string {
	if p.AuthorPublic && p.AuthorID > 0 {
		return fmt.Sprintf("/users/%d", p.AuthorID)
	}
	return ""
}

func (p Post) PostedLabel() string { return postedLabel(p.Created) }

func postedLabel(value string) string {
	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return dateLabel(value)
	}
	if location, err := time.LoadLocation("Europe/Berlin"); err == nil {
		date = date.In(location)
	}
	return date.Format("15:04") + " Uhr am " + dateLabel(date.Format(time.RFC3339))
}

func (p Post) SourceURL() string { return "https://ik.fsv-wh.de/d/" + p.ID }

func (p Post) DateISO() string {
	if len(p.Created) < 10 {
		return ""
	}
	return p.Created[:10]
}

func (p Post) DateLabel() string {
	return dateLabel(p.Created)
}

func dateLabel(value string) string {
	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return ""
	}
	months := [...]string{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"}
	return date.Format("2.") + " " + months[date.Month()-1] + " " + date.Format("2006")
}

func (p Post) TypeLabel() string {
	switch p.Type {
	case "news":
		return "Neuigkeit"
	case "event":
		return "Veranstaltung"
	default:
		return "Diskussion"
	}
}

type Filter struct {
	Query, Type string
	AuthorID    int64
}

type Comment struct {
	AuthorPublic    bool
	Votes, UserVote int
	AuthorID        int64
	CanEdit         bool
	ID              string
	Author          string `json:"user_name"`
	Text            string
	Created         string `json:"created_at"`
	Updated         string `json:"updated_at"`
}

func (c Comment) DateLabel() string { return dateLabel(c.Created) }

func (c Comment) AuthorURL() string {
	return (Post{AuthorID: c.AuthorID, AuthorPublic: c.AuthorPublic}).AuthorURL()
}

func (c Comment) Edited() bool { return c.Updated != "" && c.Updated != c.Created }

func (c Comment) UpdatedLabel() string { return dateLabel(c.Updated) }

type Thread struct {
	Post
	Updated   string
	EventDate string `json:"event_date"`
	Location  string
	Replies   []Comment `json:"comments"`
}

func (t Thread) Edited() bool { return t.Updated != "" && t.Updated != t.Created }

func (t Thread) UpdatedLabel() string { return dateLabel(t.Updated) }

func (t Thread) EventLabel() string {
	date, err := time.Parse("2006-01-02T15:04", t.EventDate)
	if err != nil {
		return ""
	}
	return dateLabel(date.Format(time.RFC3339)) + ", " + date.Format("15:04") + " Uhr"
}

func (p Post) RelativeTime() string { return relativeTime(p.Created, time.Now()) }

func (c Comment) RelativeTime() string { return relativeTime(c.Created, time.Now()) }

func (c Comment) PostedLabel() string { return postedLabel(c.Created) }

func (t Thread) EditedLabel() string { return postedLabel(t.Updated) }

func (c Comment) EditedLabel() string { return postedLabel(c.Updated) }

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
