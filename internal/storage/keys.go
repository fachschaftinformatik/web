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

func EventCoverSourceKey(eventID string) string {
	return fmt.Sprintf("events/%s/cover/source", eventID)
}

func EventCoverPreviewKey(eventID string, size int) string {
	return fmt.Sprintf("events/%s/cover/p/%d.jpg", eventID, size)
}

func MediaSourceKey(mediaID string) string {
	return fmt.Sprintf("media/%s/source", mediaID)
}

func MediaPreviewKey(mediaID string, size int) string {
	return fmt.Sprintf("media/%s/p/%d.jpg", mediaID, size)
}

func ArchiveSourceKey(fileID string) string {
	return fmt.Sprintf("archive/%s/source.pdf", fileID)
}
