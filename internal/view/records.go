package view

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/model/people"
	"github.com/fachschaftinformatik/web/internal/model/users"

	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/model/archive"
	"github.com/fachschaftinformatik/web/internal/model/events"
	"github.com/fachschaftinformatik/web/internal/model/forum"
	"github.com/fachschaftinformatik/web/internal/model/gallery"
	"github.com/fachschaftinformatik/web/internal/model/meetings"
)

func (a Assignment) VersionLabel() string { return VersionLabel(a.Version) }

func (e Exam) URL() string {
	return "/archive/files/" + e.ID
}

func (e Exam) Size() string {
	if e.Bytes < 1024*1024 {
		return fmt.Sprintf("%d kB", max(1, (e.Bytes+1023)/1024))
	}
	return strings.ReplaceAll(fmt.Sprintf("%.1f MB", float64(e.Bytes)/(1024*1024)), ".", ",")
}

func VersionLabel(value string) string {
	versions := strings.Split(value, ", ")
	for i, value := range versions {
		version := strings.TrimSpace(strings.TrimPrefix(value, "PO"))
		if len(version) == 4 && strings.Trim(version, "0123456789") == "" {
			versions[i] = version
		}
	}
	return strings.Join(versions, ", ")
}

func (e Exam) VersionLabel() string { return VersionLabel(e.Version) }

func (s RequestStatus) Label() string {
	switch archive.Status(s) {
	case archive.Draft:
		return "Entwurf"
	case archive.Pending:
		return "Eingereicht"
	case archive.Approved:
		return "Bestätigt"
	case archive.Rejected:
		return "Abgelehnt"
	}
	return ""
}

func (r Request) DateLabel() string {
	if date, err := parseDate(r.Date); err == nil {
		return date.Format("02.01.2006")
	}
	return r.Date
}

func (r Request) SubmittedLabel() string {
	value := r.CreatedAt
	if r.SubmittedAt != nil {
		value = *r.SubmittedAt
	}
	if date, err := time.Parse(time.RFC3339, value); err == nil {
		return date.Format("02.01.2006")
	}
	return value
}

func (e Event) URL() string { return "/community/veranstaltungen/" + e.ID }

func (e Event) CoverURL() string {
	if e.CoverKey != "" {
		return e.URL() + "/cover"
	}
	return ""
}

func (e Event) LegacyURL() string {
	if e.LegacyPostID.Valid {
		return "/d/" + strconv.FormatInt(e.LegacyPostID.Int64, 10)
	}
	return ""
}

func (e Event) EventLabel() string {
	if len(e.EventDate) == 10 {
		return dateLabel(e.EventDate + "T00:00:00Z")
	}
	return eventLabel(e.EventDate)
}

func (e Event) RichBody() string {
	if e.BodyFormat == "html" {
		return content.SanitizeRichText(e.Body)
	}
	return content.PlainRichText(e.Body)
}

func (e Event) Summary() string {
	text := e.Body
	if e.BodyFormat == "html" {
		text = content.RichTextVisible(text)
	}
	if r := []rune(text); len(r) > 240 {
		return string(r[:240]) + "…"
	}
	return text
}

func (e Event) SourceDateLabel() string {
	return dateLabel(e.SourcePublishedAt)
}

func (e Event) FullImageURL() string {
	if e.GalleryCoverID.Valid {
		return "/media/" + strconv.FormatInt(e.GalleryCoverID.Int64, 10) + "/original"
	}
	return e.CoverURL()
}

func (e Event) GalleryURL() string {
	if e.GalleryAlbumID.Valid {
		return "/community/events/" + strconv.FormatInt(e.GalleryAlbumID.Int64, 10)
	}
	return ""
}

func (e Event) StatusLabel() string {
	switch e.SourceStatus {
	case "cancelled":
		return "Abgesagt"
	case "rescheduled":
		return "Termin geändert"
	}
	return ""
}

func (p Post) URL() string { return "/d/" + p.ID }

func (p Post) AuthorURL() string { return authorURL(p.AuthorID, p.AuthorPublic) }
func authorURL(id int64, public bool) string {
	if public && id > 0 {
		return fmt.Sprintf("/users/%d", id)
	}
	return ""
}

func (p Post) PostedLabel() string { return postedLabel(p.Created) }

func postedLabel(value string) string {
	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return dateLabel(value)
	}
	if location, err := time.LoadLocation("Europe/Berlin"); err == nil {
		date = date.In(location)
	}
	return date.Format("15:04") + " Uhr am " + dateLabel(date.Format(time.RFC3339))
}

