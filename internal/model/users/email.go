// Package users manages login identities, access policies and directory operations.
package users

import (
	"net/mail"
	"strings"
)

// EmailError validates a complete mailbox against the account-domain policy.
// Callers may trim surrounding whitespace but must preserve the local part.
func EmailError(email string) string {
	return validateEmail(email, true)
}

func validateEmail(email string, allowStaff bool) string {
	if email == "" {
		return "Gib deine E-Mail-Adresse ein."
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Name != "" || (address.Address != email && address.String() != "<"+email+">") {
		return "Gib eine vollständige E-Mail-Adresse ein, zum Beispiel name@studmail.w-hs.de."
	}
	domain := email[strings.LastIndex(email, "@")+1:]
	if !strings.EqualFold(domain, "studmail.w-hs.de") && !(allowStaff && strings.EqualFold(domain, "fsv-wh.de")) {
		if !allowStaff {
			return "Verwende eine E-Mail-Adresse mit @studmail.w-hs.de."
		}
		return "Verwende eine E-Mail-Adresse mit @studmail.w-hs.de oder @fsv-wh.de."
	}
	return ""
}
