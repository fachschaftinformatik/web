package site

import (
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/urls"
	"github.com/fachschaftinformatik/web/model/exams"
	"github.com/fachschaftinformatik/web/model/meetings"
	"github.com/fachschaftinformatik/web/view"
)

func selectedAssignments(values []string) []exams.Assignment {
	var result []exams.Assignment
	for _, value := range values {
		module, version, ok := strings.Cut(value, "|")
		if ok {
			result = append(result, exams.Assignment{ModuleID: module, Version: version})
		}
	}
	return result
}

func (c *Handler) ExamsSubmissions(w http.ResponseWriter, r *http.Request) {

	submissions, err := c.exams.Submissions(r.Context(), w.(*sessionResponse).session.UserID, false)
	if err != nil {
		c.examsWriteError(w, r, err)
		return
	}
	c.render(w, "submissions", false, view.Page{Title: "Meine Einreichungen", Active: "exams", SubmissionList: view.SubmissionListPage{Submissions: view.Submissions(submissions)}})
}

func (c *Handler) ExamsSubmission(w http.ResponseWriter, r *http.Request) {

	userID := w.(*sessionResponse).session.UserID
	submission, err := c.exams.Submission(r.Context(), userID, r.PathValue("submission"), false)
	if err != nil {
		c.examsWriteError(w, r, err)
		return
	}
	if r.Method == http.MethodPost {
		if err := c.exams.Submit(r.Context(), userID, submission.ID); err != nil {
			c.examsWriteError(w, r, err)
			return
		}
		http.Redirect(w, r, urls.Submission(submission.ID), http.StatusSeeOther)
		return
	}
	title := "Deine Einreichung"
	if submission.Status == exams.Draft {
		title = "Prüfe deine Angaben"
	}
	if submission.Status == exams.Pending {
		title = "Einreichung eingegangen"
	}
	c.render(w, "submission", false, view.Page{Title: title, Active: "exams", Submission: &view.SubmissionPage{Entry: view.NewSubmission(submission), EditURL: urls.EditSubmission(submission.ID)}})
}

func (c *Handler) ReviewExamsSubmission(w http.ResponseWriter, r *http.Request) {

	session := w.(*sessionResponse).session
	submission, err := c.exams.Submission(r.Context(), session.UserID, r.PathValue("submission"), true)
	if err != nil {
		c.examsWriteError(w, r, err)
		return
	}
	page := view.Page{Title: "Einreichung prüfen", Active: "admin", Admin: &view.AdminPage{Section: "submissions", SectionTitle: "Einreichungen", View: "review", ReviewerOnly: !session.CanAdmin()}, Submission: &view.SubmissionPage{Entry: view.NewSubmission(submission), Reviewing: true, EditURL: urls.AdminSubmission(submission.ID) + "/edit"}}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		decision := r.PostForm.Get("decision")
		if decision != "approve" && decision != "reject" {
			c.pageError(w, "Wähle Freigeben oder Ablehnen aus.", http.StatusBadRequest)
			return
		}
		err := c.exams.ReviewVersion(r.Context(), session.UserID, submission.ID, decision == "approve", r.PostForm.Get("edit_token"))
		if errors.Is(err, exams.ErrConflict) {
			page.Submission.Error = "Die Einreichung oder die bisherige Revision wurde bereits geändert. Lade die Seite neu und prüfe den aktuellen Stand."
			page.DocumentTitle = "Fehler: Einreichung prüfen"
			status = http.StatusConflict
		} else {
			if err != nil {
				c.examsWriteError(w, r, err)
				return
			} else {
				http.Redirect(w, r, urls.AdminSubmission(submission.ID), http.StatusSeeOther)
				return
			}
		}
	}
	c.renderStatus(w, "submission", false, page, status)
}

