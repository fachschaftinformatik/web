package dto

import "github.com/fachschaftinformatik/web/internal/database"

// @Description Public user information
type UserResponse struct {
    database.User
    // Explicitly ignore sensitive fields using the json tag "-"
    Password string `json:"-"` 
}

// @Description Credentials for login
type LoginRequest struct {
    Email    string `json:"email" example:"user@studmail.w-hs.de"`
    Password string `json:"password" example:"secret123"`
}
