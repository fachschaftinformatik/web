package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

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
	"github.com/golang-jwt/jwt/v5"
	"github.com/microcosm-cc/bluemonday"
)

const (
	maxUploadSize = 256 << 20 // 256 MB
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

func (s *Server) Authenticate(w http.ResponseWriter, r *http.Request) (database.User, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return database.User{}, errors.New("authorization header missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return database.User{}, errors.New("invalid authorization header format")
	}

	tokenString := parts[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.Config.InternalJWTSecret), nil
	})

	if err != nil || !token.Valid {
		return database.User{}, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return database.User{}, errors.New("invalid token claims")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return database.User{}, errors.New("subject claim missing in token")
	}

	uid, err := id.Parse(sub)
	if err != nil {
		return database.User{}, errors.New("invalid user id in token")
	}

	user, err := s.DB.GetUser(r.Context(), int64(uid))
	if err != nil {
		return database.User{}, errors.New("user not found")
	}

	return user, nil
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
