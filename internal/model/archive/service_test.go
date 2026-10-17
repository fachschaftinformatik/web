package archive

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/storage/disk"
)

func archiveService(t *testing.T) (*Service, *sql.DB) {
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
	_, err = db.Exec(`
INSERT INTO programs (id,name) VALUES (1,'Informatik (B. Sc.)'),(2,'Informatik (M. Sc.)');
INSERT INTO program_versions (program_id,name) VALUES (1,'2017'),(1,'2023'),(2,'2023');
INSERT INTO modules (id,program_id,name,alias) VALUES (10,1,'Datenbanken','SQL'),(11,2,'Datenbanken II',NULL),(12,1,'Ohne Upload',NULL);
INSERT INTO users (id,email,name,password,active,verified) VALUES (1,'archive@fsv-wh.de','Person','unused',1,1);
INSERT INTO archive (id,group_id,user_id,module_id,version,exam_date,is_latest,access_key,mime_type,nbytes,checksum) VALUES
(100,100,1,10,'2017','2024-01-01',0,'old','application/pdf',5,'checksum'),
(101,100,1,10,'2017','2024-01-01',1,'latest','application/pdf',5,'checksum'),
(102,102,1,10,'2023','2026-01-01',1,'new','application/pdf',5,'checksum'),
(103,103,1,11,'2023','2025-01-01',1,'master','application/pdf',5,'checksum');`)
	if err != nil {
		t.Fatal(err)
	}
	storage, err := disk.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { storage.Close() })
	for _, value := range []string{"100", "101", "102", "103"} {
		if err := storage.Upload(context.Background(), "archive/"+value+"/source.pdf", strings.NewReader("%PDF-"), 5, "application/pdf"); err != nil {
			t.Fatal(err)
		}
	}
	return NewService(db, storage), db
}

func TestPersistentCatalogLatestRevisionsAndFilters(t *testing.T) {
	s, db := archiveService(t)
	catalog, err := s.Catalog(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	modules := catalog.FindModules(Filter{})
	if len(modules) != 2 || modules[0].ID != "10" || len(modules[0].Exams) != 2 || modules[0].Exams[0].ID != "102" || modules[0].Exams[1].ID != "101" {
		t.Fatalf("latest grouped uploads: %+v", modules)
	}
	for _, tc := range []struct {
		filter Filter
		count  int
	}{
		{Filter{Query: " SQL "}, 2},
		{Filter{Programs: []string{"Informatik"}, Degrees: []string{"Master"}}, 1},
		{Filter{Degrees: []string{"Bachelor", "Master"}, Versions: []string{"2017", "2023"}}, 3},
		{Filter{Degrees: []string{"Master"}, Versions: []string{"2017"}}, 0},
	} {
		if got := len(catalog.FindExams(tc.filter)); got != tc.count {
			t.Fatalf("filter %+v: got %d want %d", tc.filter, got, tc.count)
		}
	}
	if module, ok := catalog.FindModule("12", Filter{}); !ok || len(module.Exams) != 0 {
		t.Fatal("empty existing module lost its URL")
	}
	if _, err := db.Exec(`UPDATE programs SET deleted_at='2026-01-01T00:00:00Z' WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	catalog, err = s.Catalog(context.Background())
	if err != nil || len(catalog.FindExams(Filter{})) != 1 {
		t.Fatal("deleted programme exposed archive entries")
	}
}

func TestPersistentDownloadsRetainHistoricalKeysAndCheckMetadata(t *testing.T) {
	s, db := archiveService(t)
	ctx := context.Background()
	object, err := s.Open(ctx, 1, "100")
	if err != nil {
		t.Fatal(err)
	}
	value, err := io.ReadAll(object)
	object.Close()
	if err != nil || string(value) != "%PDF-" {
		t.Fatal("historical revision or V1 object key lost")
	}
	if _, err := db.Exec(`UPDATE archive SET nbytes=100 WHERE id=101`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Open(ctx, 1, "101"); !errors.Is(err, ErrStorage) {
		t.Fatal("mismatching storage metadata accepted")
	}
	if _, err := db.Exec(`UPDATE archive SET deleted_at='2026-01-01T00:00:00Z' WHERE id=100`); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"100", "999", "../100", "0"} {
		if _, err := s.Open(ctx, 1, value); !errors.Is(err, ErrNotFound) {
			t.Fatalf("download %q: %v", value, err)
		}
	}
}
