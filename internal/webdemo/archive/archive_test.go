package archive

import "testing"

func TestFindExams(t *testing.T) {
	for _, tc := range []struct {
		name   string
		filter Filter
		want   int
	}{
		{"all", Filter{}, 8},
		{"case and whitespace", Filter{Query: "  SQL  "}, 1},
		{"description", Filter{Query: "Normalisierung"}, 1},
		{"combined", Filter{Programs: []string{"Informatik"}, Versions: []string{"2023"}}, 3},
		{"multiple versions", Filter{Versions: []string{"2023", "2017"}}, 8},
		{"multiple versions with program", Filter{Programs: []string{"Informatik"}, Versions: []string{"2023", "2017"}}, 5},
		{"unknown version", Filter{Versions: []string{"unknown"}}, 0},
		{"bachelor", Filter{Degrees: []string{"Bachelor"}}, 8},
		{"master", Filter{Degrees: []string{"Master"}}, 0},
		{"both degrees", Filter{Degrees: []string{"Bachelor", "Master"}}, 8},
		{"degree combined with program and version", Filter{Degrees: []string{"Bachelor"}, Programs: []string{"Informatik"}, Versions: []string{"2023"}}, 3},
		{"unknown degree", Filter{Degrees: []string{"unknown"}}, 0},
		{"multiple programs", Filter{Programs: []string{"Informatik", "Medieninformatik"}}, 6},
		{"incompatible filters", Filter{Query: "SQL", Programs: []string{"Informatik"}}, 0},
		{"unknown program", Filter{Programs: []string{"unknown"}}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(FindExams(tc.filter)); got != tc.want {
				t.Fatalf("FindExams(%+v): got %d records, want %d", tc.filter, got, tc.want)
			}
		})
	}
}

func TestModuleUploads(t *testing.T) {
	modules := FindModules(Filter{})
	if len(modules) != 6 {
		t.Fatalf("want six modules despite multiple uploads, got %d", len(modules))
	}
	for _, tc := range []struct {
		filter Filter
		count  int
		newest string
	}{
		{Filter{}, 3, "12.02.2026"},
		{Filter{Versions: []string{"2023"}}, 2, "12.02.2026"},
		{Filter{Versions: []string{"2017"}}, 1, "22.01.2025"},
		{Filter{Degrees: []string{"Master"}}, 0, ""},
	} {
		module, ok := FindModule("algorithmen-und-datenstrukturen", tc.filter)
		if !ok || len(module.Exams) != tc.count {
			t.Fatalf("module/filter lookup: ok=%v, uploads=%d, want %d", ok, len(module.Exams), tc.count)
		}
		if tc.count > 0 && module.Exams[0].Date != tc.newest {
			t.Fatalf("wrong newest matching upload: %s", module.Exams[0].Date)
		}
	}
	if _, ok := FindModule("missing", Filter{}); ok {
		t.Fatal("unknown module found")
	}
}
