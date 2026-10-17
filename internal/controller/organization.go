package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/fachschaftinformatik/web/internal/model/organization"
	"github.com/fachschaftinformatik/web/internal/view"
)

func (c *Handler) MasterData(w http.ResponseWriter, r *http.Request, a *view.AdminPage) {
	id := r.PathValue("id")
	if id == "" {
		a.View = "master-data"
		roles := append(c.people.Responsibilities("board"), c.people.Responsibilities("departments")...)
		a.MasterData = view.MasterDataTasks(c.organization.Current(), roles)
		c.adminRender(w, "Stammdaten", a, http.StatusOK)
		return
	}
	title := ""
	for _, section := range view.MasterDataSections {
		if section.ID == id {
			title = section.Title
		}
	}
	if title == "" {
		c.notFound(w, r)
		return
	}
	d := c.organization.Current()
	fields := map[string]string{}
	if r.Method == http.MethodPost {
		get := func(key string) string { return strings.TrimSpace(r.PostForm.Get(key)) }
		switch id {
		case "identitaet":
			d.Name, d.ShortName, d.FormalName, d.Institution = get("Name"), get("ShortName"), get("FormalName"), get("Institution")
		case "kontakt":
			d.ContactEmail, d.Street, d.PostalCode, d.City, d.Country, d.Office = get("ContactEmail"), get("Street"), get("PostalCode"), get("City"), get("Country"), get("Office")
			d.ChairResponsibility, d.PublicityResponsibility = get("ChairResponsibility"), get("PublicityResponsibility")
		case "rechtliches":
			d.HostingName, d.HostingURL, d.PrivacyURL = get("HostingName"), get("HostingURL"), get("PrivacyURL")
			d.SupervisionName, d.SupervisionURL = get("SupervisionName"), get("SupervisionURL")
			d.PrivacyAuthorityName, d.PrivacyAuthorityURL = get("PrivacyAuthorityName"), get("PrivacyAuthorityURL")
			urls := r.PostForm["RegulationLinksURL"]
			for i := range d.RegulationLinks {
				d.RegulationLinks[i].URL = ""
				if i < len(urls) {
					d.RegulationLinks[i].URL = strings.TrimSpace(urls[i])
				}
			}
		case "verknuepfungen":
			d.SocialLinks = parseLinks(r, "SocialLinks")
			d.ElectionLinks = parseLinks(r, "ElectionLinks")
		}
		var err error
		fields, err = c.organization.Save(r.Context(), w.(*sessionResponse).session.UserID, d)
		if errors.Is(err, organization.ErrForbidden) {
			c.pageError(w, "Du hast keine Berechtigung für die Stammdaten.", http.StatusForbidden)
			return
		}
		if err != nil {
			c.serverError(w, err)
			return
		}
		if len(fields) == 0 {
			http.Redirect(w, r, "/verwaltung/stammdaten?gespeichert=1", http.StatusSeeOther)
			return
		}
	}
	roles := append(c.people.Responsibilities("board"), c.people.Responsibilities("departments")...)
	a.View, a.Action = "form", r.URL.Path
	a.Fields = view.MasterDataFields(id, d, roles)
	view.AttachErrors(a.Fields, fields)
	status := http.StatusOK
	if len(fields) > 0 {
		status = http.StatusUnprocessableEntity
	}
	c.adminRender(w, title, a, status)
}

func parseLinks(r *http.Request, key string) []organization.Link {
	titles, urls := r.PostForm[key+"Title"], r.PostForm[key+"URL"]
	var links []organization.Link
	for i := 0; i < max(len(titles), len(urls)); i++ {
		var link organization.Link
		if i < len(titles) {
			link.Title = strings.TrimSpace(titles[i])
		}
		if i < len(urls) {
			link.URL = strings.TrimSpace(urls[i])
		}
		if link.Title != "" || link.URL != "" {
			links = append(links, link)
		}
	}
	return links
}
