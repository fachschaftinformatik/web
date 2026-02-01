package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/api/dto"
	"github.com/fachschaftinformatik/web/internal/api/middleware"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/id"
	"github.com/fachschaftinformatik/web/internal/storage"
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
	if !s.Config.SignupsEnabled {
		s.JsonError(w, "signups_disabled", "Registrierungen sind momentan deaktiviert.", http.StatusForbidden)
		return
	}

	var payload dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	if err := s.Validate(payload); err != nil {
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
	userID := id.New()

	role := "user"
	count, _ := s.DB.ListUsers(r.Context(), database.ListUsersParams{Limit: 1, Offset: 0})
	if len(count) == 0 {
		role = "admin"
	}

	var pid *int64
	if payload.ProgramID != nil {
		v := int64(*payload.ProgramID)
		pid = &v
	}

	params := database.CreateUserParams{
		ID:                int64(userID),
		Email:             payload.Email,
		Name:              s.Sanitize(payload.Name),
		Password:          string(hashedPassword),
		Role:              role,
		Active:            1,
		ProgramID:         pid,
		VerificationToken: &verificationToken,
	}

	avatarPath, err := s.Avatars.GenerateAndStoreAvatar(r.Context(), userID.String())
	if err != nil {
		s.Log.Error("Failed to generate avatar", "userID", userID, "err", err)
	} else {
		params.AvatarUrl = &avatarPath
	}

	dbUser, err := s.DB.CreateUser(r.Context(), params)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			s.JsonError(w, "email_exists", "A user with this email already exists", http.StatusConflict)
		} else {
			s.Log.Error("Registration: Database error", "err", err)
			s.JsonError(w, "database_error", "Could not create user", http.StatusInternalServerError)
		}
		return
	}

	if s.Config.SignupsVerify && role != "admin" {
		go func() {
			if err := s.Email.SendVerificationEmail(dbUser.Email, dbUser.Name, verificationToken); err != nil {
				s.Log.Error("Failed to send verification email", "email", dbUser.Email, "err", err)
			}
		}()
	} else {
		_, err = s.DB.VerifyUser(r.Context(), dbUser.ID)
		if err != nil {
			s.Log.Error("Failed to auto-verify user", "err", err)
		}
		dbUser.Verified = 1
	}

	s.RespondJSON(w, http.StatusCreated, s.toUserResponse(dbUser))
}

