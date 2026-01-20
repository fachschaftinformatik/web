package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fachschaftinformatik/web/internal/id"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/api/middleware"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// @Summary Register a user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration Info"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (s *Server) PostAuthRegister(w http.ResponseWriter, r *http.Request) {
	s.Log.Info("Registration request received")
	var payload dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.Log.Error("Registration: Failed to decode body", "err", err)
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if err := s.Validate(payload); err != nil {
		s.Log.Error("Registration: Validation failed", "err", err)
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		s.Log.Error("Failed to hash password", "err", err)
		s.JsonError(w, "server_error", "Could not process registration", http.StatusInternalServerError)
		return
	}

	verificationToken := uuid.NewString()

	params := database.CreateUserParams{
		ID:                id.New(),
		Email:             string(payload.Email),
		Name:              s.Sanitize(payload.Name),
		Password:          string(hashedPassword),
		Role:              "user",
		Active:            1,
		Programid:         payload.Programid,
		VerificationToken: &verificationToken,
		AvatarUrl:         nil,
	}

	avatarPath, err := s.Avatars.GenerateAndStoreAvatar(r.Context(), params.ID)
	if err != nil {
		s.Log.Error("Failed to generate avatar", "userID", params.ID, "err", err)
	} else {
		params.AvatarUrl = &avatarPath
	}

	dbUser, err := s.DB.CreateUser(r.Context(), params)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			s.Log.Warn("Registration: Email already exists", "email", payload.Email)
			s.JsonError(w, "email_exists", "A user with this email already exists", http.StatusConflict)
		} else {
			s.Log.Error("Registration: Database error", "err", err)
			s.JsonError(w, "database_error", "Could not create user", http.StatusInternalServerError)
		}
		return
	}

	s.Log.Info("Registration: User created successfully", "userID", dbUser.ID)

	if s.Config.SignupsVerify {
		go func() {
			if err := s.Email.SendVerificationEmail(dbUser.Email, dbUser.Name, verificationToken); err != nil {
				s.Log.Error("Failed to send verification email", "email", dbUser.Email, "err", err)
			}
		}()
	} else {
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

		updatedUser, err := s.DB.VerifyUser(r.Context(), database.VerifyUserParams{
			ID:            dbUser.ID,
			VerifiedUntil: verifiedUntil,
		})
		if err != nil {
			s.Log.Error("Failed to auto-verify user", "err", err)
			s.JsonError(w, "database_error", "Could not verify user", http.StatusInternalServerError)
			return
		}
		dbUser = updatedUser
	}

	s.RespondJSON(w, http.StatusCreated, s.toUserResponse(dbUser))
}

