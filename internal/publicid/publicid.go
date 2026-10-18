// Package publicid assigns stable, opaque public identifiers to database records.
package publicid

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
)

type Querier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type DB interface {
	Querier
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

var ErrRetired = errors.New("public ID record has been retired")

var ErrIntegerExhausted = errors.New("public ID integer sequence exhausted")

// NextInteger reserves a monotonically increasing internal key. Allocate, insert
// the resource with this explicit key, and Ensure its public ID in the same tx.
// Only INTEGER resource namespaces have sequences; others return sql.ErrNoRows.
func NextInteger(ctx context.Context, tx *sql.Tx, namespace string) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `UPDATE public_id_sequences SET last_id=last_id+1 WHERE namespace=? AND last_id<9223372036854775807 RETURNING last_id`, namespace).Scan(&id)
	if !errors.Is(err, sql.ErrNoRows) {
		return id, err
	}
	// A missing RETURNING row means either an unknown namespace or exhaustion.
	err = tx.QueryRowContext(ctx, `SELECT last_id FROM public_id_sequences WHERE namespace=?`, namespace).Scan(&id)
	if err != nil {
		return 0, err
	}
	return 0, ErrIntegerExhausted
}

// New returns the canonical unpadded URL-safe encoding of 64 random bits.
func New() string {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		panic(fmt.Errorf("generate public ID: %w", err))
	}
	return base64.RawURLEncoding.EncodeToString(raw[:])
}

func Valid(value string) bool {
	if len(value) != 11 {
		return false
	}
	raw, err := base64.RawURLEncoding.Strict().DecodeString(value)
	return err == nil && len(raw) == 8 && base64.RawURLEncoding.EncodeToString(raw) == value
}

// Ensure allocates once. Call it in the same transaction that creates the record.
func Ensure(ctx context.Context, db DB, namespace, recordID string) (string, error) {
	return ensure(ctx, db, namespace, recordID, New)
}

func ensure(ctx context.Context, db DB, namespace, recordID string, generate func() string) (string, error) {
	if namespace == "" || recordID == "" {
		return "", errors.New("public ID namespace and record ID must not be empty")
	}
	for range 32 {
		var value string
		var retired int
		err := db.QueryRowContext(ctx, `SELECT public_id,retired FROM public_ids WHERE namespace=? AND record_id=?`, namespace, recordID).Scan(&value, &retired)
		if err == nil {
			if retired != 0 {
				return "", ErrRetired
			}
			return value, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
		value = generate()
		result, err := db.ExecContext(ctx, `INSERT INTO public_ids(namespace,record_id,public_id) VALUES(?,?,?) ON CONFLICT DO NOTHING`, namespace, recordID, value)
		if err != nil {
			return "", err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return "", err
		}
		if n == 1 {
			return value, nil
		}
	}
	return "", errors.New("public ID allocation exhausted collision retries")
}

// Resolve accepts canonical public IDs only; retired IDs are never resolved.
func Resolve(ctx context.Context, db Querier, namespace, value string) (string, error) {
	if !Valid(value) {
		return "", sql.ErrNoRows
	}
	var recordID string
	err := db.QueryRowContext(ctx, `SELECT record_id FROM public_ids WHERE namespace=? AND public_id=? AND retired=0`, namespace, value).Scan(&recordID)
	return recordID, err
}

// Lookup never allocates an ID as a side effect of rendering or reading a record.
func Lookup(ctx context.Context, db Querier, namespace, recordID string) (string, error) {
	var value string
	err := db.QueryRowContext(ctx, `SELECT public_id FROM public_ids WHERE namespace=? AND record_id=? AND retired=0`, namespace, recordID).Scan(&value)
	return value, err
}
