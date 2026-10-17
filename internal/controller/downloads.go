package controller

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/fachschaftinformatik/web/internal/model/archive"
	"github.com/fachschaftinformatik/web/internal/model/events"
	"github.com/fachschaftinformatik/web/internal/model/gallery"
	"github.com/fachschaftinformatik/web/internal/storage"
)

func (c *Handler) Document(w http.ResponseWriter, r *http.Request) {
	p, ok, err := c.meetings.GetWithError(r.PathValue("id"))
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

func (c *Handler) ArchiveFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	response := w.(*sessionResponse)
	if response.session == nil || !response.session.Authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	object, err := c.archive.Open(r.Context(), response.session.UserID, r.PathValue("id"))
	if errors.Is(err, archive.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, archive.ErrForbidden) {
		http.Error(w, "Melde dich mit einem aktiven, bestätigten Konto an.", http.StatusForbidden)
		return
	}
	if errors.Is(err, archive.ErrStorage) {
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

func (c *Handler) CalendarCover(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	object, err := c.events.OpenCover(r.Context(), r.PathValue("event"))
	if err != nil {
		if errors.Is(err, events.ErrNotFound) {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Das Titelbild konnte nicht geladen werden.", http.StatusServiceUnavailable)
		}
		return
	}
	c.serveGalleryObject(w, r, object)
}

// MediaFile serves only database-backed gallery images, never raw storage keys.
func (c *Handler) MediaFile(w http.ResponseWriter, r *http.Request) {
	c.galleryImageFile(w, r, false)
}

func (c *Handler) EventCover(w http.ResponseWriter, r *http.Request) {
	c.galleryImageFile(w, r, true)
}

func (c *Handler) galleryImageFile(w http.ResponseWriter, r *http.Request, cover bool) {
	w.Header().Set("Cache-Control", "no-store")

	open := c.gallery.Open
	if cover {
		open = c.gallery.OpenCover
	}
	object, err := open(r.Context(), r.PathValue("id"), r.PathValue("size"))
	if errors.Is(err, gallery.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, gallery.ErrStorage) {
		http.Error(w, "Das Bild konnte nicht geladen werden.", http.StatusServiceUnavailable)
		return
	}
	if err != nil {
		c.serverError(w, err)
		return
	}
	c.serveGalleryObject(w, r, object)
}

func (c *Handler) serveGalleryObject(w http.ResponseWriter, r *http.Request, object *storage.Object) {
	defer object.Close()
	seeker, ok := object.ReadCloser.(io.ReadSeeker)
	if !ok {
		// Providers may return streaming bodies. Spool those to a bounded
		// temporary file so ServeContent still supports HEAD and byte ranges.
		file, err := os.CreateTemp("", "gallery")
		if err != nil {
			c.serverError(w, err)
			return
		}
		defer os.Remove(file.Name())
		defer file.Close()
		written, err := io.Copy(file, io.LimitReader(object, object.Info.Size+1))
		if err != nil || written != object.Info.Size {
			http.Error(w, "Das Bild konnte nicht geladen werden.", http.StatusServiceUnavailable)
			return
		}
		seeker = file
	}
	w.Header().Set("Content-Type", object.Info.ContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Revalidate rather than retaining deleted photographs in a shared cache.
	w.Header().Set("Cache-Control", "public, no-cache")
	http.ServeContent(w, r, "image", time.Time{}, seeker)
}
