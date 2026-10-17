package view

import (
	"net/url"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/model/archive"
	"github.com/fachschaftinformatik/web/internal/model/meetings"
	"github.com/fachschaftinformatik/web/internal/model/people"
	"github.com/fachschaftinformatik/web/internal/model/users"
)

type AdminMemberRow struct {
	people.Person
	Tasks []string
}

type AdminTaskRow struct {
	Title, Kind, URL string
	Members          []string
}

type AdminLink struct{ ID, Label, URL, Group string }

var AdminLinks = []AdminLink{
	{"mitglieder", "Mitglieder", "/verwaltung/mitglieder", "Organisation"},
	{"funktionen", "Funktionen", "/verwaltung/funktionen", "Organisation"},
	{"sitzungen", "Sitzungen", "/verwaltung/sitzungen", "Organisation"},
	{"einreichungen", "Einreichungen", "/verwaltung/einreichungen", "Vorgänge"},
	{"benutzer", "Benutzerkonten", "/verwaltung/benutzer", "Zugänge"},
}

type AdminPage struct {
	CanResendConfirmation                                        bool
	Person                                                       *people.Person
	Inviting                                                     bool
	ReviewerOnly                                                 bool
	Requests                                                     []archive.Request
	Section, SectionTitle, View, Action, Query, Message, Success string
	Fields                                                       []AdminField
	Members                                                      []AdminMemberRow
	Tasks                                                        []AdminTaskRow
	Users                                                        []users.User
	User                                                         *users.User
	Meetings                                                     []meetings.Meeting
	Publication                                                  *meetings.Publication
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
		if a.ReviewerOnly && link.ID != "einreichungen" {
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

func (a AdminPage) RequestList() RequestListPage {
	return RequestListPage{Requests: a.Requests, Reviewing: true, CanViewUsers: !a.ReviewerOnly}
}

func (a AdminPage) UserPageURL(page int) string {
	q := url.Values{}
	if a.Query != "" {
		q.Set("q", a.Query)
	}
	if page != 1 {
		q.Set("seite", strconv.Itoa(page))
	}
	if len(q) == 0 {
		return "/verwaltung/benutzer"
	}
	return "/verwaltung/benutzer?" + q.Encode()
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
