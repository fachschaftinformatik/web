// Package home contains a public-content snapshot for the landing-page demo.
package home

type Post struct {
	Title, Summary, URL, Date, DateISO string
}

type Content struct {
	News  Post
	Forum []Post
}

// Snapshot was retrieved from the public discussions API on 16 October 2026.
func Snapshot() Content {
	return Content{
		News: Post{Title: "Show And Tell 2026", Summary: "Projektvorstellungen, Projektstände, Voting und Vorträge aus dem Fachbereich Informatik.", URL: "/d/140128508453060609", Date: "22. Januar 2026", DateISO: "2026-01-22"},
		Forum: []Post{
			{Title: "LaTeX Template für Bachelor-/Masterarbeit", Summary: "Eine Vorlage für die Abschlussarbeit und Tipps zum Einstieg in LaTeX.", URL: "/d/203823111852785665", Date: "17. Juli 2026", DateISO: "2026-07-17"},
			{Title: "BA Kolloquium, was wird abgefragt?", Summary: "Austausch zur Vorbereitung auf das Bachelor-Kolloquium.", URL: "/d/147381306706100225", Date: "11. Februar 2026", DateISO: "2026-02-11"},
		},
	}
}
