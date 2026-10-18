package view

import "github.com/fachschaftinformatik/web/internal/urls"

type Member struct {
	ID, Name, Email, PublicID  string
	UserID                     int64
	UserPublicID, AccountEmail string
	Confirmed                  bool
}
type MemberRole struct {
	ID, Group, Title, Email, PublicID string
}
type TeamMember struct{ Name, Initials, Role, Email string }
type TeamSection struct {
	Title, Group, SharedEmail string
	Members                   []TeamMember
}
type TeamGroup struct {
	ID, Title string
	Sections  []TeamSection
}

func (m Member) URL() string       { return urls.AdminMember(m.PublicID) }
func (m Member) EditURL() string   { return m.URL() + "/edit" }
func (m Member) InviteURL() string { return m.URL() + "/invite" }
func (m Member) UserURL() string   { return urls.AdminUser(m.UserPublicID) }
func (m Member) Status() string {
	if m.UserID == 0 {
		return "Ohne Benutzerkonto"
	}
	if m.Confirmed {
		return "Bestätigt"
	}
	return "Eingeladen"
}
