package web

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/account"
	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/team"
)

const uploadBodyLimit = content.MaxPDFBytes + 1<<20

type AdminOption struct {
	Value, Label string
	Selected     bool
}
type AdminField struct {
	ID, Label, Value, Kind, Hint, Error string
	Day, Month, Year                    string
	Options                             []AdminOption
}
type AdminRow struct{ Name, Detail, URL string }
type PaginationItem struct {
	Number            int
	URL               string
	Current, Ellipsis bool
}
type AdminLink struct{ ID, Label, URL string }

var adminLinks = []AdminLink{
	{"benutzer", "Benutzer", "/verwaltung/benutzer"},
	{"mitglieder", "Mitglieder", "/verwaltung/mitglieder"},
	{"referate", "Referate", "/verwaltung/referate"},
	{"projektstellen", "Projektstellen", "/verwaltung/projektstellen"},
	{"sitzungen", "Sitzungen und Protokolle", "/verwaltung/sitzungen"},
}

type AdminPage struct {
	Section, SectionTitle, View, Action, Query, Message, Success, AddLabel string
	Fields                                                                 []AdminField
	Rows                                                                   []AdminRow
	Users                                                                  []account.User
	User                                                                   *account.User
	Meetings                                                               []content.Meeting
	Publication                                                            *content.Publication
	Page, Pages, Total                                                     int
	PreviousURL, NextURL                                                   string
	ResetPreview                                                           bool
}

func (a AdminPage) Links() []AdminLink { return adminLinks }

func (a AdminPage) UserPageURL(page int) string {
	q := url.Values{}
	if a.Query != "" {
		q.Set("q", a.Query)
	}
	if page != 1 {
		q.Set("seite", strconv.Itoa(page))
	}
	if len(q) == 0 {
		return "/verwaltung/benutzer"
	}
	return "/verwaltung/benutzer?" + q.Encode()
}

func (a AdminPage) Pagination() []PaginationItem {
	var result []PaginationItem
	var previous int
	for number := 1; number <= a.Pages; number++ {
		if number != 1 && number != a.Pages && (number < a.Page-1 || number > a.Page+1) {
			continue
		}
		if number-previous == 2 {
			skipped := previous + 1
			result = append(result, PaginationItem{Number: skipped, URL: a.UserPageURL(skipped)})
		} else if number-previous > 2 {
			result = append(result, PaginationItem{Ellipsis: true})
		}
		result = append(result, PaginationItem{Number: number, URL: a.UserPageURL(number), Current: number == a.Page})
		previous = number
	}
	return result
}
func (a AdminPage) HasErrors() bool {
	if a.Message != "" {
		return true
	}
	for _, field := range a.Fields {
		if field.Error != "" {
			return true
		}
	}
	return false
}

func (c *Handler) adminRender(w http.ResponseWriter, title string, a *AdminPage, status int) {
	p := Page{Title: title, Active: "admin", Admin: a}
	if a.View == "users" && a.Pages > 1 {
		p.DocumentTitle = fmt.Sprintf("Benutzer – Seite %d von %d", a.Page, a.Pages)
	}
	if a.HasErrors() {
		p.DocumentTitle = "Fehler: " + title
	}
	c.renderStatus(w, "admin", false, p, status)
}

