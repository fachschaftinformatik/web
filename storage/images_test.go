package storage

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	"golang.org/x/image/bmp"
)

func TestGeneratePreviews(t *testing.T) {
	for _, tc := range []struct {
		name          string
		width, height int
	}{
		{"landscape", 1800, 900},
		{"portrait", 900, 1800},
		{"square", 1800, 1800},
		{"small", 80, 40},
		{"narrow", 1, 1800},
	} {
		t.Run(tc.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, tc.width, tc.height))
			var src bytes.Buffer
			if err := png.Encode(&src, img); err != nil {
				t.Fatal(err)
			}
			previews, err := GeneratePreviews(&src)
			if err != nil {
				t.Fatal(err)
			}
			if len(previews) != 6 {
				t.Fatalf("got %d previews, want 6", len(previews))
			}
			for _, size := range []int{200, 400, 600, 800, 1200, 1600} {
				cfg, err := jpeg.DecodeConfig(bytes.NewReader(previews[size]))
				if err != nil {
					t.Fatalf("preview %d: %v", size, err)
				}
				if cfg.Width < 1 || cfg.Height < 1 || cfg.Width > min(size, tc.width) || cfg.Height > min(size, tc.height) {
					t.Errorf("preview %d has invalid dimensions %dx%d", size, cfg.Width, cfg.Height)
				}
				if max(cfg.Width, cfg.Height) != min(size, max(tc.width, tc.height)) {
					t.Errorf("preview %d does not fill its longest side: %dx%d", size, cfg.Width, cfg.Height)
				}
				// Allow one pixel of rounding when preserving the aspect ratio.
				if delta := cfg.Width*tc.height - cfg.Height*tc.width; delta > max(tc.width, tc.height) || delta < -max(tc.width, tc.height) {
					t.Errorf("preview %d changed aspect ratio: %dx%d", size, cfg.Width, cfg.Height)
				}
			}
		})
	}
}

func TestGeneratePreviewsFormats(t *testing.T) {
	img := image.NewPaletted(image.Rect(0, 0, 2, 1), color.Palette{color.Black, color.White})
	img.SetColorIndex(1, 0, 1)
	for _, tc := range []struct {
		name   string
		encode func(*bytes.Buffer) error
	}{
		{"png", func(b *bytes.Buffer) error { return png.Encode(b, img) }},
		{"gif", func(b *bytes.Buffer) error { return gif.Encode(b, img, nil) }},
		{"jpeg", func(b *bytes.Buffer) error { return jpeg.Encode(b, img, nil) }},
		{"bmp", func(b *bytes.Buffer) error { return bmp.Encode(b, img) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var src bytes.Buffer
			if err := tc.encode(&src); err != nil {
				t.Fatal(err)
			}
			previews, err := GeneratePreviews(&src)
			if err != nil {
				t.Fatal(err)
			}
			preview, err := jpeg.Decode(bytes.NewReader(previews[200]))
			if err != nil {
				t.Fatal(err)
			}
			dark, _, _, _ := preview.At(0, 0).RGBA()
			light, _, _, _ := preview.At(1, 0).RGBA()
			if dark >= light {
				t.Error("preview did not preserve pixel ordering")
			}
		})
	}
}

func TestGeneratePreviewsRejectsUnsupportedInput(t *testing.T) {
	for _, src := range []string{
		"not an image",
		"II\x2a\x00\x08\x00\x00\x00", // Little-endian TIFF header.
		"MM\x00\x2a\x00\x00\x00\x08", // Big-endian TIFF header.
	} {
		previews, err := GeneratePreviews(strings.NewReader(src))
		if !errors.Is(err, image.ErrFormat) || previews != nil {
			t.Errorf("got previews %v, error %v; want unsupported format", previews, err)
		}
	}
}
