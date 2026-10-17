package controller

import (
	"net/http"
	"slices"

	"github.com/fachschaftinformatik/web/internal/model/people"
	"github.com/fachschaftinformatik/web/internal/view"
)

func responsibilityGroup(section string) string {
	if section == "referate" {
		return "departments"
	}
	return "project"
}

func (c *Handler) EditMember(w http.ResponseWriter, r *http.Request, a *view.AdminPage) {
	id := r.PathValue("id")
	p := people.Person{ID: id}
	found := id == "neu"
	for _, person := range c.people.People() {
		if person.ID == id {
			p = person
			found = true
		}
	}
	if !found {
		c.notFound(w, r)
		return
	}
	title := "Mitglied bearbeiten"
	if id == "neu" || p.UserID == 0 {
		title = "Mitglied einladen"
		a.Inviting = true
	}
	accountEmail := ""
	errors := map[string]string{}
	if r.Method == http.MethodPost {
		p.Name = r.PostForm.Get("name")
		if a.Inviting {
			accountEmail = r.PostForm.Get("email")

			var err error
			errors, err = c.people.InviteMember(r.Context(), c.identity, id, p.Name, accountEmail)
			if err != nil {
				c.serverError(w, err)
				return
			}
			if len(errors) == 0 {
				http.Redirect(w, r, "/verwaltung/mitglieder?eingeladen=1", http.StatusSeeOther)
				return
			}
		} else {

			errors = c.people.SavePerson(p)
		}
		if len(errors) == 0 {
			http.Redirect(w, r, "/verwaltung/mitglieder?gespeichert=1", http.StatusSeeOther)
			return
		}
	}
	a.View, a.Action = "form", r.URL.Path
	if id != "neu" {
		a.Person = ptr(view.Person{Person: p})
	}
	a.Fields = []view.AdminField{view.Field("name", "Name", p.Name, "text")}
	if a.Inviting {
		a.Fields = append(a.Fields, view.Field("email", "E-Mail-Adresse des Benutzerkontos", accountEmail, "email"))
		a.Fields[1].Hint = "Besteht bereits ein Benutzerkonto, wird es mit dem Mitglied verknüpft. Andernfalls wird ein Konto angelegt und eine Einladung gesendet."
	} else {

		a.Fields[0].Hint = "Der Name wird auch im verknüpften Benutzerkonto geändert."
	}

	view.AttachErrors(a.Fields, errors)
	status := http.StatusOK
	if len(errors) != 0 {
		status = http.StatusUnprocessableEntity
	}
	c.adminRender(w, title, a, status)
}

func (c *Handler) EditResponsibility(w http.ResponseWriter, r *http.Request, a *view.AdminPage) {
	id, section := r.PathValue("id"), r.PathValue("section")
	group := responsibilityGroup(section)
	if section == "aufgaben" || section == "funktionen" {
		group = "departments"
	}
	responsibility := people.Responsibility{ID: id, Group: group}
	found := id == "neu"
	for _, kind := range []string{"board", "departments", "project"} {
		if section != "aufgaben" && section != "funktionen" && kind != group {
			continue
		}
		for _, existing := range c.people.Responsibilities(kind) {
			if existing.ID == id {
				responsibility = existing
				found = true
			}
		}
	}
	if !found {
		c.notFound(w, r)
		return
	}
	title := "Funktion bearbeiten"
	if id == "neu" {
		title = "Funktion hinzufügen"
	}
	errors := map[string]string{}
	if r.Method == http.MethodPost {
		responsibility.Title, responsibility.Email, responsibility.MemberIDs = r.PostForm.Get("title"), r.PostForm.Get("email"), r.PostForm["members"]
		if value := r.PostForm.Get("group"); value != "" {
			responsibility.Group = value
		}
		errors = c.people.SaveResponsibility(responsibility)
		if len(errors) == 0 {
			http.Redirect(w, r, "/verwaltung/funktionen?gespeichert=1", http.StatusSeeOther)
			return
		}
	}
	a.View, a.Action = "form", r.URL.Path
	a.Fields = []view.AdminField{view.Field("title", "Bezeichnung", responsibility.Title, "text"), view.Field("group", "Art", responsibility.Group, "radios"), view.Field("email", "Öffentliche E-Mail-Adresse (optional)", responsibility.Email, "email"), view.Field("members", "Mitglieder", "", "checkboxes")}
	a.Fields[1].Options = []view.AdminOption{{Value: "board", Label: "Vorstand", Selected: responsibility.Group == "board"}, {Value: "departments", Label: "Referat", Selected: responsibility.Group == "departments"}, {Value: "project", Label: "Projektstelle", Selected: responsibility.Group == "project"}}
	for _, p := range c.people.People() {
		a.Fields[3].Options = append(a.Fields[3].Options, view.AdminOption{Value: p.ID, Label: p.Name, Selected: slices.Contains(responsibility.MemberIDs, p.ID)})
	}
	view.AttachErrors(a.Fields, errors)
	status := http.StatusOK
	if len(errors) != 0 {
		status = http.StatusUnprocessableEntity
	}
	c.adminRender(w, title, a, status)
}

func (c *Handler) Team(w http.ResponseWriter, r *http.Request) {
	if err := c.people.Refresh(r.Context()); err != nil {
		c.serverError(w, err)
		return
	}
	c.render(w, "team", false, view.Page{Title: "Mitglieder", Active: "team", Fachschaft: true, Team: c.people.Groups()})
}
