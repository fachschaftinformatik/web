package forum

import (
	_ "embed"
	"encoding/json"
	"slices"
	"strings"
	"time"
)

//go:embed threads.json
var threads []byte

type Comment struct {
	ID      string
	Author  string `json:"user_name"`
	Text    string
	Created string `json:"created_at"`
	Updated string `json:"updated_at"`
}

func (c Comment) DateLabel() string    { return dateLabel(c.Created) }
func (c Comment) Edited() bool         { return c.Updated != "" && c.Updated != c.Created }
func (c Comment) UpdatedLabel() string { return dateLabel(c.Updated) }

type Thread struct {
	Post
	Updated   string
	EventDate string `json:"event_date"`
	Location  string
	Replies   []Comment `json:"comments"`
}

func (t Thread) Edited() bool         { return t.Updated != "" && t.Updated != t.Created }
func (t Thread) UpdatedLabel() string { return dateLabel(t.Updated) }

func (t Thread) EventLabel() string {
	date, err := time.Parse("2006-01-02T15:04", t.EventDate)
	if err != nil {
		return ""
	}
	return dateLabel(date.Format(time.RFC3339)) + ", " + date.Format("15:04") + " Uhr"
}

// Lookup returns one complete public thread, with comments newest first.
func Lookup(id string) (Thread, bool) {
	var records []Thread
	if err := json.Unmarshal(threads, &records); err != nil {
		panic("invalid embedded forum threads: " + err.Error())
	}
	for _, post := range All() {
		if post.ID != id {
			continue
		}
		for _, thread := range records {
			if thread.ID == id {
				thread.Post = post
				slices.SortStableFunc(thread.Replies, func(a, b Comment) int { return strings.Compare(b.Created, a.Created) })
				return thread, true
			}
		}
	}
	return Thread{}, false
}
