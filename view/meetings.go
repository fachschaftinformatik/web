package view

import (
	"fmt"
	"github.com/fachschaftinformatik/web/internal/urls"
	"time"
)

type Document struct {
	Kind, File, Note, SHA256, PublicID string
	Bytes                              int
	Sources                            []string
	Stored                             bool
}
type Meeting struct {
	Year, Number int
	Date         string
	Documents    []Document
}
type Publication struct {
	ID, Owner, Date, Kind, Filename, PublicID string
	Number, Bytes                             int
	Published                                 bool
}

func (d Document) URL() string  { return urls.Document(d.PublicID) }
func (d Document) Size() string { return fmt.Sprintf("%d kB", (d.Bytes+999)/1000) }
func kindLabel(kind string) string {
	if kind == "Einladung" {
		return "Einladung zur Sitzung"
	}
	if kind == "Protokoll" {
		return "Sitzungsprotokoll"
	}
	return kind
}
func (d Document) Label() string { return kindLabel(d.Kind) }
func (m Meeting) DateLabel() string {
	date, err := time.Parse("2006-01-02", m.Date)
	if err != nil {
		return "Datum uneindeutig"
	}
	return dateLabel(date.Format(time.DateOnly))
}
func (p Publication) Size() string      { return fmt.Sprintf("%d kB", (p.Bytes+999)/1000) }
func (p Publication) KindLabel() string { return kindLabel(p.Kind) }
func (p Publication) DateLabel() string { return (Meeting{Date: p.Date}).DateLabel() }
func (p Publication) URL() string       { return urls.Document(p.PublicID) }
func (p Publication) EditURL() string   { return urls.AdminMeeting(p.PublicID) + "/edit" }
func (p Publication) CheckURL() string  { return urls.AdminMeeting(p.PublicID) + "/check" }