func (c *Handler) Admin(w http.ResponseWriter, r *http.Request) {
	section, id := r.PathValue("section"), r.PathValue("id")
	a := &AdminPage{Section: section, View: "index"}
	if section == "" {
		c.adminRender(w, "Verwaltung", a, http.StatusOK)
		return
	}
	for _, link := range adminLinks {
		if link.ID == section {
			a.SectionTitle = link.Label
		}
	}
	if a.SectionTitle == "" {
		http.NotFound(w, r)
		return
	}
	if r.URL.Query().Get("gespeichert") == "1" {
		a.Success = "Die Änderungen wurden gespeichert."
	}
	switch section {
	case "benutzer":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if id != "" {
			if user, ok := c.users.Get(id); ok {
				a.View = "user"
				a.ResetPreview = r.URL.Query().Get("zuruecksetzen") == "demo"
				a.User = &user
				c.adminRender(w, user.Name, a, http.StatusOK)
				return
			}
			http.NotFound(w, r)
			return
		}
		a.View, a.Query = "users", strings.TrimSpace(r.URL.Query().Get("q"))
		a.Page = 1
		if value := r.URL.Query().Get("seite"); value != "" {
			page, err := strconv.Atoi(value)
			if err != nil {
				http.Redirect(w, r, a.UserPageURL(1), http.StatusSeeOther)
				return
			}
			a.Page = page
		}
		page, ok := c.users.Search(a.Query, a.Page)
		if !ok {
			http.Redirect(w, r, a.UserPageURL(1), http.StatusSeeOther)
			return
		}
		a.Users, a.Total, a.Pages = page.Users, page.Total, page.Pages
		if a.Page > 1 {
			a.PreviousURL = a.UserPageURL(a.Page - 1)
		}
		if a.Page < a.Pages {
			a.NextURL = a.UserPageURL(a.Page + 1)
		}
	case "mitglieder":
		if id != "" {
			c.EditMember(w, r, a)
			return
		}
		a.View, a.AddLabel = "list", "Mitglied hinzufügen"
		for _, p := range c.team.People() {
			a.Rows = append(a.Rows, AdminRow{p.Name, p.Email, "/verwaltung/mitglieder/" + p.ID})
		}
	case "referate", "projektstellen":
		if id != "" {
			c.EditResponsibility(w, r, a)
			return
		}
		a.View, a.AddLabel = "list", "Referat hinzufügen"
		if section == "projektstellen" {
			a.AddLabel = "Projektstelle hinzufügen"
		}
		for _, responsibility := range c.team.Responsibilities(responsibilityGroup(section)) {
			var names []string
			for _, p := range c.team.People() {
				if slices.Contains(responsibility.MemberIDs, p.ID) {
					names = append(names, p.Name)
				}
			}
			a.Rows = append(a.Rows, AdminRow{responsibility.Title, strings.Join(names, ", "), "/verwaltung/" + section + "/" + responsibility.ID})
		}
	case "sitzungen":
		if id != "" {
			c.EditPublication(w, r, a)
			return
		}
		a.View = "meetings"
		a.Meetings = c.content.All()
		if r.URL.Query().Get("veroeffentlicht") == "1" {
			a.Success = "Das Dokument wurde veröffentlicht."
		}
	}
	c.adminRender(w, a.SectionTitle, a, http.StatusOK)
}

func responsibilityGroup(section string) string {
	if section == "referate" {
		return "departments"
	}
	return "project"
}

func field(id, label, value, kind string) AdminField {
	return AdminField{ID: id, Label: label, Value: value, Kind: kind}
}
func attachErrors(fields []AdminField, errors map[string]string) {
	for i := range fields {
		fields[i].Error = errors[fields[i].ID]
	}
}

func (c *Handler) EditMember(w http.ResponseWriter, r *http.Request, a *AdminPage) {
	id := r.PathValue("id")
	p := team.Person{ID: id}
	found := id == "neu"
	for _, person := range c.team.People() {
		if person.ID == id {
			p = person
			found = true
		}
	}
	if !found {
		http.NotFound(w, r)
		return
	}
	title := "Mitglied bearbeiten"
	if id == "neu" {
		title = "Mitglied hinzufügen"
	}
	errors := map[string]string{}
	if r.Method == http.MethodPost {
		p.Name, p.Email, p.Role = r.PostForm.Get("name"), r.PostForm.Get("email"), r.PostForm.Get("role")
		errors = c.team.SavePerson(p)
		if len(errors) == 0 {
			http.Redirect(w, r, "/verwaltung/mitglieder?gespeichert=1", http.StatusSeeOther)
			return
		}
	}
	a.View, a.Action = "form", r.URL.Path
	a.Fields = []AdminField{field("name", "Name", p.Name, "text"), field("email", "Öffentliche E-Mail-Adresse (optional)", p.Email, "email"), field("role", "Aufgabe im Vorstand (optional)", p.Role, "text")}
	a.Fields[2].Hint = "Wird nur bei Mitgliedern des Vorstands angezeigt. Die Zuordnung zu Referaten und Projektstellen bearbeitest du im jeweiligen Bereich."
	attachErrors(a.Fields, errors)
	status := http.StatusOK
	if len(errors) != 0 {
		status = http.StatusUnprocessableEntity
	}
	c.adminRender(w, title, a, status)
}

