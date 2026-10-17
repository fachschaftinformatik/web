package web

import (
	"errors"
	"net/http"

	"github.com/fachschaftinformatik/web/internal/content/posts"
	"github.com/fachschaftinformatik/web/internal/users"
)

func (c *Handler) PublicProfile(w http.ResponseWriter, r *http.Request) {
	if c.identity == nil || c.posts == nil {
		http.NotFound(w, r)
		return
	}
	profile, err := c.identity.PublicProfile(r.Context(), r.PathValue("id"))
	if errors.Is(err, users.ErrProfileNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		c.serverError(w, err)
		return
	}
	records, err := c.posts.Find(r.Context(), posts.Filter{AuthorID: profile.ID}, forumUser(w))
	if err != nil {
		c.serverError(w, err)
		return
	}
	c.render(w, "profile", false, Page{Title: profile.Name, Active: "forum", Community: true, Forum: ForumPage{Live: true, Posts: records}})
}