func (p Post) SourceURL() string { return "https://ik.fsv-wh.de/d/" + p.ID }

func (p Post) DateISO() string {
	if len(p.Created) < 10 {
		return ""
	}
	return p.Created[:10]
}

func (p Post) DateLabel() string {
	return dateLabel(p.Created)
}

func dateLabel(value string) string {
	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return ""
	}
	months := [...]string{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"}
	return date.Format("2.") + " " + months[date.Month()-1] + " " + date.Format("2006")
}

func (p Post) TypeLabel() string {
	switch p.Type {
	case "news":
		return "Neuigkeit"
	case "event":
		return "Veranstaltung"
	default:
		return "Diskussion"
	}
}

func (c Comment) DateLabel() string { return dateLabel(c.Created) }

func (c Comment) AuthorURL() string {
	return authorURL(c.AuthorID, c.AuthorPublic)
}

func (c Comment) UpdatedLabel() string { return dateLabel(c.Updated) }

func (t Thread) UpdatedLabel() string { return dateLabel(t.Updated) }

func (t Thread) EventLabel() string { return eventLabel(t.EventDate) }
func eventLabel(value string) string {
	date, err := time.Parse("2006-01-02T15:04", value)
	if err != nil {
		return ""
	}
	return dateLabel(date.Format(time.RFC3339)) + ", " + date.Format("15:04") + " Uhr"
}

func (p Post) RelativeTime() string { return relativeTime(p.Created, time.Now()) }

func (c Comment) RelativeTime() string { return relativeTime(c.Created, time.Now()) }

func (c Comment) PostedLabel() string { return postedLabel(c.Created) }

func (t Thread) EditedLabel() string { return postedLabel(t.Updated) }

func (c Comment) EditedLabel() string { return postedLabel(c.Updated) }

func relativeTime(value string, now time.Time) string {
	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return dateLabel(value)
	}
	age := now.Sub(date)
	if age < time.Minute {
		return "gerade eben"
	}
	for _, unit := range []struct {
		duration         time.Duration
		singular, plural string
	}{
		{365 * 24 * time.Hour, "einem Jahr", "Jahren"},
		{30 * 24 * time.Hour, "einem Monat", "Monaten"},
		{24 * time.Hour, "einem Tag", "Tagen"},
		{time.Hour, "einer Stunde", "Stunden"},
		{time.Minute, "einer Minute", "Minuten"},
	} {
		if count := int(age / unit.duration); count > 0 {
			if count == 1 {
				return "vor " + unit.singular
			}
			return fmt.Sprintf("vor %d %s", count, unit.plural)
		}
	}
	return "gerade eben"
}

func (i Image) Thumbnail() string {
	return i.previewURL("400")
}

