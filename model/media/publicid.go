package media

import (
	"context"
	"database/sql"
	"errors"
	"github.com/fachschaftinformatik/web/internal/publicid"
	"strconv"
	"strings"
)

type registryDB interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func resolve(ctx context.Context, db registryDB, namespace, ref string) (string, error) {
	if !publicid.Valid(ref) {
		return "", ErrNotFound
	}
	internal, err := publicid.Resolve(ctx, db, namespace, ref)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return internal, err
}

// Storage tokens are independent of public references. Both historical numeric
// keys and newly generated opaque keys retain their original object locations.
func sourceToken(key, prefix, suffix string) (string, bool) {
	token := strings.TrimSuffix(strings.TrimPrefix(key, prefix), suffix)
	if key != prefix+token+suffix {
		return "", false
	}
	if publicid.Valid(token) {
		return token, true
	}
	n, err := strconv.ParseInt(token, 10, 64)
	return token, err == nil && n > 0 && strconv.FormatInt(n, 10) == token
}

func (s *Service) imageReference(ctx context.Context, image *Image) error {
	image.InternalID = image.ID
	ref, err := publicid.Lookup(ctx, s.db, "media", image.InternalID)
	image.ID = ref
	return err
}

func (s *Service) albumReferences(ctx context.Context, album *Album) error {
	var err error
	album.ID, err = publicid.Lookup(ctx, s.db, "albums", album.InternalID)
	if err != nil {
		return err
	}
	if album.Cover.AlbumCover {
		album.Cover.InternalID, album.Cover.ID = album.InternalID, album.ID
	} else if album.Cover.ID != "" {
		if err := s.imageReference(ctx, &album.Cover); err != nil {
			return err
		}
	}
	for i := range album.Images {
		if err := s.imageReference(ctx, &album.Images[i]); err != nil {
			return err
		}
	}
	return nil
}
