package web

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/fachschaftinformatik/web/internal/content/files"
)

func (c *Handler) ArchiveFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	response := w.(*sessionResponse)
	if response.session == nil || !response.session.Authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if c.archive == nil {
		http.NotFound(w, r)
		return
	}
	object, err := c.archive.Open(r.Context(), r.PathValue("id"))
	if errors.Is(err, files.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, files.ErrStorage) {
		log.Printf("Archive storage: %v", err)
		http.Error(w, "Die Datei konnte nicht geladen werden. Versuche es später erneut.", http.StatusServiceUnavailable)
		return
	}
	if err != nil {
		c.serverError(w, err)
		return
	}
	defer object.Close()
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", `inline; filename="reko.pdf"`)
	if seeker, ok := object.ReadCloser.(io.ReadSeeker); ok {
		http.ServeContent(w, r, "reko.pdf", time.Time{}, seeker)
		return
	}
	w.Header().Set("Content-Length", strconv.FormatInt(object.Info.Size, 10))
	if r.Method == http.MethodHead {
		return
	}
	if _, err := io.CopyN(w, object, object.Info.Size); err != nil {
		log.Printf("Archive download: %v", err)
	}
}
