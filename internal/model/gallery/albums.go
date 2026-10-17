package gallery

const PageSize = 24

type Image struct {
	ID, Title, Description string
	Width, Height          int
	AlbumCover             bool
}

type Album struct {
	ID, Title     string
	OriginalCount int
	Images        []Image
	Cover         Image
}
