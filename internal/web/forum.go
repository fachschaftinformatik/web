package web

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/fachschaftinformatik/web/internal/content/posts"
)

type ForumPage struct {
	Pagination              ListPagination
	Live, CanWrite, CanEdit bool
	Filter                  posts.Filter
	Posts                   []posts.Post
	Thread                  *posts.Thread
	BackURL                 string
	Comment                 CommentForm
}

type CommentForm struct {
	ID, Text, Error string
}

func (f CommentForm) Anchor() string {
	if f.ID != "" {
		return "edit-comment-" + f.ID
	}
	return "comment-text"
}

func (p ForumPage) CommentURL(id string) string {
	base := p.Thread.URL()
	if p.Pagination.Page > 1 {
		base = p.Pagination.PageURL(p.Pagination.Page)
	}
	return base + "#comment-" + id
}

func (c *Handler) Discussion(w http.ResponseWriter, r *http.Request) {
	c.discussion(w, r, CommentForm{})
}

func (c *Handler) discussion(w http.ResponseWriter, r *http.Request, draft CommentForm) {
	var thread posts.Thread
	var ok bool
	var total int
	if c.posts != nil {
		var err error
		if draft.ID != "" {
			page, err := c.posts.CommentPage(r.Context(), r.PathValue("post"), draft.ID)
			if err != nil {
				c.forumError(w, r, err)
				return
			}
			query := r.URL.Query()
			query.Set("page", strconv.Itoa(page))
			r = r.Clone(r.Context())
			urlCopy := *r.URL
			r.URL = &urlCopy
			r.URL.RawQuery = query.Encode()
		}
		thread, total, err = c.posts.LookupPage(r.Context(), r.PathValue("post"), requestedPage(r), forumUser(w))
		if err != nil {
			c.forumError(w, r, err)
			return
		}
		ok = true
	} else {
		thread, ok = posts.Lookup(r.PathValue("post"))
	}
	if !ok {
		c.notFound(w, r)
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
	forum := ForumPage{Thread: &thread, BackURL: back, Live: c.posts != nil, Pagination: listPagination(r, total, 20), Comment: draft}
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
	status := http.StatusOK
	if draft.Error != "" {
		status = http.StatusUnprocessableEntity
	}
	c.renderStatus(w, "thread", false, Page{Title: thread.Title, Active: active, Community: true, Forum: forum}, status)
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
		var total int
		data.Forum.Posts, total, err = c.posts.FindPage(r.Context(), filter, requestedPage(r), 20, forumUser(w))
		data.Forum.Pagination = listPagination(r, total, 20)
		if err != nil {
			c.serverError(w, err)
			return
		}
		data.Forum.Live = true
	} else {
		data.Forum.Posts = posts.Find(filter)
		data.Forum.Pagination = listPagination(r, len(data.Forum.Posts), 20)
		start := (data.Forum.Pagination.Page - 1) * 20
		data.Forum.Posts = data.Forum.Posts[start:min(start+20, len(data.Forum.Posts))]
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
		c.notFound(w, r)
	case errors.Is(err, posts.ErrForbidden):
		c.pageError(w, "Du hast keine Berechtigung für diese Aktion.", http.StatusForbidden)
	default:
		c.serverError(w, err)
	}
}

func (c *Handler) Comment(w http.ResponseWriter, r *http.Request) {
	if c.posts == nil {
		c.notFound(w, r)
		return
	}
	user := forumUser(w)
	if user == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 128<<10)
	if err := r.ParseForm(); err != nil {
		c.pageError(w, "Das Formular konnte nicht gelesen werden.", http.StatusBadRequest)
		return
	}
	comment, err := c.posts.SaveComment(r.Context(), user, r.PathValue("post"), r.PathValue("comment"), r.PostForm.Get("text"))
	if err != nil {
		var validation *posts.ValidationError
		if errors.As(err, &validation) {
			c.discussion(w, r, CommentForm{ID: r.PathValue("comment"), Text: r.PostForm.Get("text"), Error: validation.Fields[0].Message})
			return
		}
		c.forumError(w, r, err)
		return
	}
	page, err := c.posts.CommentPage(r.Context(), r.PathValue("post"), comment)
	if err != nil {
		c.forumError(w, r, err)
		return
	}
	back := "/d/" + r.PathValue("post")
	if page > 1 {
		back += "?page=" + strconv.Itoa(page)
	}
	http.Redirect(w, r, back+"#comment-"+comment, http.StatusSeeOther)
}
