package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/model/users"
)

func TestMemberInvitationRegistrationJourney(t *testing.T) {
	b, db := identityBrowser(t)
	if _, err := db.Exec(`INSERT INTO users(id,email,name,password,role,active,verified) VALUES (42,'admin@fsv-wh.de','Admin','unused','admin',1,1)`); err != nil {
		t.Fatal(err)
	}
	session, err := users.NewService(db, "http://example.org", users.DomainPolicy{Registration: []string{"studmail.w-hs.de"}, Accounts: []string{"studmail.w-hs.de", "fsv-wh.de"}}).CreateSession(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	b.client.Jar.SetCookies(mustURL(t, b.server.URL), []*http.Cookie{{Name: sessionCookie, Value: session.ID, Path: "/"}})
	status, body, _ := b.get("/verwaltung/mitglieder/neu")
	if status != 200 || strings.Contains(body, "Aufgabe im Vorstand") {
		t.Fatal("invitation form unavailable or still asks for board role")
	}
	csrf := csrfFrom(t, body)
	if status, _, _ := b.post("/verwaltung/mitglieder/neu", url.Values{"csrf": {csrf}, "name": {"Neue Person"}, "email": {"new@studmail.w-hs.de"}}); status != 303 {
		t.Fatal("invitation failed")
	}
	if status, body, _ := b.get("/verwaltung/mitglieder"); status != 200 || !strings.Contains(body, "new@studmail.w-hs.de") || !strings.Contains(body, "Eingeladen") {
		t.Fatal("pending invitation not visible")
	}
	var link string
	if err := db.QueryRow(`SELECT link FROM mail_outbox WHERE recipient='new@studmail.w-hs.de'`).Scan(&link); err != nil {
		t.Fatal(err)
	}
	target, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	invitee := newAdminBrowser(t, b.server)
	status, body, _ = invitee.get(target.RequestURI())
	if status != 200 {
		t.Fatal("invitation registration unavailable")
	}
	var verified int
	if err := db.QueryRow(`SELECT verified FROM users WHERE email='new@studmail.w-hs.de'`).Scan(&verified); err != nil || verified != 0 {
		t.Fatal("GET consumed invitation")
	}
	form := url.Values{"csrf": {csrfFrom(t, body)}, "token": {target.Query().Get("token")}, "password": {"secure 🔐🔐🔐"}}
	if status, body, _ := invitee.post("/invite", form); status != 200 || !strings.Contains(body, "Dein Konto ist bestätigt") {
		t.Fatal("invitation confirmation failed")
	}
	_, body, _ = invitee.get("/login")
	if status, _, headers := invitee.post("/login", url.Values{"csrf": {csrfFrom(t, body)}, "email": {"new@studmail.w-hs.de"}, "password": {"secure 🔐🔐🔐"}}); status != 303 || headers.Get("Location") != "/verwaltung" {
		t.Fatal("confirmed member could not log in as reviewer")
	}
	if status, _, _ := invitee.get("/verwaltung/einreichungen"); status != 200 {
		t.Fatal("reviewer cannot access submissions")
	}
	for _, path := range []string{"/verwaltung/benutzer", "/verwaltung/mitglieder", "/verwaltung/funktionen", "/verwaltung/funktionen/neu", "/verwaltung/sitzungen"} {
		if status, _, _ := invitee.get(path); status != 403 {
			t.Fatalf("invitation granted administration of %s", path)
		}
	}
	if status, _, _ := invitee.get(target.RequestURI()); status != 422 {
		t.Fatal("consumed invitation remained usable")
	}
}

func TestExistingMemberInvitationAndAccountEditing(t *testing.T) {
	b, db := identityBrowser(t)
	if _, err := db.Exec(`INSERT INTO users(id,email,name,password,role,active,verified) VALUES (42,'admin@fsv-wh.de','Admin','unused','admin',1,1)`); err != nil {
		t.Fatal(err)
	}
	session, err := users.NewService(db, "http://example.org", users.DomainPolicy{Registration: []string{"studmail.w-hs.de"}, Accounts: []string{"studmail.w-hs.de", "fsv-wh.de"}}).CreateSession(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	b.client.Jar.SetCookies(mustURL(t, b.server.URL), []*http.Cookie{{Name: sessionCookie, Value: session.ID, Path: "/"}})
	status, body, _ := b.get("/verwaltung/mitglieder/1")
	if status != 200 || !strings.Contains(body, "E-Mail-Adresse des Benutzerkontos") {
		t.Fatal("existing member cannot be linked to an account")
	}
	csrf := csrfFrom(t, body)
	if status, _, _ := b.post("/verwaltung/mitglieder/1", url.Values{"csrf": {csrf}, "name": {"Linked Member"}, "email": {"linked@studmail.w-hs.de"}}); status != 303 {
		t.Fatal("existing member invitation failed")
	}
	var userID int64
	if err := db.QueryRow(`SELECT user_id FROM people WHERE id='1'`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM people WHERE user_id=?`, userID).Scan(&count); err != nil || count != 1 {
		t.Fatal("invitation duplicated the membership")
	}
	if status, body, _ := b.get("/team"); status != 200 || !strings.Contains(body, "Linked Member") {
		t.Fatal("pending confirmation hid the public member")
	}
	status, body, _ = b.get("/verwaltung/mitglieder/1")
	accountPath := fmt.Sprintf("/verwaltung/benutzer/%d", userID)
	if status != 200 || !strings.Contains(body, accountPath) || strings.Contains(body, `name="email"`) {
		t.Fatal("linked member is not using the account identity")
	}
	if status, _, _ := b.post("/verwaltung/mitglieder/1", url.Values{"csrf": {csrfFrom(t, body)}, "name": {"Updated Member"}, "email": {"unexpected@studmail.w-hs.de"}}); status != 303 {
		t.Fatal("linked member name edit failed")
	}
	if status, body, _ := b.get(accountPath); status != 200 || !strings.Contains(body, "Updated Member") || !strings.Contains(body, "linked@studmail.w-hs.de") || !strings.Contains(body, `href="/verwaltung/mitglieder/1"`) {
		t.Fatal("account and membership are not connected")
	}
}
