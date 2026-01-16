package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/avatars"
	"github.com/fachschaftinformatik/web/internal/buckets"
	"github.com/fachschaftinformatik/web/internal/config"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/email"
	"github.com/fachschaftinformatik/web/internal/sid"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieName = "__Host-session"
	csrfCookieName    = "__Host-csrf"
	sessionDuration   = 24 * time.Hour
	csrfDuration      = 15 * time.Minute
)

// @Description User account information
type UserResponse struct {
	database.User

	// Do not export the password hash
	Password          string  `json:"-"`
	VerificationToken *string `json:"verification_token,omitempty"`
}

type PublicUserResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	Active    int64   `json:"active"`
	Verified  int64   `json:"verified"`
	Programid int64   `json:"programid"`
	CreatedAt string  `json:"created_at"`
	Theme     string  `json:"theme"`
	Private   int64   `json:"private"`
	AvatarUrl *string `json:"avatar_url"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"user@studmail.w-hs.de"`
	Password string `json:"password" example:"secret123"`
}

type RegisterRequest struct {
	Email     string `json:"email" example:"user@studmail.w-hs.de"`
	Name      string `json:"name" example:"Max Mustermann"`
	Password  string `json:"password" example:"secret123"`
	Programid int    `json:"programid" example:"1"`
}

