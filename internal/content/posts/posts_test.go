package posts

import (
	"testing"
)

func TestFindPosts(t *testing.T) {
	for _, tc := range []struct {
		name   string
		filter Filter
		count  int
		first  string
	}{
		{"newest first", Filter{}, 5, "203823111852785665"},
		{"case and multiple terms", Filter{Query: "  LATEX   Abschlussarbeit "}, 1, "203823111852785665"},
		{"body search", Filter{Query: "Neovim"}, 1, "203823111852785665"},
		{"tag search", Filter{Query: "Hilfe benötigt"}, 1, "147381306706100225"},
		{"discussions", Filter{Type: "discussion"}, 3, "203823111852785665"},
		{"news", Filter{Type: "news"}, 1, "140128508453060609"},
		{"events", Filter{Type: "event"}, 1, "140127153315381249"},
		{"combined filters", Filter{Query: "LaTeX", Type: "event"}, 0, ""},
		{"no results", Filter{Query: "unfindable-topic"}, 0, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			posts := Find(tc.filter)
			if len(posts) != tc.count || (len(posts) > 0 && posts[0].ID != tc.first) {
				t.Fatalf("unexpected posts: %+v", posts)
			}
		})
	}
}

func TestThreadSnapshots(t *testing.T) {
	for _, post := range All() {
		thread, ok := Lookup(post.ID)
		if !ok || thread.Title != post.Title || len(thread.Replies) != post.Comments {
			t.Fatalf("incomplete thread %s: %+v", post.ID, thread)
		}
		previous := ""
		for _, comment := range thread.Replies {
			if (previous != "" && comment.Created > previous) || comment.Author == "" || comment.Text == "" || comment.ID == "" {
				t.Fatalf("invalid comment: %+v", comment)
			}
			previous = comment.Created
		}
	}
	if _, ok := Lookup("unknown"); ok {
		t.Fatal("unknown thread found")
	}
}
