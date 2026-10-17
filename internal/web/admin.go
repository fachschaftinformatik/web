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

	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/content/files"
	"github.com/fachschaftinformatik/web/internal/people"
	"github.com/fachschaftinformatik/web/internal/users"
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
type AdminMemberRow struct {
	people.Person
	Tasks []string
}
type AdminTaskRow struct {
	Title, Kind, URL string
	Members          []string
}
type PaginationItem struct {
	Number            int
	URL               string
	Current, Ellipsis bool
}
type AdminLink struct{ ID, Label, URL, Group string }

var adminLinks = []AdminLink{
	{"mitglieder", "Mitglieder", "/verwaltung/mitglieder", "Organisation"},
	{"funktionen", "Funktionen", "/verwaltung/funktionen", "Organisation"},
	{"sitzungen", "Sitzungen", "/verwaltung/sitzungen", "Organisation"},
	{"einreichungen", "Einreichungen", "/verwaltung/einreichungen", "Vorgänge"},
	{"benutzer", "Benutzerkonten", "/verwaltung/benutzer", "Zugänge"},
}

type AdminPage struct {
	CanResendConfirmation                                        bool
	Person                                                       *people.Person
	Inviting                                                     bool
	ReviewerOnly                                                 bool
	Requests                                                     []files.Request
	Section, SectionTitle, View, Action, Query, Message, Success string
	Fields                                                       []AdminField
	Members                                                      []AdminMemberRow
	Tasks                                                        []AdminTaskRow
	Users                                                        []users.User
	User                                                         *users.User
	Meetings                                                     []content.Meeting
	Publication                                                  *content.Publication
	Page, Pages, Total                                           int
	PreviousURL, NextURL                                         string
}

type AdminGroup struct {
	Title string
	Links []AdminLink
}

func (a AdminPage) Groups() []AdminGroup {
	var groups []AdminGroup
	for _, link := range adminLinks {
		if a.ReviewerOnly && link.ID != "einreichungen" {
			continue
		}
		if len(groups) == 0 || groups[len(groups)-1].Title != link.Group {
			groups = append(groups, AdminGroup{Title: link.Group})
		}
		group := &groups[len(groups)-1]
		group.Links = append(group.Links, link)
	}
	return groups
}

func (a AdminPage) RequestList() RequestListPage {
	return RequestListPage{Requests: a.Requests, Reviewing: true, CanViewUsers: !a.ReviewerOnly}
}

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
	a.CanResendConfirmation = c.identity != nil && a.User != nil && !a.User.Verified
	a.ReviewerOnly = !w.(*sessionResponse).session.CanAdmin()
	p := Page{Title: title, Active: "admin", Admin: a}
	if a.View == "users" && a.Pages > 1 {
		p.DocumentTitle = fmt.Sprintf("Benutzerkonten – Seite %d von %d", a.Page, a.Pages)
	}
	if a.HasErrors() {
		p.DocumentTitle = "Fehler: " + title
	}
	c.renderStatus(w, "admin", false, p, status)
}

