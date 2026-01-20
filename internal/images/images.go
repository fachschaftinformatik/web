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
)

// GeneratePreview resizes an image to fit within maxW x maxH while preserving aspect ratio.
// It returns the preview as a JPEG byte slice.
func GeneratePreview(src io.Reader, maxW, maxH int) ([]byte, error) {
	img, format, err := image.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize the image preserving aspect ratio
	// Fit into maxW x maxH
	thumb := imaging.Fit(img, maxW, maxH, imaging.Lanczos)

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 80})
	if err != nil {
		return nil, fmt.Errorf("failed to encode jpeg (%s): %w", format, err)
	}

	return buf.Bytes(), nil
}
