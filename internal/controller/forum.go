package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/content"

	"github.com/fachschaftinformatik/web/internal/model/forum"
	"github.com/fachschaftinformatik/web/internal/view"
)

func (c *Handler) EditDiscussion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	user := forumUser(w)

	if user == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := c.forum.CanWrite(r.Context(), user, 0, false); err != nil {
		c.forumError(w, r, err)
		return
	}

	editor := view.ContentEditor{Action: "/d/new", BackURL: "/discussions", SourceURL: "https://ik.fsv-wh.de/d/new", Draft: view.EditorDraft{Type: "discussion"}}

	title := "Beitrag verfassen"
	if r.URL.Path == view.NewsURL+"/new" {
		editor.Announcement = true
		editor.Draft.Type = "news"
		editor.Action = view.NewsURL + "/new"
		editor.BackURL = view.NewsURL
		title = "Neuigkeit veröffentlichen"
	}
	if id := r.PathValue("post"); id != "" {
		thread, err := c.forum.Lookup(r.Context(), id)
		if err != nil {
			c.forumError(w, r, err)
			return
		}
		if err = c.forum.CanWrite(r.Context(), user, thread.AuthorID, false); err != nil {
			c.forumError(w, r, err)
			return
		}

		editor.Editing = true
		if thread.Type == "event" {
			if event, err := c.events.Lookup(r.Context(), id); err == nil {
				http.Redirect(w, r, view.Event(event).URL()+"/edit", http.StatusSeeOther)
				return
			}
		}
		editor.Announcement = thread.Type == "news"
		editor.Event = thread.Type == "event"
		editor.Action, editor.BackURL, editor.SourceURL = view.NewThread(thread).URL()+"/edit", view.NewThread(thread).URL(), view.NewThread(thread).SourceURL()+"/edit"
		editor.Draft = view.EditorDraft{Title: thread.Title, Body: thread.Body, Type: thread.Type, Tags: strings.Join(thread.Tags, ", "), Location: thread.Location}
		if thread.BodyFormat == "html" {
			editor.RichBody = content.SanitizeRichText(thread.Body)
			editor.Draft.Body = content.RichTextPlain(editor.RichBody)
		}
		if date, err := time.Parse("2006-01-02T15:04", thread.EventDate); err == nil {
			editor.Draft.Date, editor.Draft.Time = date.Format("02.01.2006"), date.Format("15:04")
		}
		title = "Beitrag bearbeiten"
		if editor.Event {
			title = "Veranstaltung bearbeiten"
		}
		if editor.Announcement {
			title = "Neuigkeit bearbeiten"
		}
	}
	if editor.Event || editor.Announcement {
		if err := c.forum.CanPublish(r.Context(), user); err != nil {
			c.forumError(w, r, err)
			return
		}
	}
	status := http.StatusOK
	if editor.RichBody == "" {
		editor.RichBody = content.PlainRichText(editor.Draft.Body)
	}
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, 128<<10)
		parseErr := r.ParseForm()
		if parseErr != nil {
			c.pageError(w, "Das Formular konnte nicht gelesen werden. Versuche es mit einem kürzeren Beitrag erneut.", http.StatusBadRequest)
			return
		}
		editor.Draft = view.EditorDraft{Title: r.PostForm.Get("title"), Body: r.PostForm.Get("body"), Type: editor.Draft.Type, Tags: r.PostForm.Get("tags"), Date: r.PostForm.Get("date"), Time: r.PostForm.Get("time"), Location: r.PostForm.Get("location")}
		if r.PostForm.Has("rich_body") {
			editor.RichBody = content.SanitizeRichText(r.PostForm.Get("rich_body"))
			editor.Draft.Body = content.RichTextPlain(editor.RichBody)
		} else {
			editor.RichBody = content.PlainRichText(editor.Draft.Body)
		}
		validation := forum.Draft(editor.Draft)
		if r.PostForm.Has("rich_body") {
			validation.Body = content.RichTextVisible(editor.RichBody)
		}
		editor.Errors = validation.Validate()
		if len(editor.Errors) > 0 {
			status = http.StatusUnprocessableEntity
		} else {
			post, err := c.forum.Save(r.Context(), user, r.PathValue("post"), forum.Draft(editor.Draft), editor.RichBody)
			if err == nil {
				http.Redirect(w, r, "/d/"+post, http.StatusSeeOther)
				return
			}
			var validation *forum.ValidationError
			if errors.As(err, &validation) {
				editor.Errors = validation.Fields
				status = http.StatusUnprocessableEntity
			} else {
				c.forumError(w, r, err)
				return
			}
		}

	}
	data := view.Page{Title: title, Active: "forum", Community: true, Editor: &editor}
	if editor.Event {
		data.Active = "calendar"
	}
	if editor.Announcement {
		data.Active = "news"
	}
	if len(editor.Errors) > 0 {
		data.DocumentTitle = "Fehler: " + title
	}
	c.renderStatus(w, "editor", false, data, status)
}

func (c *Handler) Discussion(w http.ResponseWriter, r *http.Request) {
	c.discussion(w, r, view.CommentForm{})
}

