package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/model/gallery"
	"github.com/fachschaftinformatik/web/internal/view"
)

func (c *Handler) Images(w http.ResponseWriter, r *http.Request) {
	albums, total, err := c.gallery.AlbumPage(r.Context(), requestedPage(r), 12)
	if err != nil {
		c.serverError(w, err)
		return
	}
	g := view.ImagesPage{Albums: view.Albums(albums), Pagination: listPagination(r, total, 12)}
	data := view.Page{Title: "Fotos", Active: "events", Community: true}
	if id := r.PathValue("album"); id != "" {
		a, err := c.gallery.Album(r.Context(), id)
		if errors.Is(err, gallery.ErrNotFound) {
			c.notFound(w, r)
			return
		}
		if err != nil {
			c.serverError(w, err)
			return
		}
		album := view.NewAlbum(a)
		g.Album, g.Detail, g.Total = album, true, a.OriginalCount
		data.Title = a.Title
		if imageID := r.PathValue("image"); imageID != "" {
			index, err := c.gallery.PhotoPosition(r.Context(), id, imageID)
			if errors.Is(err, gallery.ErrNotFound) {
				c.notFound(w, r)
				return
			}
			if err != nil {
				c.serverError(w, err)
				return
			}
			start := max(0, index-1)
			photos, err := c.gallery.Photos(r.Context(), a, start, 3)
			if err != nil {
				c.serverError(w, err)
				return
			}
			if index-start >= len(photos) {
				c.notFound(w, r)
				return
			}
			g.Image, g.Viewer, g.Number = view.NewImage(photos[index-start]), true, index+1
			g.Return = album.PageURL(index/gallery.PageSize+1) + "#image-" + imageID
			if index > 0 {
				g.Previous = album.ImageURL(view.NewImage(photos[0]))
			}
			if index+1 < g.Total && index-start+1 < len(photos) {
				g.Next = album.ImageURL(view.NewImage(photos[index-start+1]))
			}
			data.DocumentTitle = a.Title + " – Bild " + strconv.Itoa(g.Number) + " von " + strconv.Itoa(g.Total)
		} else {
			g.Pagination = listPagination(r, g.Total, gallery.PageSize)
			if raw := r.URL.Query().Get("page"); raw != "" {
				n, err := strconv.Atoi(raw)
				if err != nil || n < 1 || n > g.Pagination.Pages {
					http.Redirect(w, r, album.URL(), http.StatusSeeOther)
					return
				}
			}
			g.Page, g.Pages = g.Pagination.Page, g.Pagination.Pages
			if g.Pages > 1 {
				data.DocumentTitle = a.Title + " – Seite " + strconv.Itoa(g.Page) + " von " + strconv.Itoa(g.Pages)
			}
			photos, err := c.gallery.Photos(r.Context(), a, (g.Page-1)*gallery.PageSize, gallery.PageSize)
			if err != nil {
				c.serverError(w, err)
				return
			}
			g.PageNumbers = view.GalleryPageNumbers(g.Page, g.Pages)
			g.Images = view.Images(photos)
			g.Previous, g.Next = g.Pagination.PreviousURL(), g.Pagination.NextURL()
		}
	}
	data.Images = g
	c.render(w, "images", false, data)
}
