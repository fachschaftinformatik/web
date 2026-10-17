package web

import "net/http"

func (c *Handler) notFound(w http.ResponseWriter, _ *http.Request) {
	c.pageError(w, "Prüfe die Adresse. Wenn du einem Link auf unserer Website gefolgt bist, gib uns bitte Bescheid.", http.StatusNotFound)
}

func (c *Handler) pageError(w http.ResponseWriter, message string, status int) {
	title := "Die Anfrage konnte nicht verarbeitet werden"
	switch status {
	case http.StatusNotFound:
		title = "Seite nicht gefunden"
	case http.StatusForbidden:
		title = "Zugriff nicht möglich"
	case http.StatusInternalServerError, http.StatusServiceUnavailable:
		title = "Der Dienst ist gerade nicht verfügbar"
	}
	w.Header().Set("Cache-Control", "no-store")
	c.renderStatus(w, "error", false, Page{Title: title, Active: "error", Message: message}, status)
}
