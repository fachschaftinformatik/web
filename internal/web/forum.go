package web

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/fachschaftinformatik/web/internal/forum"
)

type ForumPage struct {
	Filter  forum.Filter
	Posts   []forum.Post
	Thread  *forum.Thread
	BackURL string
}

func (c *Handler) Discussion(w http.ResponseWriter, r *http.Request) {
	thread, ok := forum.Lookup(r.PathValue("post"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	values := url.Values{}
	if q := strings.TrimSpace(r.URL.Query().Get("q")); q != "" {
		values.Set("q", q)
	}
	switch kind := r.URL.Query().Get("type"); kind {
	case "discussion", "news", "event":
		values.Set("type", kind)
	}
	back := "/discussions"
	if len(values) > 0 {
		back += "?" + values.Encode()
	}
	c.render(w, "thread", false, Page{Title: thread.Title, Active: "forum", Community: true, Forum: ForumPage{Thread: &thread, BackURL: back}})
}

func (c *Handler) Forum(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := forum.Filter{Query: strings.TrimSpace(query.Get("q")), Type: query.Get("type")}
	switch filter.Type {
	case "discussion", "news", "event":
	default:
		filter.Type = ""
	}
	data := Page{
		Title: "Forum", Active: "forum", Community: true,
		Forum: ForumPage{Filter: filter, Posts: forum.Find(filter)},
	}
	if filter.Query != "" || filter.Type != "" {
		data.DocumentTitle = "Forum – Suchergebnisse"
	}
	c.render(w, "forum", false, data)
}
