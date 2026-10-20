package site

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/urls"
	"github.com/fachschaftinformatik/web/model/media"
	"github.com/fachschaftinformatik/web/view"
)

func (c *Handler) Media(w http.ResponseWriter, r *http.Request) {
	albums, total, err := c.media.AlbumPage(r.Context(), requestedPage(r), 12)
	if err != nil {
		c.serverError(w, err)
		return
	}
	g := view.MediaPage{Albums: mapRecords(albums, albumView), Pagination: listPagination(r, total, 12), PageSize: media.PageSize}
	g.CanManage = c.media.CanManage(r.Context(), postsUser(w)) == nil
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
		album := albumView(a)
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
			g.Image, g.Viewer, g.Number = imageView(photos[index-start]), true, index+1
			g.Return = album.PageURL(index/media.PageSize+1) + "#image-" + imageID
			if index > 0 {
				g.Previous = album.ImageURL(imageView(photos[0]))
			}
			if index+1 < g.Total && index-start+1 < len(photos) {
				g.Next = album.ImageURL(imageView(photos[index-start+1]))
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
			g.Images = mapRecords(photos, imageView)
			g.Previous, g.Next = g.Pagination.PreviousURL(), g.Pagination.NextURL()
		}
	}
	data.Media = g
	c.render(w, "media", false, data)
}

func (c *Handler) ManageMedia(w http.ResponseWriter, r *http.Request) {
	items, total, err := c.media.GalleryPage(r.Context(), postsUser(w), requestedPage(r), 12)
	if err != nil {
		c.mediaManagementError(w, r, err)
		return
	}
	if !c.validPageRange(w, r, total, 12) {
		return
	}
	g := view.MediaPage{Managing: true, Saved: r.URL.Query().Get("saved") == "1", Pagination: listPagination(r, total, 12)}
	for _, item := range items {
		g.GalleryItems = append(g.GalleryItems, view.GalleryItem{Image: imageView(item.Image), AlbumTitle: item.AlbumTitle, Visible: item.Visible})
	}
	c.render(w, "media", false, view.Page{Title: "Startseitenbilder auswählen", Active: "media", Community: true, Media: g})
}

func (c *Handler) MediaVisibility(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		c.pageError(w, "Das Formular konnte nicht gelesen werden.", http.StatusBadRequest)
		return
	}
	page, err := strconv.Atoi(r.PostForm.Get("page"))
	value := r.PostForm.Get("homepage_visible")
	if err != nil || page < 1 || (value != "" && value != "1") {
		c.pageError(w, "Prüfe die Auswahl für die Startseite.", http.StatusBadRequest)
		return
	}
	if err := c.media.SetHomepageVisibility(r.Context(), postsUser(w), r.PathValue("image"), value == "1"); err != nil {
		c.mediaManagementError(w, r, err)
		return
	}
	http.Redirect(w, r, urls.ManageMediaPath+"?page="+strconv.Itoa(page)+"&saved=1", http.StatusSeeOther)
}

func (c *Handler) mediaManagementError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, media.ErrNotFound) {
		c.notFound(w, r)
	} else {
		c.postsError(w, r, err)
	}
}
