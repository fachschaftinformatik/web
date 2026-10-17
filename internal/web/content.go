package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/fachschaftinformatik/web/internal/content"
)

type MeetingsPage struct {
	Year     int
	Years    []int
	Meetings []content.Meeting
	Next     *content.Meeting
}

func (c *Handler) Meetings(w http.ResponseWriter, r *http.Request) {
	all, err := c.content.AllWithError()
	if err != nil {
		c.serverError(w, err)
		return
	}
	var years []int
	var next *content.Meeting
	today := time.Now().Format("2006-01-02")
	for _, meeting := range all {
		if len(years) == 0 || years[len(years)-1] != meeting.Year {
			years = append(years, meeting.Year)
		}
		if meeting.Date >= today && (next == nil || meeting.Date < next.Date) {
			copy := meeting
			next = &copy
		}
	}
	year := years[0]
	if value := r.URL.Query().Get("jahr"); value != "" {
		var err error
		year, err = strconv.Atoi(value)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}
	var records []content.Meeting
	for _, meeting := range all {
		if meeting.Year == year {
			records = append(records, meeting)
		}
	}
	if len(records) == 0 {
		http.NotFound(w, r)
		return
	}
	c.render(w, "content", false, Page{
		Title: "Sitzungen und Protokolle", Active: "meetings", Fachschaft: true,
		Meetings: MeetingsPage{Year: year, Years: years, Meetings: records, Next: next},
	})
}
