package web

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/fachschaftinformatik/web/internal/content/media"
	"github.com/fachschaftinformatik/web/internal/data/blob"
)

type ImagesPage struct {
	Pagination                 ListPagination
	Albums                     []media.Album
	Album                      media.Album
	Images                     []media.Image
	Image                      media.Image
	Detail, Viewer             bool
	Number, Total, Page, Pages int
	PageNumbers                []int
	Previous, Next, Return     string
}

func (g ImagesPage) ImageNumber(index int) int {
	return (g.Page-1)*media.PageSize + index + 1
}

func (g ImagesPage) CarouselAlbums() []media.Album {
	var albums []media.Album
	for _, album := range g.Albums {
		if album.CoverImage().Source != "" {
			albums = append(albums, album)
		}
	}
	return albums
}

func (c *Handler) Images(w http.ResponseWriter, r *http.Request) {
	if c.media != nil {
		c.pagedImages(w, r)
		return
	}
	g := ImagesPage{Albums: media.Albums()}
	data := Page{Title: "Fotos", Active: "events", Community: true}
	if id := r.PathValue("album"); id != "" {
		var album media.Album
		found := false
		for _, candidate := range g.Albums {
			if candidate.ID == id {
				album, found = candidate, true
				break
			}
		}
		if !found {
			http.NotFound(w, r)
			return
		}
		g.Album, g.Detail, g.Total = album, true, len(album.Images)
		data.Title = album.Title
		if imageID := r.PathValue("image"); imageID != "" {
			index := -1
			for i, image := range album.Images {
				if image.ID == imageID {
					index = i
					break
				}
			}
			if index < 0 {
				http.NotFound(w, r)
				return
			}
			g.Image, g.Viewer, g.Number = album.Images[index], true, index+1
			g.Return = album.ReturnURL(index)
			if index > 0 {
				g.Previous = album.ImageURL(album.Images[index-1])
			}
			if index+1 < len(album.Images) {
				g.Next = album.ImageURL(album.Images[index+1])
			}
			data.DocumentTitle = album.Title + " – Bild " + strconv.Itoa(g.Number) + " von " + strconv.Itoa(g.Total)
		} else {
			g.Pages = max(1, (g.Total+media.PageSize-1)/media.PageSize)
			g.Page = 1
			if value := r.URL.Query().Get("page"); value != "" {
				page, err := strconv.Atoi(value)
				if err != nil || page < 1 || page > g.Pages {
					http.Redirect(w, r, album.URL(), http.StatusSeeOther)
					return
				}
				g.Page = page
			}
			g.PageNumbers = galleryPageNumbers(g.Page, g.Pages)
			start := (g.Page - 1) * media.PageSize
			g.Images = album.Images[start:min(start+media.PageSize, g.Total)]
			if g.Pages > 1 {
				data.DocumentTitle = data.Title + " – Seite " + strconv.Itoa(g.Page) + " von " + strconv.Itoa(g.Pages)
			}
			if g.Page > 1 {
				g.Previous = album.PageURL(g.Page - 1)
			}
			if g.Page < g.Pages {
				g.Next = album.PageURL(g.Page + 1)
			}
		}
	}
	data.Images = g
	c.render(w, "images", false, data)
}

