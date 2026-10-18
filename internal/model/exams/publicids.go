package exams

import (
	"context"
	"strconv"
	"strings"

	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/publicid"
)

func lookupID(ctx context.Context, db database.DBTX, namespace string, id int64) (string, error) {
	return publicid.Lookup(ctx, db, namespace, strconv.FormatInt(id, 10))
}

func resolveID(ctx context.Context, db database.DBTX, namespace, value string) (int64, error) {
	id, err := publicid.Resolve(ctx, db, namespace, value)
	if err != nil {
		return 0, ErrNotFound
	}
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, ErrNotFound
	}
	return n, nil
}

func resolveSubmission(ctx context.Context, db database.DBTX, value string) (string, error) {
	id, err := publicid.Resolve(ctx, db, "submissions", value)
	if err != nil {
		return "", ErrNotFound
	}
	return id, nil
}

func DegreeValue(label string) string { return strings.ToLower(label) }

func DegreeLabel(value string) string {
	switch value {
	case "bachelor":
		return "Bachelor"
	case "master":
		return "Master"
	default:
		return value
	}
}
