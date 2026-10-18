package view

import (
	"github.com/fachschaftinformatik/web/internal/model/members"
	"github.com/fachschaftinformatik/web/internal/model/organization"
	"strings"
)

type OrganizationSettingsSection struct{ ID, Title, Description string }

var OrganizationSettingsSections = []OrganizationSettingsSection{
	{"identity", "Identität", "Name und Hochschule"},
	{"contact", "Kontakt", "E-Mail, Anschrift und Zuständigkeiten"},
	{"legal", "Rechtliches", "Betrieb, Aufsicht, Datenschutz und Ordnungen"},
	{"links", "Links", "Soziale Kanäle, Wahlen und Selbstverwaltung"},
}

type OrganizationSettingsTask struct {
	ID, Title, Missing string
}

// OrganizationSettingsTasks derives completeness from saved values, rather than visits to
// an edit page. Optional office details do not affect completeness.
func OrganizationSettingsTasks(d organization.Details, roles []members.MemberRole) []OrganizationSettingsTask {
	errors := d.Validate()
	var tasks []OrganizationSettingsTask
	for _, section := range OrganizationSettingsSections {
		task := OrganizationSettingsTask{ID: section.ID, Title: section.Title}
		var missing []string
		for _, field := range OrganizationSettingsFields(section.ID, d, roles) {
			incomplete := errors[field.ID] != ""
			if field.Kind == "radios" {
				assigned := false
				for _, option := range field.Options {
					if option.Selected && option.Value != "" {
						assigned = true
					}
				}
				incomplete = incomplete || !assigned
			}
			// Link forms include one empty row for adding another entry.
			if field.Kind == "links" && ((!field.FixedLinkTitles && len(field.Links) == 1) || (field.FixedLinkTitles && len(field.Links) == 0)) {
				incomplete = true
			}
			if incomplete {
				missing = append(missing, field.Label)
			}
		}
		task.Missing = strings.Join(missing, ", ")
		tasks = append(tasks, task)
	}
	return tasks
}

func OrganizationSettingsFields(section string, d organization.Details, roles []members.MemberRole) []AdminField {
	var fields []AdminField
	switch section {
	case "identity":
		fields = []AdminField{Field("Name", "Name der Website", d.Name, "text"), Field("ShortName", "Kurzbezeichnung", d.ShortName, "text"), Field("FormalName", "Offizielle Bezeichnung", d.FormalName, "text"), Field("Institution", "Hochschule", d.Institution, "text")}
	case "contact":
		fields = []AdminField{Field("ContactEmail", "Allgemeine E-Mail-Adresse", d.ContactEmail, "email"), Field("Street", "Straße und Hausnummer", d.Street, "text"), Field("PostalCode", "Postleitzahl", d.PostalCode, "text"), Field("City", "Ort", d.City, "text"), Field("Country", "Land", d.Country, "text"), Field("Office", "Büroraum (optional)", d.Office, "text")}
		for _, choice := range []struct{ id, label, group, value string }{{"ChairRole", "Funktion für den Vorsitz", "board", d.ChairRole}, {"PublicityRole", "Funktion für Öffentlichkeitsarbeit", "departments", d.PublicityRole}} {
			field := Field(choice.id, choice.label, "", "radios")
			field.Hint = "Die Kontaktdaten werden aus dieser Funktion übernommen."
			field.Options = append(field.Options, AdminOption{Value: "", Label: "Nicht zugeordnet", Selected: choice.value == ""})
			for _, role := range roles {
				if role.Group == choice.group {
					field.Options = append(field.Options, AdminOption{Value: role.PublicID, Label: role.Title, Selected: role.ID == choice.value})
				}
			}
			fields = append(fields, field)
		}
	case "legal":
		fields = []AdminField{Field("HostingName", "Name des technischen Betreibers", d.HostingName, "text"), Field("HostingURL", "Impressum des Betreibers", d.HostingURL, "url"), Field("PrivacyURL", "Datenschutzinformationen des Betreibers", d.PrivacyURL, "url"), Field("SupervisionName", "Name der Rechtsaufsicht", d.SupervisionName, "text"), Field("SupervisionURL", "Website der Rechtsaufsicht", d.SupervisionURL, "url"), Field("PrivacyAuthorityName", "Name der Datenschutzaufsicht", d.PrivacyAuthorityName, "text"), Field("PrivacyAuthorityURL", "Website der Datenschutzaufsicht", d.PrivacyAuthorityURL, "url")}
		fields = append(fields, AdminField{ID: "RegulationLinks", Label: "Satzungen und Ordnungen", Kind: "links", FixedLinkTitles: true, Links: d.RegulationLinks})
	case "links":
		for _, group := range []struct {
			id, label string
			links     []organization.Link
		}{{"SocialLinks", "Soziale Kanäle", d.SocialLinks}, {"ElectionLinks", "Wahlen und Selbstverwaltung", d.ElectionLinks}} {
			fields = append(fields, AdminField{ID: group.id, Label: group.label, Kind: "links", Links: append(append([]organization.Link{}, group.links...), organization.Link{})})
		}
	}
	return fields
}