func (c *Handler) EditResponsibility(w http.ResponseWriter, r *http.Request, a *AdminPage) {
	id, group := r.PathValue("id"), responsibilityGroup(a.Section)
	responsibility := team.Responsibility{ID: id, Group: group}
	found := id == "neu"
	for _, existing := range c.team.Responsibilities(group) {
		if existing.ID == id {
			responsibility = existing
			found = true
		}
	}
	if !found {
		http.NotFound(w, r)
		return
	}
	label := "Referat"
	if group == "project" {
		label = "Projektstelle"
	}
	title := label + " bearbeiten"
	if id == "neu" {
		title = label + " hinzufügen"
	}
	errors := map[string]string{}
	if r.Method == http.MethodPost {
		responsibility.Title, responsibility.Email, responsibility.MemberIDs = r.PostForm.Get("title"), r.PostForm.Get("email"), r.PostForm["members"]
		errors = c.team.SaveResponsibility(responsibility)
		if len(errors) == 0 {
			http.Redirect(w, r, "/verwaltung/"+a.Section+"?gespeichert=1", http.StatusSeeOther)
			return
		}
	}
	a.View, a.Action = "form", r.URL.Path
	a.Fields = []AdminField{field("title", "Bezeichnung", responsibility.Title, "text"), field("email", "Öffentliche E-Mail-Adresse (optional)", responsibility.Email, "email"), field("members", "Mitglieder", "", "checkboxes")}
	a.Fields[2].Hint = "Wähle die zuständigen Personen aus. Entferne ein Häkchen, um eine Zuordnung aufzuheben."
	for _, p := range c.team.People() {
		a.Fields[2].Options = append(a.Fields[2].Options, AdminOption{p.ID, p.Name, slices.Contains(responsibility.MemberIDs, p.ID)})
	}
	attachErrors(a.Fields, errors)
	status := http.StatusOK
	if len(errors) != 0 {
		status = http.StatusUnprocessableEntity
	}
	c.adminRender(w, title, a, status)
}

