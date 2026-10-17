package web

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/url"

	"github.com/fachschaftinformatik/web/internal/content/files"
	"github.com/fachschaftinformatik/web/internal/content/posts"
	"github.com/fachschaftinformatik/web/internal/people"
)

//go:embed templates/*.html assets/*
var resources embed.FS

type Page struct {
	Title, Active string
	DocumentTitle string
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
	Editor        *ForumEditor
	SignedIn      bool
	Voting        bool
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

type ArchivePage struct {
	Module             *files.Module
	Filter             files.Filter
	Programs           []FilterOption
	Versions           []FilterOption
	Degrees            []FilterOption
	Modules            []files.Module
	Page, Pages, Total int
}

func (p ArchivePage) querySuffix() string {
	q := url.Values{}
	q.Set("applied", "1")
	if p.Filter.Query != "" {
		q.Set("q", p.Filter.Query)
	}
	for key, values := range map[string][]string{"degree": p.Filter.Degrees, "program": p.Filter.Programs, "version": p.Filter.Versions} {
		for _, value := range values {
			q.Add(key, value)
		}
	}
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}

func (p ArchivePage) SearchURL() string { return "/archive" + p.querySuffix() }

func (p ArchivePage) ModuleURL(module files.Module) string {
	return "/archive/" + url.PathEscape(module.ID) + p.querySuffix()
}

func (p ArchivePage) Filtered() bool {
	return p.Filter.Query != "" || len(p.Filter.Programs)+len(p.Filter.Degrees)+len(p.Filter.Versions) > 0
}

func (p ArchivePage) FilterCount() int {
	count := 0
	for _, program := range p.Programs {
		if program.Selected {
			count++
		}
	}
	for _, version := range p.Versions {
		if version.Selected {
			count++
		}
	}
	for _, degree := range p.Degrees {
		if degree.Selected {
			count++
		}
	}
	return count
}

type FilterOption struct {
	Name     string
	Selected bool
}

type Renderer struct {
	pages map[string]*template.Template
}

func newRenderer() (*Renderer, error) {
	r := &Renderer{pages: make(map[string]*template.Template)}
	for _, name := range []string{"home", "team", "contact", "sitemap", "archive", "archivepreferences", "module", "upload", "request", "requests", "images", "community", "fachschaft", "content", "elections", "regulations", "login", "register", "forgot", "token", "invite", "forum", "thread", "profile", "news", "events", "editor", "admin"} {
		t, err := template.New("layout").Funcs(template.FuncMap{"voteControl": voteControl, "requestScope": requestScope, "assignmentSummary": assignmentSummary, "scopeField": scopeField, "poLabel": files.VersionLabel, "forumText": forumText, "rawForumText": rawForumText, "richHTML": func(source string) template.HTML { return template.HTML(posts.SanitizeRichText(source)) }}).ParseFS(resources, "templates/layout.html", "templates/vote.html", "templates/requestlist.html", "templates/archivescope.html", "templates/assignmentsummary.html", "templates/adminnav.html", "templates/"+name+".html")
		if err != nil {
			return nil, err
		}
		r.pages[name] = t
	}
	return r, nil
}

func (r *Renderer) Render(w io.Writer, page string, fragment bool, data Page) error {
	t, ok := r.pages[page]
	if !ok {
		return fmt.Errorf("unknown view %q", page)
	}
	name := "layout"
	if fragment {
		name = "results"
	}
	return t.ExecuteTemplate(w, name, data)
}

func Assets() fs.FS {
	assets, _ := fs.Sub(resources, "assets")
	return assets
}
