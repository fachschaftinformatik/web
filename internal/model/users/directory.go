package users

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strconv"
	"strings"
)

// User is a login identity, independent of a public person profile.
type User struct {
	ID, Name, Email, Permission string
	Program, Status             string
	Verified                    bool
}

// Directory is an immutable account display snapshot. It contains neither
// credentials nor authentication decisions.
type Directory struct {
	users []User
}

func NewDirectory(records []User) *Directory {
	records = slices.Clone(records)
	for i := range records {
		records[i].Email = strings.ToLower(records[i].Email)
	}
	return &Directory{users: records}
}

func (d *Directory) Get(id string) (User, bool) {
	for _, user := range d.users {
		if user.ID == id {
			return user, true
		}
	}
	return User{}, false
}

const PerPage = 20

type Page struct {
	Users                []User
	Number, Pages, Total int
}

// Search applies the shared account search and page bounds. Empty results have
// one valid, empty page. Out-of-range page numbers are rejected, not clamped.
func (d *Directory) Search(query string, number int) (Page, bool) {
	records := search(d.users, query)
	page := Page{Number: number, Pages: max(1, (len(records)+PerPage-1)/PerPage), Total: len(records)}
	if number < 1 || number > page.Pages {
		return page, false
	}
	start := (number - 1) * PerPage
	page.Users = records[start:min(start+PerPage, len(records))]
	return page, true
}

func search(all []User, query string) []User {
	var result []User
	query = strings.ToLower(strings.TrimSpace(query))
	for _, user := range all {
		if strings.Contains(strings.ToLower(user.Name+" "+user.Email), query) {
			result = append(result, user)
		}
	}
	return result
}

type Program struct{ ID, Name string }

var ErrProfileNotFound = errors.New("public profile not found")

// PublicProfile contains only information a user has chosen to make public.
type PublicProfile struct {
	ID   int64
	Name string
}

func (s *Service) PublicProfile(ctx context.Context, value string) (PublicProfile, error) {
	var profile PublicProfile
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 || strconv.FormatInt(id, 10) != value {
		return profile, ErrProfileNotFound
	}
	err = s.db.QueryRowContext(ctx, `SELECT id,name FROM users WHERE id=? AND private=0 AND deleted_at IS NULL`, id).Scan(&profile.ID, &profile.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return profile, ErrProfileNotFound
	}
	return profile, err
}