func (i Image) Source() string { return i.previewURL("1200") }
func (i Image) previewURL(size string) string {
	if i.ID == "" {
		return ""
	}
	if i.AlbumCover {
		return "/media/events/" + i.ID + "/cover/" + size
	}
	return "/media/" + i.ID + "/" + size
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

func (a Album) URL() string { return "/community/events/" + a.ID }

func (a Album) OriginalURL() string { return "https://ik.fsv-wh.de/events/" + a.ID }

func (a Album) ImageURL(image Image) string { return a.URL() + "/" + image.ID }

func (a Album) PageURL(page int) string { return fmt.Sprintf("%s?page=%d", a.URL(), page) }

func (a Album) ReturnURL(index int) string {
	return fmt.Sprintf("%s#image-%s", a.PageURL(index/gallery.PageSize+1), a.Images[index].ID)
}

func (d Document) URL() string {
	if d.Stored {
		return "/documents/" + d.File
	}
	return "/assets/content/" + d.File
}

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
	months := [...]string{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"}
	return fmt.Sprintf("%d. %s %d", date.Day(), months[date.Month()-1], date.Year())
}

func (p Publication) Size() string { return fmt.Sprintf("%d kB", (len(p.Data)+999)/1000) }

func (p Publication) KindLabel() string { return kindLabel(p.Kind) }

func (p Publication) DateLabel() string {
	return NewMeeting(meetings.Meeting{Date: p.Date}).DateLabel()
}

// These adapters contain website routes and display formatting. Domain records
// remain independent of the HTML representation.
type Post struct{ forum.Post }
type Comment struct{ forum.Comment }
type Thread struct {
	Post
	Updated, EventDate, Location string
	Replies                      []Comment
}
type Event events.Event
type Exam struct {
	archive.Exam
	Assignments []Assignment
}
type Module struct {
	archive.Module
	Exams []Exam
}
type Assignment archive.Assignment
type RequestStatus archive.Status
type Request struct {
	archive.Request
	Documents   []Exam
	Assignments []Assignment
	Status      RequestStatus
}
type Image struct{ gallery.Image }
type Album struct {
	gallery.Album
	Images []Image
	Cover  Image
}
type Document meetings.Document
type Meeting struct {
	meetings.Meeting
	Documents []Document
}
type Publication meetings.Publication

func NewPost(p forum.Post) Post { return Post{p} }
func NewThread(t forum.Thread) Thread {
	return Thread{Post: NewPost(t.Post), Updated: t.Updated, EventDate: t.EventDate, Location: t.Location, Replies: mapRecords(t.Replies, func(c forum.Comment) Comment { return Comment{c} })}
}
func (t Thread) Edited() bool { return t.Updated != "" && t.Updated != t.Created }
func NewExam(e archive.Exam) Exam {
	return Exam{Exam: e, Assignments: mapRecords(e.Assignments, func(a archive.Assignment) Assignment { return Assignment(a) })}
}
func NewModule(m archive.Module) Module {
	return Module{Module: m, Exams: mapRecords(m.Exams, NewExam)}
}
func NewRequest(r archive.Request) Request {
	return Request{Request: r, Documents: mapRecords(r.Documents, NewExam), Assignments: mapRecords(r.Assignments, func(a archive.Assignment) Assignment { return Assignment(a) }), Status: RequestStatus(r.Status)}
}
func NewImage(i gallery.Image) Image { return Image{i} }
func NewAlbum(a gallery.Album) Album {
	return Album{Album: a, Images: mapRecords(a.Images, NewImage), Cover: NewImage(a.Cover)}
}
func NewMeeting(m meetings.Meeting) Meeting {
	return Meeting{Meeting: m, Documents: mapRecords(m.Documents, func(d meetings.Document) Document { return Document(d) })}
}
func Posts(v []forum.Post) []Post { return mapRecords(v, NewPost) }
func Events(v []events.Event) []Event {
	return mapRecords(v, func(e events.Event) Event { return Event(e) })
}
func Exams(v []archive.Exam) []Exam           { return mapRecords(v, NewExam) }
func Modules(v []archive.Module) []Module     { return mapRecords(v, NewModule) }
func Requests(v []archive.Request) []Request  { return mapRecords(v, NewRequest) }
func Images(v []gallery.Image) []Image        { return mapRecords(v, NewImage) }
func Albums(v []gallery.Album) []Album        { return mapRecords(v, NewAlbum) }
func Meetings(v []meetings.Meeting) []Meeting { return mapRecords(v, NewMeeting) }
func mapRecords[T, U any](values []T, convert func(T) U) []U {
	if values == nil {
		return nil
	}
	result := make([]U, len(values))
	for i, v := range values {
		result[i] = convert(v)
	}
	return result
}
func (p Post) Summary() string { return summary(p.Body, p.BodyFormat) }
func summary(body, format string) string {
	if format == "html" {
		body = content.RichTextVisible(body)
	}
	if r := []rune(body); len(r) > 240 {
		return string(r[:240]) + "…"
	}
	return body
}
func (e Exam) Date() string {
	if date, err := parseDate(e.Exam.Date); err == nil {
		return date.Format("02.01.2006")
	}
	return e.Exam.Date
}
func parseDate(value string) (time.Time, error) {
	var err error
	for _, layout := range []string{"2006-01-02", "02.01.2006", time.RFC3339} {
		var date time.Time
		date, err = time.Parse(layout, value)
		if err == nil {
			return date, nil
		}
	}
	return time.Time{}, err
}

type Person struct{ people.Person }

func (p Person) Status() string {
	if p.UserID == 0 {
		return "Ohne Benutzerkonto"
	}
	if p.Confirmed {
		return "Bestätigt"
	}
	return "Eingeladen"
}

type User struct{ users.User }

func (u User) Permission() string {
	switch u.Role {
	case "admin":
		return "Administration"
	case "editor":
		return "Redaktion"
	case "user":
		return "Studierende"
	}
	return u.Role
}
func (u User) Status() string {
	if u.Active {
		return "Aktiv"
	}
	return "Inaktiv"
}
func Users(values []users.User) []User {
	return mapRecords(values, func(u users.User) User { return User{u} })
}