type UpdateProfileRequest struct {
	Name      string `json:"name" example:"Max Mustermann"`
	Programid int    `json:"programid" example:"1"`
	Theme     string `json:"theme" example:"dark"`
	Private   bool   `json:"private"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type CsrfResponse struct {
	Csrf string `json:"csrf"`
}

type Server struct {
	DB            database.Querier
	Log           *log.Logger
	Config        *config.Config
	Email         *email.Sender
	Avatars       *avatars.Service
	Store         *buckets.Client
	SecureCookies bool
}

func NewServer(db database.Querier, logger *log.Logger, cfg *config.Config, emailSender *email.Sender, store *buckets.Client, avatarService *avatars.Service) *Server {
	return &Server{
		DB:            db,
		Log:           logger,
		Config:        cfg,
		Email:         emailSender,
		Store:         store,
		Avatars:       avatarService,
		SecureCookies: cfg.SecureCookies,
	}
}

// PostAuthRegister registers a new user
// @Summary Register a user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration Info"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/register [post]
func (s *Server) PostAuthRegister(w http.ResponseWriter, r *http.Request) {
	var payload RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.jsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		s.Log.Printf("Failed to hash password: %v", err)
		s.jsonError(w, "server_error", "Could not process registration", http.StatusInternalServerError)
		return
	}

	verificationToken := uuid.NewString()

	params := database.CreateUserParams{
		ID:                sid.New(),
		Email:             string(payload.Email),
		Name:              payload.Name,
		Password:          string(hashedPassword),
		Role:              "user",
		Active:            1,
		Programid:         int64(payload.Programid),
		VerificationToken: &verificationToken,
		AvatarUrl:         nil, // Updated via GenerateAndStoreAvatar
	}

	avatarPath, err := s.Avatars.GenerateAndStoreAvatar(r.Context(), params.ID)
	if err != nil {
		s.Log.Printf("Failed to generate avatar for %s: %v", params.ID, err)
	} else {
		params.AvatarUrl = &avatarPath
	}

	dbUser, err := s.DB.CreateUser(r.Context(), params)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			s.jsonError(w, "email_exists", "A user with this email already exists", http.StatusConflict)
		} else {
			s.Log.Printf("Failed to create user: %v", err)
			s.jsonError(w, "database_error", "Could not create user", http.StatusInternalServerError)
		}
		return
	}

	if s.Config.SignupsVerify {
		go func() {
			if err := s.Email.SendVerificationEmail(dbUser.Email, dbUser.Name, verificationToken); err != nil {
				s.Log.Printf("Failed to send verification email to %s: %v", dbUser.Email, err)
			}
		}()
	} else {
		var verifiedUntil *string
		now := time.Now()

		if strings.HasSuffix(dbUser.Email, "@studmail.w-hs.de") {
			year := now.Year()

			// TODO: Actually handle this at some point
			march1 := time.Date(year, time.March, 1, 0, 0, 0, 0, time.UTC)
			oct1 := time.Date(year, time.October, 1, 0, 0, 0, 0, time.UTC)

			var nextDate time.Time
			if now.Before(march1) {
				nextDate = march1
			} else if now.Before(oct1) {
				nextDate = oct1
			} else {
				nextDate = time.Date(year+1, time.March, 1, 0, 0, 0, 0, time.UTC)
			}
			tStr := nextDate.Format(time.RFC3339)
			verifiedUntil = &tStr
		} else {
			verifiedUntil = nil
		}

		updatedUser, err := s.DB.VerifyUser(r.Context(), database.VerifyUserParams{
			ID:            dbUser.ID,
			VerifiedUntil: verifiedUntil,
		})
		if err != nil {
			s.Log.Printf("Failed to auto-verify user: %v", err)
			s.jsonError(w, "database_error", "Could not verify user", http.StatusInternalServerError)
			return
		}
		dbUser = updatedUser
	}

	dbUser.AvatarUrl = s.formatAvatarURL(dbUser.AvatarUrl, dbUser.ID)
	s.respondJSON(w, http.StatusCreated, UserResponse{User: dbUser})
}

// PostAuthLogin logs in a user
// @Summary Log in
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Credentials"
// @Success 200 {object} UserResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (s *Server) PostAuthLogin(w http.ResponseWriter, r *http.Request) {
	var payload LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.jsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	dbUser, err := s.DB.GetUserByEmail(r.Context(), string(payload.Email))
	if err != nil {
		s.jsonError(w, "invalid_credentials", "Invalid email or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(payload.Password))
	if err != nil {
		s.jsonError(w, "invalid_credentials", "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if dbUser.Verified == 0 {
		newToken := uuid.NewString()
		if err := s.DB.UpdateUserToken(r.Context(), database.UpdateUserTokenParams{
			ID:                dbUser.ID,
			VerificationToken: &newToken,
		}); err != nil {
			s.Log.Printf("Failed to update token for user %s: %v", dbUser.ID, err)
		} else {
			go func() {
				if err := s.Email.SendVerificationEmail(dbUser.Email, dbUser.Name, newToken); err != nil {
					s.Log.Printf("Failed to resend verification email to %s: %v", dbUser.Email, err)
				}
			}()
		}

		s.jsonError(w, "email_not_verified", "Du musst erst deine E-Mail bestätigen. Wir haben dir eine neue E-Mail gesendet.", http.StatusForbidden)
		return
	}

	sessionID := uuid.NewString()
	expiresAt := time.Now().Add(sessionDuration)

	_, err = s.DB.CreateSession(r.Context(), database.CreateSessionParams{
		ID:        sessionID,
		Userid:    dbUser.ID,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
	if err != nil {
		s.Log.Printf("Failed to create session: %v", err)
		s.jsonError(w, "server_error", "Could not create session", http.StatusInternalServerError)
		return
	}

	s.setCookie(w, sessionCookieName, sessionID, sessionDuration, true)
	dbUser.AvatarUrl = s.formatAvatarURL(dbUser.AvatarUrl, dbUser.ID)
	s.respondJSON(w, http.StatusOK, UserResponse{User: dbUser})
}

// GetAuthVerify verifies an email
// @Summary Verify user email
// @Tags Auth
// @Param token query string true "Verification Token"
// @Success 302
// @Router /auth/verify [get]
func (s *Server) GetAuthVerify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		s.jsonError(w, "invalid_token", "Missing token", http.StatusBadRequest)
		return
	}

	dbUser, err := s.DB.GetUserByVerificationToken(r.Context(), &token) // FIXED: Pointer
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.jsonError(w, "invalid_token", "Invalid verification token", http.StatusBadRequest)
		} else {
			s.Log.Printf("Failed to lookup token: %v", err)
			s.jsonError(w, "server_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	var verifiedUntil *string
	now := time.Now()

	if strings.HasSuffix(dbUser.Email, "@studmail.w-hs.de") {
		year := now.Year()
		march1 := time.Date(year, time.March, 1, 0, 0, 0, 0, time.UTC)
		oct1 := time.Date(year, time.October, 1, 0, 0, 0, 0, time.UTC)

		var nextDate time.Time
		if now.Before(march1) {
			nextDate = march1
		} else if now.Before(oct1) {
			nextDate = oct1
		} else {
			nextDate = time.Date(year+1, time.March, 1, 0, 0, 0, 0, time.UTC)
		}
		tStr := nextDate.Format(time.RFC3339)
		verifiedUntil = &tStr
	} else {
		verifiedUntil = nil
	}

	_, err = s.DB.VerifyUser(r.Context(), database.VerifyUserParams{
		ID:            dbUser.ID,
		VerifiedUntil: verifiedUntil,
	})
	if err != nil {
		s.Log.Printf("Failed to verify user: %v", err)
		s.jsonError(w, "server_error", "Verification failed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/login?verified=true", s.Config.Domain), http.StatusFound)
}

// GetAuthMe gets the current user
// @Summary Get current user
// @Tags Auth
// @Success 200 {object} UserResponse
// @Router /auth/me [get]
func (s *Server) GetAuthMe(w http.ResponseWriter, r *http.Request) {
	_, dbUser, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}
	dbUser.AvatarUrl = s.formatAvatarURL(dbUser.AvatarUrl, dbUser.ID)
	s.respondJSON(w, http.StatusOK, UserResponse{User: dbUser})
}

// PutAuthMe updates the current user's profile
// @Summary Update current user profile
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Update Profile Info"
// @Success 200 {object} UserResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/me [put]
func (s *Server) PutAuthMe(w http.ResponseWriter, r *http.Request) {
	_, authUser, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if err := s.checkCSRF(r); err != nil {
		s.jsonError(w, "invalid_csrf", err.Error(), http.StatusForbidden)
		return
	}

	var payload UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.jsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	updatedUser, err := s.DB.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:        authUser.ID,
		Name:      payload.Name,
		Programid: int64(payload.Programid),
		Theme:     payload.Theme,
		Private:   s.boolToInt(payload.Private),
	})
	if err != nil {
		s.Log.Printf("Failed to update user profile: %v", err)
		s.jsonError(w, "database_error", "Could not update user profile", http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, http.StatusOK, UserResponse{User: updatedUser})
}

// PostAuthLogout logs out
// @Summary Log out
// @Tags Auth
// @Param X-CSRF-Token header string true "CSRF Token"
// @Success 204
// @Router /auth/logout [post]
func (s *Server) PostAuthLogout(w http.ResponseWriter, r *http.Request) {
	session, _, err := s.authenticate(w, r)
	if err != nil {
		s.setCookie(w, sessionCookieName, "", -time.Hour, true)
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if err := s.checkCSRF(r); err != nil {
		s.jsonError(w, "invalid_csrf", err.Error(), http.StatusForbidden)
		return
	}

	if err = s.DB.DeleteSession(r.Context(), session.ID); err != nil {
		s.Log.Printf("Failed to delete session: %v", err)
	}

	s.setCookie(w, sessionCookieName, "", -time.Hour, true)
	s.setCookie(w, csrfCookieName, "", -time.Hour, false)
	w.WriteHeader(http.StatusNoContent)
}

// GetAuthCsrf issues a CSRF token
// @Summary Issue CSRF token
// @Tags Auth
// @Success 200 {object} CsrfResponse
// @Router /auth/csrf [get]
func (s *Server) GetAuthCsrf(w http.ResponseWriter, r *http.Request) {
	csrfToken := uuid.NewString()
	s.setCookie(w, csrfCookieName, csrfToken, csrfDuration, false)
	s.respondJSON(w, http.StatusOK, CsrfResponse{Csrf: csrfToken})
}

// GetUsers lists users (admin only)
// @Summary List users
// @Tags Users
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} UserResponse
// @Router /users [get]
func (s *Server) GetUsers(w http.ResponseWriter, r *http.Request) {
	_, dbUser, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if dbUser.Role != "admin" {
		s.jsonError(w, "forbidden", "Insufficient permissions", http.StatusForbidden)
		return
	}

	// TODO
	limit := int64(32)
	offset := int64(0)

	dbUsers, err := s.DB.ListUsers(r.Context(), database.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		s.jsonError(w, "database_error", "Could not list users", http.StatusInternalServerError)
		return
	}

	apiUsers := make([]UserResponse, 0, len(dbUsers))
	for _, user := range dbUsers {
		user.AvatarUrl = s.formatAvatarURL(user.AvatarUrl, user.ID)
		apiUsers = append(apiUsers, UserResponse{User: user})
	}

	s.respondJSON(w, http.StatusOK, apiUsers)
}

// GetUsersId gets a specific user
// @Summary Get user by ID
// @Tags Users
// @Param id path string true "User ID"
// @Summary Get user profile
// @Tags Auth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} PublicUserResponse
// @Router /users/{id} [get]
func (s *Server) GetUsersId(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Optional authentication
	_, authUser, authErr := s.authenticate(w, r)

	dbUser, err := s.DB.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.jsonError(w, "not_found", "User not found", http.StatusNotFound)
		} else {
			s.jsonError(w, "database_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	// If the requester is the user themselves or an admin, return the full profile
	if authErr == nil && (authUser.ID == id || authUser.Role == "admin") {
		// Ensure token is hidden even for self if you want, but usually self can see it?
		// Actually Model already has json tag for it. Let's just null it out if strictly private.
		dbUser.AvatarUrl = s.formatAvatarURL(dbUser.AvatarUrl, dbUser.ID)
		resp := UserResponse{User: dbUser}
		if authUser.Role != "admin" {
			resp.VerificationToken = nil // Hide token even from self if not needed
		}
		s.respondJSON(w, http.StatusOK, resp)
		return
	}

	// Otherwise check for privacy
	if dbUser.Private == 1 {
		dbUser.AvatarUrl = s.formatAvatarURL(dbUser.AvatarUrl, dbUser.ID)
		s.respondJSON(w, http.StatusOK, PublicUserResponse{
			ID:        dbUser.ID,
			Name:      "Anonym",
			Role:      dbUser.Role,
			Private:   1,
			CreatedAt: dbUser.CreatedAt,
			AvatarUrl: dbUser.AvatarUrl,
		})
		return
	}

	// Otherwise return public profile
	dbUser.AvatarUrl = s.formatAvatarURL(dbUser.AvatarUrl, dbUser.ID)
	s.respondJSON(w, http.StatusOK, PublicUserResponse{
		ID:        dbUser.ID,
		Name:      dbUser.Name,
		Role:      dbUser.Role,
		Active:    dbUser.Active,
		Verified:  dbUser.Verified,
		Programid: dbUser.Programid,
		CreatedAt: dbUser.CreatedAt,
		Theme:     dbUser.Theme,
		Private:   dbUser.Private,
		AvatarUrl: dbUser.AvatarUrl,
	})
}

func (s *Server) jsonError(w http.ResponseWriter, err, msg string, status int) {
	s.respondJSON(w, status, ErrorResponse{Error: err, Message: msg})
}

func (s *Server) respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			s.Log.Printf("Failed to encode JSON response: %v", err)
		}
	}
}

func (s *Server) setCookie(w http.ResponseWriter, name string, value string, maxAge time.Duration, httpOnly bool) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: httpOnly,
		Secure:   s.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

func (s *Server) authenticate(w http.ResponseWriter, r *http.Request) (database.Session, database.User, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return database.Session{}, database.User{}, errors.New("session cookie not found")
	}

	sessionID := cookie.Value
	ctx := r.Context()

	session, err := s.DB.GetSession(ctx, sessionID)
	if err != nil {
		return database.Session{}, database.User{}, errors.New("invalid session")
	}

	expiresAt, _ := time.Parse(time.RFC3339, session.ExpiresAt)
	if expiresAt.Before(time.Now()) {
		return database.Session{}, database.User{}, errors.New("session expired")
	}

	user, err := s.DB.GetUser(ctx, session.Userid)
	if err != nil {
		return database.Session{}, database.User{}, errors.New("user not found")
	}

	return session, user, nil
}

func (s *Server) checkCSRF(r *http.Request) error {
	headerToken := r.Header.Get("X-CSRF-Token")
	cookie, err := r.Cookie(csrfCookieName)
	if headerToken == "" || err != nil || headerToken != cookie.Value {
		return errors.New("invalid CSRF token")
	}
	return nil
}
func (s *Server) boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// GetAvatar serves the user avatar from the bucket
// @Summary Get user avatar
// @Tags Auth
// @Param userId path string true "User ID"
// @Param filename path string true "Filename"
// @Produce image/svg+xml
// @Success 200 {file} binary
// @Router /auth/avatars/{userId}/{filename} [get]
func (s *Server) GetAvatar(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	filename := chi.URLParam(r, "filename")
	objectName := fmt.Sprintf("avatars/%s/%s", userID, filename)

	obj, err := s.Store.GetObject(r.Context(), objectName)
	if err != nil {
		// If not found, generate it on the fly and store it
		data, genErr := s.Avatars.FetchDicebearSVG(r.Context(), userID)
		if genErr != nil {
			s.Log.Printf("Failed to fetch default avatar for %s: %v", userID, genErr)
			s.jsonError(w, "not_found", "Avatar not found and could not be generated", http.StatusNotFound)
			return
		}

		// Store it in the bucket
		uploadErr := s.Avatars.StoreAvatar(r.Context(), objectName, data)
		if uploadErr != nil {
			s.Log.Printf("Failed to store generated avatar for %s: %v", userID, uploadErr)
			// Still serve the data even if storage failed
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		w.Write(data)
		return
	}
	defer obj.Close()

	info, err := obj.Stat()
	if err != nil {
		// If Stat fails (e.g. object not found), generate it on the fly and store it
		data, genErr := s.Avatars.FetchDicebearSVG(r.Context(), userID)
		if genErr != nil {
			s.Log.Printf("Failed to fetch default avatar for %s: %v", userID, genErr)
			s.jsonError(w, "not_found", "Avatar not found and could not be generated", http.StatusNotFound)
			return
		}

		// Store it in the bucket
		// We use the same objectName that was requested
		uploadErr := s.Avatars.StoreAvatar(r.Context(), objectName, data)
		if uploadErr != nil {
			s.Log.Printf("Failed to store generated avatar for %s: %v", userID, uploadErr)
			// Still serve the data even if storage failed
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		w.Write(data)
		return
	}

	w.Header().Set("Content-Type", info.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size))
	w.Header().Set("ETag", info.ETag)

	// If it's a UUID-based avatar (contains .svg and is not the fallback), we can cache it forever
	if strings.HasSuffix(filename, ".svg") && !strings.Contains(filename, "generated") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=3600") // 1 hour for others
	}

	if r.Header.Get("If-None-Match") == info.ETag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	io.Copy(w, obj)
}

func (s *Server) formatAvatarURL(path *string, userID string) *string {
	if path != nil && *path != "" {
		// If it already starts with /api/auth/avatars, return as is (legacy or full url)
		if strings.HasPrefix(*path, "/api/auth/avatars/") {
			return path
		}

		// The path is stored as avatars/<userid>/<uuid>.svg
		// The route is /api/auth/avatars/<userid>/<uuid>.svg which maps to GetAvatar handler
		// Handler expects userID and filename param.
		// Route: /auth/avatars/{userId}/{filename}
		// We need to transform avatars/<userid>/<uuid>.svg -> /api/auth/avatars/<userid>/<uuid>.svg
		// Simply removing "avatars/" prefix from stored path gives us <userid>/<uuid>.svg which fits URL structure

		cleanPath := strings.TrimPrefix(*path, "avatars/")
		url := fmt.Sprintf("/api/auth/avatars/%s", cleanPath)
		return &url
	}

	// Fallback URL served by us, which will generate it on the fly if needed
	url := fmt.Sprintf("/api/auth/avatars/%s/generated_v4.svg", userID)
	return &url
}
