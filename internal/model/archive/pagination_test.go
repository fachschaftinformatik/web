package archive

import (
	"context"
	"errors"
	"testing"
)

func TestSQLArchivePagesKeepGroupsFiltersAndLatestRevisions(t *testing.T) {
	s, _ := archiveService(t)
	ctx := context.Background()
	catalog, total, err := s.CatalogPage(ctx, Filter{}, 1, 1)
	modules := catalog.FindModules(Filter{})
	if err != nil || total != 2 || len(modules) != 1 || modules[0].ID != "10" || len(modules[0].Exams) != 2 {
		t.Fatalf("group page: %+v %d %v", modules, total, err)
	}
	catalog, total, err = s.CatalogPage(ctx, Filter{Programs: []string{"Informatik"}, Degrees: []string{"Master"}}, 1, 1)
	modules = catalog.FindModules(Filter{})
	if err != nil || total != 1 || len(modules) != 1 || modules[0].ID != "11" {
		t.Fatalf("filtered page: %+v %d %v", modules, total, err)
	}
	for page, want := range []string{"102", "101"} {
		module, total, err := s.ModulePage(ctx, "10", Filter{}, page+1, 1)
		if err != nil || total != 2 || len(module.Exams) != 1 || module.Exams[0].ID != want {
			t.Fatalf("module page: %+v %d %v", module, total, err)
		}
	}
	module, total, err := s.ModulePage(ctx, "12", Filter{}, 1, 20)
	if err != nil || module.ID != "12" || total != 0 || len(module.Exams) != 0 {
		t.Fatalf("empty module: %+v %d %v", module, total, err)
	}
	if _, _, err := s.ModulePage(ctx, "999", Filter{}, 1, 20); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing module: %v", err)
	}
}
