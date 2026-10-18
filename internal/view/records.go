package view

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/model/members"
	"github.com/fachschaftinformatik/web/internal/model/users"
	"github.com/fachschaftinformatik/web/internal/urls"

	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/model/events"
	"github.com/fachschaftinformatik/web/internal/model/exams"
	"github.com/fachschaftinformatik/web/internal/model/media"
	"github.com/fachschaftinformatik/web/internal/model/meetings"
	"github.com/fachschaftinformatik/web/internal/model/posts"
)

func (a Assignment) VersionLabel() string { return VersionLabel(a.Version) }

func (e Exam) URL() string {
	return urls.ExamDocument(e.ID)
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

func (s SubmissionStatus) Label() string {
	switch exams.Status(s) {
	case exams.Draft:
		return "Entwurf"
	case exams.Pending:
		return "In Prüfung"
	case exams.Approved:
		return "Freigegeben"
	case exams.Rejected:
		return "Abgelehnt"
	}
	return ""
}

func (r Submission) DateLabel() string {
	if date, err := parseDate(r.Date); err == nil {
		return dateLabel(date.Format(time.DateOnly))
	}
	return r.Date
}

func (r Submission) SubmittedLabel() string {
	value := r.CreatedAt
	if r.SubmittedAt != nil {
		value = *r.SubmittedAt
	}
	if label := dateLabel(value); label != "" {
		return label
	}
	return value
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

func (e Event) EventLabel() string {
	return eventLabel(e.EventDate)
}

func (e Event) DateLabel() string { return dateLabel(e.EventDate) }

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

func (e Event) SourceDateLabel() string {
	return dateLabel(e.SourcePublishedAt)
}

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
	switch e.SourceStatus {
	case "rescheduled":
		return "Termin geändert"
	}
	return ""
}

func (p Post) URL() string                            { return urls.Post(p.ID) }
func (p Post) EditURL() string                        { return urls.EditPost(p.ID) }
func (p Post) VoteURL() string                        { return urls.PostVote(p.ID) }
func (p Post) CommentsURL() string                    { return urls.PostComments(p.ID) }
func (t Thread) CommentVoteURL(comment string) string { return urls.CommentVote(t.ID, comment) }
func (t Thread) EditCommentURL(comment string) string { return urls.EditComment(t.ID, comment) }

func (p Post) AuthorURL() string { return authorURL(p.AuthorPublicID, p.AuthorPublic) }
func authorURL(id string, public bool) string {
	// Public account profiles are not currently exposed. Keep author names
	// subject to the existing account visibility rules without linking them.
	return ""
}

func (p Post) PostedLabel() string { return postedLabel(p.Created) }

func postedLabel(value string) string {
	return formatDate(value, displayLocale, dateStylePosted)
}

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
	return formatDate(value, displayLocale, dateStyleDate)
}

func (p Post) TypeLabel() string {
	switch p.Type {
	case "news":
		return "Ankündigung"
	case "event":
		return "Veranstaltung"
	default:
		return "Diskussion"
	}
}

func (c Comment) DateLabel() string { return dateLabel(c.Created) }

func (c Comment) AuthorURL() string {
	return authorURL(c.AuthorPublicID, c.AuthorPublic)
}

func (c Comment) UpdatedLabel() string { return dateLabel(c.Updated) }

func (t Thread) UpdatedLabel() string { return dateLabel(t.Updated) }

