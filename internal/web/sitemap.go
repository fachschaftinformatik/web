package web

import (
	"net/http"

	"github.com/fachschaftinformatik/web/internal/content/media"
)

func (c *Handler) Sitemap(w http.ResponseWriter, r *http.Request) {
	albums := media.Albums()
	if c.media != nil {
		var err error
		albums, err = c.media.Catalog(r.Context())
		if err != nil {
			c.serverError(w, err)
			return
		}
	}
	c.render(w, "sitemap", false, Page{Title: "Seitenübersicht", Active: "sitemap", Images: ImagesPage{Albums: albums}})
}