// @Summary Log in
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Credentials"
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (s *Server) PostAuthLogin(w http.ResponseWriter, r *http.Request) {
	s.Log.Info("Login request received")
	var payload dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	dbUser, err := s.DB.GetUserByEmail(r.Context(), string(payload.Email))
	if err != nil {
		s.JsonError(w, "invalid_credentials", "Invalid email or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(payload.Password))
	if err != nil {
		s.JsonError(w, "invalid_credentials", "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if dbUser.Active == 0 {
		s.JsonError(w, "account_disabled", "Dein Account wurde deaktiviert.", http.StatusForbidden)
		return
	}

	if dbUser.Verified == 0 {
		newToken := uuid.NewString()
		if err := s.DB.UpdateUserToken(r.Context(), database.UpdateUserTokenParams{
			ID:                dbUser.ID,
			VerificationToken: &newToken,
		}); err != nil {
			s.Log.Error("Failed to update token for user", "userID", dbUser.ID, "err", err)
		} else {
			go func() {
				if err := s.Email.SendVerificationEmail(dbUser.Email, dbUser.Name, newToken); err != nil {
					s.Log.Error("Failed to resend verification email", "email", dbUser.Email, "err", err)
				}
			}()
		}

		s.JsonError(w, "email_not_verified", "Du musst erst deine E-Mail bestätigen. Wir haben dir eine neue E-Mail gesendet.", http.StatusForbidden)
		return
	}

	// OWASP: Session Fixation Protection
	s.SetCookie(w, SessionCookieName, "", -time.Hour, true)
	s.SetCookie(w, CsrfCookieName, "", -time.Hour, false)

	sessionID := id.New()

	duration := SessionDuration
	if payload.Remember {
		duration = 30 * 24 * time.Hour // 30 days
	}
	expiresAt := time.Now().Add(duration)

	userAgent := r.UserAgent()
	ipAddress := r.RemoteAddr

	_, err = s.DB.CreateSession(r.Context(), database.CreateSessionParams{
		ID:        sessionID,
		Userid:    dbUser.ID,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		UserAgent: &userAgent,
		IpAddress: &ipAddress,
	})
	if err != nil {
		s.Log.Error("Failed to create session", "err", err)
		s.JsonError(w, "server_error", "Could not create session", http.StatusInternalServerError)
		return
	}

	s.SetCookie(w, SessionCookieName, sessionID, duration, true)
	s.RespondJSON(w, http.StatusOK, s.toUserResponse(dbUser))
}

// @Summary Verify user email
// @Tags Auth
// @Param token query string true "Verification Token"
// @Success 302
// @Router /auth/verify [get]
func (s *Server) GetAuthVerify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, fmt.Sprintf("%s/login?verified=false&error=invalid_token", s.Config.Domain), http.StatusFound)
		return
	}

	dbUser, err := s.DB.GetUserByVerificationToken(r.Context(), &token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Redirect(w, r, fmt.Sprintf("%s/login?verified=false&error=invalid_token", s.Config.Domain), http.StatusFound)
		} else {
			s.Log.Error("Failed to lookup token", "err", err)
			http.Redirect(w, r, fmt.Sprintf("%s/login?verified=false&error=server_error", s.Config.Domain), http.StatusFound)
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
		s.Log.Error("Failed to verify user", "err", err)
		http.Redirect(w, r, fmt.Sprintf("%s/login?verified=false&error=server_error", s.Config.Domain), http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/login?verified=true", s.Config.Domain), http.StatusFound)
}

// @Summary Get current user
// @Tags Auth
// @Success 200 {object} dto.UserResponse
// @Router /auth/me [get]
func (s *Server) GetAuthMe(w http.ResponseWriter, r *http.Request) {
	user, _ := s.User(r)
	s.RespondJSON(w, http.StatusOK, s.toUserResponse(user))
}

// @Summary Update current user profile
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.UpdateProfileRequest true "Update Profile Info"
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/me [put]
func (s *Server) PutAuthMe(w http.ResponseWriter, r *http.Request) {
	authUser, _ := s.User(r)

	var payload dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if err := s.Validate(payload); err != nil {
		s.JsonError(w, "invalid_input", err.Error(), http.StatusBadRequest)
		return
	}

	updatedUser, err := s.DB.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:        authUser.ID,
		Name:      s.Sanitize(payload.Name),
		Programid: payload.Programid,
		Theme:     payload.Theme,
		Private:   s.BoolToInt(payload.Private),
	})
	if err != nil {
		s.Log.Error("Failed to update user profile", "err", err)
		s.JsonError(w, "database_error", "Could not update user profile", http.StatusInternalServerError)
		return
	}
	s.RespondJSON(w, http.StatusOK, s.toUserResponse(updatedUser))
}

// @Summary Log out
// @Tags Auth
// @Param X-CSRF-Token header string true "CSRF Token"
// @Success 204
// @Router /auth/logout [post]
func (s *Server) PostAuthLogout(w http.ResponseWriter, r *http.Request) {
	session, ok := r.Context().Value(middleware.SessionKey).(database.Session)
	if !ok {
		s.JsonError(w, "unauthorized", "Not logged in", http.StatusUnauthorized)
		return
	}

	if err := s.DB.DeleteSession(r.Context(), session.ID); err != nil {
		s.Log.Error("Failed to delete session", "err", err)
	}

	s.SetCookie(w, SessionCookieName, "", -time.Hour, true)
	s.SetCookie(w, CsrfCookieName, "", -time.Hour, false)
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Issue CSRF token
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.CsrfResponse
// @Router /auth/csrf [get]
func (s *Server) GetAuthCsrf(w http.ResponseWriter, r *http.Request) {
	csrfToken := uuid.NewString()
	s.SetCookie(w, CsrfCookieName, csrfToken, CsrfDuration, false)
	s.RespondJSON(w, http.StatusOK, dto.CsrfResponse{Csrf: csrfToken})
}

// @Summary List users
// @Tags Users
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} dto.UserResponse
// @Router /users [get]
func (s *Server) GetUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if limit <= 0 {
		limit = 32
	}
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)

	dbUsers, err := s.DB.ListUsers(r.Context(), database.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		s.Log.Error("Failed to list users", "err", err)
		s.JsonError(w, "database_error", "Could not list users", http.StatusInternalServerError)
		return
	}

	apiUsers := make([]dto.UserResponse, 0, len(dbUsers))
	for _, user := range dbUsers {
		apiUsers = append(apiUsers, s.toUserResponse(user))
	}

	s.RespondJSON(w, http.StatusOK, apiUsers)
}