func (t Thread) EventLabel() string { return eventLabel(t.EventDate) }
func eventLabel(value string) string {
	return formatDate(value, displayLocale, dateStyleEvent)
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

func (a Album) URL() string { return urls.Album(a.ID) }

func (a Album) ImageURL(image Image) string { return urls.MediaItem(image.ID) }

func (a Album) PageURL(page int) string {
	if page <= 1 {
		return a.URL()
	}
	return fmt.Sprintf("%s?page=%d", a.URL(), page)
}

func (a Album) ReturnURL(index int) string {
	return fmt.Sprintf("%s#image-%s", a.PageURL(index/media.PageSize+1), a.Images[index].ID)
}

func (d Document) URL() string {
	return urls.Document(d.PublicID)
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
	return dateLabel(date.Format(time.DateOnly))
}

func (p Publication) Size() string { return fmt.Sprintf("%d kB", (len(p.Data)+999)/1000) }

func (p Publication) KindLabel() string { return kindLabel(p.Kind) }

func (p Publication) DateLabel() string {
	return NewMeeting(meetings.Meeting{Date: p.Date}).DateLabel()
}

// These adapters contain website routes and display formatting. Domain records
// remain independent of the HTML representation.
type Post struct{ posts.Post }
type Comment struct{ posts.Comment }
type Thread struct {
	Post
	Updated, EventDate, Location string
	Replies                      []Comment
}
type Event events.Event
type Exam struct {
	exams.Exam
	Assignments []Assignment
}
type Module struct {
	exams.Module
	Exams []Exam
}

func (m Module) URL() string { return urls.ExamModule(m.ID) }

type Assignment exams.Assignment
type SubmissionStatus exams.Status
type Submission struct {
	exams.Submission
	Documents   []Exam
	Assignments []Assignment
	Status      SubmissionStatus
}

func (r Submission) ListTitle() string {
	if r.ModuleCount > 1 {
		return fmt.Sprintf("%s und %d weitere", r.Title, r.ModuleCount-1)
	}
	return r.Title
}

type Image struct{ media.Image }
type Album struct {
	media.Album
	Images []Image
	Cover  Image
}
type Document meetings.Document
type Meeting struct {
	meetings.Meeting
	Documents []Document
}
type Publication meetings.Publication

func (p Publication) URL() string      { return urls.Document(p.PublicID) }
func (p Publication) EditURL() string  { return urls.AdminMeeting(p.PublicID) + "/edit" }
func (p Publication) CheckURL() string { return urls.AdminMeeting(p.PublicID) + "/check" }

func (r Submission) URL() string       { return urls.Submission(r.ID) }
func (r Submission) CheckURL() string  { return urls.CheckSubmission(r.ID) }
func (r Submission) ReviewURL() string { return urls.ReviewSubmission(r.ID) }
func (r Submission) AdminURL() string  { return urls.AdminSubmission(r.ID) }
func (r Submission) UserURL() string   { return urls.AdminUser(r.UserPublicID) }

func (e Exam) RevisionURL() string {
	return "/exams/submissions/create?" + url.Values{"revision": {e.ID}}.Encode()
}

func NewPost(p posts.Post) Post { return Post{p} }
func NewThread(t posts.Thread) Thread {
	return Thread{Post: NewPost(t.Post), Updated: t.Updated, EventDate: t.EventDate, Location: t.Location, Replies: mapRecords(t.Replies, func(c posts.Comment) Comment { return Comment{c} })}
}
func (t Thread) Edited() bool { return t.Updated != "" && t.Updated != t.Created }
func NewExam(e exams.Exam) Exam {
	return Exam{Exam: e, Assignments: mapRecords(e.Assignments, func(a exams.Assignment) Assignment { return Assignment(a) })}
}
func NewModule(m exams.Module) Module {
	return Module{Module: m, Exams: mapRecords(m.Exams, NewExam)}
}
func NewSubmission(r exams.Submission) Submission {
	return Submission{Submission: r, Documents: mapRecords(r.Documents, NewExam), Assignments: mapRecords(r.Assignments, func(a exams.Assignment) Assignment { return Assignment(a) }), Status: SubmissionStatus(r.Status)}
}
func NewImage(i media.Image) Image { return Image{i} }
func NewAlbum(a media.Album) Album {
	return Album{Album: a, Images: mapRecords(a.Images, NewImage), Cover: NewImage(a.Cover)}
}
func NewMeeting(m meetings.Meeting) Meeting {
	return Meeting{Meeting: m, Documents: mapRecords(m.Documents, func(d meetings.Document) Document { return Document(d) })}
}
func Posts(v []posts.Post) []Post { return mapRecords(v, NewPost) }
func Events(v []events.Event) []Event {
	return mapRecords(v, func(e events.Event) Event { return Event(e) })
}
func Exams(v []exams.Exam) []Exam                   { return mapRecords(v, NewExam) }
func Modules(v []exams.Module) []Module             { return mapRecords(v, NewModule) }
func Submissions(v []exams.Submission) []Submission { return mapRecords(v, NewSubmission) }
func Images(v []media.Image) []Image                { return mapRecords(v, NewImage) }
func Albums(v []media.Album) []Album                { return mapRecords(v, NewAlbum) }
func Meetings(v []meetings.Meeting) []Meeting       { return mapRecords(v, NewMeeting) }
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
	body = strings.Join(strings.Fields(body), " ")
	if r := []rune(body); len(r) > 240 {
		end := 240
		if r[end] != ' ' {
			for i := end - 1; i > 0; i-- {
				if r[i] == ' ' {
					end = i
					break
				}
			}
		}
		return strings.TrimSpace(string(r[:end])) + "…"
	}
	return body
}
func (e Exam) Date() string {
	if date, err := parseDate(e.Exam.Date); err == nil {
		return dateLabel(date.Format(time.DateOnly))
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

type Member struct{ members.Member }

func (m Member) URL() string       { return urls.AdminMember(m.PublicID) }
func (m Member) EditURL() string   { return m.URL() + "/edit" }
func (m Member) InviteURL() string { return m.URL() + "/invite" }
func (m Member) UserURL() string   { return urls.AdminUser(m.UserPublicID) }

func (p Member) Status() string {
	if p.UserID == 0 {
		return "Ohne Benutzerkonto"
	}
	if p.Confirmed {
		return "Bestätigt"
	}
	return "Eingeladen"
}

type User struct{ users.User }

func (u User) URL() string        { return urls.AdminUser(u.PublicID) }
func (u User) ResetURL() string   { return u.URL() + "/reset" }
func (u User) ConfirmURL() string { return u.URL() + "/confirm" }

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
