package web

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/users"
)

func TestMemberInvitationRegistrationJourney(t *testing.T) {
	b, db := identityBrowser(t)
	if _, err := db.Exec(`INSERT INTO users(id,email,name,password,role,active,verified) VALUES (42,'admin@fsv-wh.de','Admin','unused','admin',1,1)`); err != nil {
		t.Fatal(err)
	}
	session, err := users.NewService(db, "http://example.org").CreateSession(context.Background(), 42)
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
