package web

import (
	"net/http"
	"slices"
	"strings"

	"github.com/fachschaftinformatik/web/internal/people"
)

type LegalPage struct {
	Kind  string
	Chair []people.Person
}

func (c *Handler) Legal(w http.ResponseWriter, r *http.Request) {
	kind := strings.TrimPrefix(r.URL.Path, "/")
	titles := map[string]string{"impressum": "Impressum", "datenschutz": "Datenschutz", "barrierefreiheit": "Barrierefreiheit"}
	if err := c.people.Refresh(r.Context()); err != nil {
		c.serverError(w, err)
		return
	}
	legal := LegalPage{Kind: kind}
	for _, role := range c.people.Responsibilities("board") {
		switch strings.ToLower(strings.TrimSpace(role.Title)) {
		case "vorsitz", "vorsitzende", "vorsitzender":
			for _, person := range c.people.People() {
				if slices.Contains(role.MemberIDs, person.ID) {
					legal.Chair = append(legal.Chair, person)
				}
			}
		}
	}
	c.render(w, "legal", false, Page{Title: titles[kind], Active: "legal", Legal: legal})
}
