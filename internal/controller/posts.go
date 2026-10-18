package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/fachschaftinformatik/web/internal/content"

	"github.com/fachschaftinformatik/web/internal/model/posts"
	"github.com/fachschaftinformatik/web/internal/urls"
	"github.com/fachschaftinformatik/web/internal/view"
)

func (c *Handler) EditDiscussion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	user := postsUser(w)

	if user == 0 {
		http.Redirect(w, r, urls.LoginPath, http.StatusSeeOther)
		return
	}
	if err := c.posts.CanWrite(r.Context(), user, 0, false); err != nil {
		c.postsError(w, r, err)
		return
	}

	editor := view.ContentEditor{Action: urls.CreatePostPath, BackURL: urls.PostsPath, Draft: view.EditorDraft{Type: "discussion"}}

	title := "Beitrag verfassen"
	if r.PathValue("post") == "" && r.URL.Query().Get("type") == "news" {
		editor.Announcement = true
		editor.Draft.Type = "news"
		editor.Action = urls.CreatePostPath + "?type=news"
		editor.BackURL = urls.NewsPath
		title = "Ankündigung veröffentlichen"
	}
	if id := r.PathValue("post"); id != "" {
		thread, err := c.posts.Lookup(r.Context(), id)
		if err != nil {
			c.postsError(w, r, err)
			return
		}
		if err = c.posts.CanWrite(r.Context(), user, thread.AuthorID, false); err != nil {
			c.postsError(w, r, err)
			return
		}

		editor.Editing = true
		if thread.Type == "event" {
			if event, err := c.events.ForPost(r.Context(), id); err == nil {
				http.Redirect(w, r, urls.EditEvent(event.ID), http.StatusSeeOther)
				return
			}
		}
		editor.Announcement = thread.Type == "news"
		editor.Event = thread.Type == "event"
		editor.Action, editor.BackURL = urls.EditPost(thread.ID), urls.Post(thread.ID)
		editor.Draft = view.EditorDraft{Title: thread.Title, Body: thread.Body, Type: thread.Type, Tags: strings.Join(thread.Tags, ", "), Location: thread.Location}
		if thread.BodyFormat == "html" {
			editor.RichBody = content.SanitizeRichText(thread.Body)
			editor.Draft.Body = content.RichTextPlain(editor.RichBody)
		}
		editor.Draft.Date, editor.Draft.Time = view.ScheduleFields(thread.EventDate)
		title = "Beitrag bearbeiten"
		if editor.Event {
			title = "Veranstaltung bearbeiten"
		}
		if editor.Announcement {
			title = "Ankündigung bearbeiten"
		}
	}
	if editor.Event || editor.Announcement {
		if err := c.posts.CanPublish(r.Context(), user); err != nil {
			c.postsError(w, r, err)
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
		tags := r.PostForm.Get("tags")
		if editor.Announcement {
			tags = editor.Draft.Tags
		}
		editor.Draft = view.EditorDraft{Title: r.PostForm.Get("title"), Body: r.PostForm.Get("body"), Type: editor.Draft.Type, Tags: tags, Date: r.PostForm.Get("date"), Time: r.PostForm.Get("time"), Location: r.PostForm.Get("location")}
		if editor.Event {
			editor.Draft.Date = eventFormDate(r)
		}
		if r.PostForm.Has("rich_body") {
			editor.RichBody = content.SanitizeRichText(r.PostForm.Get("rich_body"))
			editor.Draft.Body = content.RichTextPlain(editor.RichBody)
		} else {
			editor.RichBody = content.PlainRichText(editor.Draft.Body)
		}
		validation := posts.Draft(editor.Draft)
		if r.PostForm.Has("rich_body") {
			validation.Body = content.RichTextVisible(editor.RichBody)
		}
		editor.Errors = validation.Validate()
		if len(editor.Errors) > 0 {
			status = http.StatusUnprocessableEntity
		} else {
			post, err := c.posts.Save(r.Context(), user, r.PathValue("post"), posts.Draft(editor.Draft), editor.RichBody)
			if err == nil {
				http.Redirect(w, r, urls.Post(post), http.StatusSeeOther)
				return
			}
			var validation *posts.ValidationError
			if errors.As(err, &validation) {
				editor.Errors = validation.Fields
				status = http.StatusUnprocessableEntity
			} else {
				c.postsError(w, r, err)
				return
			}
		}

	}
	data := view.Page{Title: title, Active: "posts", Community: true, Editor: &editor}
	if editor.Event {
		data.Active = "events"
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
		page, err := c.posts.CommentPage(r.Context(), r.PathValue("post"), draft.ID)
		if err != nil {
			c.postsError(w, r, err)
			return
		}
		query := r.URL.Query()
		query.Set("page", strconv.Itoa(page))
		r = r.Clone(r.Context())
		urlCopy := *r.URL
		r.URL = &urlCopy
		r.URL.RawQuery = query.Encode()
	}
	thread, total, err := c.posts.LookupPage(r.Context(), r.PathValue("post"), requestedPage(r), postsUser(w))
	if err != nil {
		c.postsError(w, r, err)
		return
	}
	if !c.validPageRange(w, r, total, 20) {
		return
	}
	values := url.Values{}
	if q := strings.TrimSpace(r.URL.Query().Get("search")); q != "" {
		values.Set("search", q)
	}
	back := urls.PostsPath
	if len(values) > 0 {
		back += "?" + values.Encode()
	}
	page := view.PostsPage{Thread: ptr(view.NewThread(thread)), BackURL: back, Pagination: listPagination(r, total, 20, "search"), Comment: draft}

	if user := postsUser(w); user != 0 {
		page.CanWrite = c.posts.CanWrite(r.Context(), user, 0, false) == nil
		page.CanEdit = c.posts.CanWrite(r.Context(), user, thread.AuthorID, false) == nil
		for i := range thread.Replies {
			page.Thread.Replies[i].CanEdit = c.posts.CanWrite(r.Context(), user, thread.Replies[i].AuthorID, true) == nil
		}
	}

	active := "posts"
	if thread.Type == "event" {
		active = "events"
		page.BackURL = urls.EventsPath
	}
	if thread.Type == "news" {
		active = "news"
		page.BackURL = urls.NewsPath
	}
	status := http.StatusOK
	if draft.Error != "" {
		status = http.StatusUnprocessableEntity
	}
	c.renderStatus(w, "thread", false, view.Page{Title: thread.Title, Active: active, Community: true, Posts: page}, status)
}

