package storage

import (
	"fmt"
)

func AvatarSourceKey(userID string) string {
	return fmt.Sprintf("avatars/%s/source", userID)
}

func AvatarPreviewKey(userID string, size int) string {
	return fmt.Sprintf("avatars/%s/p/%d.jpg", userID, size)
}

// AlbumCoverSourceKey retains the original object layout for media albums.
func AlbumCoverSourceKey(albumID string) string {
	return fmt.Sprintf("events/%s/cover/source", albumID)
}

func AlbumCoverPreviewKey(albumID string, size int) string {
	return fmt.Sprintf("events/%s/cover/p/%d.jpg", albumID, size)
}

func MediaSourceKey(mediaID string) string {
	return fmt.Sprintf("media/%s/source", mediaID)
}

func MediaPreviewKey(mediaID string, size int) string {
	return fmt.Sprintf("media/%s/p/%d.jpg", mediaID, size)
}

// ExamSourceKey retains existing exam object keys across URL and model renames.
func ExamSourceKey(documentID string) string {
	return fmt.Sprintf("archive/%s/source.pdf", documentID)
}
