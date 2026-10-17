package web

import (
	"github.com/fachschaftinformatik/web/internal/content/posts"
	"net/http"
)

const newsURL = "/community/neuigkeiten"

func (c *Handler) News(w http.ResponseWriter, r *http.Request) {
	page := ForumPage{Live: c.posts != nil}
	if c.posts != nil {
		var err error
		page.Posts, err = c.posts.Find(r.Context(), posts.Filter{Type: "news"}, forumUser(w))
		if err != nil {
			c.serverError(w, err)
			return
		}
		page.CanWrite = c.posts.CanPublish(r.Context(), forumUser(w)) == nil
	} else {
		page.Posts = posts.Find(posts.Filter{Type: "news"})
	}
	c.render(w, "news", false, Page{Title: "Neuigkeiten", Active: "news", Community: true, Forum: page})
}
