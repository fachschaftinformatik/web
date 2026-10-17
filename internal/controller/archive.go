package controller

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/model/archive"
	"github.com/fachschaftinformatik/web/internal/view"
)

func (c *Handler) preferredArchiveFilter(w http.ResponseWriter, r *http.Request) (archive.Filter, error) {
	filter := archiveFilter(r)
	q := r.URL.Query()
	if q.Has("applied") || q.Has("program") || q.Has("degree") || q.Has("version") {
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
	page := view.Page{Title: "Voreinstellungen für das Archiv", Active: "archive", Upload: view.UploadPage{Action: r.URL.Path}}
	if r.Method == http.MethodPost {
		filter = archive.Filter{Programs: r.PostForm["program"], Degrees: r.PostForm["degree"], Versions: r.PostForm["version"]}
		if err := c.archive.SavePreferences(r.Context(), user, filter); err != nil {
			var fields archive.ValidationErrors
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
	page.Upload.Programs = view.UploadOptions(catalog.Programs(), filter.Programs)
	page.Upload.Degrees = view.UploadOptions(catalog.Degrees(), filter.Degrees)
	page.Upload.Versions = view.UploadOptions(catalog.Versions(), filter.Versions)
	c.renderStatus(w, "archivepreferences", false, page, status)
}

func archiveFilter(r *http.Request) archive.Filter {
	q := r.URL.Query()
	return archive.Filter{Query: q.Get("q"), Programs: q["program"], Versions: slices.DeleteFunc(q["version"], func(version string) bool { return version == "" }), Degrees: q["degree"]}
}

func (c *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	filter, err := c.preferredArchiveFilter(w, r)
	if err != nil {
		c.serverError(w, err)
		return
	}
	pageNumber, _ := strconv.Atoi(r.URL.Query().Get("seite"))
	pageNumber = max(1, pageNumber)

	catalog, total, err := c.archive.CatalogPage(r.Context(), filter, pageNumber, view.ArchivePageSize)

	if err != nil {
		c.serverError(w, err)
		return
	}
	programs := make([]view.FilterOption, 0)
	for _, program := range catalog.Programs() {
		programs = append(programs, view.FilterOption{Name: program, Selected: slices.Contains(filter.Programs, program)})
	}
	versions := make([]view.FilterOption, 0)
	for _, version := range catalog.Versions() {
		versions = append(versions, view.FilterOption{Name: version, Selected: slices.Contains(filter.Versions, version)})
	}
	degrees := make([]view.FilterOption, 0)
	for _, degree := range catalog.Degrees() {
		degrees = append(degrees, view.FilterOption{Name: degree, Selected: slices.Contains(filter.Degrees, degree)})
	}
	data := view.Page{
		Title: "Altklausuren", Active: "archive",
		Archive: view.ArchivePage{Filter: filter, Programs: programs, Versions: versions, Degrees: degrees, Modules: catalog.FindModules(filter)},
	}

	data.Archive.Total = total
	data.Archive.Pages = max(1, (total+view.ArchivePageSize-1)/view.ArchivePageSize)
	data.Archive.Page = min(pageNumber, data.Archive.Pages)

	if raw := r.URL.Query().Get("seite"); raw != "" && raw != fmt.Sprint(data.Archive.Page) {
		http.Redirect(w, r, data.Archive.PageURL(1), http.StatusSeeOther)
		return
	}
	if data.Archive.Pages > 1 {
		data.DocumentTitle = fmt.Sprintf("Altklausuren (Seite %d von %d)", data.Archive.Page, data.Archive.Pages)
	}
	w.Header().Set("Vary", "HX-Request, HX-History-Restore-Request")
	fragment := r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-History-Restore-Request") != "true"
	c.render(w, "archive", fragment, data)
}

func (c *Handler) ArchiveModule(w http.ResponseWriter, r *http.Request) {
	filter, err := c.preferredArchiveFilter(w, r)
	if err != nil {
		c.serverError(w, err)
		return
	}

	module, total, err := c.archive.ModulePage(r.Context(), r.PathValue("module"), filter, requestedPage(r), 20)
	if errors.Is(err, archive.ErrNotFound) {
		c.notFound(w, r)
		return
	}
	if err != nil {
		c.serverError(w, err)
		return
	}
	c.render(w, "module", false, view.Page{Title: module.Title, Active: "archive", Archive: view.ArchivePage{Module: &module, Filter: filter, ListPagination: listPagination(r, total, 20)}})
}

func requestedPage(r *http.Request) int {
	n, _ := strconv.Atoi(r.URL.Query().Get("page"))
	return max(1, n)
}

func listPagination(r *http.Request, total, size int) view.ListPagination {
	pages := max(1, (total+size-1)/size)
	return view.ListPagination{Page: min(requestedPage(r), pages), Pages: pages, Total: total, Path: r.URL.Path, Query: r.URL.Query()}
}
