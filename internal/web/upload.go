package web

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/content/files"
)

type UploadPage struct {
	RequestID                                string
	Existing                                 []files.Exam
	ModuleID, Version, Date, Comment, Action string
	Revision                                 bool
	Modules                                  []files.Module
	Versions                                 []string
	Errors                                   map[string]string
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
	page := Page{Title: "Reko einreichen", Active: "archive", Upload: UploadPage{Action: r.URL.Path, Modules: catalog.Modules(), Versions: catalog.Versions()}}
	form := &page.Upload
	userID := w.(*sessionResponse).session.UserID
	revisionOf := r.PathValue("id")
	requestID := r.PathValue("request")
	if requestID != "" {
		request, err := c.archive.Request(r.Context(), userID, requestID, false)
		if err != nil {
			c.archiveWriteError(w, r, err)
			return
		}
		if request.Status != files.Draft {
			c.archiveWriteError(w, r, files.ErrConflict)
			return
		}
		form.RequestID, form.Existing = request.ID, request.Documents
		form.ModuleID, form.Version, form.Comment, form.Date = request.ModuleID, request.Version, request.Comment, request.DateLabel()
		revisionOf = request.RevisionOf
		form.Revision = revisionOf != ""
	}
	if revisionOf != "" && requestID == "" {
		base, err := c.archive.Revision(r.Context(), userID, revisionOf)
		if err != nil {
			c.archiveWriteError(w, r, err)
			return
		}
		form.Revision, form.ModuleID, form.Version, form.Comment = true, base.ModuleID, base.Version, base.Comment
		if date, err := time.Parse("2006-01-02", base.Date); err == nil {
			form.Date = date.Format("02.01.2006")
		}
		page.Title = "Neue Revision einreichen"
	}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		form.ModuleID = strings.TrimSpace(r.PostForm.Get("module_id"))
		form.Version = strings.TrimSpace(r.PostForm.Get("version"))
		form.Date = strings.TrimSpace(r.PostForm.Get("date"))
		form.Comment = strings.TrimSpace(r.PostForm.Get("comment"))
		submission := files.Submission{ModuleID: form.ModuleID, Version: form.Version, Comment: form.Comment, RevisionOf: revisionOf, RequestID: requestID}
		if date, err := time.Parse("02.01.2006", form.Date); err == nil {
			submission.Date = date.Format("2006-01-02")
		}
		form.Errors = map[string]string{}
		total := 0
		if r.MultipartForm != nil {
			for _, header := range r.MultipartForm.File["files"] {
				if header.Size > content.MaxPDFBytes {
					form.Errors["files"] = "Die PDF-Datei darf höchstens 8 MB groß sein."
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
					form.Errors["files"] = "Die PDF-Dateien dürfen zusammen höchstens 32 MB groß sein."
					break
				}
				submission.Documents = append(submission.Documents, value)
			}
		}
		if len(form.Errors) == 0 {
			result, err := c.archive.SaveDraft(r.Context(), userID, submission)
			var fields files.ValidationErrors
			if errors.As(err, &fields) {
				form.Errors = fields
			} else if err != nil {
				c.archiveWriteError(w, r, err)
				return
			} else {
				http.Redirect(w, r, "/archive/requests/"+result.RequestID+"/review", http.StatusSeeOther)
				return
			}
		}
		status, page.DocumentTitle = http.StatusUnprocessableEntity, "Fehler: "+page.Title
	}
	c.renderStatus(w, "upload", false, page, status)
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
