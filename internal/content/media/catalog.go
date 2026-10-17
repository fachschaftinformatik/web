package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/fachschaftinformatik/web/internal/data/blob"
)

var ErrNotFound = errors.New("Das Bild wurde nicht gefunden.")
var ErrStorage = errors.New("Der Bildspeicher ist nicht erreichbar.")

// Only passive raster formats are public gallery images. In particular this
// excludes uploaded SVG/HTML and never exposes avatars or arbitrary bucket keys.
const galleryTypes = `('image/jpeg','image/png','image/gif','image/webp','image/bmp')`

// Catalog uses one snapshot for albums and their complete image lists. The
// historical newest-first order is made deterministic with the stable IDs.
func (s *Service) Catalog(ctx context.Context) ([]Album, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT e.id,e.title,COALESCE(e.cover_path,''),m.id,COALESCE(m.title,''),COALESCE(m.description,'')
FROM events e LEFT JOIN media m ON m.event_id=e.id AND m.deleted_at IS NULL AND m.mime_type IN `+galleryTypes+`
WHERE e.deleted_at IS NULL AND (SELECT COUNT(*) FROM (SELECT 1 FROM media photos WHERE photos.event_id=e.id AND photos.deleted_at IS NULL AND photos.mime_type IN `+galleryTypes+` LIMIT 3))=3
ORDER BY e.created_at DESC,e.id DESC,m.created_at DESC,m.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	albums := []Album{}
	for rows.Next() {
		var eventID int64
		var title, coverPath, imageTitle, description string
		var imageID sql.NullInt64
		if err := rows.Scan(&eventID, &title, &coverPath, &imageID, &imageTitle, &description); err != nil {
			return nil, err
		}
		id := strconv.FormatInt(eventID, 10)
		if len(albums) == 0 || albums[len(albums)-1].ID != id {
			if strings.TrimSpace(title) == "" {
				title = "Fotoalbum"
			}
			album := Album{ID: id, Title: title}
			if coverPath == blob.EventCoverSourceKey(id) {
				prefix := "/media/events/" + id + "/cover/"
				album.Cover = Image{ID: id, Title: title, Description: title, Source: prefix + "1200", thumbnail: prefix + "400"}
			}
			albums = append(albums, album)
		}
		if !imageID.Valid {
			continue
		}
		album := &albums[len(albums)-1]
		imageIDString := strconv.FormatInt(imageID.Int64, 10)
		album.Images = append(album.Images, galleryPhoto(imageIDString, imageTitle, description, len(album.Images)+1, album.Title))
		album.OriginalCount = len(album.Images)
	}
	return albums, rows.Err()
}

// Covers are resolved through their public event, never through a caller-supplied
// storage key. Only raster previews are exposed, including for legacy covers.
func (s *Service) OpenCover(ctx context.Context, value, size string) (*blob.Object, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 || strconv.FormatInt(id, 10) != value {
		return nil, ErrNotFound
	}
	switch size {
	case "200", "400", "600", "800", "1200", "1600":
	default:
		return nil, ErrNotFound
	}
	var path string
	err = s.db.QueryRowContext(ctx, "SELECT COALESCE(cover_path,'') FROM events WHERE id=? AND deleted_at IS NULL", id).Scan(&path)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && path != blob.EventCoverSourceKey(value)) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if s.storage == nil {
		return nil, ErrStorage
	}
	pixels, _ := strconv.Atoi(size)
	object, err := s.storage.GetObject(ctx, blob.EventCoverPreviewKey(value, pixels))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStorage, err)
	}
	if object.Info.Size <= 0 {
		object.Close()
		return nil, ErrStorage
	}
	object.Info.ContentType = "image/jpeg"
	return object, nil
}

// Open resolves a database image before constructing its V1 object key. The
// source size is checked against the catalog, and preview MIME is always JPEG.
func (s *Service) Open(ctx context.Context, value, size string) (*blob.Object, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 || strconv.FormatInt(id, 10) != value {
		return nil, ErrNotFound
	}
	key := blob.MediaSourceKey(value)
	if size != "original" {
		switch size {
		case "200", "400", "600", "800", "1200", "1600":
			pixels, _ := strconv.Atoi(size)
			key = blob.MediaPreviewKey(value, pixels)
		default:
			return nil, ErrNotFound
		}
	}
	var mimeType string
	var nbytes int64
	err = s.db.QueryRowContext(ctx, `SELECT m.mime_type,m.nbytes FROM media m JOIN events e ON e.id=m.event_id
WHERE m.id=? AND m.deleted_at IS NULL AND e.deleted_at IS NULL AND m.mime_type IN `+galleryTypes, id).Scan(&mimeType, &nbytes)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if s.storage == nil {
		return nil, ErrStorage
	}
	object, err := s.storage.GetObject(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStorage, err)
	}
	if object.Info.Size <= 0 || (size == "original" && object.Info.Size != nbytes) {
		object.Close()
		return nil, fmt.Errorf("%w: invalid stored size for media %s", ErrStorage, value)
	}
	// Disk objects named 'source' do not carry a MIME extension. Use the
	// allowlisted database MIME rather than trusting bucket metadata.
	object.Info.ContentType = mimeType
	if size != "original" {
		object.Info.ContentType = "image/jpeg"
	}
	return object, nil
}
