package controller

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fachschaftinformatik/web/internal/storage/disk"
)

func TestArchivePaginationKeepsFiltersAndGroupsModules(t *testing.T) {
	storage, err := disk.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	b, db := identityBrowserWithStorage(t, storage)
	if _, err := db.Exec(`INSERT INTO programs(id,name) VALUES(1,'Informatik (B. Sc.)'); INSERT INTO program_versions(program_id,name) VALUES(1,'PO2023'); INSERT INTO users(id,email,name,password) VALUES(1,'pagination@fsv-wh.de','Person','unused')`); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 45; i++ {
		if _, err := db.Exec(`INSERT INTO modules(id,program_id,name) VALUES(?,1,?)`, i, fmt.Sprintf("Modul %02d", i)); err != nil {
			t.Fatal(err)
		}
		for j := 0; j < 2; j++ {
			id := i*10 + j
			if _, err := db.Exec(`INSERT INTO archive(id,group_id,user_id,module_id,version,exam_date,access_key,mime_type,nbytes,checksum) VALUES(?,?,1,?,'PO2023','2026-01-01',?,'application/pdf',5,?)`, id, id, i, fmt.Sprint(id), fmt.Sprint(id)); err != nil {
				t.Fatal(err)
			}
		}
	}
	query := "?applied=1&degree=Bachelor&program=Informatik&version=PO2023&q=Modul"
	for _, tc := range []struct {
		page, rows      int
		present, absent string
	}{{1, 20, "Modul 45", "Modul 25"}, {2, 20, "Modul 25", "Modul 45"}, {3, 5, "Modul 01", "Modul 25"}} {
		status, body, _ := b.get("/archive" + query + fmt.Sprintf("&seite=%d", tc.page))
		if status != 200 || strings.Count(body, `scope="row"`) != tc.rows || !strings.Contains(body, tc.present) || strings.Contains(body, tc.absent) || !strings.Contains(body, "45 Module gefunden") || !strings.Contains(body, "Alle 2 Rekos ansehen") {
			t.Fatalf("page %d: %d %s", tc.page, status, body)
		}
		if !strings.Contains(body, fmt.Sprintf("Altklausuren (Seite %d von 3)", tc.page)) || !strings.Contains(body, `aria-current="page"`) {
			t.Fatal("missing page title or current state")
		}
		if !strings.Contains(body, `degree=Bachelor&amp;program=Informatik&amp;q=Modul`) {
			t.Fatal("filter context lost")
		}
		if tc.page == 1 && strings.Contains(body, `rel="prev"`) || tc.page == 3 && strings.Contains(body, `rel="next"`) {
			t.Fatal("invalid boundary navigation")
		}
	}
	r, err := http.NewRequest("GET", b.server.URL+"/archive"+query+"&seite=2", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("HX-Request", "true")
	response, err := b.client.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	fragment, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(fragment), "<!doctype") || !strings.Contains(string(fragment), "Modul 25") || !strings.Contains(string(fragment), "data-document-title") {
		t.Fatal("fragment lost pagination or title")
	}
	for _, page := range []string{"0", "-1", "999", "invalid"} {
		status, _, headers := b.get("/archive" + query + "&seite=" + page)
		if status != 303 || strings.Contains(headers.Get("Location"), "seite=") || !strings.Contains(headers.Get("Location"), "program=Informatik") {
			t.Fatalf("invalid page %s: %d %s", page, status, headers.Get("Location"))
		}
	}
	status, body, _ := b.get("/archive?applied=1&q=Modul+01")
	if status != 200 || !strings.Contains(body, "1 Modul gefunden") || strings.Contains(body, `class="govuk-pagination"`) {
		t.Fatal("single-result search has pagination")
	}
}
