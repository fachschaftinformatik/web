package gallery

import (
	"fmt"
	"strings"
)

const PageSize = 24

type Image struct {
	ID, Title, Description, Source string
	Width, Height                  int
	thumbnail                      string
}

type Album struct {
	ID, Title     string
	OriginalCount int
	Images        []Image
	Cover         Image
}

func (i Image) Thumbnail() string {
	if i.thumbnail != "" {
		return i.thumbnail
	}
	return "/assets/images/" + i.ID + "-400.jpg"
}

func (i Image) ThumbnailSources() string {
	if i.thumbnail != "" {
		return fmt.Sprintf("%s 1x, %s/800 2x, %s 3x", i.Thumbnail(), strings.TrimSuffix(i.Source, "/1200"), i.Source)
	}
	return fmt.Sprintf("%s 1x, %s 3x", i.Thumbnail(), i.Source)
}

func (a Album) CoverImage() Image {
	if a.Cover.Source != "" {
		return a.Cover
	}
	if len(a.Images) > 0 {
		return a.Images[0]
	}
	return Image{}
}

func (a Album) URL() string                 { return "/community/events/" + a.ID }
func (a Album) OriginalURL() string         { return "https://ik.fsv-wh.de/events/" + a.ID }
func (a Album) ImageURL(image Image) string { return a.URL() + "/" + image.ID }
func (a Album) PageURL(page int) string     { return fmt.Sprintf("%s?page=%d", a.URL(), page) }
func (a Album) ReturnURL(index int) string {
	return fmt.Sprintf("%s#image-%s", a.PageURL(index/PageSize+1), a.Images[index].ID)
}
