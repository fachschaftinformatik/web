package view

import (
	"github.com/fachschaftinformatik/web/internal/model/people"
)

type Breadcrumb struct {
	Title, URL string
}

// Breadcrumbs describes ancestors only; the current page is already the h1.
func (p Page) Breadcrumbs() []Breadcrumb {
	if p.Active == "home" {
		return nil
	}
	items := []Breadcrumb{{"Startseite", "/"}}
	if p.Admin != nil {
		if p.Admin.Section != "" {
			items = append(items, Breadcrumb{"Verwaltung", "/verwaltung"})
		}
		if p.Admin.View == "form" || p.Admin.View == "user" || p.Admin.View == "review" || p.Admin.View == "reset" {
			items = append(items, Breadcrumb{p.Admin.SectionTitle, "/verwaltung/" + p.Admin.Section})
		}
		if p.Admin.View == "reset" {
			items = append(items, Breadcrumb{p.Admin.User.Name, "/verwaltung/benutzer/" + p.Admin.User.ID})
		}
	}
	if p.Archive.Module != nil {
		items = append(items, Breadcrumb{"Altklausuren", p.Archive.SearchURL()})
	}
	if p.Active == "archive" && (p.Upload.Action != "" || p.Request != nil || p.RequestList.Requests != nil) {
		items = append(items, Breadcrumb{"Altklausuren", "/archive"})
		if p.Request != nil {
			items = append(items, Breadcrumb{"Meine Einreichungen", "/archive/requests"})
		}
	}
	if p.Fachschaft && p.Active != "fachschaft" {
		items = append(items, Breadcrumb{"Fachschaft", "/fachschaft"})
	}
	if p.Community && p.Active != "community" {
		items = append(items, Breadcrumb{"Community", "/community"})
	}
	if p.Active == "forum" && (p.Forum.Thread != nil || p.Editor != nil) {
		url := "/discussions"
		if p.Forum.BackURL != "" {
			url = p.Forum.BackURL
		}
		items = append(items, Breadcrumb{"Forum", url})
		if p.Editor != nil && p.Editor.Editing {
			items = append(items, Breadcrumb{p.Editor.Draft.Title, p.Editor.BackURL})
		}
	}
	if p.Active == "calendar" && (p.Forum.Thread != nil || p.Editor != nil) {
		items = append(items, Breadcrumb{"Veranstaltungen", EventsURL})
		if p.Editor != nil && p.Editor.Editing {
			items = append(items, Breadcrumb{p.Editor.Draft.Title, p.Editor.BackURL})
		}
	}
	if p.Active == "calendar" && p.Events.Detail != nil {
		items = append(items, Breadcrumb{"Veranstaltungen", EventsURL})
	}
	if p.Active == "news" && (p.Forum.Thread != nil || p.Editor != nil) {
		items = append(items, Breadcrumb{"Neuigkeiten", NewsURL})
	}
	if p.Images.Detail {
		items = append(items, Breadcrumb{"Fotos", "/community/events"})
		if p.Images.Viewer {
			items = append(items, Breadcrumb{p.Images.Album.Title, p.Images.Return})
		}
	}
	if p.Active == "register" || p.Active == "forgot" {
		items = append(items, Breadcrumb{"Anmelden", "/login"})
	}
	return items
}

const EventsURL = "/community/veranstaltungen"

type LegalPage struct {
	Kind  string
	Chair []people.Person
}

const NewsURL = "/community/neuigkeiten"

type Page struct {
	Title, Active string
	DocumentTitle string
	Message       string
	Legal         LegalPage
	Community     bool
	Fachschaft    bool
	Team          []people.Group
	Archive       ArchivePage
	Images        ImagesPage
	Home          HomePage
	Login         LoginPage
	Register      RegisterPage
	Forgot        ForgotPage
	Meetings      MeetingsPage
	Forum         ForumPage
	Events        EventsPage
	Editor        *ContentEditor
	SignedIn      bool
	CanAdmin      bool
	Token         TokenPage
	Upload        UploadPage
	Request       *RequestPage
	RequestList   RequestListPage
	Invite        InvitationPage
	UserID        int64
	Role          string
	CSRF          string
	Admin         *AdminPage
}
