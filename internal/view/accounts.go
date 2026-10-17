package view

import (
	"github.com/fachschaftinformatik/web/internal/model/users"
)

type ForgotPage struct {
	Email, EmailError string
	Submitted         bool
}

type InvitationPage struct {
	Value, Name, Email, Error, PasswordError string
	Done                                     bool
}

// LoginPage never retains a submitted password.
type LoginPage struct {
	Email, EmailError, PasswordError string
}

// RegisterPage never retains a submitted password.
type RegisterPage struct {
	Name, Email, NameError, EmailError, PasswordError string
	ProgramID, ProgramError                           string
	Programs                                          []users.Program
	Submitted                                         bool
}

type TokenPage struct {
	Value, Kind, Error string
	Done               bool
}
