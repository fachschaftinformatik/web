package view

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"

	"github.com/fachschaftinformatik/web/internal/model/archive"
	"github.com/fachschaftinformatik/web/internal/model/forum"
)

//go:embed templates/*.html assets/*
var resources embed.FS

type Renderer struct {
	pages map[string]*template.Template
}

func NewRenderer() (*Renderer, error) {
	r := &Renderer{pages: make(map[string]*template.Template)}
	for _, name := range []string{"home", "team", "contact", "sitemap", "legal", "error", "archive", "archivepreferences", "module", "upload", "request", "requests", "images", "community", "fachschaft", "content", "elections", "regulations", "login", "register", "forgot", "token", "invite", "forum", "thread", "profile", "news", "events", "editor", "admin"} {
		t, err := template.New("layout").Funcs(template.FuncMap{"voteControl": voteControl, "requestScope": func(r archive.Request) archive.Filter { return r.Scope() }, "assignmentSummary": assignmentSummary, "scopeField": scopeField, "poLabel": archive.VersionLabel, "rawForumText": rawForumText, "richHTML": func(source string) template.HTML { return template.HTML(forum.SanitizeRichText(source)) }}).ParseFS(resources, "templates/layout.html", "templates/vote.html", "templates/requestlist.html", "templates/archivescope.html", "templates/assignmentsummary.html", "templates/adminnav.html", "templates/eventcards.html", "templates/"+name+".html")
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
