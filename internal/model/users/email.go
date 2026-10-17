// Package users manages login identities, access policies and directory operations.
package users

import (
	"net/mail"
	"slices"
	"strings"
)

type DomainPolicy struct{ Registration, Accounts []string }

func (s *Service) Domains() DomainPolicy {
	return DomainPolicy{Registration: slices.Clone(s.domains.Registration), Accounts: slices.Clone(s.domains.Accounts)}
}

// EmailError validates a complete mailbox against the account-domain policy.
// Callers may trim surrounding whitespace but must preserve the local part.
func (s *Service) EmailError(email string) string { return validateEmail(email, s.domains.Accounts) }

func validateEmail(email string, domains []string) string {
	if email == "" {
		return "Gib deine E-Mail-Adresse ein."
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Name != "" || (address.Address != email && address.String() != "<"+email+">") {
		return "Gib eine vollständige E-Mail-Adresse ein."
	}
	domain := email[strings.LastIndex(email, "@")+1:]
	for _, allowed := range domains {
		if strings.EqualFold(domain, allowed) {
			return ""
		}
	}
	if len(domains) == 0 {
		return "Die Anmeldung mit dieser E-Mail-Domain ist nicht freigeschaltet."
	}
	labels := make([]string, len(domains))
	for i, d := range domains {
		labels[i] = "@" + d
	}
	return "Verwende eine E-Mail-Adresse mit " + strings.Join(labels, " oder ") + "."
}
