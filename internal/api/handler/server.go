package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/api/middleware"
	"github.com/fachschaftinformatik/web/internal/avatars"
	"github.com/fachschaftinformatik/web/internal/config"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/email"
	"github.com/fachschaftinformatik/web/internal/id"
	"github.com/fachschaftinformatik/web/internal/storage"
	"github.com/go-chi/httplog/v2"
	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

const (
	SessionCookieName = "__Host-session"
	CsrfCookieName    = "__Host-csrf"
	SessionDuration   = 24 * time.Hour
	CsrfDuration      = 24 * time.Hour
	maxUploadSize     = 256 << 20 // 256 MB
)

type Server struct {
	DB            database.Querier
	Log           *httplog.Logger
	Config        *config.Config
	Email         *email.Sender
	Avatars       *avatars.Service
	Store         storage.Provider
	SecureCookies bool
	policy        *bluemonday.Policy
	validator     *validator.Validate
}

func NewServer(db database.Querier, logger *httplog.Logger, cfg *config.Config, emailSender *email.Sender, store storage.Provider, avatarService *avatars.Service) *Server {
	return &Server{
		DB:            db,
		Log:           logger,
		Config:        cfg,
		Email:         emailSender,
		Store:         store,
		Avatars:       avatarService,
		SecureCookies: cfg.SecureCookies,
		policy:        bluemonday.UGCPolicy(),
		validator:     validator.New(),
	}
}

func (s *Server) Sanitize(str string) string {
	return s.policy.Sanitize(str)
}

func (s *Server) Validate(i interface{}) error {
	return s.validator.Struct(i)
}

func (s *Server) User(r *http.Request) (database.User, bool) {
	u, ok := r.Context().Value(middleware.UserKey).(database.User)
	return u, ok
}

func (s *Server) JsonError(w http.ResponseWriter, err, msg string, status int) {
	s.RespondJSON(w, status, dto.ErrorResponse{Error: err, Message: msg})
}

func (s *Server) RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			s.Log.Error("Failed to encode JSON response", "err", err)
		}
	}
}

func (s *Server) toUserResponse(user database.User) dto.UserResponse {
	var programID *id.ID
	if user.ProgramID != nil {
		pid := id.ID(*user.ProgramID)
		programID = &pid
	}

	return dto.UserResponse{
		ID:        id.ID(user.ID),
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		Active:    user.Active,
		Verified:  user.Verified,
		ProgramID: programID,
		CreatedAt: user.CreatedAt,
		Theme:     user.Theme,
		Private:   user.Private,
		AvatarUrl: s.FormatAvatarURL(user.AvatarUrl, id.ID(user.ID), 0), // Use 0 for private flag here since it's the owner's view
	}
}

func (s *Server) ToPublicUserResponse(user database.User) dto.PublicUserResponse {
	name := user.Name
	if user.Private == 1 {
		name = "Anonym"
	}

	var programID *id.ID
	if user.ProgramID != nil {
		pid := id.ID(*user.ProgramID)
		programID = &pid
	}

	return dto.PublicUserResponse{
		ID:        id.ID(user.ID),
		Name:      name,
		Role:      user.Role,
		Active:    user.Active,
		Verified:  user.Verified,
		ProgramID: programID,
		CreatedAt: user.CreatedAt,
		Theme:     user.Theme,
		Private:   user.Private,
		AvatarUrl: s.FormatAvatarURL(user.AvatarUrl, id.ID(user.ID), user.Private),
	}
}

func (s *Server) SetCookie(w http.ResponseWriter, name string, value string, maxAge time.Duration, httpOnly bool) {
	actualName := name
	if !s.SecureCookies {
		actualName = strings.TrimPrefix(name, "__Host-")
	}

	cookie := &http.Cookie{
		Name:     actualName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: httpOnly,
		Secure:   s.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

func (s *Server) Authenticate(w http.ResponseWriter, r *http.Request) (database.Session, database.User, error) {
	name := SessionCookieName
	if !s.SecureCookies {
		name = strings.TrimPrefix(name, "__Host-")
	}

	cookie, err := r.Cookie(name)
	if err != nil {
		return database.Session{}, database.User{}, errors.New("session cookie not found")
	}

	sessionID := cookie.Value
	ctx := r.Context()

	session, err := s.DB.GetSession(ctx, sessionID)
	if err != nil {
		return database.Session{}, database.User{}, errors.New("invalid session")
	}

	if session.UserAgent != nil && *session.UserAgent != r.UserAgent() {
		s.Log.Warn("Session user agent mismatch (potential hijack or browser update)", "sessionID", session.ID, "expectedUA", *session.UserAgent, "gotUA", r.UserAgent())
	}

	expiresAt, err := time.Parse(time.RFC3339, session.ExpiresAt)
	if err != nil || expiresAt.Before(time.Now()) {
		return database.Session{}, database.User{}, errors.New("session expired or invalid")
	}

	lastSeen, err := time.Parse(time.RFC3339, session.LastSeen)
	if err == nil && time.Since(lastSeen) > 5*time.Minute {
		createdAt, err := time.Parse(time.RFC3339, session.CreatedAt)
		if err != nil {
			createdAt = time.Now()
		}

		extension := SessionDuration
		if expiresAt.Sub(createdAt) > 25*time.Hour {
			extension = 30 * 24 * time.Hour
		}

		newExpires := time.Now().UTC().Add(extension).Format(time.RFC3339)
		s.DB.SlideSession(ctx, database.SlideSessionParams{
			ID:        session.ID,
			ExpiresAt: newExpires,
		})
	}

	user, err := s.DB.GetUser(ctx, session.UserID)
	if err != nil {
		return database.Session{}, database.User{}, errors.New("user not found")
	}

	return session, user, nil
}

func (s *Server) CheckCSRF(r *http.Request) error {
	headerToken := r.Header.Get("X-CSRF-Token")

	name := CsrfCookieName
	if !s.SecureCookies {
		name = strings.TrimPrefix(name, "__Host-")
	}

	cookie, err := r.Cookie(name)
	if headerToken == "" || err != nil || headerToken != cookie.Value {
		return errors.New("invalid CSRF token")
	}
	return nil
}

func (s *Server) BoolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func (s *Server) FormatAvatarURL(path *string, userID id.ID, private int64) *string {
	if private == 1 {
		url := "/api/v1/auth/avatars/" + userID.String() + "/generated_v4.svg?private=true"
		return &url
	}

	if path != nil && *path != "" {
		if strings.HasPrefix(*path, "/api/v1/auth/avatars/") {
			return path
		}
		// Standardized path is avatars/{userId}/source
		url := "/api/v1/auth/avatars/" + userID.String() + "/source"
		return &url
	}

	url := "/api/v1/auth/avatars/" + userID.String() + "/generated_v4.svg"
	return &url
}
