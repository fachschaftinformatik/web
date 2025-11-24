package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/api"
	"github.com/fachschaftinformatik/web/internal/config"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/email"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieName = "__Host-session"
	csrfCookieName    = "__Host-csrf"
	sessionDuration   = 24 * time.Hour
	csrfDuration      = 15 * time.Minute
)

type Server struct {
	DB            database.Querier
	Log           *log.Logger
	Config        *config.Config
	Email         *email.Sender
	SecureCookies bool
}

func NewServer(db database.Querier, logger *log.Logger, cfg *config.Config, emailSender *email.Sender) *Server {
	return &Server{
		DB:            db,
		Log:           logger,
		Config:        cfg,
		Email:         emailSender,
		SecureCookies: cfg.SecureCookies,
	}
}

func (s *Server) PostAuthRegister(w http.ResponseWriter, r *http.Request) {
	var payload api.UserRegister
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
		ID:                uuid.NewString(),
		Email:             string(payload.Email),
		Name:              payload.Name,
		Password:          string(hashedPassword),
		Role:              "user",
		Active:            1,
		Programid:         int64(payload.Programid),
		VerificationToken: sql.NullString{String: verificationToken, Valid: true},
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

	// Send verification email
	go func() {
		if err := s.Email.SendVerificationEmail(dbUser.Email, dbUser.Name, verificationToken); err != nil {
			s.Log.Printf("Failed to send verification email to %s: %v", dbUser.Email, err)
		}
	}()

	apiUser, err := dbUserToAPI(dbUser)
	if err != nil {
		s.jsonError(w, "server_error", "Could not process user data", http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, http.StatusCreated, apiUser)
}

func (s *Server) PostAuthLogin(w http.ResponseWriter, r *http.Request) {
	var payload api.UserLogin
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
		// Logic to resend token if needed could go here
		// For now, just regenerate a token and resend if one exists or if expired.
		// Simplest approach: Always generate a new one and send it if they try to login.
		newToken := uuid.NewString()
		if err := s.DB.UpdateUserToken(r.Context(), database.UpdateUserTokenParams{
			ID:                dbUser.ID,
			VerificationToken: sql.NullString{String: newToken, Valid: true},
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

	apiUser, err := dbUserToAPI(dbUser)
	if err != nil {
		s.jsonError(w, "server_error", "Could not process user data", http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, http.StatusOK, apiUser)
}

func (s *Server) GetAuthVerify(w http.ResponseWriter, r *http.Request, params api.GetAuthVerifyParams) {
	dbUser, err := s.DB.GetUserByVerificationToken(r.Context(), sql.NullString{String: params.Token, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.jsonError(w, "invalid_token", "Invalid verification token", http.StatusBadRequest)
		} else {
			s.Log.Printf("Failed to lookup token: %v", err)
			s.jsonError(w, "server_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Calculate Verified Until
	var verifiedUntil sql.NullString
	now := time.Now()

	if strings.HasSuffix(dbUser.Email, "@studmail.w-hs.de") {
		// Determine next March 1st or October 1st
		year := now.Year()
		march1 := time.Date(year, time.March, 1, 0, 0, 0, 0, time.UTC)
		oct1 := time.Date(year, time.October, 1, 0, 0, 0, 0, time.UTC)

		var nextDate time.Time
		if now.Before(march1) {
			nextDate = march1
		} else if now.Before(oct1) {
			nextDate = oct1
		} else {
			// After Oct 1st, next date is March 1st next year
			nextDate = time.Date(year+1, time.March, 1, 0, 0, 0, 0, time.UTC)
		}
		verifiedUntil = sql.NullString{String: nextDate.Format(time.RFC3339), Valid: true}
	} else if strings.HasSuffix(dbUser.Email, "@fachschaftinformatik.de") {
		// Forever verified (NULL)
		verifiedUntil = sql.NullString{Valid: false}
	} else {
		// Default fallback for other domains (if any allowed in future)
		// For now, treat as students or maybe block? Assuming studmail behavior or just verified until forever for now to be safe?
		// Prompt said "users with @studmail...". Let's assume others are external and verify once.
		verifiedUntil = sql.NullString{Valid: false}
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

	// Instead of JSON, we should probably redirect to the frontend dashboard or a success page
	// But the prompt asked to hook it up. The link in email points to API.
	// We can redirect to the frontend login page with a success parameter.
	http.Redirect(w, r, fmt.Sprintf("%s/login?verified=true", s.Config.PublicURL), http.StatusFound)
}

func (s *Server) GetAuthMe(w http.ResponseWriter, r *http.Request) {
	_, dbUser, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	apiUser, err := dbUserToAPI(dbUser)
	if err != nil {
		s.jsonError(w, "server_error", "Could not process user data", http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, http.StatusOK, apiUser)
}

func (s *Server) PostAuthLogout(w http.ResponseWriter, r *http.Request, params api.PostAuthLogoutParams) {
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

func (s *Server) GetAuthCsrf(w http.ResponseWriter, r *http.Request) {
	csrfToken := uuid.NewString()
	s.setCookie(w, csrfCookieName, csrfToken, csrfDuration, false)
	s.respondJSON(w, http.StatusOK, map[string]string{"csrf": csrfToken})
}

func (s *Server) GetUsers(w http.ResponseWriter, r *http.Request, params api.GetUsersParams) {
	_, dbUser, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if dbUser.Role != "admin" {
		s.jsonError(w, "forbidden", "You do not have permission to access this resource", http.StatusForbidden)
		return
	}

	limit := int64(32)
	offset := int64(0)
	if params.Limit != nil {
		limit = int64(*params.Limit)
	}
	if params.Offset != nil {
		offset = int64(*params.Offset)
	}

	dbUsers, err := s.DB.ListUsers(r.Context(), database.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		s.Log.Printf("Failed to list users: %v", err)
		s.jsonError(w, "database_error", "Could not list users", http.StatusInternalServerError)
		return
	}

	apiUsers := make([]api.User, 0, len(dbUsers))
	for _, user := range dbUsers {
		apiUser, err := dbUserToAPI(user)
		if err != nil {
			s.jsonError(w, "server_error", "Could not process user data", http.StatusInternalServerError)
			return
		}
		apiUsers = append(apiUsers, apiUser)
	}

	s.respondJSON(w, http.StatusOK, apiUsers)
}

func (s *Server) GetUsersId(w http.ResponseWriter, r *http.Request, id string) {
	_, authUser, err := s.authenticate(w, r)
	if err != nil {
		s.jsonError(w, "unauthorized", err.Error(), http.StatusUnauthorized)
		return
	}

	if authUser.ID != id && authUser.Role != "admin" {
		s.jsonError(w, "forbidden", "You do not have permission to access this resource", http.StatusForbidden)
		return
	}

	dbUser, err := s.DB.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.jsonError(w, "not_found", "User not found", http.StatusNotFound)
		} else {
			s.Log.Printf("Failed to get user: %v", err)
			s.jsonError(w, "database_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	apiUser, err := dbUserToAPI(dbUser)
	if err != nil {
		s.jsonError(w, "server_error", "Could not process user data", http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, http.StatusOK, apiUser)
}

func (s *Server) GetPrograms(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.ListProgramsWithVersions(r.Context())
	if err != nil {
		s.Log.Printf("Failed to list programs: %v", err)
		s.jsonError(w, "database_error", "Could not fetch programs", http.StatusInternalServerError)
		return
	}

	programMap := make(map[int64]*api.Program)
	var orderedPrograms []*api.Program

	for _, row := range rows {
		prog, exists := programMap[row.ID]
		if !exists {
			prog = &api.Program{
				Id:       int(row.ID),
				Name:     row.Name,
				Versions: []string{},
			}
			programMap[row.ID] = prog
			orderedPrograms = append(orderedPrograms, prog)
		}
		prog.Versions = append(prog.Versions, row.Version)
	}

	response := make([]api.Program, 0, len(orderedPrograms))
	for _, p := range orderedPrograms {
		response = append(response, *p)
	}

	s.respondJSON(w, http.StatusOK, response)
}

func (s *Server) GetProgramsId(w http.ResponseWriter, r *http.Request, id int) {
	rows, err := s.DB.GetProgramWithVersions(r.Context(), int64(id))
	if err != nil {
		s.Log.Printf("Failed to get program: %v", err)
		s.jsonError(w, "database_error", "Could not fetch program", http.StatusInternalServerError)
		return
	}

	if len(rows) == 0 {
		s.jsonError(w, "not_found", "Program not found", http.StatusNotFound)
		return
	}

	prog := api.Program{
		Id:       int(rows[0].ID),
		Name:     rows[0].Name,
		Versions: make([]string, 0, len(rows)),
	}

	for _, row := range rows {
		prog.Versions = append(prog.Versions, row.Version)
	}

	s.respondJSON(w, http.StatusOK, prog)
}

func (s *Server) jsonError(w http.ResponseWriter, err, msg string, status int) {
	s.respondJSON(w, status, api.Error{
		Error:   err,
		Message: msg,
	})
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
		if errors.Is(err, sql.ErrNoRows) {
			return database.Session{}, database.User{}, errors.New("invalid session")
		}
		s.Log.Printf("Auth: DB.GetSession error: %v", err)
		return database.Session{}, database.User{}, errors.New("database error")
	}

	expiresAt, err := time.Parse(time.RFC3339, session.ExpiresAt)
	if err != nil || expiresAt.Before(time.Now()) {
		return database.Session{}, database.User{}, errors.New("session expired")
	}

	user, err := s.DB.GetUser(ctx, session.Userid)
	if err != nil {
		s.Log.Printf("Auth: DB.GetUser error: %v", err)
		return database.Session{}, database.User{}, errors.New("user not found for session")
	}

	newExpiresAt := time.Now().Add(sessionDuration)
	if _, err = s.DB.SlideSession(ctx, database.SlideSessionParams{
		ID:        sessionID,
		ExpiresAt: newExpiresAt.Format(time.RFC3339),
	}); err != nil {
		s.Log.Printf("Auth: Failed to slide session: %v", err)
	}

	s.setCookie(w, sessionCookieName, sessionID, sessionDuration, true)

	return session, user, nil
}

func (s *Server) checkCSRF(r *http.Request) error {
	headerToken := r.Header.Get("X-CSRF-Token")
	if headerToken == "" {
		return errors.New("X-CSRF-Token header missing")
	}

	cookie, err := r.Cookie(csrfCookieName)
	if err != nil {
		return errors.New("CSRF cookie missing")
	}

	if headerToken != cookie.Value {
		return errors.New("invalid CSRF token")
	}

	return nil
}

func dbUserToAPI(user database.User) (api.User, error) {
	apiUser := api.User{
		Active:       api.UserActive(user.Active),
		Programid:    int(user.Programid),
		Email:        types.Email(user.Email),
		Id:           user.ID,
		Name:         user.Name,
		Role:         api.UserRole(user.Role),
		Verified:     api.UserVerified(user.Verified),
	}

	var err error
	apiUser.CreatedAt, err = time.Parse(time.RFC3339, user.CreatedAt)
	if err != nil {
		return api.User{}, fmt.Errorf("could not parse CreatedAt: %w", err)
	}

	apiUser.UpdatedAt, err = time.Parse(time.RFC3339, user.UpdatedAt)
	if err != nil {
		return api.User{}, fmt.Errorf("could not parse UpdatedAt: %w", err)
	}

	if t, err := convertNullTime(user.VerifiedAt); err == nil && t != nil {
		apiUser.VerifiedAt = t
	} else if err != nil {
		return api.User{}, err
	}

	if t, err := convertNullTime(user.VerifiedUntil); err == nil && t != nil {
		apiUser.VerifiedUntil = t
	} else if err != nil {
		return api.User{}, err
	}

	return apiUser, nil
}

func convertNullTime(ns sql.NullString) (*time.Time, error) {
	if !ns.Valid {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, ns.String)
	if err != nil {
		return nil, fmt.Errorf("could not parse timestamp: %w", err)
	}
	return &t, nil
}
