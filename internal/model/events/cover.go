package events

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"github.com/fachschaftinformatik/web/internal/datetime"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"strconv"
	"strings"
	"time"

	"github.com/fachschaftinformatik/web/internal/publicid"
	"github.com/fachschaftinformatik/web/internal/storage"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const MaxCoverBytes = 6 << 20

var ErrCoverStorage = errors.New("Das Titelbild konnte nicht gespeichert oder geladen werden.")

type CoverUpload struct {
	Data   []byte
	Alt    string
	Remove bool
}

func permission(ctx context.Context, db querier, user int64) error {
	var role string
	err := db.QueryRowContext(ctx, `SELECT role FROM users WHERE id=? AND active=1 AND verified=1 AND deleted_at IS NULL AND role IN ('admin','editor')`, user).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	}
	return err
}

func (s *Service) CanPublish(ctx context.Context, user int64) error {
	return permission(ctx, s.db, user)
}

func prepareCover(data []byte) ([]byte, int, int, error) {
	return prepareImage(data, 1600)
}

func prepareImage(data []byte, maxSide int) ([]byte, int, int, error) {
	invalid := func() ([]byte, int, int, error) {
		return nil, 0, 0, &ValidationError{Fields: []FieldError{{Field: "cover", Message: "Wähle ein JPEG-, PNG- oder WebP-Bild mit höchstens 6 MB und 24 Megapixeln."}}}
	}
	if len(data) == 0 || len(data) > MaxCoverBytes {
		return invalid()
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png" && format != "webp") || config.Width <= 0 || config.Height <= 0 || config.Width > 12000 || config.Height > 12000 || int64(config.Width)*int64(config.Height) > 24000000 {
		return invalid()
	}
	source, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return invalid()
	}
	w, h := config.Width, config.Height
	if max(w, h) > maxSide {
		factor := float64(maxSide) / float64(max(w, h))
		w = max(1, int(float64(w)*factor))
		h = max(1, int(float64(h)*factor))
	}
	target := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(target, target.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	xdraw.CatmullRom.Scale(target, target.Bounds(), source, source.Bounds(), draw.Over, nil)
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, target, &jpeg.Options{Quality: 88}); err != nil {
		return nil, 0, 0, err
	}
	return encoded.Bytes(), w, h, nil
}

func (s *Service) OpenCover(ctx context.Context, post string) (*storage.Object, error) {
	post, resolveErr := resolve(ctx, s.db, "events", post)
	if resolveErr != nil {
		return nil, resolveErr
	}
	var key string
	err := s.db.QueryRowContext(ctx, `SELECT cover_key FROM calendar_events WHERE id=? AND deleted_at IS NULL AND cover_key!=''`, post).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !validCoverKey(key) {
		return nil, ErrNotFound
	}
	if s.storage == nil {
		return nil, ErrCoverStorage
	}
	object, err := s.storage.GetObject(ctx, key)
	if err != nil {
		return nil, ErrCoverStorage
	}
	object.Info.ContentType = "image/jpeg"
	return object, nil
}

func validCoverKey(key string) bool {
	value := strings.TrimSuffix(strings.TrimPrefix(key, "calendar/covers/"), ".jpg")
	if publicid.Valid(value) && key == "calendar/covers/"+value+".jpg" {
		return true
	}
	id, err := strconv.ParseInt(value, 10, 64)
	return err == nil && id > 0 && key == "calendar/covers/"+strconv.FormatInt(id, 10)+".jpg"
}
func parseDate(value string) (string, error) {
	d, err := time.Parse("2.1.2006", strings.TrimSpace(value))
	return d.Format("2006-01-02"), err
}

func schedule(d Draft) (string, error) {
	if strings.TrimSpace(d.Date) == "" {
		return "", nil
	}
	date, err := parseDate(d.Date)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(d.Time) == "" {
		return date, nil
	}
	stamp, err := datetime.LocalTimestamp(date + "T" + strings.TrimSpace(d.Time))
	if err != nil {
		return "", &ValidationError{Fields: []FieldError{{Field: "time", Message: "Diese Uhrzeit ist wegen der Zeitumstellung nicht eindeutig oder existiert nicht. Wähle eine andere Uhrzeit."}}}
	}
	return stamp, nil
}
