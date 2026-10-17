package web

import (
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/content/files"
)

type UploadPage struct {
	RequestID, Date, Comment, Action string
	Existing                         []files.Exam
	Revision, ChooseScope            bool
	ReviewEditing                    bool
	EditToken                        string
	Modules                          []string
	DateInput                        archiveDate
	Scope                            files.Filter
	Programs, Degrees, Versions      []FilterOption
	Groups                           []AssignmentGroup
	Errors                           map[string]string
}

func (u UploadPage) ErrorFields() []string {
	var fields []string
	for _, field := range []string{"degree", "program", "version", "module_id", "date", "comment", "files"} {
		if u.Errors[field] != "" {
			fields = append(fields, field)
		}
	}
	return fields
}

func (c *Handler) UploadArchive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if c.archive == nil {
		http.NotFound(w, r)
		return
	}
	catalog, err := c.archive.Catalog(r.Context())
	if err != nil {
		c.serverError(w, err)
		return
	}
	page := Page{Title: "Reko einreichen", Active: "archive", Upload: UploadPage{Action: r.URL.Path, Errors: map[string]string{}}}
	form := &page.Upload
	form.ReviewEditing = strings.HasPrefix(r.URL.Path, "/verwaltung/einreichungen/")
	if form.ReviewEditing {
		page.Title, page.Active = "Einreichung bearbeiten", "admin"
		page.Admin = &AdminPage{Section: "einreichungen", SectionTitle: "Einreichungen", View: "review", ReviewerOnly: !w.(*sessionResponse).session.CanAdmin()}
	}
	userID := w.(*sessionResponse).session.UserID
	revisionOf, requestID := r.PathValue("id"), r.PathValue("request")
	var selected []files.Assignment
	if requestID != "" {
		request, err := c.archive.Request(r.Context(), userID, requestID, form.ReviewEditing)
		if err != nil {
			c.archiveWriteError(w, r, err)
			return
		}
		if !form.ReviewEditing && request.Status != files.Draft {
			c.archiveWriteError(w, r, files.ErrConflict)
			return
		}
		form.RequestID, form.Existing = request.ID, request.Documents
		form.EditToken = request.EditToken
		form.Date, form.Comment = request.DateLabel(), request.Comment
		selected, revisionOf = request.Assignments, request.RevisionOf
	} else if revisionOf != "" {
		base, err := c.archive.Revision(r.Context(), userID, revisionOf)
		if err != nil {
			c.archiveWriteError(w, r, err)
			return
		}
		selected = base.Assignments
		form.Comment = base.Comment
		if date, err := time.Parse("2006-01-02", base.Date); err == nil {
			form.Date = date.Format("02.01.2006")
		}
		page.Title = "Neue Revision einreichen"
	}
	form.Revision = revisionOf != ""
	for _, a := range selected {
		form.Scope.Programs = uniqueChoice(form.Scope.Programs, a.Program)
		form.Scope.Degrees = uniqueChoice(form.Scope.Degrees, a.Degree)
		form.Scope.Versions = uniqueChoice(form.Scope.Versions, a.Version)
	}
	if len(selected) == 0 {
		preferred, err := c.archive.Preferences(r.Context(), userID)
		if err != nil {
			c.archiveWriteError(w, r, err)
			return
		}
		form.Scope = preferred
	}
	query := r.URL.Query()
	if form.ReviewEditing {
		if r.Method == http.MethodPost {
			form.EditToken = r.PostForm.Get("edit_token")
		} else if query.Has("edit_token") {
			form.EditToken = query.Get("edit_token")
		}
	}
	explicitScope := query.Get("scope") == "1" || r.Method == http.MethodPost
	if explicitScope {
		values := query
		if r.Method == http.MethodPost {
			values = r.PostForm
		}
		form.Scope = files.Filter{Programs: values["program"], Degrees: values["degree"], Versions: values["version"]}
	}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		form.Date, form.Comment = strings.TrimSpace(r.PostForm.Get("date")), strings.TrimSpace(r.PostForm.Get("comment"))
		selected = selectedAssignments(r.PostForm["assignment"])
		// Keep older clients functional; the rendered form selects module titles.
		if r.PostForm.Get("scope") == "" && r.PostForm.Get("module_id") != "" {
			selected = []files.Assignment{{ModuleID: r.PostForm.Get("module_id"), Version: r.PostForm.Get("version")}}
			for _, m := range catalog.Modules() {
				if m.ID == selected[0].ModuleID {
					program, degree := files.ProgramName(m.Program)
					form.Scope = files.Filter{Programs: []string{program}, Degrees: []string{degree}, Versions: []string{selected[0].Version}}
				}
			}
		}
	}
	if date, err := time.Parse("02.01.2006", form.Date); err == nil {
		form.DateInput = archiveDate{Day: date.Format("2"), Month: date.Format("1"), Year: date.Format("2006")}
	}
	if r.Method == http.MethodPost && (r.PostForm.Has("date-day") || r.PostForm.Has("date-month") || r.PostForm.Has("date-year")) {
		form.DateInput = archiveDate{Day: strings.TrimSpace(r.PostForm.Get("date-day")), Month: strings.TrimSpace(r.PostForm.Get("date-month")), Year: strings.TrimSpace(r.PostForm.Get("date-year"))}
	}
	form.ChooseScope = (len(selected) == 0 && !explicitScope) || query.Get("choose_scope") == "1"
	if explicitScope {
		for _, item := range []struct {
			field         string
			values, known []string
			message       string
		}{
			{"program", form.Scope.Programs, catalog.Programs(), "Wähle mindestens einen Studiengang aus."},
			{"degree", form.Scope.Degrees, catalog.Degrees(), "Wähle mindestens einen Abschluss aus."},
			{"version", form.Scope.Versions, catalog.Versions(), "Wähle mindestens eine Prüfungsordnung aus."},
		} {
			if len(item.values) == 0 || slices.ContainsFunc(item.values, func(value string) bool { return !slices.Contains(item.known, value) }) {
				form.Errors[item.field] = item.message
			}
		}
		if len(form.Errors) > 0 {
			form.ChooseScope = true
			status = 422
		}
	}
	form.Programs = uploadOptions(catalog.Programs(), form.Scope.Programs)
	form.Degrees = uploadOptions(catalog.Degrees(), form.Scope.Degrees)
	form.Versions = uploadOptions(catalog.Versions(), form.Scope.Versions)
	var allowed map[string]files.Assignment
	form.Groups, allowed = uploadGroups(catalog, form.Scope, selected)
	if !form.ChooseScope && len(form.Groups) == 0 {
		form.ChooseScope = true
		form.Errors["version"] = "Für diese Auswahl gibt es keine passenden Module. Ändere die Auswahl."
		status = 422
	}
	values := query
	if r.Method == http.MethodPost {
		values = r.PostForm
	}
	form.Modules = values["module"]
	moduleSelection := r.Method == http.MethodPost && (values.Get("module_selection") == "1" || values.Has("module"))
	if len(form.Modules) == 0 && len(selected) > 0 && !moduleSelection {
		for _, group := range form.Groups {
			if group.Selected {
				form.Modules = append(form.Modules, group.Title)
			}
		}
	}
	for i := range form.Groups {
		form.Groups[i].Selected = slices.Contains(form.Modules, form.Groups[i].Title)
	}
	if moduleSelection {
		selected = nil
		for _, title := range form.Modules {
			if !slices.ContainsFunc(form.Groups, func(g AssignmentGroup) bool { return g.Title == title }) {
				form.Errors["module_id"] = "Wähle gültige Module aus."
			}
		}
		// Expand only catalogue-backed combinations, never the Cartesian product
		// of the independent scope fields. Iterating groups keeps review stable.
		for _, group := range form.Groups {
			if !group.Selected {
				continue
			}
			for _, degree := range group.Degrees {
				for _, program := range degree.Programs {
					for _, choice := range program.Choices {
						selected = append(selected, allowed[choice.Value])
					}
				}
			}
		}
	}
	if r.Method == http.MethodPost && !form.ChooseScope {
		submission := files.Submission{Comment: form.Comment, RevisionOf: revisionOf, RequestID: requestID}
		var dateError string
		submission.Date, dateError = form.DateInput.parse()
		if dateError != "" {
			form.Errors["date"] = dateError
		}
		if len(selected) == 0 || !moduleSelection && len(selected) != len(r.PostForm["assignment"]) && r.PostForm.Get("scope") != "" {
			if form.Errors["module_id"] == "" {
				form.Errors["module_id"] = "Wähle mindestens ein passendes Modul aus."
			}
		}
		for _, a := range selected {
			if value, ok := allowed[assignmentKey(a)]; ok {
				submission.Assignments = append(submission.Assignments, value)
			} else {
				form.Errors["module_id"] = "Wähle gültige Module und dazugehörige Prüfungsordnungen aus."
			}
		}
		total := 0
		uploaded := false
		if r.MultipartForm != nil {
			for _, header := range r.MultipartForm.File["files"] {
				uploaded = true
				if header.Size > content.MaxPDFBytes {
					form.Errors["files"] = "Die Datei darf höchstens 8 MB groß sein."
					break
				}
				file, err := header.Open()
				if err != nil {
					c.serverError(w, err)
					return
				}
				value, err := io.ReadAll(io.LimitReader(file, content.MaxPDFBytes+1))
				file.Close()
				if err != nil {
					c.serverError(w, err)
					return
				}
				total += len(value)
				if total > files.MaxUploadBytes {
					form.Errors["files"] = "Wähle genau eine Datei aus."
					break
				}
				submission.Documents = append(submission.Documents, value)
			}
		}
		for field, message := range submission.Validate() {
			if field == "version" && len(form.Scope.Versions) > 0 {
				continue
			}
			if form.Errors[field] == "" {
				form.Errors[field] = message
			}
		}
		if len(form.Errors) == 0 {
			var result files.UploadResult
			var err error
			if form.ReviewEditing {
				result, err = c.archive.EditRequest(r.Context(), userID, submission, form.EditToken)
			} else {
				result, err = c.archive.SaveDraft(r.Context(), userID, submission)
			}
			var fields files.ValidationErrors
			if errors.As(err, &fields) {
				form.Errors = fields
			} else if err != nil {
				c.archiveWriteError(w, r, err)
				return
			} else {
				next := "/archive/requests/" + result.RequestID + "/review"
				if form.ReviewEditing {
					next = "/verwaltung/einreichungen/" + result.RequestID
				}
				if r.Header.Get("HX-Request") == "true" {
					w.Header().Set("HX-Redirect", next)
					w.WriteHeader(200)
					return
				}
				http.Redirect(w, r, next, http.StatusSeeOther)
				return
			}
		}
		if uploaded && form.Errors["files"] == "" && r.Header.Get("HX-Request") != "true" {
			form.Errors["files"] = "Wähle die Datei erneut aus. Deine übrigen Angaben sind erhalten geblieben."
		}
		status = 422
	}
	if status == 422 {
		page.DocumentTitle = "Fehler: " + page.Title
	}
	c.renderStatus(w, "upload", false, page, status)
}

func uniqueChoice(values []string, value string) []string {
	if value != "" && !slices.Contains(values, value) {
		return append(values, value)
	}
	return values
}

func (c *Handler) archiveWriteError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, files.ErrConflict) {
		http.Error(w, files.ErrConflict.Error(), http.StatusConflict)
		return
	}
	if errors.Is(err, files.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, files.ErrForbidden) {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if errors.Is(err, files.ErrStorage) {
		http.Error(w, "Die Dateien konnten nicht gespeichert werden. Versuche es später erneut.", http.StatusServiceUnavailable)
		return
	}
	c.serverError(w, err)
}
