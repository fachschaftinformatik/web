package web

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/url"

	"github.com/fachschaftinformatik/web/internal/archive"
	"github.com/fachschaftinformatik/web/internal/forum"
	"github.com/fachschaftinformatik/web/internal/home"
	"github.com/fachschaftinformatik/web/internal/team"
)

//go:embed templates/*.html assets/*
var files embed.FS

type Page struct {
	Title, Active string
	DocumentTitle string
	Community     bool
	Fachschaft    bool
	Team          []team.Group
	Archive       ArchivePage
	Gallery       GalleryPage
	Home          home.Content
	Login         LoginPage
	Register      RegisterPage
	Forgot        ForgotPage
	Meetings      MeetingsPage
	Forum         ForumPage
	Editor        *ForumEditor
	SignedIn      bool
	CSRF          string
	Admin         *AdminPage
}

type ArchivePage struct {
	Module   *archive.Module
	Filter   archive.Filter
	Programs []FilterOption
	Versions []FilterOption
	Degrees  []FilterOption
	Modules  []archive.Module
}

func (p ArchivePage) querySuffix() string {
	q := url.Values{}
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

func (p ArchivePage) ModuleURL(module archive.Module) string {
	return "/archive/" + url.PathEscape(module.ID) + p.querySuffix()
}

func (p ArchivePage) Filtered() bool { return p.querySuffix() != "" }

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
	for _, name := range []string{"home", "team", "contact", "archive", "archive-module", "gallery", "community", "fachschaft", "meetings", "elections", "regulations", "login", "register", "forgot", "forum", "thread", "forum-editor", "admin"} {
		t, err := template.New("layout").Funcs(template.FuncMap{"forumText": forumText, "richHTML": func(source string) template.HTML { return template.HTML(forum.SanitizeRichText(source)) }}).ParseFS(files, "templates/layout.html", "templates/"+name+".html")
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
	assets, _ := fs.Sub(files, "assets")
	return assets
}
