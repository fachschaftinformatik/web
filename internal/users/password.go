package users

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

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