func (c *Handler) Posts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := posts.Filter{Query: strings.TrimSpace(query.Get("search")), Type: "discussion"}
	data := view.Page{
		Title: "Forum", Active: "posts", Community: true,
		Posts: view.PostsPage{Filter: filter},
	}

	records, total, err := c.posts.FindPage(r.Context(), filter, requestedPage(r), 20, postsUser(w))
	data.Posts.Posts = view.Posts(records)
	data.Posts.Pagination = listPagination(r, total, 20, "search")
	if err != nil {
		c.serverError(w, err)
		return
	}

	if !c.validPageRange(w, r, total, 20) {
		return
	}
	if filter.Query != "" {
		data.DocumentTitle = "Forum – Suchergebnisse"
	}
	if data.Posts.Pagination.Pages > 1 {
		title := data.DocumentTitle
		if title == "" {
			title = data.Title
		}
		data.DocumentTitle = title + " – Seite " + strconv.Itoa(data.Posts.Pagination.Page) + " von " + strconv.Itoa(data.Posts.Pagination.Pages)
	}
	c.render(w, "posts", false, data)
}

func postsUser(w http.ResponseWriter) int64 {
	if response, ok := w.(*sessionResponse); ok && response.session != nil && response.session.Authenticated {
		return response.session.UserID
	}
	return 0
}

func (c *Handler) postsError(w http.ResponseWriter, r *http.Request, err error) {
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

	user := postsUser(w)
	if user == 0 {
		http.Redirect(w, r, urls.LoginPath, http.StatusSeeOther)
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
			c.discussion(w, r, view.CommentForm{ID: r.PathValue("comment"), Text: r.PostForm.Get("text"), Error: validation.Fields[0].Message})
			return
		}
		c.postsError(w, r, err)
		return
	}
	page, err := c.posts.CommentPage(r.Context(), r.PathValue("post"), comment)
	if err != nil {
		c.postsError(w, r, err)
		return
	}
	back := urls.Post(r.PathValue("post"))
	if page > 1 {
		back += "?page=" + strconv.Itoa(page)
	}
	http.Redirect(w, r, back+"#comment-"+comment, http.StatusSeeOther)
}

func (c *Handler) News(w http.ResponseWriter, r *http.Request) {
	filter := posts.Filter{Type: "news", Query: strings.TrimSpace(r.URL.Query().Get("search"))}
	page := view.PostsPage{Filter: filter}

	records, total, err := c.posts.FindPage(r.Context(), filter, requestedPage(r), 20, postsUser(w))
	page.Posts = view.Posts(records)
	page.Pagination = listPagination(r, total, 20, "search")
	if err != nil {
		c.serverError(w, err)
		return
	}
	page.CanWrite = c.posts.CanPublish(r.Context(), postsUser(w)) == nil
	if !c.validPageRange(w, r, total, 20) {
		return
	}

	data := view.Page{Title: "Ankündigungen", Active: "news", Community: true, Posts: page}
	if filter.Query != "" {
		data.DocumentTitle = "Ankündigungen – Suchergebnisse"
	}
	if page.Pagination.Pages > 1 {
		title := data.DocumentTitle
		if title == "" {
			title = data.Title
		}
		data.DocumentTitle = title + " – Seite " + strconv.Itoa(page.Pagination.Page) + " von " + strconv.Itoa(page.Pagination.Pages)
	}
	c.render(w, "news", false, data)
}

func (c *Handler) Vote(w http.ResponseWriter, r *http.Request) {

	user := postsUser(w)
	if user == 0 {
		http.Redirect(w, r, urls.LoginPath, http.StatusSeeOther)
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
	if err := c.posts.Vote(r.Context(), user, post, comment, value); err != nil {
		var validation *posts.ValidationError
		if errors.As(err, &validation) {
			c.pageError(w, validation.Fields[0].Message, http.StatusUnprocessableEntity)
			return
		}
		c.postsError(w, r, err)
		return
	}
	if r.Header.Get("Accept") == "application/json" {
		page := 1
		if comment != "" {
			var err error
			page, err = c.posts.CommentPage(r.Context(), post, comment)
			if err != nil {
				c.postsError(w, r, err)
				return
			}
		}
		thread, _, err := c.posts.LookupPage(r.Context(), post, page, user)
		if err != nil {
			c.postsError(w, r, err)
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
	back := urls.Post(post)
	if comment != "" {
		page, err := c.posts.CommentPage(r.Context(), post, comment)
		if err != nil {
			c.postsError(w, r, err)
			return
		}
		if page > 1 {
			back += "?page=" + strconv.Itoa(page)
		}
		back += "#comment-" + comment
	}
	// Only local landing/post destinations are accepted after voting.
	if u, err := url.Parse(r.PostForm.Get("return")); err == nil && u.Scheme == "" && u.Host == "" && (u.Path == urls.HomePath || u.Path == urls.PostsPath || u.Path == urls.Post(post)) {
		back = u.String()
	}
	http.Redirect(w, r, back, http.StatusSeeOther)
}
