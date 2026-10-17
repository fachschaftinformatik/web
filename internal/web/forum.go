package web

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/fachschaftinformatik/web/internal/content/posts"
)

type ForumPage struct {
	Live, CanWrite, CanEdit bool
	Filter                  posts.Filter
	Posts                   []posts.Post
	Thread                  *posts.Thread
	BackURL                 string
}

func (c *Handler) Discussion(w http.ResponseWriter, r *http.Request) {
	var thread posts.Thread
	var ok bool
	if c.posts != nil {
		var err error
		thread, err = c.posts.Lookup(r.Context(), r.PathValue("post"), forumUser(w))
		if err != nil {
			c.forumError(w, r, err)
			return
		}
		ok = true
	} else {
		thread, ok = posts.Lookup(r.PathValue("post"))
	}
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
	forum := ForumPage{Thread: &thread, BackURL: back, Live: c.posts != nil}
	if c.posts != nil {
		if user := forumUser(w); user != 0 {
			forum.CanWrite = c.posts.CanWrite(r.Context(), user, 0, false) == nil
			forum.CanEdit = c.posts.CanWrite(r.Context(), user, thread.AuthorID, false) == nil
			for i := range thread.Replies {
				thread.Replies[i].CanEdit = c.posts.CanWrite(r.Context(), user, thread.Replies[i].AuthorID, true) == nil
			}
		}
	}
	active := "forum"
	if thread.Type == "event" {
		active = "calendar"
		forum.BackURL = eventsURL
	}
	if thread.Type == "news" {
		active = "news"
		forum.BackURL = newsURL
	}
	c.render(w, "thread", false, Page{Title: thread.Title, Active: active, Community: true, Forum: forum})
}

func (c *Handler) Forum(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if query.Get("type") == "news" {
		http.Redirect(w, r, newsURL, http.StatusSeeOther)
		return
	}
	if query.Get("type") == "event" {
		http.Redirect(w, r, eventsURL, http.StatusSeeOther)
		return
	}
	filter := posts.Filter{Query: strings.TrimSpace(query.Get("q")), Type: "discussion"}
	data := Page{
		Title: "Forum", Active: "forum", Community: true,
		Forum: ForumPage{Filter: filter},
	}
	if c.posts != nil {
		var err error
		data.Forum.Posts, err = c.posts.Find(r.Context(), filter, forumUser(w))
		if err != nil {
			c.serverError(w, err)
			return
		}
		data.Forum.Live = true
	} else {
		data.Forum.Posts = posts.Find(filter)
	}
	if filter.Query != "" {
		data.DocumentTitle = "Forum – Suchergebnisse"
	}
	c.render(w, "forum", false, data)
}

func forumUser(w http.ResponseWriter) int64 {
	if response, ok := w.(*sessionResponse); ok && response.session != nil && response.session.Authenticated {
		return response.session.UserID
	}
	return 0
}

func (c *Handler) forumError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, posts.ErrNotFound):
		http.NotFound(w, r)
	case errors.Is(err, posts.ErrForbidden):
		http.Error(w, "Du hast keine Berechtigung für diese Aktion.", http.StatusForbidden)
	default:
		c.serverError(w, err)
	}
}

func (c *Handler) Comment(w http.ResponseWriter, r *http.Request) {
	if c.posts == nil {
		http.NotFound(w, r)
		return
	}
	user := forumUser(w)
	if user == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 128<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Das Formular konnte nicht gelesen werden.", http.StatusBadRequest)
		return
	}
	comment, err := c.posts.SaveComment(r.Context(), user, r.PathValue("post"), r.PathValue("comment"), r.PostForm.Get("text"))
	if err != nil {
		var validation *posts.ValidationError
		if errors.As(err, &validation) {
			http.Error(w, validation.Fields[0].Message, http.StatusUnprocessableEntity)
			return
		}
		c.forumError(w, r, err)
		return
	}
	http.Redirect(w, r, "/d/"+r.PathValue("post")+"#comment-"+comment, http.StatusSeeOther)
}
