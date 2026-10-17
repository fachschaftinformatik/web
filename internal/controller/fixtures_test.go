package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/model/forum"
	"github.com/fachschaftinformatik/web/internal/model/users"
	"golang.org/x/crypto/bcrypt"
)

// testApplication always uses the same migrated SQLite runtime as production.
func testApplication(t *testing.T, records ...users.User) (*Handler, *sql.DB) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sqlite.db")
	if err := database.Migrate(path, log.New(io.Discard, "", 0)); err != nil {
		t.Fatal(err)
	}
	db, err := database.NewConnection(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	hash, err := bcrypt.GenerateFromPassword([]byte("test password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO users(id,email,name,password,role,active,verified) VALUES (900,'admin@fsv-wh.de','Administrator',?,'admin',1,1),(901,'other@fsv-wh.de','Other administrator',?,'admin',1,1)`, string(hash), string(hash))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) == 0 {
		records = []users.User{{ID: "42", Name: "Kim Muster", Email: "kim@studmail.w-hs.de", Active: true}, {ID: "43", Name: "Alex Beispiel", Email: "alex@studmail.w-hs.de", Active: true}}
	}
	for _, user := range records {
		if _, err := db.Exec(`INSERT INTO users(id,email,name,password,role,active,verified) VALUES(?,?,?,?,'user',1,?)`, user.ID, strings.ToLower(strings.TrimSpace(user.Email)), user.Name, string(hash), user.Verified); err != nil {
			t.Fatal(err)
		}
	}
	handler, err := New(persistentServices(t, db, nil), false)
	if err != nil {
		t.Fatal(err)
	}
	return handler, db
}

func testHandler(t *testing.T) (*Handler, error) {
	t.Helper()
	handler, db := testApplication(t)
	seedPublicContent(t, db)
	return handler, nil
}

func seedPublicContent(t *testing.T, db *sql.DB) {
	t.Helper()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	var posts []forum.Post
	var threads []forum.Thread
	for path, target := range map[string]any{"snapshot.json": &posts, "threads.json": &threads} {
		data, err := os.ReadFile("../model/forum/testdata/" + path)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(data, target); err != nil {
			t.Fatal(err)
		}
	}
	for _, post := range posts {
		exec(`INSERT INTO discussion_posts(id,title,body,user_id,type,created_at) VALUES(?,?,?,900,?,?)`, post.ID, post.Title, html.UnescapeString(post.Body), post.Type, post.Created)
		for _, tag := range post.Tags {
			exec(`INSERT INTO discussion_post_tags(post_id,tag) VALUES(?,?)`, post.ID, tag)
		}
	}
	for _, thread := range threads {
		exec(`UPDATE discussion_posts SET event_date=?,location=? WHERE id=?`, thread.EventDate, thread.Location, thread.ID)
		for _, comment := range thread.Replies {
			exec(`INSERT INTO discussion_comments(id,post_id,user_id,text,created_at,updated_at) VALUES(?,?,900,?,?,?)`, comment.ID, thread.ID, html.UnescapeString(comment.Text), comment.Created, comment.Updated)
		}
	}
	exec(`INSERT INTO programs(id,name) VALUES(1,'Informatik (B. Sc.)'),(2,'Wirtschaftsinformatik (B. Sc.)'),(3,'Medieninformatik (B. Sc.)');
INSERT INTO program_versions(program_id,name) VALUES(1,'2017'),(1,'2023'),(2,'2017'),(2,'2023'),(3,'2023');
INSERT INTO modules(id,program_id,name) VALUES(10,1,'Algorithmen und Datenstrukturen'),(11,2,'Datenbanken'),(12,1,'Mathematik I'),(13,3,'Web Engineering'),(14,1,'Betriebssysteme'),(15,2,'Software Engineering');
UPDATE modules SET alias='SQL' WHERE id=11;`)
	for i := 0; i < 6; i++ {
		version := "2023"
		if i >= 4 {
			version = "2017"
		}
		exec(`INSERT INTO archive(id,group_id,user_id,module_id,version,exam_date,access_key,mime_type,nbytes,checksum) VALUES(?,?,900,?,?, '2026-01-01',?, 'application/pdf',5,'test')`, 100+i, 100+i, 10+i, version, fmt.Sprint(i))
	}
	album := "140125680309370881"
	exec(`INSERT INTO events(id,title,created_at) VALUES(?,'LAN Party 2025','2025-01-01')`, album)
	for _, id := range []string{"140126343856652289", "140126331227602945", "140126318313340929", "140126304396640257", "140126291885031425", "140126279717355521"} {
		exec(`INSERT INTO media(id,event_id,user_id,title,description,access_key,mime_type,nbytes,created_at) VALUES(?,?,900,'Foto','Teilnehmende',?,'image/jpeg',5,'2025-01-01')`, id, album, id)
	}
}

func (b *adminBrowser) loginAs(email string) {
	b.t.Helper()
	_, body, _ := b.get("/login")
	token := csrfFrom(b.t, body)
	status, _, headers := b.post("/login", url.Values{"email": {email}, "password": {"test password"}, "csrf": {token}})
	if status != 303 || headers.Get("Location") != "/verwaltung" {
		b.t.Fatalf("login failed: %d", status)
	}
	_, body, _ = b.get("/verwaltung")
	b.csrf = csrfFrom(b.t, body)
	if b.csrf == token {
		b.t.Fatal("session not rotated at login")
	}
}
