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
	"github.com/fachschaftinformatik/web/internal/database"
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
	SecureCookies bool
}

func NewServer(db database.Querier, logger *log.Logger, secureCookies bool) *Server {
	return &Server{
		DB:            db,
		Log:           logger,
		SecureCookies: secureCookies,
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

	params := database.CreateUserParams{
		ID:        uuid.NewString(),
		Email:     string(payload.Email),
		Name:      payload.Name,
		Password:  string(hashedPassword),
		Role:      "user",
		Active:    1,
		Programid: int64(payload.Programid),
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

	// Since we filter by ID, all rows belong to the same program.
	// We take the first row for the main program details.
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
