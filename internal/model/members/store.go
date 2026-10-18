package members

import (
	"context"
	"database/sql"
	"log"
	"net/mail"
	"slices"
	"strings"
	"sync"

	"github.com/fachschaftinformatik/web/internal/model/users"
	"github.com/fachschaftinformatik/web/internal/publicid"
)

type TeamMember struct {
	Name, Initials, Role, Email string
}

type TeamSection struct {
	Title, Group string
	Members      []TeamMember
}

type Group struct {
	ID, Title string
	Sections  []TeamSection
}

// SharedEmail avoids repeating a role's shared address for each member.
func (s TeamSection) SharedEmail() string {
	if len(s.Members) < 2 || s.Members[0].Email == "" {
		return ""
	}
	email := s.Members[0].Email
	for _, member := range s.Members[1:] {
		if member.Email != email {
			return ""
		}
	}
	return email
}

type Member struct {
	ID, Name, Email string
	PublicID        string
	UserID          int64
	UserPublicID    string
	AccountEmail    string
	Confirmed       bool
}

type MemberRole struct {
	ID, Group, Title, Email string
	PublicID                string
	MemberIDs               []string
}

// Store shares public profiles and assignments between public and admin views.
type Store struct {
	db      *sql.DB
	mu      sync.RWMutex
	members []Member
	roles   []MemberRole
}

func (s *Store) Members() []Member {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.members)
}

func (s *Store) Roles(group string) []MemberRole {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []MemberRole
	for _, r := range s.roles {
		if r.Group == group {
			r.MemberIDs = slices.Clone(r.MemberIDs)
			result = append(result, r)
		}
	}
	return result
}

func validEmail(value string) bool {
	if value == "" {
		return true
	}
	a, err := mail.ParseAddress(value)
	return err == nil && a.Address == value
}

func (s *Store) SaveMember(p Member) map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	errors := map[string]string{}
	p.Name, p.Email = strings.TrimSpace(p.Name), strings.TrimSpace(p.Email)
	if message := users.NameError(p.Name); message != "" {
		errors["name"] = message
	}
	if !validEmail(p.Email) {
		errors["email"] = "Gib eine gültige E-Mail-Adresse ein."
	}
	if len(errors) != 0 {
		return errors
	}
	if p.ID == "create" {
		p.ID = publicid.New()
		p.PublicID = p.ID
		s.members = append(s.members, p)
		if err := s.persist(); err != nil {
			s.members = s.members[:len(s.members)-1]
			log.Printf("Save member: %v", err)
			return map[string]string{"name": "Das Mitglied konnte nicht gespeichert werden."}
		}
		return nil
	}
	for i := range s.members {
		if s.members[i].ID == p.ID {
			old := s.members[i]
			p.PublicID = old.PublicID
			// Account linkage is not editable through public profile fields.
			p.UserID, p.UserPublicID, p.AccountEmail, p.Confirmed = old.UserID, old.UserPublicID, old.AccountEmail, old.Confirmed
			if old.UserID != 0 {
				if err := s.saveLinked(context.Background(), p); err != nil {
					log.Printf("Save member: %v", err)
					return map[string]string{"name": "Das Mitglied konnte nicht gespeichert werden."}
				}
				s.members[i] = p
				return nil
			}
			s.members[i] = p
			if err := s.persist(); err != nil {
				s.members[i] = old
				log.Printf("Save member: %v", err)
				return map[string]string{"name": "Das Mitglied konnte nicht gespeichert werden."}
			}
			return nil
		}
	}
	return map[string]string{"name": "Dieses Mitglied wurde nicht gefunden."}
}

func (s *Store) SaveRole(r MemberRole) map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	errors := map[string]string{}
	r.Title, r.Email = strings.TrimSpace(r.Title), strings.TrimSpace(r.Email)
	if r.Title == "" {
		errors["title"] = "Gib eine Bezeichnung ein."
	}
	if !validEmail(r.Email) {
		errors["email"] = "Gib eine gültige E-Mail-Adresse ein."
	}
	if r.Group != "board" && r.Group != "departments" && r.Group != "project" {
		errors["group"] = "Wähle Vorstand, Referat oder Projektstelle aus."
	}
	for _, id := range r.MemberIDs {
		if !slices.ContainsFunc(s.members, func(p Member) bool { return p.ID == id }) {
			errors["members"] = "Wähle vorhandene Mitglieder aus."
		}
	}
	if len(errors) != 0 {
		return errors
	}
	var members []string
	for _, member := range r.MemberIDs {
		if !slices.Contains(members, member) {
			members = append(members, member)
		}
	}
	r.MemberIDs = members
	if r.ID == "create" {
		r.ID = publicid.New()
		r.PublicID = r.ID
		s.roles = append(s.roles, r)
		if err := s.persist(); err != nil {
			s.roles = s.roles[:len(s.roles)-1]
			log.Printf("Save role: %v", err)
			return map[string]string{"title": "Die Funktion konnte nicht gespeichert werden."}
		}
		return nil
	}
	for i, old := range s.roles {
		if old.ID == r.ID {
			r.PublicID = old.PublicID
			s.roles[i] = r
			if err := s.persist(); err != nil {
				s.roles[i] = old
				log.Printf("Save role: %v", err)
				return map[string]string{"title": "Die Funktion konnte nicht gespeichert werden."}
			}
			return nil
		}
	}
	return map[string]string{"title": "Dieser Eintrag wurde nicht gefunden."}
}

func (s *Store) Groups() []Group {
	s.mu.RLock()
	defer s.mu.RUnlock()
	groups := []Group{{ID: "board", Title: "Vorstand"}, {ID: "departments", Title: "Referate"}, {ID: "project", Title: "Projektstellen"}}
	for i := range groups {
		for _, r := range s.roles {
			if r.Group != groups[i].ID {
				continue
			}
			section := TeamSection{Title: r.Title}
			if r.Group == "board" {
				section.Title = "Vorstand"
			}
			for _, id := range r.MemberIDs {
				for _, p := range s.members {
					if p.ID != id {
						continue
					}
					email, role := p.Email, r.Title
					if r.Email != "" {
						email = r.Email
					}
					section.Members = append(section.Members, TeamMember{Name: p.Name, Role: role, Email: email})
				}
			}
			groups[i].Sections = append(groups[i].Sections, section)
		}
	}
	return groups
}

func (s *Store) persist() error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := writeState(ctx, tx, s.members, s.roles); err != nil {
		return err
	}
	return tx.Commit()
}
