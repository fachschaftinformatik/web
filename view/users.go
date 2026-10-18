package view

import "github.com/fachschaftinformatik/web/internal/urls"

type User struct {
	ID                                           int64
	Name, Email, Role, PublicID, Program, Degree string
	Active, Verified                             bool
}
type Program struct{ ID, Name string }

func (u User) URL() string        { return urls.AdminUser(u.PublicID) }
func (u User) ResetURL() string   { return u.URL() + "/reset" }
func (u User) ConfirmURL() string { return u.URL() + "/confirm" }
func (u User) Permission() string {
	switch u.Role {
	case "admin":
		return "Administration"
	case "editor":
		return "Redaktion"
	case "user":
		return "Studierende"
	}
	return u.Role
}
func (u User) Status() string {
	if u.Active {
		return "Aktiv"
	}
	return "Inaktiv"
}
