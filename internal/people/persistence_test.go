package people

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/fachschaftinformatik/web/internal/data"
	"github.com/fachschaftinformatik/web/internal/users"
)

func membershipStore(t *testing.T) (*Store, *users.Service, *sql.DB) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sqlite.db")
	if err := data.Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	db, err := data.NewConnection(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	store, err := Open(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	return store, users.NewService(db, "https://example.org"), db
}

func inviteToken(t *testing.T, db *sql.DB) string {
	t.Helper()
	var link string
	if err := db.QueryRow(`SELECT link FROM mail_outbox WHERE kind='verify'`).Scan(&link); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	if u.Path != "/invite" {
		t.Fatal("invitation does not lead to registration completion")
	}
	return u.Query().Get("token")
}

func TestPersistentInvitationConfirmationAndAssignments(t *testing.T) {
	store, identity, db := membershipStore(t)
	ctx := context.Background()
	initial := len(store.People())
	if fields, err := store.Invite(ctx, identity, "Neue Person", "NEW@STUDMAIL.W-HS.DE"); err != nil || len(fields) != 0 {
		t.Fatalf("invite: %v %v", fields, err)
	}
	people := store.People()
	member := people[len(people)-1]
	if member.UserID == 0 || member.Confirmed || member.Status() != "Eingeladen" || member.InviteEmail != "new@studmail.w-hs.de" || member.Email != "" {
		t.Fatalf("invited member: %+v", member)
	}
	token := inviteToken(t, db)
	if fields, err := store.Invite(ctx, identity, "Duplicate Person", "new@studmail.w-hs.de"); err != nil || fields["email"] == "" {
		t.Fatal("duplicate membership accepted")
	}
	if _, err := identity.Invitation(ctx, token); err != nil {
		t.Fatal("duplicate invite invalidated existing link")
	}
	previous := token
	if queued, err := store.Resend(ctx, identity, member.ID); err != nil || !queued {
		t.Fatal("resending invitation failed")
	}
	token = inviteToken(t, db)
	if previous == token {
		t.Fatal("resend reused the invitation token")
	}
	if _, err := identity.Invitation(ctx, previous); !errors.Is(err, users.ErrToken) {
		t.Fatal("resend left old invitation usable")
	}
	if fields := store.SaveResponsibility(Responsibility{ID: "neu", Group: "project", Title: "Neue Aufgabe", MemberIDs: []string{member.ID}}); len(fields) != 0 {
		t.Fatal(fields)
	}
	for _, group := range store.Groups() {
		for _, section := range group.Sections {
			for _, p := range section.Members {
				if p.Name == member.Name {
					t.Fatal("unconfirmed invitation appeared publicly")
				}
			}
		}
	}
	restarted, err := Open(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if len(restarted.People()) != initial+1 {
		t.Fatal("restart lost invitations or reseeded profiles")
	}
	if err := identity.ConsumeToken(ctx, token, "verify", ""); !errors.Is(err, users.ErrToken) {
		t.Fatal("invitation bypassed password setup")
	}
	if err := identity.AcceptInvitation(ctx, token, "eight 🔐🔐🔐"); err != nil {
		t.Fatal(err)
	}
	if err := identity.AcceptInvitation(ctx, token, "another password"); !errors.Is(err, users.ErrToken) {
		t.Fatal("invitation link reused")
	}
	userID, err := identity.Login(ctx, "new@studmail.w-hs.de", "eight 🔐🔐🔐")
	if err != nil || userID != member.UserID {
		t.Fatal("confirmed invitation cannot log in")
	}
	session, err := identity.CreateSession(ctx, userID)
	if err != nil || !session.CanReview() || session.CanAdmin() {
		t.Fatal("invitation role policy incorrect")
	}
	if err := restarted.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	confirmed := false
	for _, p := range restarted.People() {
		if p.ID == member.ID {
			confirmed = p.Confirmed && p.Status() == "Bestätigt"
		}
	}
	if !confirmed {
		t.Fatal("confirmation not visible in membership administration")
	}
	found := false
	for _, group := range restarted.Groups() {
		for _, section := range group.Sections {
			for _, p := range section.Members {
				if p.Name == member.Name {
					found = true
					if p.Email != "" {
						t.Fatal("private invitation address published")
					}
				}
			}
		}
	}
	if !found {
		t.Fatal("confirmed task assignment not published or persisted")
	}
}

func TestExistingConfirmedAccountAndFailedInvitationArePreserved(t *testing.T) {
	store, identity, db := membershipStore(t)
	ctx := context.Background()
	hash, err := users.HashPassword("existing password")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users(id,email,name,password,role,active,verified) VALUES (42,'existing@fsv-wh.de','Existing',?,'admin',1,1)`, hash); err != nil {
		t.Fatal(err)
	}
	if fields, err := store.Invite(ctx, identity, "Existing Profile", "existing@fsv-wh.de"); err != nil || len(fields) != 0 {
		t.Fatal(fields, err)
	}
	var role, password string
	if err := db.QueryRow(`SELECT role,password FROM users WHERE id=42`).Scan(&role, &password); err != nil || role != "admin" || password != hash {
		t.Fatal("inviting a confirmed account changed credentials or downgraded role")
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM mail_outbox`).Scan(&count); err != nil || count != 0 {
		t.Fatal("confirmed account received password-setup invitation")
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_person BEFORE INSERT ON people BEGIN SELECT RAISE(FAIL,'forced membership failure'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Invite(ctx, identity, "Failed Profile", "failed@studmail.w-hs.de"); err == nil {
		t.Fatal("failed membership insertion reported success")
	}
	if err := db.QueryRow(`SELECT count(*) FROM users WHERE email='failed@studmail.w-hs.de'`).Scan(&count); err != nil || count != 0 {
		t.Fatal("failed invitation left a user account")
	}
	if err := db.QueryRow(`SELECT count(*) FROM mail_outbox`).Scan(&count); err != nil || count != 0 {
		t.Fatal("failed invitation queued mail")
	}
}
