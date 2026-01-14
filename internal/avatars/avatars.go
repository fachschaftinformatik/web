package avatars

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/fachschaftinformatik/web/internal/buckets"
	"github.com/google/uuid"
)

type Service struct {
	buckets *buckets.Client
}

func NewService(buckets *buckets.Client) *Service {
	return &Service{
		buckets: buckets,
	}
}

func (s *Service) FetchDicebearSVG(ctx context.Context, userID string) ([]byte, error) {
	url := fmt.Sprintf("https://api.dicebear.com/9.x/identicon/svg?seed=%s&backgroundColor=ffffff", userID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch avatar from dicebear: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dicebear returned status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read avatar data: %w", err)
	}
	return data, nil
}

// StoreAvatar stores the provided data in the bucket at the specified objectName.
func (s *Service) StoreAvatar(ctx context.Context, objectName string, data []byte) error {
	reader := bytes.NewReader(data)
	return s.buckets.Upload(ctx, objectName, reader, int64(len(data)), "image/svg+xml")
}

// GenerateAndStoreAvatar fetches an identicon from DiceBear and stores it in the bucket.
// Returns the object name (path) in the bucket.
func (s *Service) GenerateAndStoreAvatar(ctx context.Context, userID string) (string, error) {
	data, err := s.FetchDicebearSVG(ctx, userID)
	if err != nil {
		return "", err
	}

	// 2. Prepare storage path: avatars/<userid>/<avatarid>.svg
	avatarID := uuid.New().String()
	objectName := fmt.Sprintf("avatars/%s/%s.svg", userID, avatarID)

	// 3. Upload to bucket
	err = s.StoreAvatar(ctx, objectName, data)
	if err != nil {
		return "", fmt.Errorf("failed to upload avatar to storage: %w", err)
	}

	return objectName, nil
}
