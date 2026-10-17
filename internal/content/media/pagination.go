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

const eligibleAlbum = `e.deleted_at IS NULL AND (SELECT COUNT(*) FROM (SELECT 1 FROM media photos WHERE photos.event_id=e.id AND photos.deleted_at IS NULL AND photos.mime_type IN ` + galleryTypes + ` LIMIT 3))=3`
const albumSummary = `SELECT e.id,e.title,COALESCE(e.cover_path,''),(SELECT COUNT(*) FROM media m WHERE m.event_id=e.id AND m.deleted_at IS NULL AND m.mime_type IN ` + galleryTypes + `),(SELECT CAST(m.id AS TEXT) FROM media m WHERE m.event_id=e.id AND m.deleted_at IS NULL AND m.mime_type IN ` + galleryTypes + ` ORDER BY m.created_at DESC,m.id DESC LIMIT 1) FROM events e WHERE ` + eligibleAlbum

func galleryPhoto(id, title, description string, number int, albumTitle string) Image {
	if strings.TrimSpace(title) == "" {
		title = fmt.Sprintf("%s – Bild %d", albumTitle, number)
	}
	if strings.TrimSpace(description) == "" {
		description = title
	}
	return Image{ID: id, Title: title, Description: description, Source: "/media/" + id + "/1200", thumbnail: "/media/" + id + "/400"}
}
func scanAlbum(row interface{ Scan(...any) error }) (Album, error) {
	var a Album
	var id int64
	var cover, first string
	err := row.Scan(&id, &a.Title, &cover, &a.OriginalCount, &first)
	if errors.Is(err, sql.ErrNoRows) {
		return a, ErrNotFound
	}
	if err != nil {
		return a, err
	}
	a.ID = strconv.FormatInt(id, 10)
	if strings.TrimSpace(a.Title) == "" {
		a.Title = "Fotoalbum"
	}
	if cover == blob.EventCoverSourceKey(a.ID) {
		prefix := "/media/events/" + a.ID + "/cover/"
		a.Cover = Image{ID: a.ID, Title: a.Title, Description: a.Title, Source: prefix + "1200", thumbnail: prefix + "400"}
	} else {
		a.Cover = galleryPhoto(first, a.Title, a.Title, 1, a.Title)
	}
	return a, nil
}

// AlbumPage loads only summaries and covers, never every album's photographs.
func (s *Service) AlbumPage(ctx context.Context, page, size int) ([]Album, int, error) {
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM events e WHERE `+eligibleAlbum).Scan(&total); err != nil {
		return nil, 0, err
	}
	size = min(100, max(1, size))
	page = max(1, min(page, max(1, (total+size-1)/size)))
	rows, err := s.db.QueryContext(ctx, albumSummary+` ORDER BY e.created_at DESC,e.id DESC LIMIT ? OFFSET ?`, size, (page-1)*size)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var albums []Album
	for rows.Next() {
		a, err := scanAlbum(rows)
		if err != nil {
			return nil, 0, err
		}
		albums = append(albums, a)
	}
	return albums, total, rows.Err()
}

func (s *Service) Album(ctx context.Context, id string) (Album, error) {
	return scanAlbum(s.db.QueryRowContext(ctx, albumSummary+` AND e.id=?`, id))
}

func (s *Service) Photos(ctx context.Context, a Album, offset, limit int) ([]Image, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT CAST(id AS TEXT),COALESCE(title,''),COALESCE(description,'') FROM media WHERE event_id=? AND deleted_at IS NULL AND mime_type IN `+galleryTypes+` ORDER BY created_at DESC,id DESC LIMIT ? OFFSET ?`, a.ID, min(100, max(1, limit)), max(0, offset))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var images []Image
	for rows.Next() {
		var id, title, description string
		if err := rows.Scan(&id, &title, &description); err != nil {
			return nil, err
		}
		images = append(images, galleryPhoto(id, title, description, offset+len(images)+1, a.Title))
	}
	return images, rows.Err()
}

// PhotoPosition resolves membership and ordinal without loading all photographs.
func (s *Service) PhotoPosition(ctx context.Context, album, photo string) (int, error) {
	var position int
	err := s.db.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM media preceding WHERE preceding.event_id=m.event_id AND preceding.deleted_at IS NULL AND preceding.mime_type IN `+galleryTypes+` AND (preceding.created_at>m.created_at OR (preceding.created_at=m.created_at AND preceding.id>m.id))) FROM media m WHERE m.id=? AND m.event_id=? AND m.deleted_at IS NULL AND m.mime_type IN `+galleryTypes, photo, album).Scan(&position)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return position, err
}
