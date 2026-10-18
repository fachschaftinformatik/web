package view

import (
	"time"

	"github.com/fachschaftinformatik/web/internal/model/members"
	"github.com/fachschaftinformatik/web/internal/model/organization"
)

func (p Page) LegalUpdatedLabel() string {
	value := p.Legal.Updated
	if value == "" {
		value = p.Organization.LegalUpdated
	}
	d, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return ""
	}
	return dateLabel(d.Format(time.RFC3339))
}

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
			items = append(items, Breadcrumb{"Verwaltung", "/admin"})
		}
		if p.Admin.View == "form" || p.Admin.View == "user" || p.Admin.View == "review" || p.Admin.View == "reset" {
			items = append(items, Breadcrumb{p.Admin.SectionTitle, "/admin/" + p.Admin.Section})
		}
		if p.Admin.View == "reset" {
			items = append(items, Breadcrumb{p.Admin.User.Name, p.Admin.User.URL()})
		}
	}
	if p.Exams.Module != nil {
		items = append(items, Breadcrumb{"Altklausuren", p.Exams.SearchURL()})
	}
	if p.Active == "exams" && (p.Upload.Action != "" || p.Submission != nil || p.SubmissionList.Submissions != nil) {
		items = append(items, Breadcrumb{"Altklausuren", "/exams"})
		if p.Submission != nil {
			items = append(items, Breadcrumb{"Meine Einreichungen", "/exams/submissions"})
		}
	}
	if p.About && p.Active != "about" {
		items = append(items, Breadcrumb{"Über uns", "/about"})
	}
	if p.Community && p.Active != "community" {
		items = append(items, Breadcrumb{"Community", "/community"})
	}
	if p.Active == "posts" && (p.Posts.Thread != nil || p.Editor != nil) {
		url := "/posts"
		if p.Posts.BackURL != "" {
			url = p.Posts.BackURL
		}
		items = append(items, Breadcrumb{"Forum", url})
		if p.Editor != nil && p.Editor.Editing {
			items = append(items, Breadcrumb{p.Editor.Draft.Title, p.Editor.BackURL})
		}
	}
	if p.Active == "events" && (p.Posts.Thread != nil || p.Editor != nil) {
		items = append(items, Breadcrumb{"Veranstaltungen", EventsURL})
		if p.Editor != nil && p.Editor.Editing {
			items = append(items, Breadcrumb{p.Editor.Draft.Title, p.Editor.BackURL})
		}
	}
	if p.Active == "events" && p.Events.Detail != nil {
		items = append(items, Breadcrumb{"Veranstaltungen", EventsURL})
	}
	if p.Active == "news" && (p.Posts.Thread != nil || p.Editor != nil) {
		items = append(items, Breadcrumb{"Ankündigungen", NewsURL})
	}
	if p.Media.Detail {
		items = append(items, Breadcrumb{"Fotos", "/media"})
		if p.Media.Viewer {
			items = append(items, Breadcrumb{p.Media.Album.Title, p.Media.Return})
		}
	}
	if p.Active == "register" || p.Active == "forgot" {
		items = append(items, Breadcrumb{"Anmelden", "/login"})
	}
	return items
}

const EventsURL = "/events"

type LegalPage struct {
	Updated string
	Kind    string
	Chair   []members.Member
}

const NewsURL = "/news"

type Page struct {
	CanonicalURL                                                    string
	Organization                                                    organization.Details
	RegistrationDomains, AccountDomains, ChairEmail, PublicityEmail string
	Title, Active                                                   string
	DocumentTitle                                                   string
	Message                                                         string
	Legal                                                           LegalPage
	Community                                                       bool
	About                                                           bool
	Team                                                            []members.Group
	Exams                                                           ExamsPage
	Media                                                           MediaPage
	Home                                                            HomePage
	Login                                                           LoginPage
	Register                                                        RegisterPage
	Forgot                                                          ForgotPage
	Meetings                                                        MeetingsPage
	Posts                                                           PostsPage
	Events                                                          EventsPage
	Editor                                                          *ContentEditor
	SignedIn                                                        bool
	CanAdmin                                                        bool
	Token                                                           TokenPage
	Upload                                                          UploadPage
	Submission                                                      *SubmissionPage
	SubmissionList                                                  SubmissionListPage
	Invite                                                          InvitationPage
	UserID                                                          int64
	Role                                                            string
	CSRF                                                            string
	Admin                                                           *AdminPage
}
