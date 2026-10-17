package forum

type Post struct {
	AuthorID                               int64
	AuthorPublic                           bool
	ID, Title, Body, Author, Created, Type string
	BodyFormat                             string
	Comments                               int
	Votes, UserVote                        int
	Tags                                   []string
}

type Filter struct {
	Query, Type string
	AuthorID    int64
}

type Comment struct {
	AuthorPublic    bool
	Votes, UserVote int
	AuthorID        int64
	CanEdit         bool
	ID              string
	Author          string `json:"user_name"`
	Text            string
	Created         string `json:"created_at"`
	Updated         string `json:"updated_at"`
}

func (c Comment) Edited() bool { return c.Updated != "" && c.Updated != c.Created }

type Thread struct {
	Post
	Updated   string
	EventDate string `json:"event_date"`
	Location  string
	Replies   []Comment `json:"comments"`
}

func (t Thread) Edited() bool { return t.Updated != "" && t.Updated != t.Created }
