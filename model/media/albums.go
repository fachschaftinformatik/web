package media

const PageSize = 24

type Image struct {
	InternalID             string
	ID, Title, Description string
	Width, Height          int
	AlbumCover             bool
}

type Album struct {
	InternalID    string
	ID, Title     string
	OriginalCount int
	Images        []Image
	Cover         Image
}