func (c *Handler) discussion(w http.ResponseWriter, r *http.Request, draft view.CommentForm) {
	if draft.ID != "" {
		page, err := c.forum.CommentPage(r.Context(), r.PathValue("post"), draft.ID)
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
	thread, total, err := c.forum.LookupPage(r.Context(), r.PathValue("post"), requestedPage(r), forumUser(w))
	if err != nil {
		c.forumError(w, r, err)
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
	forum := view.ForumPage{Thread: ptr(view.NewThread(thread)), BackURL: back, Pagination: listPagination(r, total, 20), Comment: draft}

	if user := forumUser(w); user != 0 {
		forum.CanWrite = c.forum.CanWrite(r.Context(), user, 0, false) == nil
		forum.CanEdit = c.forum.CanWrite(r.Context(), user, thread.AuthorID, false) == nil
		for i := range thread.Replies {
			forum.Thread.Replies[i].CanEdit = c.forum.CanWrite(r.Context(), user, thread.Replies[i].AuthorID, true) == nil
		}
	}

	active := "forum"
	if thread.Type == "event" {
		active = "calendar"
		forum.BackURL = view.EventsURL
	}
	if thread.Type == "news" {
		active = "news"
		forum.BackURL = view.NewsURL
	}
	status := http.StatusOK
	if draft.Error != "" {
		status = http.StatusUnprocessableEntity
	}
	c.renderStatus(w, "thread", false, view.Page{Title: thread.Title, Active: active, Community: true, Forum: forum}, status)
}

func (c *Handler) Forum(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if query.Get("type") == "news" {
		http.Redirect(w, r, view.NewsURL, http.StatusSeeOther)
		return
	}
	if query.Get("type") == "event" {
		http.Redirect(w, r, view.EventsURL, http.StatusSeeOther)
		return
	}
	filter := forum.Filter{Query: strings.TrimSpace(query.Get("q")), Type: "discussion"}
	data := view.Page{
		Title: "Forum", Active: "forum", Community: true,
		Forum: view.ForumPage{Filter: filter},
	}

	records, total, err := c.forum.FindPage(r.Context(), filter, requestedPage(r), 20, forumUser(w))
	data.Forum.Posts = view.Posts(records)
	data.Forum.Pagination = listPagination(r, total, 20)
	if err != nil {
		c.serverError(w, err)
		return
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
	case errors.Is(err, forum.ErrNotFound):
		c.notFound(w, r)
	case errors.Is(err, forum.ErrForbidden):
		c.pageError(w, "Du hast keine Berechtigung für diese Aktion.", http.StatusForbidden)
	default:
		c.serverError(w, err)
	}
}

func (c *Handler) Comment(w http.ResponseWriter, r *http.Request) {

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
	comment, err := c.forum.SaveComment(r.Context(), user, r.PathValue("post"), r.PathValue("comment"), r.PostForm.Get("text"))
	if err != nil {
		var validation *forum.ValidationError
		if errors.As(err, &validation) {
			c.discussion(w, r, view.CommentForm{ID: r.PathValue("comment"), Text: r.PostForm.Get("text"), Error: validation.Fields[0].Message})
			return
		}
		c.forumError(w, r, err)
		return
	}
	page, err := c.forum.CommentPage(r.Context(), r.PathValue("post"), comment)
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

func (c *Handler) News(w http.ResponseWriter, r *http.Request) {
	page := view.ForumPage{}

	records, total, err := c.forum.FindPage(r.Context(), forum.Filter{Type: "news"}, requestedPage(r), 20, forumUser(w))
	page.Posts = view.Posts(records)
	page.Pagination = listPagination(r, total, 20)
	if err != nil {
		c.serverError(w, err)
		return
	}
	page.CanWrite = c.forum.CanPublish(r.Context(), forumUser(w)) == nil

	c.render(w, "news", false, view.Page{Title: "Neuigkeiten", Active: "news", Community: true, Forum: page})
}

func (c *Handler) Vote(w http.ResponseWriter, r *http.Request) {

	user := forumUser(w)
	if user == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if err := r.ParseForm(); err != nil {
		c.pageError(w, "Das Formular konnte nicht gelesen werden.", http.StatusBadRequest)
		return
	}
	value, err := strconv.Atoi(r.PostForm.Get("vote"))
	if err != nil {
		c.pageError(w, "Ungültige Bewertung.", http.StatusUnprocessableEntity)
		return
	}
	post, comment := r.PathValue("post"), r.PathValue("comment")
	if err := c.forum.Vote(r.Context(), user, post, comment, value); err != nil {
		var validation *forum.ValidationError
		if errors.As(err, &validation) {
			c.pageError(w, validation.Fields[0].Message, http.StatusUnprocessableEntity)
			return
		}
		c.forumError(w, r, err)
		return
	}
	if r.Header.Get("Accept") == "application/json" {
		page := 1
		if comment != "" {
			var err error
			page, err = c.forum.CommentPage(r.Context(), post, comment)
			if err != nil {
				c.forumError(w, r, err)
				return
			}
		}
		thread, _, err := c.forum.LookupPage(r.Context(), post, page, user)
		if err != nil {
			c.forumError(w, r, err)
			return
		}
		score, selected := thread.Votes, thread.UserVote
		if comment != "" {
			for _, reply := range thread.Replies {
				if reply.ID == comment {
					score, selected = reply.Votes, reply.UserVote
					break
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(struct {
			Score    int `json:"score"`
			Selected int `json:"selected"`
		}{score, selected})
		return
	}
	back := "/d/" + post
	if comment != "" {
		page, err := c.forum.CommentPage(r.Context(), post, comment)
		if err != nil {
			c.forumError(w, r, err)
			return
		}
		if page > 1 {
			back += "?page=" + strconv.Itoa(page)
		}
		back += "#comment-" + comment
	}
	// Only local landing/forum/thread destinations are accepted after voting.
	if u, err := url.Parse(r.PostForm.Get("return")); err == nil && u.Scheme == "" && u.Host == "" && (u.Path == "/" || u.Path == "/discussions" || u.Path == "/d/"+post) {
		back = u.String()
	}
	http.Redirect(w, r, back, http.StatusSeeOther)
}
