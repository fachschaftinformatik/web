package files

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/fachschaftinformatik/web/internal/content"
	"github.com/fachschaftinformatik/web/internal/data/blob"
)

func storeObject(ctx context.Context, tx *sql.Tx, storage writableStorage, value []byte, attempts *[]string) (string, error) {
	digest := sha256.Sum256(value)
	checksum := hex.EncodeToString(digest[:])
	var key string
	var size int64
	err := tx.QueryRowContext(ctx, "SELECT blob_key,nbytes FROM archive_objects WHERE checksum=?", checksum).Scan(&key, &size)
	if err == nil {
		object, err := storage.GetObject(ctx, key)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrStorage, err)
		}
		hash := sha256.New()
		n, readErr := io.Copy(hash, io.LimitReader(object, content.MaxPDFBytes+1))
		readErr = errors.Join(readErr, object.Close())
		if readErr != nil || n != int64(len(value)) || size != n || hex.EncodeToString(hash.Sum(nil)) != checksum {
			return "", fmt.Errorf("%w: Prüfsumme der vorhandenen Datei stimmt nicht überein", ErrStorage)
		}
		return checksum, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	// A failed transaction must only remove its own upload, even if another
	// process concurrently retries the same checksum after the rollback.
	key = "archive/objects/" + checksum + "/" + rand.Text() + ".pdf"
	*attempts = append(*attempts, key)
	if err := storage.Upload(ctx, key, bytes.NewReader(value), int64(len(value)), "application/pdf"); err != nil {
		return "", fmt.Errorf("%w: %v", ErrStorage, err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO archive_objects(checksum,blob_key,nbytes) VALUES(?,?,?)", checksum, key, len(value)); err != nil {
		return "", err
	}
	return checksum, nil
}

// CleanupBlobs removes unreferenced draft objects and compacts verified legacy
// copies in bounded batches. Shared targets and old download URLs stay intact.
func (s *Service) CleanupBlobs(ctx context.Context) error {
	storage, ok := s.storage.(writableStorage)
	if !ok {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "UPDATE archive_blob_cleanup SET blob_key=blob_key WHERE 0"); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, "SELECT blob_key FROM archive_blob_cleanup WHERE NOT EXISTS(SELECT 1 FROM archive_objects o WHERE o.blob_key=archive_blob_cleanup.blob_key) LIMIT 16")
	if err != nil {
		return err
	}
	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			rows.Close()
			return err
		}
		keys = append(keys, key)
	}
	err = errors.Join(rows.Err(), rows.Close())
	if err != nil {
		return err
	}
	var failures error
	for _, key := range keys {
		if err := storage.Delete(ctx, key); err != nil {
			failures = errors.Join(failures, err)
			continue
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM archive_blob_cleanup WHERE blob_key=?", key); err != nil {
			return err
		}
	}
	rows, err = tx.QueryContext(ctx, `SELECT c.blob_key,c.checksum,o.blob_key,o.nbytes FROM archive_legacy_copies c JOIN archive_objects o ON o.checksum=c.checksum ORDER BY c.blob_key LIMIT 16`)
	if err != nil {
		return err
	}
	type redundant struct {
		key, checksum, target string
		size                  int64
	}
	var copies []redundant
	for rows.Next() {
		var copy redundant
		if err := rows.Scan(&copy.key, &copy.checksum, &copy.target, &copy.size); err != nil {
			rows.Close()
			return err
		}
		copies = append(copies, copy)
	}
	err = errors.Join(rows.Err(), rows.Close())
	if err != nil {
		return err
	}
	for _, copy := range copies {
		// Verify both byte streams before discarding a legacy copy. The shared
		// target is immutable and remains referenced by every old download URL.
		if err := verifyStoredObject(ctx, storage, copy.target, copy.checksum, copy.size); err != nil {
			failures = errors.Join(failures, err)
			continue
		}
		if err := verifyStoredObject(ctx, storage, copy.key, copy.checksum, copy.size); err != nil && !errors.Is(err, blob.ErrNotFound) {
			failures = errors.Join(failures, err)
			continue
		}
		if err := storage.Delete(ctx, copy.key); err != nil {
			failures = errors.Join(failures, err)
			continue
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM archive_legacy_copies WHERE blob_key=?", copy.key); err != nil {
			return err
		}
	}
	return errors.Join(failures, tx.Commit())
}

func verifyStoredObject(ctx context.Context, storage Storage, key, checksum string, size int64) error {
	object, err := storage.GetObject(ctx, key)
	if err != nil {
		return err
	}
	hash := sha256.New()
	n, err := io.Copy(hash, io.LimitReader(object, size+1))
	err = errors.Join(err, object.Close())
	if err != nil {
		return err
	}
	if n != size || hex.EncodeToString(hash.Sum(nil)) != checksum {
		return fmt.Errorf("%w: checksum or size mismatch for %s", ErrStorage, key)
	}
	return nil
}
