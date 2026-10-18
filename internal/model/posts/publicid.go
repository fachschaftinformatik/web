package posts

import (
	"context"
	"database/sql"
	"errors"
	"github.com/fachschaftinformatik/web/internal/publicid"
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