// @Summary Get user profile
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.PublicUserResponse
// @Router /users/{id} [get]
func (s *Server) GetUsersId(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	authUser, hasAuth := s.User(r)

	dbUser, err := s.DB.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.JsonError(w, "not_found", "User not found", http.StatusNotFound)
		} else {
			s.Log.Error("Database error", "err", err)
			s.JsonError(w, "database_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	if hasAuth && (authUser.ID == id || authUser.Role == "admin") {
		s.RespondJSON(w, http.StatusOK, s.toUserResponse(dbUser))
		return
	}

	if dbUser.Private == 1 {
		s.RespondJSON(w, http.StatusOK, dto.PublicUserResponse{
			ID:        dbUser.ID,
			Name:      "Anonym",
			Role:      dbUser.Role,
			Private:   1,
			CreatedAt: dbUser.CreatedAt,
			AvatarUrl: s.FormatAvatarURL(dbUser.AvatarUrl, dbUser.ID, 1),
		})
		return
	}

	s.RespondJSON(w, http.StatusOK, dto.PublicUserResponse{
		ID:        dbUser.ID,
		Name:      dbUser.Name,
		Role:      dbUser.Role,
		Active:    dbUser.Active,
		Verified:  dbUser.Verified,
		Programid: dbUser.Programid,
		CreatedAt: dbUser.CreatedAt,
		Theme:     dbUser.Theme,
		Private:   dbUser.Private,
		AvatarUrl: s.FormatAvatarURL(dbUser.AvatarUrl, dbUser.ID, dbUser.Private),
	})
}

// @Summary Get user avatar
// @Tags Auth
// @Param userId path string true "User ID"
// @Param filename path string true "Filename"
// @Produce image/svg+xml
// @Success 200 {file} binary
// @Router /auth/avatars/{userId}/{filename} [get]
func (s *Server) GetAvatar(w http.ResponseWriter, r *http.Request) {
	userID := filepath.Base(chi.URLParam(r, "userId"))
	filename := filepath.Base(chi.URLParam(r, "filename"))

	if !strings.HasSuffix(filename, ".svg") {
		s.JsonError(w, "bad_request", "Only SVG avatars are supported", http.StatusBadRequest)
		return
	}

	objectName := fmt.Sprintf("avatars/%s/%s", userID, filename)

	obj, err := s.Store.GetObject(r.Context(), objectName)
	if err != nil {
		data := s.Avatars.GenerateSVG(userID)

		// Store it in the bucket asynchronously or just serve it
		go func() {
			if err := s.Avatars.StoreAvatar(context.Background(), objectName, data); err != nil {
				s.Log.Error("Failed to cache generated avatar", "userID", userID, "err", err)
			}
		}()

		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'unsafe-inline'")
		w.Write(data)
		return
	}
	defer obj.Close()

	info, err := obj.Stat()
	if err != nil {
		data := s.Avatars.GenerateSVG(userID)

		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'unsafe-inline'")
		w.Write(data)
		return
	}

	w.Header().Set("Content-Type", info.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size))
	w.Header().Set("ETag", info.ETag)
	w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'unsafe-inline'")

	if strings.HasSuffix(filename, ".svg") && !strings.Contains(filename, "generated") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}

	if r.Header.Get("If-None-Match") == info.ETag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	io.Copy(w, obj)
}
