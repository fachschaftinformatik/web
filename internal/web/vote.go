package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/content/posts"
)

type VoteControl struct {
	Page                  Page
	Action, Return, Label string
	Score, Selected       int
	Up, Down              int
}

func voteControl(page Page, action, back, label string, score, selected int) VoteControl {
	v := VoteControl{Page: page, Action: action, Return: back, Label: label, Score: score, Selected: selected, Up: 1, Down: -1}
	if selected == 1 {
		v.Up = 0
	}
	if selected == -1 {
		v.Down = 0
	}
	return v
}

func (c *Handler) Vote(w http.ResponseWriter, r *http.Request) {
	if c.posts == nil {
		http.NotFound(w, r)
		return
	}
	user := forumUser(w)
	if user == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Das Formular konnte nicht gelesen werden.", http.StatusBadRequest)
		return
	}
	value, err := strconv.Atoi(r.PostForm.Get("vote"))
	if err != nil {
		http.Error(w, "Ungültige Bewertung.", http.StatusUnprocessableEntity)
		return
	}
	post, comment := r.PathValue("post"), r.PathValue("comment")
	if err := c.posts.Vote(r.Context(), user, post, comment, value); err != nil {
		var validation *posts.ValidationError
		if errors.As(err, &validation) {
			http.Error(w, validation.Fields[0].Message, http.StatusUnprocessableEntity)
			return
		}
		c.forumError(w, r, err)
		return
	}
	if r.Header.Get("Accept") == "application/json" {
		thread, err := c.posts.Lookup(r.Context(), post, user)
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
		back += "#comment-" + comment
	}
	// Only local landing/forum/thread destinations are accepted after voting.
	if u, err := url.Parse(r.PostForm.Get("return")); err == nil && u.Scheme == "" && u.Host == "" && (u.Path == "/" || u.Path == "/discussions" || u.Path == "/d/"+post) {
		back = u.String()
	}
	http.Redirect(w, r, back, http.StatusSeeOther)
}
