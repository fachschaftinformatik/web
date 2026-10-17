package web

import (
	"context"
	"github.com/fachschaftinformatik/web/internal/content/posts"
)

type HomePost struct {
	Title, Summary, URL, Date, DateISO string
}

type HomePage struct {
	News  HomePost
	Forum []HomePost
}

func (c *Handler) home(ctx context.Context) (HomePage, error) {
	if c.posts == nil {
		return homeContent(), nil
	}
	records, err := c.posts.Find(ctx, posts.Filter{})
	if err != nil {
		return HomePage{}, err
	}
	var page HomePage
	for _, post := range records {
		item := HomePost{Title: post.Title, Summary: post.Summary, URL: post.URL(), Date: post.DateLabel(), DateISO: post.DateISO()}
		if post.Type == "news" && page.News.URL == "" {
			page.News = item
		}
		if post.Type == "discussion" && len(page.Forum) < 2 {
			page.Forum = append(page.Forum, item)
		}
	}
	return page, nil
}

// homeContent was retrieved from the public discussions API on 16 October 2026.
func homeContent() HomePage {
	return HomePage{
		News: HomePost{Title: "Show And Tell 2026", Summary: "Projektvorstellungen, Projektstände, Voting und Vorträge aus dem Fachbereich Informatik.", URL: "/d/140128508453060609", Date: "22. Januar 2026", DateISO: "2026-01-22"},
		Forum: []HomePost{
			{Title: "LaTeX Template für Bachelor-/Masterarbeit", Summary: "Eine Vorlage für die Abschlussarbeit und Tipps zum Einstieg in LaTeX.", URL: "/d/203823111852785665", Date: "17. Juli 2026", DateISO: "2026-07-17"},
			{Title: "BA Kolloquium, was wird abgefragt?", Summary: "Austausch zur Vorbereitung auf das Bachelor-Kolloquium.", URL: "/d/147381306706100225", Date: "11. Februar 2026", DateISO: "2026-02-11"},
		},
	}
}
