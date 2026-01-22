package images

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

// GeneratePreviews creates 6 previews in varying resolutions (200, 400, 600, 800, 1200, 1600).
func GeneratePreviews(src io.Reader) (map[int][]byte, error) {
	img, _, err := image.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	sizes := []int{200, 400, 600, 800, 1200, 1600}
	previews := make(map[int][]byte)

	for _, size := range sizes {
		// Fit into size x size preserving aspect ratio
		thumb := imaging.Fit(img, size, size, imaging.Lanczos)
		var buf bytes.Buffer
		// Use JPEG for encoding
		err = jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 80})
		if err != nil {
			return nil, fmt.Errorf("failed to encode jpeg for size %d: %w", size, err)
		}
		previews[size] = buf.Bytes()
	}

	return previews, nil
}
