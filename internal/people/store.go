package people

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/mail"
	"slices"
	"strings"
	"sync"

	"github.com/fachschaftinformatik/web/internal/users"
)

type Person struct {
	ID, Name, Email string
	UserID          int64
	AccountEmail    string
	Confirmed       bool
}

func (p Person) Status() string {
	if p.UserID == 0 {
		return "Ohne Benutzerkonto"
	}
	if p.Confirmed {
		return "Bestätigt"
	}
	return "Eingeladen"
}

type Responsibility struct {
	ID, Group, Title, Email string
	MemberIDs               []string
}

// Store shares public profiles and assignments between public and admin views.
type Store struct {
	db               *sql.DB
	mu               sync.RWMutex
	people           []Person
	responsibilities []Responsibility
}

func NewStore() *Store {
	s := &Store{}
	ids := map[string]string{}
	for _, group := range Groups() {
		sections := group.Sections
		if group.ID == "board" {
			sections = nil
			for _, section := range group.Sections {
				for _, member := range section.Members {
					sections = append(sections, TeamSection{Title: member.Role, Members: []Member{member}})
				}
			}
		}
		for _, section := range sections {
			r := Responsibility{ID: fmt.Sprint(len(s.responsibilities) + 1), Group: group.ID, Title: section.Title, Email: section.SharedEmail()}
			for _, member := range section.Members {
				id, ok := ids[member.Name]
				if !ok {
					id = fmt.Sprint(len(s.people) + 1)
					ids[member.Name] = id
					s.people = append(s.people, Person{ID: id, Name: member.Name, Email: member.Email})
				}
				r.MemberIDs = append(r.MemberIDs, id)
			}
			s.responsibilities = append(s.responsibilities, r)
		}
	}
	return s
}

func (s *Store) People() []Person {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.people)
}

func (s *Store) Responsibilities(group string) []Responsibility {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Responsibility
	for _, r := range s.responsibilities {
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

func (s *Store) SavePerson(p Person) map[string]string {
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
	if p.ID == "neu" {
		p.ID = fmt.Sprint(len(s.people) + 1)
		s.people = append(s.people, p)
		if err := s.persist(); err != nil {
			s.people = s.people[:len(s.people)-1]
			log.Printf("Save member: %v", err)
			return map[string]string{"name": "Das Mitglied konnte nicht gespeichert werden."}
		}
		return nil
	}
	for i := range s.people {
		if s.people[i].ID == p.ID {
			old := s.people[i]
			// Account linkage is not editable through public profile fields.
			p.UserID, p.AccountEmail, p.Confirmed = old.UserID, old.AccountEmail, old.Confirmed
			if s.db != nil && old.UserID != 0 {
				if err := s.saveLinked(context.Background(), p); err != nil {
					log.Printf("Save member: %v", err)
					return map[string]string{"name": "Das Mitglied konnte nicht gespeichert werden."}
				}
				s.people[i] = p
				return nil
			}
			s.people[i] = p
			if err := s.persist(); err != nil {
				s.people[i] = old
				log.Printf("Save member: %v", err)
				return map[string]string{"name": "Das Mitglied konnte nicht gespeichert werden."}
			}
			return nil
		}
	}
	return map[string]string{"name": "Dieses Mitglied wurde nicht gefunden."}
}

func (s *Store) SaveResponsibility(r Responsibility) map[string]string {
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
		if !slices.ContainsFunc(s.people, func(p Person) bool { return p.ID == id }) {
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
	if r.ID == "neu" {
		r.ID = fmt.Sprint(len(s.responsibilities) + 1)
		s.responsibilities = append(s.responsibilities, r)
		if err := s.persist(); err != nil {
			s.responsibilities = s.responsibilities[:len(s.responsibilities)-1]
			log.Printf("Save task: %v", err)
			return map[string]string{"title": "Die Funktion konnte nicht gespeichert werden."}
		}
		return nil
	}
	for i, old := range s.responsibilities {
		if old.ID == r.ID {
			s.responsibilities[i] = r
			if err := s.persist(); err != nil {
				s.responsibilities[i] = old
				log.Printf("Save task: %v", err)
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
		for _, r := range s.responsibilities {
			if r.Group != groups[i].ID {
				continue
			}
			section := TeamSection{Title: r.Title}
			if r.Group == "board" {
				section.Title = "Vorstand"
			}
			for _, id := range r.MemberIDs {
				for _, p := range s.people {
					if p.ID != id {
						continue
					}
					email, role := p.Email, r.Title
					if r.Email != "" {
						email = r.Email
					}
					section.Members = append(section.Members, Member{Name: p.Name, Role: role, Email: email})
				}
			}
			groups[i].Sections = append(groups[i].Sections, section)
		}
	}
	return groups
}

func (s *Store) persist() error {
	if s.db == nil {
		return nil
	}
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := writeState(ctx, tx, s.people, s.responsibilities); err != nil {
		return err
	}
	return tx.Commit()
}
