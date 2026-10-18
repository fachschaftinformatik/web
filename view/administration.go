package view

import (
	"net/url"
	"strconv"
)

type AdminMemberRow struct {
	Member
	Roles []string
}

type AdminRoleRow struct {
	Title, Kind, URL string
	Members          []string
}

func MemberRoleKind(group string) string {
	if group == "board" {
		return "Vorstand"
	}
	if group == "departments" {
		return "Referat"
	}
	return "Projektstelle"
}

type AdminLink struct{ ID, Label, URL, Group string }

var AdminLinks = []AdminLink{
	{"settings", "Stammdaten", "/admin/settings", "Organisation"},
	{"members", "Mitglieder", "/admin/members", "Organisation"},
	{"roles", "Funktionen", "/admin/roles", "Organisation"},
	{"meetings", "Sitzungen", "/admin/meetings", "Organisation"},
	{"submissions", "Einreichungen", "/admin/submissions", "Vorgänge"},
	{"users", "Benutzerkonten", "/admin/users", "Zugänge"},
}

type AdminPage struct {
	OrganizationSettings                                         []OrganizationSettingsTask
	CanResendConfirmation                                        bool
	Member                                                       *Member
	Inviting                                                     bool
	ReviewerOnly                                                 bool
	Submissions                                                  []Submission
	Section, SectionTitle, View, Action, Query, Message, Success string
	Fields                                                       []AdminField
	Members                                                      []AdminMemberRow
	Roles                                                        []AdminRoleRow
	Users                                                        []User
	User                                                         *User
	Meetings                                                     []Meeting
	Publication                                                  *Publication
	Page, Pages, Total                                           int
	PreviousURL, NextURL                                         string
}

type AdminGroup struct {
	Title string
	Links []AdminLink
}

func (a AdminPage) Groups() []AdminGroup {
	var groups []AdminGroup
	for _, link := range AdminLinks {
		if a.ReviewerOnly && link.ID != "submissions" {
			continue
		}
		if len(groups) == 0 || groups[len(groups)-1].Title != link.Group {
			groups = append(groups, AdminGroup{Title: link.Group})
		}
		group := &groups[len(groups)-1]
		group.Links = append(group.Links, link)
	}
	return groups
}

func (a AdminPage) SubmissionList() SubmissionListPage {
	return SubmissionListPage{Submissions: a.Submissions, Reviewing: true, CanViewUsers: !a.ReviewerOnly}
}

func (a AdminPage) UserPageURL(page int) string {
	q := url.Values{}
	if a.Query != "" {
		q.Set("search", a.Query)
	}
	if page != 1 {
		q.Set("page", strconv.Itoa(page))
	}
	if len(q) == 0 {
		return "/admin/users"
	}
	return "/admin/users?" + q.Encode()
}

func (a AdminPage) Pagination() []PaginationItem {
	return paginationItems(a.Page, a.Pages, a.UserPageURL)
}

func (a AdminPage) HasErrors() bool {
	if a.Message != "" {
		return true
	}
	for _, Field := range a.Fields {
		if Field.Error != "" {
			return true
		}
	}
	return false
}
