package people

import (
	"fmt"
	"net/mail"
	"slices"
	"strings"
	"sync"
)

type Person struct{ ID, Name, Email, Role string }
type Responsibility struct {
	ID, Group, Title, Email string
	MemberIDs               []string
}

// Store is a process-local copy of the public team, shared by public and admin views.
type Store struct {
	mu               sync.RWMutex
	people           []Person
	responsibilities []Responsibility
}

func NewStore() *Store {
	s := &Store{}
	ids := map[string]string{}
	for _, group := range Groups() {
		for _, section := range group.Sections {
			r := Responsibility{ID: fmt.Sprint(len(s.responsibilities) + 1), Group: group.ID, Title: section.Title, Email: section.SharedEmail()}
			for _, member := range section.Members {
				id, ok := ids[member.Name]
				if !ok {
					id = fmt.Sprint(len(s.people) + 1)
					ids[member.Name] = id
					role := ""
					if group.ID == "board" {
						role = member.Role
					}
					s.people = append(s.people, Person{id, member.Name, member.Email, role})
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
	p.Name, p.Email, p.Role = strings.TrimSpace(p.Name), strings.TrimSpace(p.Email), strings.TrimSpace(p.Role)
	if p.Name == "" {
		errors["name"] = "Gib einen Namen ein."
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
		return nil
	}
	for i := range s.people {
		if s.people[i].ID == p.ID {
			s.people[i] = p
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
	if r.Group != "departments" && r.Group != "project" {
		errors["group"] = "Wähle Referat oder Projektstelle aus."
	}
	for _, id := range r.MemberIDs {
		if !slices.ContainsFunc(s.people, func(p Person) bool { return p.ID == id }) {
			errors["members"] = "Wähle vorhandene Mitglieder aus."
		}
	}
	if len(errors) != 0 {
		return errors
	}
	r.MemberIDs = slices.Clone(r.MemberIDs)
	if r.ID == "neu" {
		r.ID = fmt.Sprint(len(s.responsibilities) + 1)
		s.responsibilities = append(s.responsibilities, r)
		return nil
	}
	for i, old := range s.responsibilities {
		if old.ID == r.ID && old.Group != "board" {
			s.responsibilities[i] = r
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
			for _, id := range r.MemberIDs {
				for _, p := range s.people {
					if p.ID != id {
						continue
					}
					email, role := p.Email, r.Title
					if r.Email != "" {
						email = r.Email
					}
					if r.Group == "board" {
						role = p.Role
					}
					section.Members = append(section.Members, Member{Name: p.Name, Role: role, Email: email})
				}
			}
			groups[i].Sections = append(groups[i].Sections, section)
		}
	}
	return groups
}
