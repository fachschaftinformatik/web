package web

import (
	"net/http"

	"github.com/fachschaftinformatik/web/internal/content/media"
)

func (c *Handler) Sitemap(w http.ResponseWriter, r *http.Request) {
	albums := media.Albums()
	total := len(albums)
	if c.media != nil {
		var err error
		albums, total, err = c.media.AlbumPage(r.Context(), requestedPage(r), 100)
		if err != nil {
			c.serverError(w, err)
			return
		}
	}
	c.render(w, "sitemap", false, Page{Title: "Seitenübersicht", Active: "sitemap", Images: ImagesPage{Albums: albums, Pagination: listPagination(r, total, 100)}})
}
