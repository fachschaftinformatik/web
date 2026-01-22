package avatars

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/fachschaftinformatik/web/internal/storage"
)

type Service struct {
	storage storage.Provider
}

func NewService(storage storage.Provider) *Service {
	return &Service{
		storage: storage,
	}
}

// GenerateSVG creates a symmetrical 5x5 identicon SVG based on the userID (Snowflake).
// It adheres strictly to the DiceBear identicon style using row-based patterns.
func (s *Service) GenerateSVG(userID string) []byte {
	// Use SHA-256 of the userID string to get a stable seed
	hash := sha256.Sum256([]byte(userID))

	// Determine theme color from first 2 bytes
	hue := float64(binary.BigEndian.Uint16(hash[0:2]) % 360)
	// Material-like vibrant colors: Saturation 60-80%, Lightness 45-55%
	mainColor := hslToRGB(hue, 0.70, 0.50)
	hexColor := fmt.Sprintf("#%02x%02x%02x", mainColor.R, mainColor.G, mainColor.B)

	// DiceBear Identicon Patterns (Symmetrical 5-cell rows)
	// 0: [0,0,0,0,0], 1: [1,1,1,1,1], 2: [0,1,1,1,0], 3: [1,0,1,0,1],
	// 4: [0,1,0,1,0], 5: [1,1,0,1,1], 6: [0,0,1,0,0], 7: [1,0,0,0,1]
	patterns := [][]bool{
		{false, false, false, false, false},
		{true, true, true, true, true},
		{false, true, true, true, false},
		{true, false, true, false, true},
		{false, true, false, true, false},
		{true, true, false, true, true},
		{false, false, true, false, false},
		{true, false, false, false, true},
	}

	var sb strings.Builder
	sb.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 5 5" fill="none" shape-rendering="crispEdges" width="512" height="512">`)
	sb.WriteString(`<rect width="5" height="5" fill="#ffffff"/>`)

	// Start from byte 2 to avoid using the same bits as the hue
	// We pick one of 8 patterns for each of the 5 rows.
	// 5 rows * 3 bits = 15 bits.
	for row := 0; row < 5; row++ {
		// Extract 3 bits for pattern selection (8 patterns)
		byteIdx := 2 + (row * 3 / 8)
		bitIdx := uint(row * 3 % 8)

		var val uint8
		if bitIdx <= 5 {
			val = (hash[byteIdx] >> bitIdx) & 0x07
		} else {
			val = (hash[byteIdx] >> bitIdx) | (hash[byteIdx+1] << (8 - bitIdx))
			val &= 0x07
		}

		pattern := patterns[val]
		for col, filled := range pattern {
			if filled {
				sb.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="1" height="1" fill="%s"/>`, col, row, hexColor))
			}
		}
	}

	sb.WriteString(`</svg>`)
	return []byte(sb.String())
}

func (s *Service) StoreAvatar(ctx context.Context, objectName string, data []byte) error {
	reader := bytes.NewReader(data)
	return s.storage.Upload(ctx, objectName, reader, int64(len(data)), "image/svg+xml")
}

func (s *Service) GenerateAndStoreAvatar(ctx context.Context, userID string) (string, error) {
	data := s.GenerateSVG(userID)

	objectName := storage.AvatarSourceKey(userID)

	err := s.StoreAvatar(ctx, objectName, data)
	if err != nil {
		return "", fmt.Errorf("failed to upload avatar to storage: %w", err)
	}

	return objectName, nil
}

func hslToRGB(h, s, l float64) color.RGBA {
	h /= 360
	var r, g, b float64

	if s == 0 {
		r, g, b = l, l, l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q
		r = hueToRGB(p, q, h+1.0/3.0)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-1.0/3.0)
	}

	return color.RGBA{
		R: uint8(math.Round(r * 255)),
		G: uint8(math.Round(g * 255)),
		B: uint8(math.Round(b * 255)),
		A: 255,
	}
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}