func (c *Handler) pagedImages(w http.ResponseWriter, r *http.Request) {
	albums, total, err := c.media.AlbumPage(r.Context(), requestedPage(r), 12)
	if err != nil {
		c.serverError(w, err)
		return
	}
	g := ImagesPage{Albums: albums, Pagination: listPagination(r, total, 12)}
	data := Page{Title: "Fotos", Active: "events", Community: true}
	if id := r.PathValue("album"); id != "" {
		a, err := c.media.Album(r.Context(), id)
		if errors.Is(err, media.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			c.serverError(w, err)
			return
		}
		g.Album, g.Detail, g.Total = a, true, a.OriginalCount
		data.Title = a.Title
		if imageID := r.PathValue("image"); imageID != "" {
			index, err := c.media.PhotoPosition(r.Context(), id, imageID)
			if errors.Is(err, media.ErrNotFound) {
				http.NotFound(w, r)
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
				http.NotFound(w, r)
				return
			}
			g.Image, g.Viewer, g.Number = photos[index-start], true, index+1
			g.Return = a.PageURL(index/media.PageSize+1) + "#image-" + imageID
			if index > 0 {
				g.Previous = a.ImageURL(photos[0])
			}
			if index+1 < g.Total && index-start+1 < len(photos) {
				g.Next = a.ImageURL(photos[index-start+1])
			}
			data.DocumentTitle = a.Title + " – Bild " + strconv.Itoa(g.Number) + " von " + strconv.Itoa(g.Total)
		} else {
			g.Pagination = listPagination(r, g.Total, media.PageSize)
			if raw := r.URL.Query().Get("page"); raw != "" {
				n, err := strconv.Atoi(raw)
				if err != nil || n < 1 || n > g.Pagination.Pages {
					http.Redirect(w, r, a.URL(), http.StatusSeeOther)
					return
				}
			}
			g.Page, g.Pages = g.Pagination.Page, g.Pagination.Pages
			if g.Pages > 1 {
				data.DocumentTitle = a.Title + " – Seite " + strconv.Itoa(g.Page) + " von " + strconv.Itoa(g.Pages)
			}
			g.Images, err = c.media.Photos(r.Context(), a, (g.Page-1)*media.PageSize, media.PageSize)
			if err != nil {
				c.serverError(w, err)
				return
			}
			g.PageNumbers = galleryPageNumbers(g.Page, g.Pages)
			g.Previous, g.Next = g.Pagination.PreviousURL(), g.Pagination.NextURL()
		}
	}
	data.Images = g
	c.render(w, "images", false, data)
}

// Zero marks an ellipsis between the current neighbourhood and an end page.
func galleryPageNumbers(page, pages int) []int {
	var numbers []int
	for number := 1; number <= pages; number++ {
		if pages <= 7 || number == 1 || number == pages || (number >= page-1 && number <= page+1) || (page <= 4 && number < page) || (page >= pages-3 && number > page) {
			if len(numbers) > 0 && number-numbers[len(numbers)-1] > 1 {
				numbers = append(numbers, 0)
			}
			numbers = append(numbers, number)
		}
	}
	return numbers
}

// MediaFile serves only database-backed gallery images, never raw storage keys.
func (c *Handler) MediaFile(w http.ResponseWriter, r *http.Request) {
	c.galleryImageFile(w, r, false)
}

func (c *Handler) EventCover(w http.ResponseWriter, r *http.Request) {
	c.galleryImageFile(w, r, true)
}

func (c *Handler) galleryImageFile(w http.ResponseWriter, r *http.Request, cover bool) {
	w.Header().Set("Cache-Control", "no-store")
	if c.media == nil {
		http.NotFound(w, r)
		return
	}
	open := c.media.Open
	if cover {
		open = c.media.OpenCover
	}
	object, err := open(r.Context(), r.PathValue("id"), r.PathValue("size"))
	if errors.Is(err, media.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, media.ErrStorage) {
		http.Error(w, "Das Bild konnte nicht geladen werden.", http.StatusServiceUnavailable)
		return
	}
	if err != nil {
		c.serverError(w, err)
		return
	}
	c.serveGalleryObject(w, r, object)
}

func (c *Handler) serveGalleryObject(w http.ResponseWriter, r *http.Request, object *blob.Object) {
	defer object.Close()
	seeker, ok := object.ReadCloser.(io.ReadSeeker)
	if !ok {
		// Providers may return streaming bodies. Spool those to a bounded
		// temporary file so ServeContent still supports HEAD and byte ranges.
		file, err := os.CreateTemp("", "gallery")
		if err != nil {
			c.serverError(w, err)
			return
		}
		defer os.Remove(file.Name())
		defer file.Close()
		written, err := io.Copy(file, io.LimitReader(object, object.Info.Size+1))
		if err != nil || written != object.Info.Size {
			http.Error(w, "Das Bild konnte nicht geladen werden.", http.StatusServiceUnavailable)
			return
		}
		seeker = file
	}
	w.Header().Set("Content-Type", object.Info.ContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Revalidate rather than retaining deleted photographs in a shared cache.
	w.Header().Set("Cache-Control", "public, no-cache")
	http.ServeContent(w, r, "image", time.Time{}, seeker)
}
