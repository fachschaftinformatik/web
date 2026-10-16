package web

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
			items = append(items, Breadcrumb{"Verwaltung", "/verwaltung"})
		}
		if p.Admin.View == "form" || p.Admin.View == "user" || p.Admin.View == "review" || p.Admin.View == "reset" {
			items = append(items, Breadcrumb{p.Admin.SectionTitle, "/verwaltung/" + p.Admin.Section})
		}
		if p.Admin.View == "reset" {
			items = append(items, Breadcrumb{p.Admin.User.Name, "/verwaltung/benutzer/" + p.Admin.User.ID})
		}
	}
	if p.Archive.Module != nil {
		items = append(items, Breadcrumb{"Altklausuren", p.Archive.SearchURL()})
	}
	if p.Fachschaft && p.Active != "fachschaft" {
		items = append(items, Breadcrumb{"Fachschaft", "/fachschaft"})
	}
	if p.Community && p.Active != "community" {
		items = append(items, Breadcrumb{"Community", "/community"})
	}
	if p.Active == "forum" && (p.Forum.Thread != nil || p.Editor != nil) {
		url := "/discussions"
		if p.Forum.BackURL != "" {
			url = p.Forum.BackURL
		}
		items = append(items, Breadcrumb{"Forum", url})
		if p.Editor != nil && p.Editor.Editing {
			items = append(items, Breadcrumb{p.Editor.Draft.Title, p.Editor.BackURL})
		}
	}
	if p.Gallery.Detail {
		items = append(items, Breadcrumb{"Galerie", "/community/events"})
		if p.Gallery.Viewer {
			items = append(items, Breadcrumb{p.Gallery.Album.Title, p.Gallery.Return})
		}
	}
	if p.Active == "register" || p.Active == "forgot" {
		items = append(items, Breadcrumb{"Anmelden", "/login"})
	}
	return items
}
