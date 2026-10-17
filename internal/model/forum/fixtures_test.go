package forum

import (
	"embed"
	"encoding/json"
	"slices"
	"strings"
)

var _ embed.FS

//go:embed testdata/snapshot.json
var snapshot []byte

//go:embed testdata/threads.json
var threads []byte

func All() []Post {
	var posts []Post
	if err := json.Unmarshal(snapshot, &posts); err != nil {
		panic("invalid embedded forum snapshot: " + err.Error())
	}
	slices.SortStableFunc(posts, func(a, b Post) int {
		return strings.Compare(b.Created, a.Created)
	})
	return posts
}

func Find(filter Filter) []Post {
	var result []Post
	terms := strings.Fields(strings.ToLower(filter.Query))
	for _, post := range All() {
		if filter.Type != "" && post.Type != filter.Type {
			continue
		}
		text := strings.ToLower(strings.Join([]string{post.Title, post.Body, post.Summary, strings.Join(post.Tags, " ")}, " "))
		if slices.ContainsFunc(terms, func(term string) bool { return !strings.Contains(text, term) }) {
			continue
		}
		result = append(result, post)
	}
	return result
}

func Lookup(id string) (Thread, bool) {
	var records []Thread
	if err := json.Unmarshal(threads, &records); err != nil {
		panic("invalid embedded forum threads: " + err.Error())
	}
	for _, post := range All() {
		if post.ID != id {
			continue
		}
		for _, thread := range records {
			if thread.ID == id {
				thread.Post = post
				slices.SortStableFunc(thread.Replies, func(a, b Comment) int { return strings.Compare(b.Created, a.Created) })
				return thread, true
			}
		}
	}
	return Thread{}, false
}
