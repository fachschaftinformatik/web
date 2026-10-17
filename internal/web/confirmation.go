package web

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
)

func (c *Handler) ResendUserConfirmation(w http.ResponseWriter, r *http.Request) {
	if c.identity == nil {
		http.NotFound(w, r)
		return
	}
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	queued, err := c.identity.ResendConfirmation(r.Context(), userID)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		c.serverError(w, err)
		return
	}
	target := "/verwaltung/benutzer/" + strconv.FormatInt(userID, 10)
	if queued {
		target += "?bestaetigung=1"
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}
