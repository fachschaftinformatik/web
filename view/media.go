package view

import (
	"fmt"
	"github.com/fachschaftinformatik/web/internal/urls"
)

type Image struct {
	ID, Title, Description string
	Width, Height          int
	AlbumCover             bool
}
type Album struct {
	ID, Title     string
	OriginalCount int
	Images        []Image
	Cover         Image
	PageSize      int
}

func (i Image) Thumbnail() string { return i.previewURL("400") }
func (i Image) Source() string    { return i.previewURL("1200") }
func (i Image) previewURL(size string) string {
	if i.ID == "" {
		return ""
	}
	if i.AlbumCover {
		return urls.AlbumCover(i.ID, size)
	}
	return urls.Image(i.ID, size)
}
func (i Image) ThumbnailSources() string {
	return fmt.Sprintf("%s 1x, %s 2x, %s 3x", i.Thumbnail(), i.previewURL("800"), i.Source())
}
func (a Album) CoverImage() Image {
	if a.Cover.ID != "" {
		return a.Cover
	}
	if len(a.Images) > 0 {
		return a.Images[0]
	}
	return Image{}
}
func (a Album) URL() string                 { return urls.Album(a.ID) }
func (a Album) ImageURL(image Image) string { return urls.MediaItem(image.ID) }
func (a Album) PageURL(page int) string {
	if page <= 1 {
		return a.URL()
	}
	return fmt.Sprintf("%s?page=%d", a.URL(), page)
}
func (a Album) ReturnURL(index int) string {
	return fmt.Sprintf("%s#image-%s", a.PageURL(index/max(1, a.PageSize)+1), a.Images[index].ID)
}
