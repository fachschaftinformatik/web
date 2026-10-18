package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/model/media"
	"github.com/fachschaftinformatik/web/internal/view"
)

func (c *Handler) Media(w http.ResponseWriter, r *http.Request) {
	albums, total, err := c.media.AlbumPage(r.Context(), requestedPage(r), 12)
	if err != nil {
		c.serverError(w, err)
		return
	}
	g := view.MediaPage{Albums: view.Albums(albums), Pagination: listPagination(r, total, 12)}
	data := view.Page{Title: "Fotos", Active: "media", Community: true}
	if r.PathValue("image") == "" && r.PathValue("album") == "" && !c.validPageRange(w, r, total, 12) {
		return
	}
	if imageID := r.PathValue("image"); imageID != "" {
		albumID, err := c.media.ItemAlbum(r.Context(), imageID)
		if errors.Is(err, media.ErrNotFound) {
			c.notFound(w, r)
			return
		}
		if err != nil {
			c.serverError(w, err)
			return
		}
		r.SetPathValue("album", albumID)
	}
	if id := r.PathValue("album"); id != "" {
		a, err := c.media.Album(r.Context(), id)
		if errors.Is(err, media.ErrNotFound) {
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
			index, err := c.media.PhotoPosition(r.Context(), id, imageID)
			if errors.Is(err, media.ErrNotFound) {
				c.notFound(w, r)
				return
			}
			if err != nil {
				c.serverError(w, err)
				return
			}
			start := max(0, index-1)
			photos, err := c.media.Photos(r.Context(), a, start, 3)
			if err != nil {
				c.serverError(w, err)
				return
			}
			if index-start >= len(photos) {
				c.notFound(w, r)
				return
			}
			g.Image, g.Viewer, g.Number = view.NewImage(photos[index-start]), true, index+1
			g.Return = album.PageURL(index/media.PageSize+1) + "#image-" + imageID
			if index > 0 {
				g.Previous = album.ImageURL(view.NewImage(photos[0]))
			}
			if index+1 < g.Total && index-start+1 < len(photos) {
				g.Next = album.ImageURL(view.NewImage(photos[index-start+1]))
			}
			data.DocumentTitle = a.Title + " – Bild " + strconv.Itoa(g.Number) + " von " + strconv.Itoa(g.Total)
		} else {
			g.Pagination = listPagination(r, g.Total, media.PageSize)
			if !c.validPageRange(w, r, g.Total, media.PageSize) {
				return
			}
			g.Page, g.Pages = g.Pagination.Page, g.Pagination.Pages
			if g.Pages > 1 {
				data.DocumentTitle = a.Title + " – Seite " + strconv.Itoa(g.Page) + " von " + strconv.Itoa(g.Pages)
			}
			photos, err := c.media.Photos(r.Context(), a, (g.Page-1)*media.PageSize, media.PageSize)
			if err != nil {
				c.serverError(w, err)
				return
			}
			g.PageNumbers = view.MediaPageNumbers(g.Page, g.Pages)
			g.Images = view.Images(photos)
			g.Previous, g.Next = g.Pagination.PreviousURL(), g.Pagination.NextURL()
		}
	}
	data.Media = g
	c.render(w, "media", false, data)
}
