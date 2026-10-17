package users

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// RegistrationErrors validates account creation using the GOV.UK password guidance.
// Passwords are deliberately not trimmed: whitespace can be part of a password.
func (s *Service) RegistrationErrors(name, email, password string) (nameError, emailError, passwordError string) {
	nameError = NameError(name)
	emailError = validateEmail(email, s.domains.Registration)
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

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	b64 := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=2$%s$%s", b64.EncodeToString(salt), b64.EncodeToString(key)), nil
}

// VerifyPassword retains V1 bcrypt compatibility. New hashes use Argon2id,
// which accepts the full UTF-8 password without bcrypt's 72-byte limit.
func VerifyPassword(hash, password string) bool {
	if strings.HasPrefix(hash, "$2") {
		return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
	}
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=65536,t=3,p=2" {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != 16 {
		return false
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(key) != 32 {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	return subtle.ConstantTimeCompare(actual, key) == 1
}
