package users

import (
	"strings"
	"unicode/utf8"
)

// RegistrationErrors validates account creation using the GOV.UK password guidance.
// Passwords are deliberately not trimmed: whitespace can be part of a password.
func RegistrationErrors(name, email, password string) (nameError, emailError, passwordError string) {
	nameError = NameError(name)
	emailError = validateEmail(email, false)
	passwordError = PasswordError(password)
	return
}

func NameError(name string) string {
	switch length := utf8.RuneCountInString(strings.TrimSpace(name)); {
	case length == 0:
		return "Bitte gib einen Namen an."
	case length < 2 || length > 64:
		return "Der Name muss zwischen 2 und 64 Zeichen lang sein."
	}
	return ""
}

func PasswordError(password string) string {
	switch length := utf8.RuneCountInString(password); {
	case length == 0:
		return "Bitte lege ein Passwort fest."
	case length < 8:
		return "Das Passwort muss mindestens 8 Zeichen lang sein."
	}
	return ""
}