func (c *Handler) Admin(w http.ResponseWriter, r *http.Request) {
	if err := c.people.Refresh(r.Context()); err != nil {
		c.serverError(w, err)
		return
	}
	section, id := r.PathValue("section"), r.PathValue("id")
	a := &AdminPage{Section: section, View: "index"}
	if section == "referate" || section == "projektstellen" || section == "aufgaben" {
		if id == "" {
			target := "/verwaltung/funktionen"
			if r.URL.Query().Get("gespeichert") == "1" {
				target += "?gespeichert=1"
			}
			http.Redirect(w, r, target, http.StatusSeeOther)
			return
		}
		a.Section, a.SectionTitle = "funktionen", "Funktionen"
		c.EditResponsibility(w, r, a)
		return
	}
	if section == "" {
		c.adminRender(w, "Verwaltung", a, http.StatusOK)
		return
	}
	if section == "antraege" {
		target := "/verwaltung/einreichungen"
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusTemporaryRedirect)
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
	if r.URL.Query().Get("eingeladen") == "1" {
		a.Success = "Das Mitglied wurde hinzugefügt."
	}
	if r.URL.Query().Get("einladung") == "1" {
		a.Success = "Die Einladung wurde zum Versand vorgemerkt."
	}
	switch section {
	case "einreichungen":
		if c.archive == nil {
			http.NotFound(w, r)
			return
		}
		requests, err := c.archive.Requests(r.Context(), w.(*sessionResponse).session.UserID, true)
		if err != nil {
			c.archiveWriteError(w, r, err)
			return
		}
		a.View, a.Requests = "requests", requests
	case "benutzer":
		directory := c.users
		if c.identity != nil {
			var err error
			directory, err = c.identity.Directory(r.Context())
			if err != nil {
				c.serverError(w, err)
				return
			}
		}
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if id != "" {
			if user, ok := directory.Get(id); ok {
				a.View = "user"
				if c.identity != nil && r.URL.Query().Get("zuruecksetzen") == "1" {
					a.Success = "Die E-Mail zum Zurücksetzen des Passworts wurde zum Versand vorgemerkt."
				}
				a.User = &user
				for _, person := range c.people.People() {
					if person.UserID != 0 && strconv.FormatInt(person.UserID, 10) == user.ID {
						a.Person = &person
						break
					}
				}
				if c.identity != nil && r.URL.Query().Get("bestaetigung") == "1" {
					a.Success = "Die E-Mail zur Bestätigung wurde zum Versand vorgemerkt."
					if a.Person != nil {
						a.Success = "Die Einladung wurde zum Versand vorgemerkt."
					}
				}
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
		page, ok := directory.Search(a.Query, a.Page)
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
	case "mitglieder", "funktionen":
		if id != "" {
			if section == "funktionen" {
				c.EditResponsibility(w, r, a)
			} else {
				c.EditMember(w, r, a)
			}
			return
		}
		a.View = "people"
		if section == "funktionen" {
			a.View = "functions"
		}
		members := c.people.People()
		responsibilities := append(c.people.Responsibilities("board"), c.people.Responsibilities("departments")...)
		responsibilities = append(responsibilities, c.people.Responsibilities("project")...)
		for _, person := range members {
			row := AdminMemberRow{Person: person}
			for _, task := range responsibilities {
				if slices.Contains(task.MemberIDs, person.ID) {
					row.Tasks = append(row.Tasks, task.Title+" ("+responsibilityKind(task.Group)+")")
				}
			}
			a.Members = append(a.Members, row)
		}
		for _, task := range responsibilities {
			var names []string
			for _, person := range members {
				if slices.Contains(task.MemberIDs, person.ID) {
					names = append(names, person.Name)
				}
			}
			a.Tasks = append(a.Tasks, AdminTaskRow{Title: task.Title, Kind: responsibilityKind(task.Group), Members: names, URL: "/verwaltung/funktionen/" + task.ID})
		}
	case "sitzungen":
		if id != "" {
			c.EditPublication(w, r, a)
			return
		}
		a.View = "meetings"
		var err error
		a.Meetings, err = c.content.AllWithError()
		if err != nil {
			c.serverError(w, err)
			return
		}
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

func responsibilityKind(group string) string {
	if group == "board" {
		return "Vorstand"
	}
	if group == "departments" {
		return "Referat"
	}
	return "Projektstelle"
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
	p := people.Person{ID: id}
	found := id == "neu"
	for _, person := range c.people.People() {
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
	if id == "neu" || c.identity != nil && p.UserID == 0 {
		title = "Mitglied einladen"
		a.Inviting = true
	}
	accountEmail := ""
	errors := map[string]string{}
	if r.Method == http.MethodPost {
		p.Name = r.PostForm.Get("name")
		if a.Inviting {
			accountEmail = r.PostForm.Get("email")
			if c.identity == nil {
				http.Error(w, "Die Einladung konnte nicht verarbeitet werden.", http.StatusServiceUnavailable)
				return
			}
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
			if c.identity == nil {
				p.Email = r.PostForm.Get("email")
			}
			errors = c.people.SavePerson(p)
		}
		if len(errors) == 0 {
			http.Redirect(w, r, "/verwaltung/mitglieder?gespeichert=1", http.StatusSeeOther)
			return
		}
	}
	a.View, a.Action = "form", r.URL.Path
	if id != "neu" {
		a.Person = &p
	}
	a.Fields = []AdminField{field("name", "Name", p.Name, "text")}
	if a.Inviting {
		a.Fields = append(a.Fields, field("email", "E-Mail-Adresse des Benutzerkontos", accountEmail, "email"))
		a.Fields[1].Hint = "Besteht bereits ein Benutzerkonto, wird es mit dem Mitglied verknüpft. Andernfalls wird ein Konto angelegt und eine Einladung gesendet."
	} else if c.identity == nil {
		a.Fields = append(a.Fields, field("email", "Öffentliche E-Mail-Adresse (optional)", p.Email, "email"))
	} else {
		a.Fields[0].Hint = "Der Name wird auch im verknüpften Benutzerkonto geändert."
	}
	attachErrors(a.Fields, errors)
	status := http.StatusOK
	if len(errors) != 0 {
		status = http.StatusUnprocessableEntity
	}
	c.adminRender(w, title, a, status)
}

func (c *Handler) EditResponsibility(w http.ResponseWriter, r *http.Request, a *AdminPage) {
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
		http.NotFound(w, r)
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
	a.Fields = []AdminField{field("title", "Bezeichnung", responsibility.Title, "text"), field("group", "Art", responsibility.Group, "radios"), field("email", "Öffentliche E-Mail-Adresse (optional)", responsibility.Email, "email"), field("members", "Mitglieder", "", "checkboxes")}
	a.Fields[1].Options = []AdminOption{{Value: "board", Label: "Vorstand", Selected: responsibility.Group == "board"}, {Value: "departments", Label: "Referat", Selected: responsibility.Group == "departments"}, {Value: "project", Label: "Projektstelle", Selected: responsibility.Group == "project"}}
	for _, p := range c.people.People() {
		a.Fields[3].Options = append(a.Fields[3].Options, AdminOption{p.ID, p.Name, slices.Contains(responsibility.MemberIDs, p.ID)})
	}
	attachErrors(a.Fields, errors)
	status := http.StatusOK
	if len(errors) != 0 {
		status = http.StatusUnprocessableEntity
	}
	c.adminRender(w, title, a, status)
}

func (c *Handler) EditPublication(w http.ResponseWriter, r *http.Request, a *AdminPage) {
	owner := c.publicationOwner(w)
	id := r.PathValue("id")
	p := content.Publication{ID: id, Owner: owner, Date: r.URL.Query().Get("datum")}
	p.Number, _ = strconv.Atoi(r.URL.Query().Get("nummer"))
	if id != "neu" {
		var ok bool
		var err error
		p, ok, err = c.content.GetWithError(id)
		if err != nil {
			c.serverError(w, err)
			return
		}
		if !ok || p.Owner != owner || p.Published {
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
			if id == "neu" {
				p.ID = "neu"
				p.Data = nil
				if errors["file"] == "" {
					errors["file"] = "Die Datei wurde nicht gespeichert. Wähle sie erneut aus."
				}
			}
		} else if len(errors) == 0 {
			http.Redirect(w, r, "/verwaltung/sitzungen/"+p.ID+"/pruefen", http.StatusSeeOther)
			return
		}
	}
	a.View, a.Action, a.Publication = "form", "/verwaltung/sitzungen/"+p.ID, &p
	a.Fields = []AdminField{
		{ID: "date", Label: "Sitzungsdatum", Kind: "date", Day: day, Month: month, Year: year},
		field("number", "Sitzungsnummer", number, "text"),
		{ID: "kind", Label: "Dokumentart", Kind: "radios", Options: []AdminOption{{"Einladung", "Einladung zur Sitzung", p.Kind == "Einladung"}, {"Protokoll", "Sitzungsprotokoll", p.Kind == "Protokoll"}}},
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
	owner := c.publicationOwner(w)
	p, ok, err := c.content.GetWithError(r.PathValue("id"))
	if err != nil {
		c.serverError(w, err)
		return
	}
	if !ok || p.Owner != owner {
		http.NotFound(w, r)
		return
	}
	a := &AdminPage{Section: "sitzungen", SectionTitle: "Sitzungen und Protokolle", View: "review", Publication: &p, Action: r.URL.Path}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		if err := c.content.Publish(p.ID, owner); err != nil {
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
	p, ok, err := c.content.GetWithError(r.PathValue("id"))
	if err != nil {
		c.serverError(w, err)
		return
	}
	s := w.(*sessionResponse).session
	if !ok || !p.Published && (s == nil || !s.Authenticated || c.publicationOwner(w) != p.Owner) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `inline; filename="dokument.pdf"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	http.ServeContent(w, r, "dokument.pdf", time.Time{}, bytes.NewReader(p.Data))
}

func (c *Handler) publicationOwner(w http.ResponseWriter) string {
	s := w.(*sessionResponse).session
	if s == nil || !s.Authenticated {
		return ""
	}
	if c.identity != nil {
		if s.UserID <= 0 {
			return ""
		}
		return strconv.FormatInt(s.UserID, 10)
	}
	return s.ID
}

// ResetUserPassword confirms the destination before queuing a reset email.
func (c *Handler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	directory := c.users
	if c.identity != nil {
		var err error
		directory, err = c.identity.Directory(r.Context())
		if err != nil {
			c.serverError(w, err)
			return
		}
	}
	if user, ok := directory.Get(r.PathValue("id")); ok {
		if r.Method == http.MethodPost {
			if c.identity != nil {
				if err := c.identity.RequestReset(r.Context(), user.Email); err != nil {
					c.serverError(w, err)
					return
				}
				http.Redirect(w, r, "/verwaltung/benutzer/"+url.PathEscape(user.ID)+"?zuruecksetzen=1", http.StatusSeeOther)
				return
			}
			http.Error(w, "Die Anfrage konnte nicht verarbeitet werden. Versuche es später erneut.", http.StatusServiceUnavailable)
			return
		}
		c.adminRender(w, "Passwort zurücksetzen", &AdminPage{Section: "benutzer", SectionTitle: "Benutzer", View: "reset", User: &user}, http.StatusOK)
		return
	}
	http.NotFound(w, r)
}
