package view

import (
	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/urls"
)

type Event struct {
	ID, Title, Body, BodyFormat, EventDate, Location, Created, Updated string
	LegacyPostPublicID, AlbumPublicID, AlbumCoverPublicID              string
	CoverKey, CoverAlt                                                 string
	CoverWidth, CoverHeight                                            int
	SourceURL, SourcePublishedAt, SourceStatus                         string
}

func (e Event) URL() string     { return urls.Event(e.ID) }
func (e Event) EditURL() string { return urls.EditEvent(e.ID) }
func (e Event) CoverURL() string {
	if e.CoverKey != "" {
		return urls.EventCover(e.ID)
	}
	return ""
}
func (e Event) LegacyURL() string {
	if e.LegacyPostPublicID != "" {
		return urls.Post(e.LegacyPostPublicID)
	}
	return ""
}
func (e Event) EventLabel() string { return eventLabel(e.EventDate) }
func (e Event) DateLabel() string  { return dateLabel(e.EventDate) }
func (e Event) RichBody() string {
	if e.BodyFormat == "html" {
		return content.SanitizeRichText(e.Body)
	}
	return content.MarkdownRichText(e.Body)
}
func (e Event) Summary() string {
	text := content.RichTextVisible(e.RichBody())
	if r := []rune(text); len(r) > 240 {
		return string(r[:240]) + "…"
	}
	return text
}
func (e Event) SourceDateLabel() string { return dateLabel(e.SourcePublishedAt) }
func (e Event) FullImageURL() string {
	if e.AlbumCoverPublicID != "" {
		return urls.Image(e.AlbumCoverPublicID, "original")
	}
	return e.CoverURL()
}
func (e Event) AlbumURL() string {
	if e.AlbumPublicID != "" {
		return urls.Album(e.AlbumPublicID)
	}
	return ""
}
func (e Event) StatusLabel() string {
	if e.SourceStatus == "rescheduled" {
		return "Termin geändert"
	}
	return ""
}
