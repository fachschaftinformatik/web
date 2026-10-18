package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/fachschaftinformatik/web/internal/publicid"
)

const eligibleAlbum = `e.deleted_at IS NULL AND (SELECT COUNT(*) FROM (SELECT 1 FROM media photos WHERE photos.event_id=e.id AND photos.deleted_at IS NULL AND photos.mime_type IN ` + imageTypes + ` LIMIT 3))=3`
const albumSummary = `SELECT e.id,e.title,COALESCE(e.cover_path,''),(SELECT COUNT(*) FROM media m WHERE m.event_id=e.id AND m.deleted_at IS NULL AND m.mime_type IN ` + imageTypes + `),(SELECT CAST(m.id AS TEXT) FROM media m WHERE m.event_id=e.id AND m.deleted_at IS NULL AND m.mime_type IN ` + imageTypes + ` ORDER BY m.created_at DESC,m.id DESC LIMIT 1) FROM events e WHERE ` + eligibleAlbum

func albumPhoto(id, title, description string, number int, albumTitle string) Image {
	if strings.TrimSpace(title) == "" {
		title = fmt.Sprintf("%s – Bild %d", albumTitle, number)
	}
	// Instagram's publisher/date strings are provenance, not image alternatives.
	if strings.TrimSpace(description) == "" || strings.HasPrefix(description, "Photo by ") || strings.HasPrefix(description, "Video by ") {
		description = title
	}
	return Image{ID: id, Title: title, Description: description}
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
	a.InternalID = a.ID
	if strings.TrimSpace(a.Title) == "" {
		a.Title = "Fotoalbum"
	}
	if _, ok := sourceToken(cover, "events/", "/cover/source"); ok {
		a.Cover = Image{ID: a.ID, Title: a.Title, Description: a.Title, AlbumCover: true}
	} else {
		a.Cover = albumPhoto(first, a.Title, a.Title, 1, a.Title)
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
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, 0, err
	}
	for i := range albums {
		if err := s.albumReferences(ctx, &albums[i]); err != nil {
			return nil, 0, err
		}
	}
	return albums, total, nil
}

func (s *Service) Album(ctx context.Context, id string) (Album, error) {
	internal, err := resolve(ctx, s.db, "albums", id)
	if err != nil {
		return Album{}, err
	}
	a, err := scanAlbum(s.db.QueryRowContext(ctx, albumSummary+` AND e.id=?`, internal))
	if err == nil {
		err = s.albumReferences(ctx, &a)
	}
	return a, err
}

// ItemAlbum resolves an item's owning public album from its stable ID. Album
// visibility follows the same photo-count policy as the album listing.
func (s *Service) ItemAlbum(ctx context.Context, value string) (string, error) {
	id, err := resolve(ctx, s.db, "media", value)
	if err != nil {
		return "", err
	}
	var album string
	err = s.db.QueryRowContext(ctx, `SELECT CAST(e.id AS TEXT) FROM media m JOIN events e ON e.id=m.event_id
		WHERE m.id=? AND m.deleted_at IS NULL AND m.mime_type IN `+imageTypes+` AND `+eligibleAlbum, id).Scan(&album)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return publicid.Lookup(ctx, s.db, "albums", album)
}

func (s *Service) Photos(ctx context.Context, a Album, offset, limit int) ([]Image, error) {
	internal, err := resolve(ctx, s.db, "albums", a.ID)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT CAST(id AS TEXT),COALESCE(title,''),COALESCE(description,'') FROM media WHERE event_id=? AND deleted_at IS NULL AND mime_type IN `+imageTypes+` ORDER BY created_at DESC,id DESC LIMIT ? OFFSET ?`, internal, min(100, max(1, limit)), max(0, offset))
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
		images = append(images, albumPhoto(id, title, description, offset+len(images)+1, a.Title))
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	for i := range images {
		if err := s.imageReference(ctx, &images[i]); err != nil {
			return nil, err
		}
	}
	return images, nil
}

// PhotoPosition resolves membership and ordinal without loading all photographs.
func (s *Service) PhotoPosition(ctx context.Context, album, photo string) (int, error) {
	var resolveErr error
	album, resolveErr = resolve(ctx, s.db, "albums", album)
	if resolveErr != nil {
		return 0, resolveErr
	}
	photo, resolveErr = resolve(ctx, s.db, "media", photo)
	if resolveErr != nil {
		return 0, resolveErr
	}
	var position int
	err := s.db.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM media preceding WHERE preceding.event_id=m.event_id AND preceding.deleted_at IS NULL AND preceding.mime_type IN `+imageTypes+` AND (preceding.created_at>m.created_at OR (preceding.created_at=m.created_at AND preceding.id>m.id))) FROM media m WHERE m.id=? AND m.event_id=? AND m.deleted_at IS NULL AND m.mime_type IN `+imageTypes, photo, album).Scan(&position)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return position, err
}