func (c *Handler) EditPublication(w http.ResponseWriter, r *http.Request, a *AdminPage) {
	s := w.(*sessionResponse).session
	id := r.PathValue("id")
	p := content.Publication{ID: id, Owner: s.ID, Date: r.URL.Query().Get("datum")}
	p.Number, _ = strconv.Atoi(r.URL.Query().Get("nummer"))
	if id != "neu" {
		var ok bool
		p, ok = c.content.Get(id)
		if !ok || p.Owner != s.ID || p.Published {
			http.NotFound(w, r)
			return
		}
	}
	errors := map[string]string{}
	day, month, year := "", "", ""
	if date, err := time.Parse("2006-01-02", p.Date); err == nil {
		day, month, year = strconv.Itoa(date.Day()), strconv.Itoa(int(date.Month())), strconv.Itoa(date.Year())
	}
	number := ""
	if p.Number > 0 {
		number = strconv.Itoa(p.Number)
	}
	if r.Method == http.MethodPost {
		if id == "neu" {
			p.ID = randomToken()
		}
		day, month, year = r.PostForm.Get("day"), r.PostForm.Get("month"), r.PostForm.Get("year")
		d, _ := strconv.Atoi(day)
		m, _ := strconv.Atoi(month)
		y, _ := strconv.Atoi(year)
		p.Date = fmt.Sprintf("%04d-%02d-%02d", y, m, d)
		number = r.PostForm.Get("number")
		p.Number, _ = strconv.Atoi(number)
		p.Kind = r.PostForm.Get("kind")
		file, header, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			data, readErr := io.ReadAll(io.LimitReader(file, content.MaxPDFBytes+1))
			candidate := p
			candidate.Data = data
			if readErr != nil {
				errors["file"] = "Die Datei konnte nicht gelesen werden. Wähle sie erneut aus."
			} else if msg := content.ValidatePublication(candidate)["file"]; msg != "" {
				errors["file"] = msg
			} else {
				p.Data, p.Filename = data, filepath.Base(header.Filename)
			}
		} else if err != http.ErrMissingFile {
			errors["file"] = "Die Datei konnte nicht gelesen werden. Wähle sie erneut aus."
		}
		for key, message := range content.ValidatePublication(p) {
			if errors[key] == "" {
				errors[key] = message
			}
		}
		if err := c.content.Save(p); err != nil {
			a.Message = err.Error()
		} else if len(errors) == 0 {
			http.Redirect(w, r, "/verwaltung/sitzungen/"+p.ID+"/pruefen", http.StatusSeeOther)
			return
		}
	}
	a.View, a.Action, a.Publication = "form", "/verwaltung/sitzungen/"+p.ID, &p
	a.Fields = []AdminField{
		{ID: "date", Label: "Sitzungsdatum", Kind: "date", Hint: "Zum Beispiel 27 10 2026.", Day: day, Month: month, Year: year},
		field("number", "Sitzungsnummer", number, "text"),
		{ID: "kind", Label: "Dokumentart", Kind: "radios", Options: []AdminOption{{"Einladung", "Einladung", p.Kind == "Einladung"}, {"Protokoll", "Protokoll", p.Kind == "Protokoll"}}},
		{ID: "file", Label: "PDF-Datei", Kind: "file", Hint: "Wähle eine PDF-Datei mit höchstens 8 MB aus."},
	}
	attachErrors(a.Fields, errors)
	status := http.StatusOK
	if len(errors) != 0 || a.Message != "" {
		status = http.StatusUnprocessableEntity
	}
	c.adminRender(w, "Dokument hinzufügen", a, status)
}

func (c *Handler) ReviewPublication(w http.ResponseWriter, r *http.Request) {
	s := w.(*sessionResponse).session
	p, ok := c.content.Get(r.PathValue("id"))
	if !ok || p.Owner != s.ID {
		http.NotFound(w, r)
		return
	}
	a := &AdminPage{Section: "sitzungen", SectionTitle: "Sitzungen und Protokolle", View: "review", Publication: &p, Action: r.URL.Path}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		if err := c.content.Publish(p.ID, s.ID); err != nil {
			a.Message = err.Error()
			status = http.StatusUnprocessableEntity
		} else {
			http.Redirect(w, r, "/verwaltung/sitzungen?veroeffentlicht=1", http.StatusSeeOther)
			return
		}
	}
	if p.Published {
		http.Redirect(w, r, "/verwaltung/sitzungen", http.StatusSeeOther)
		return
	}
	c.adminRender(w, "Veröffentlichung prüfen", a, status)
}

func (c *Handler) Document(w http.ResponseWriter, r *http.Request) {
	p, ok := c.content.Get(r.PathValue("id"))
	s := w.(*sessionResponse).session
	if !ok || !p.Published && (s == nil || !s.Authenticated || s.ID != p.Owner) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `inline; filename="dokument.pdf"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	http.ServeContent(w, r, "dokument.pdf", time.Time{}, bytes.NewReader(p.Data))
}

// ResetUserPassword previews the future mail action; it does not change a
// password, create a reset token or send mail to production-copy accounts.
func (c *Handler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	if user, ok := c.users.Get(r.PathValue("id")); ok {
		if r.Method == http.MethodPost {
			http.Redirect(w, r, "/verwaltung/benutzer/"+url.PathEscape(user.ID)+"?zuruecksetzen=demo", http.StatusSeeOther)
			return
		}
		c.adminRender(w, "Passwort zurücksetzen", &AdminPage{Section: "benutzer", SectionTitle: "Benutzer", View: "reset", User: &user}, http.StatusOK)
		return
	}
	http.NotFound(w, r)
}
