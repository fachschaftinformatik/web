package controller

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/model/events"
	"github.com/fachschaftinformatik/web/internal/model/forum"
	"github.com/fachschaftinformatik/web/internal/model/meetings"
	"github.com/fachschaftinformatik/web/internal/view"
)

const uploadBodyLimit = meetings.MaxPDFBytes + 1<<20

func (c *Handler) home(ctx context.Context, viewer ...int64) (view.HomePage, error) {
	var upcoming []events.Event

	calendar, _, err := c.events.FindPage(ctx, time.Now(), 1, 3, true)
	if err != nil {
		return view.HomePage{}, err
	}
	upcoming, _ = events.Group(calendar, time.Now())
	upcoming = upcoming[:min(3, len(upcoming))]

	records, _, err := c.forum.FindPage(ctx, forum.Filter{Type: "discussion"}, 1, 2, viewer...)
	if err != nil {
		return view.HomePage{}, err
	}
	page := view.HomePage{Events: upcoming}
	news, _, err := c.forum.FindPage(ctx, forum.Filter{Type: "news"}, 1, 1, viewer...)
	if err != nil {
		return view.HomePage{}, err
	}
	records = append(records, news...)
	for _, post := range records {
		item := view.HomePost{Title: post.Title, Summary: post.Summary, URL: post.URL(), Date: post.DateLabel(), DateISO: post.DateISO(), Votes: post.Votes, UserVote: post.UserVote}
		if post.Type == "news" && page.News.URL == "" {
			page.News = item
		}
		if post.Type == "discussion" && len(page.Forum) < 2 {
			page.Forum = append(page.Forum, item)
		}
	}
	return page, nil
}

func (c *Handler) Legal(w http.ResponseWriter, r *http.Request) {
	kind := strings.TrimPrefix(r.URL.Path, "/")
	titles := map[string]string{"impressum": "Impressum", "datenschutz": "Datenschutz", "barrierefreiheit": "Barrierefreiheit"}
	if err := c.people.Refresh(r.Context()); err != nil {
		c.serverError(w, err)
		return
	}
	legal := view.LegalPage{Kind: kind}
	for _, role := range c.people.Responsibilities("board") {
		switch strings.ToLower(strings.TrimSpace(role.Title)) {
		case "vorsitz", "vorsitzende", "vorsitzender":
			for _, person := range c.people.People() {
				if slices.Contains(role.MemberIDs, person.ID) {
					legal.Chair = append(legal.Chair, person)
				}
			}
		}
	}
	c.render(w, "legal", false, view.Page{Title: titles[kind], Active: "legal", Legal: legal})
}

func (c *Handler) Sitemap(w http.ResponseWriter, r *http.Request) {
	albums, total, err := c.gallery.AlbumPage(r.Context(), requestedPage(r), 100)
	if err != nil {
		c.serverError(w, err)
		return
	}

	c.render(w, "sitemap", false, view.Page{Title: "Seitenübersicht", Active: "sitemap", Images: view.ImagesPage{Albums: albums, Pagination: listPagination(r, total, 100)}})
}
