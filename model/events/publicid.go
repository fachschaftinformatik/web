package events

import (
	"context"
	"database/sql"
	"errors"
	"github.com/fachschaftinformatik/web/internal/publicid"
	"strconv"
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

func (s *Service) references(ctx context.Context, e *Event) error {
	var err error
	e.ID, err = publicid.Lookup(ctx, s.db, "events", e.InternalID)
	if err != nil {
		return err
	}
	for _, link := range []struct {
		id        sql.NullInt64
		namespace string
		out       *string
	}{
		{e.LegacyPostID, "posts", &e.LegacyPostPublicID},
		{e.AlbumID, "albums", &e.AlbumPublicID},
		{e.AlbumCoverID, "media", &e.AlbumCoverPublicID},
	} {
		if link.id.Valid {
			*link.out, err = publicid.Lookup(ctx, s.db, link.namespace, strconv.FormatInt(link.id.Int64, 10))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
