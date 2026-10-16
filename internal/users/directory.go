package users

import (
	"database/sql"
	"net/url"
	"path/filepath"
	"slices"
	"strings"

	_ "modernc.org/sqlite"
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

// Examples returns fictional users for the administration demo.
func Examples() []User {
	return []User{
		{ID: "admin", Name: "Demo-Administration", Email: "admin@fsv-wh.de", Permission: "FSV", Status: "Aktiv", Verified: true},
		{ID: "1", Name: "Alex Beispiel", Email: "alex.beispiel@studmail.w-hs.de", Permission: "Mitglied", Status: "Aktiv", Verified: true},
		{ID: "2", Name: "Kim Muster", Email: "kim.muster@studmail.w-hs.de", Permission: "FSV", Status: "Aktiv", Verified: true},
		{ID: "3", Name: "Robin Beispiel", Email: "robin.beispiel@studmail.w-hs.de", Permission: "Mitglied", Status: "Inaktiv"},
	}
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

// Load reads only display data. Passwords, tokens and production sessions are
// never selected. The source is opened read-only and closed after the snapshot.
func Load(path string) ([]User, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	u := url.URL{Scheme: "file", Path: abs, RawQuery: "mode=ro"}
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT CAST(u.id AS TEXT), u.name, u.email, u.role, COALESCE(p.name,''), CASE WHEN u.active = 1 THEN 'Aktiv' ELSE 'Inaktiv' END, u.verified = 1 FROM users u LEFT JOIN programs p ON p.id = u.program_id WHERE u.deleted_at IS NULL ORDER BY u.name COLLATE NOCASE, u.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Permission, &user.Program, &user.Status, &user.Verified); err != nil {
			return nil, err
		}
		switch user.Permission {
		case "admin":
			user.Permission = "Administration"
		case "fsv":
			user.Permission = "FSV"
		case "user":
			user.Permission = "Mitglied"
		case "editor":
			user.Permission = "Redaktion"
		}
		records = append(records, user)
	}
	return records, rows.Err()
}