func (c *Handler) UploadExams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	catalog, err := c.exams.Catalog(r.Context())
	if err != nil {
		c.serverError(w, err)
		return
	}
	page := view.Page{Title: "Altklausur einreichen", Active: "exams", Upload: view.UploadPage{Action: r.URL.Path, Errors: map[string]string{}}}
	form := &page.Upload
	form.ReviewEditing = r.PathValue("section") == "submissions"
	if form.ReviewEditing {
		page.Title, page.Active = "Einreichung bearbeiten", "admin"
		page.Admin = &view.AdminPage{Section: "submissions", SectionTitle: "Einreichungen", View: "review", ReviewerOnly: !w.(*sessionResponse).session.CanAdmin()}
	}
	userID := w.(*sessionResponse).session.UserID
	revisionOf, submissionID := r.URL.Query().Get("revision"), r.PathValue("submission")
	if r.Method == http.MethodPost {
		revisionOf = r.PostForm.Get("revision")
	}
	var selected []exams.Assignment
	if submissionID != "" {
		submission, err := c.exams.Submission(r.Context(), userID, submissionID, form.ReviewEditing)
		if err != nil {
			c.examsWriteError(w, r, err)
			return
		}
		if !form.ReviewEditing && submission.Status != exams.Draft {
			c.examsWriteError(w, r, exams.ErrConflict)
			return
		}
		form.SubmissionID, form.Existing = submission.ID, view.Exams(submission.Documents)
		form.EditToken = submission.EditToken
		form.Date, form.Comment = view.NewSubmission(submission).DateLabel(), submission.Comment
		selected, revisionOf = submission.Assignments, submission.RevisionOf
	} else {
		if revisionOf != "" {
			base, err := c.exams.Revision(r.Context(), userID, revisionOf)
			if err != nil {
				c.examsWriteError(w, r, err)
				return
			}
			selected = base.Assignments
			form.Comment = base.Comment
			if date, err := time.Parse("2006-01-02", base.Date); err == nil {
				form.Date = date.Format("02.01.2006")
			}
			page.Title = "Neue Revision einreichen"
		}
	}
	form.Revision = revisionOf != ""
	form.RevisionID = revisionOf
	form.Scope = exams.Scope(selected)
	if len(selected) == 0 {
		preferred, err := c.exams.Settings(r.Context(), userID)
		if err != nil {
			c.examsWriteError(w, r, err)
			return
		}
		form.Scope = preferred
	}
	query := r.URL.Query()
	if form.ReviewEditing && r.Method == http.MethodPost {
		form.EditToken = r.PostForm.Get("edit_token")
	}
	explicitScope := query.Has("program") || query.Has("degree") || query.Has("version") || r.Method == http.MethodPost
	if explicitScope {
		values := query
		if r.Method == http.MethodPost {
			values = r.PostForm
		}
		form.Scope = exams.Filter{Programs: slices.DeleteFunc(values["program"], func(value string) bool { return value == "all" || value == "" }), Degrees: values["degree"], Versions: values["version"]}
	}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		form.Date, form.Comment = strings.TrimSpace(r.PostForm.Get("date")), strings.TrimSpace(r.PostForm.Get("comment"))
		selected = selectedAssignments(r.PostForm["assignment"])
	}
	if date, err := time.Parse("02.01.2006", form.Date); err == nil {
		form.DateInput = exams.DateInput{Day: date.Format("2"), Month: date.Format("1"), Year: date.Format("2006")}
	}
	if r.Method == http.MethodPost && (r.PostForm.Has("date-day") || r.PostForm.Has("date-month") || r.PostForm.Has("date-year")) {
		form.DateInput = exams.DateInput{Day: strings.TrimSpace(r.PostForm.Get("date-day")), Month: strings.TrimSpace(r.PostForm.Get("date-month")), Year: strings.TrimSpace(r.PostForm.Get("date-year"))}
	}
	form.ChooseScope = len(selected) == 0 && !explicitScope
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
	form.Programs = view.UploadOptions(catalog.Programs(), form.Scope.Programs)
	form.Degrees = view.UploadOptions(catalog.Degrees(), form.Scope.Degrees)
	form.Versions = view.UploadOptions(catalog.Versions(), form.Scope.Versions)
	for i := range form.Programs {
		form.Programs[i].Label = catalog.ProgramLabel(form.Programs[i].Name)
	}
	for i := range form.Degrees {
		form.Degrees[i].Label = exams.DegreeLabel(form.Degrees[i].Name)
	}
	available := catalog.Assignments(form.Scope)
	allowed := make(map[string]exams.Assignment, len(available))
	for _, a := range available {
		allowed[a.Key()] = a
	}
	form.Groups = view.UploadGroups(available, selected)
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
		var fields exams.ValidationErrors
		selected, fields = exams.ResolveModules(available, form.Modules)
		for field, message := range fields {
			form.Errors[field] = message
		}
	}

	if r.Method == http.MethodPost && !form.ChooseScope {
		submission := exams.SubmissionDraft{Comment: form.Comment, RevisionOf: revisionOf, SubmissionID: submissionID}
		var dateError string
		submission.Date, dateError = form.DateInput.Parse()
		if dateError != "" {
			form.Errors["date"] = dateError
		}
		if len(selected) == 0 || !moduleSelection && len(selected) != len(r.PostForm["assignment"]) {
			if form.Errors["module_id"] == "" {
				form.Errors["module_id"] = "Wähle mindestens ein passendes Modul aus."
			}
		}
		for _, a := range selected {
			if value, ok := allowed[a.Key()]; ok {
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
				if header.Size > meetings.MaxPDFBytes {
					form.Errors["files"] = "Die Datei darf höchstens 8 MB groß sein."
					break
				}
				file, err := header.Open()
				if err != nil {
					c.serverError(w, err)
					return
				}
				value, err := io.ReadAll(io.LimitReader(file, meetings.MaxPDFBytes+1))
				file.Close()
				if err != nil {
					c.serverError(w, err)
					return
				}
				total += len(value)
				if total > exams.MaxUploadBytes {
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
			var result exams.UploadResult
			var err error
			if form.ReviewEditing {
				result, err = c.exams.EditSubmission(r.Context(), userID, submission, form.EditToken)
			} else {
				result, err = c.exams.SaveDraft(r.Context(), userID, submission)
			}
			var fields exams.ValidationErrors
			if errors.As(err, &fields) {
				form.Errors = fields
			} else {
				if err != nil {
					c.examsWriteError(w, r, err)
					return
				} else {
					next := urls.CheckSubmission(result.SubmissionID)
					if form.ReviewEditing {
						next = urls.AdminSubmission(result.SubmissionID)
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

func (c *Handler) examsWriteError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, exams.ErrConflict) {
		c.pageError(w, exams.ErrConflict.Error(), http.StatusConflict)
		return
	}
	if errors.Is(err, exams.ErrNotFound) {
		c.notFound(w, r)
		return
	}
	if errors.Is(err, exams.ErrForbidden) {
		c.pageError(w, err.Error(), http.StatusForbidden)
		return
	}
	if errors.Is(err, exams.ErrStorage) {
		c.pageError(w, "Die Dateien konnten nicht gespeichert werden. Versuche es später erneut.", http.StatusServiceUnavailable)
		return
	}
	c.serverError(w, err)
}
