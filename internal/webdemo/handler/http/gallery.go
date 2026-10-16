package web

import (
	"net/http"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/webdemo/gallery"
)

type GalleryPage struct {
	Albums                     []gallery.Album
	Album                      gallery.Album
	Images                     []gallery.Image
	Image                      gallery.Image
	Detail, Viewer             bool
	Number, Total, Page, Pages int
	PageNumbers                []int
	Previous, Next, Return     string
}

func (c *Handler) Gallery(w http.ResponseWriter, r *http.Request) {
	g := GalleryPage{Albums: gallery.Albums()}
	data := Page{Title: "Galerie", Active: "events", Community: true}
	if id := r.PathValue("album"); id != "" {
		album, found := gallery.Find(id)
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
			data.Title = g.Image.Title
		} else {
			g.Pages = max(1, (g.Total+gallery.PageSize-1)/gallery.PageSize)
			for number := 1; number <= g.Pages; number++ {
				g.PageNumbers = append(g.PageNumbers, number)
			}
			g.Page = 1
			if value := r.URL.Query().Get("page"); value != "" {
				page, err := strconv.Atoi(value)
				if err != nil || page < 1 || page > g.Pages {
					http.Redirect(w, r, album.URL(), http.StatusSeeOther)
					return
				}
				g.Page = page
			}
			start := (g.Page - 1) * gallery.PageSize
			g.Images = album.Images[start:min(start+gallery.PageSize, g.Total)]
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
	data.Gallery = g
	c.render(w, "gallery", false, data)
}
