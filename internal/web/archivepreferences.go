package web

import (
	"errors"
	"github.com/fachschaftinformatik/web/internal/content/files"
	"net/http"
)

func (c *Handler) preferredArchiveFilter(w http.ResponseWriter, r *http.Request) (files.Filter, error) {
	filter := archiveFilter(r)
	q := r.URL.Query()
	if c.archive == nil || q.Has("applied") || q.Has("program") || q.Has("degree") || q.Has("version") {
		return filter, nil
	}
	session := w.(*sessionResponse).session
	if session == nil || !session.Authenticated {
		return filter, nil
	}
	preferred, err := c.archive.Preferences(r.Context(), session.UserID)
	preferred.Query = filter.Query
	return preferred, err
}

func (c *Handler) ArchivePreferences(w http.ResponseWriter, r *http.Request) {
	if c.archive == nil {
		c.notFound(w, r)
		return
	}
	user := w.(*sessionResponse).session.UserID
	filter, err := c.archive.Preferences(r.Context(), user)
	if err != nil {
		c.archiveWriteError(w, r, err)
		return
	}
	catalog, err := c.archive.Catalog(r.Context())
	if err != nil {
		c.serverError(w, err)
		return
	}
	status := 200
	page := Page{Title: "Voreinstellungen für das Archiv", Active: "archive", Upload: UploadPage{Action: r.URL.Path}}
	if r.Method == http.MethodPost {
		filter = files.Filter{Programs: r.PostForm["program"], Degrees: r.PostForm["degree"], Versions: r.PostForm["version"]}
		if err := c.archive.SavePreferences(r.Context(), user, filter); err != nil {
			var fields files.ValidationErrors
			if !errors.As(err, &fields) {
				c.archiveWriteError(w, r, err)
				return
			}
			page.Upload.Errors = fields
			status = 422
			page.DocumentTitle = "Fehler: " + page.Title
		} else {
			http.Redirect(w, r, "/archive", http.StatusSeeOther)
			return
		}
	}
	page.Upload.Programs = uploadOptions(catalog.Programs(), filter.Programs)
	page.Upload.Degrees = uploadOptions(catalog.Degrees(), filter.Degrees)
	page.Upload.Versions = uploadOptions(catalog.Versions(), filter.Versions)
	c.renderStatus(w, "archivepreferences", false, page, status)
}
