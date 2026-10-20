package view

import "testing"

func TestBuildIdentity(t *testing.T) {
	const revision = "027766547f9f8698936c0c2c517dc87586faba24"
	for _, tc := range []struct {
		name, version, revision, label, short string
	}{
		{"development", "dev", "", "dev", ""},
		{"release", "v1.3.1", revision, "1.3.1", "0277665"},
		{"commit build", revision, revision, "0277665", ""},
		{"development commit", "dev", revision, "dev", "0277665"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := Page{BuildVersion: tc.version, BuildRevision: tc.revision}
			if p.BuildLabel() != tc.label || p.BuildShortRevision() != tc.short {
				t.Fatalf("got %q (%q)", p.BuildLabel(), p.BuildShortRevision())
			}
		})
	}
}