// @Summary Request password reset
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "{email: string}"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Router /auth/forgot [post]
func (s *Server) PostAuthForgot(w http.ResponseWriter, r *http.Request) {
	var payload struct{ Email string `json:"email"` }
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Email == "" {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	// Do not reveal whether email exists. If user exists, create token and send email.
	user, err := s.DB.GetUserByEmail(r.Context(), payload.Email)
	if err == nil {
		token := uuid.NewString()
		expires := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
		t := token
		if err := s.DB.SetPasswordResetToken(r.Context(), database.SetPasswordResetTokenParams{
			PasswordResetToken:   &t,
			PasswordResetExpires: &expires,
			ID:                   user.ID,
		}); err != nil {
			s.Log.Error("Failed to set password reset token", "email", payload.Email, "err", err)
			s.JsonError(w, "internal_error", "Failed to process request", http.StatusInternalServerError)
			return
		}
		go func(emailAddr, name, tk string) {
			if err := s.Email.SendPasswordResetEmail(emailAddr, name, tk); err != nil {
				s.Log.Error("Failed to send password reset email", "email", emailAddr, "err", err)
			}
		}(user.Email, user.Name, token)
	}

	// Always return 204 (no content) to avoid disclosing account existence
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Reset password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "{token: string, password: string}"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /auth/reset [post]
func (s *Server) PostAuthReset(w http.ResponseWriter, r *http.Request) {
	var payload struct{
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Token == "" || payload.Password == "" {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	tok := payload.Token
	dbUser, err := s.DB.GetUserByPasswordResetToken(r.Context(), &tok)
	if err != nil {
		s.JsonError(w, "invalid_token", "Invalid or expired token", http.StatusBadRequest)
		return
	}

	// expiry is enforced by the query in GetUserByPasswordResetToken

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		s.JsonError(w, "server_error", "Could not process password", http.StatusInternalServerError)
		return
	}

	updatedUser, err := s.DB.UpdateUserPassword(r.Context(), database.UpdateUserPasswordParams{
		Password: string(hashedPassword),
		ID:       dbUser.ID,
	})
	if err != nil {
		s.Log.Error("Failed to update password", "err", err)
		s.JsonError(w, "server_error", "Could not reset password", http.StatusInternalServerError)
		return
	}

	// clear token (already cleared by UpdateUserPassword)

	s.RespondJSON(w, http.StatusOK, s.toUserResponse(updatedUser))
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
	var payload dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.JsonError(w, "invalid_request_body", "Could not decode JSON body", http.StatusBadRequest)
		return
	}

	dbUser, err := s.DB.GetUserByEmail(r.Context(), payload.Email)
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
		s.JsonError(w, "email_not_verified", "Du musst erst deine E-Mail bestätigen.", http.StatusForbidden)
		return
	}

	s.SetCookie(w, SessionCookieName, "", -time.Hour, true)

	sessionID := id.New().String()
	duration := SessionDuration
	if payload.Remember {
		duration = 30 * 24 * time.Hour
	}
	expiresAt := time.Now().Add(duration)

	ua := r.UserAgent()
	ip := r.RemoteAddr

	_, err = s.DB.CreateSession(r.Context(), database.CreateSessionParams{
		ID:        sessionID,
		UserID:    dbUser.ID,
		ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
		UserAgent: &ua,
		IpAddress: &ip,
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
		http.Redirect(w, r, fmt.Sprintf("%s/login?verified=false&error=not_found", s.Config.Domain), http.StatusFound)
		return
	}

	_, err = s.DB.VerifyUser(r.Context(), dbUser.ID)
	if err != nil {
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

	var pid *int64
	if payload.ProgramID != nil {
		v := int64(*payload.ProgramID)
		pid = &v
	}

	updatedUser, err := s.DB.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:        authUser.ID,
		Name:      s.Sanitize(payload.Name),
		ProgramID: pid,
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
// @Success 204
// @Router /auth/logout [post]
func (s *Server) PostAuthLogout(w http.ResponseWriter, r *http.Request) {
	session, ok := r.Context().Value(middleware.SessionKey).(database.Session)
	if !ok {
		s.JsonError(w, "unauthorized", "Not logged in", http.StatusUnauthorized)
		return
	}

	_ = s.DB.DeleteSession(r.Context(), session.ID)

	s.SetCookie(w, SessionCookieName, "", -time.Hour, true)
	s.SetCookie(w, CsrfCookieName, "", -time.Hour, false)

	// Thoroughly flush browser state
	w.Header().Set("Clear-Site-Data", "\"cookies\", \"cache\"")
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
	s.RespondJSON(w, http.StatusOK, dto.CsrfResponse{
		Csrf:           csrfToken,
		SignupsEnabled: s.Config.SignupsEnabled,
	})
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
// @Param userId path string true "User ID"
// @Success 200 {object} dto.PublicUserResponse
// @Router /users/{userId} [get]
func (s *Server) GetUsersId(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userId")
	uid, err := id.Parse(idStr)
	if err != nil {
		s.JsonError(w, "invalid_id", "Invalid user ID", http.StatusBadRequest)
		return
	}

	authUser, hasAuth := s.User(r)

	dbUser, err := s.DB.GetUser(r.Context(), int64(uid))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.JsonError(w, "not_found", "User not found", http.StatusNotFound)
		} else {
			s.Log.Error("Database error", "err", err)
			s.JsonError(w, "database_error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	if hasAuth && (authUser.ID == dbUser.ID || authUser.Role == "admin") {
		s.RespondJSON(w, http.StatusOK, s.toUserResponse(dbUser))
		return
	}

	s.RespondJSON(w, http.StatusOK, s.ToPublicUserResponse(dbUser))
}

// @Summary Upload avatar
// @Tags Auth
// @Accept mpfd
// @Produce json
// @Param file formData file true "Avatar Image"
// @Success 200 {object} dto.UserResponse
// @Router /auth/me/avatar [post]
func (s *Server) PostAuthAvatar(w http.ResponseWriter, r *http.Request) {
	user, _ := s.User(r)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		s.JsonError(w, "bad_request", "Upload failed", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.JsonError(w, "bad_request", "File upload failed", http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		s.JsonError(w, "bad_request", "Only images allowed", http.StatusBadRequest)
		return
	}

	userIDStr := id.ID(user.ID).String()
	objectKey := storage.AvatarSourceKey(userIDStr)

	if err := s.Store.Upload(r.Context(), objectKey, file, header.Size, contentType); err != nil {
		s.Log.Error("S3 Upload failed", "err", err)
		s.JsonError(w, "server_error", "Upload failed", http.StatusInternalServerError)
		return
	}

	updatedUser, err := s.DB.UpdateUserAvatar(r.Context(), database.UpdateUserAvatarParams{
		ID:        user.ID,
		AvatarUrl: &objectKey,
	})
	if err != nil {
		s.Log.Error("Failed to update user avatar", "err", err)
		s.JsonError(w, "database_error", "Failed to update profile", http.StatusInternalServerError)
		return
	}

	s.RespondJSON(w, http.StatusOK, s.toUserResponse(updatedUser))
}

// @Summary Get user avatar
// @Tags Auth
// @Param userId path string true "User ID"
// @Param filename path string true "Filename"
// @Param size query int false "Size (ignored, always serves source)"
// @Produce image/svg+xml,image/jpeg,image/png
// @Success 200 {file} binary
// @Router /auth/avatars/{userId}/{filename} [get]
func (s *Server) GetAvatar(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")

	// Always serve from the 'source' path in storage
	objectName := storage.AvatarSourceKey(userIDStr)
	obj, err := s.Store.GetObject(r.Context(), objectName)

	if err == nil {
		defer obj.Close()
		w.Header().Set("Content-Type", obj.Info.ContentType)
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = io.Copy(w, obj)
		return
	}

	http.Error(w, "Avatar not found", http.StatusNotFound)
}
