package content

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
	"strconv"
	"time"

	"github.com/fachschaftinformatik/web/internal/data/blob"
)

// NewPersistentStore uses the migrated database and a private blob bucket.
func NewPersistentStore(db *sql.DB, provider blob.Provider) (*Store, error) {
	if db == nil || provider == nil {
		return nil, errors.New("Datenbank und Dateispeicher fehlen.")
	}
	if _, err := db.ExecContext(context.Background(), "SELECT id FROM meeting_publications LIMIT 0"); err != nil {
		return nil, err
	}
	return &Store{db: db, blobs: provider}, nil
}

func (s *Store) savePersistent(p Publication) (err error) {
	ownerID, ownerErr := strconv.ParseInt(p.Owner, 10, 64)
	if p.ID == "" || ownerErr != nil || ownerID <= 0 || strconv.FormatInt(ownerID, 10) != p.Owner || len(p.Data) > MaxPDFBytes {
		return errors.New("Der Entwurf konnte nicht gespeichert werden.")
	}
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Acquire SQLite's writer lock before reading ownership or uploading bytes.
	if _, err = tx.ExecContext(ctx, "UPDATE meeting_publications SET id=id WHERE id=?", p.ID); err != nil {
		return err
	}
	var owner string
	var oldKey, oldDigest string
	var published bool
	err = tx.QueryRowContext(ctx, "SELECT owner,published,blob_key,sha256 FROM meeting_publications WHERE id=?", p.ID).Scan(&owner, &published, &oldKey, &oldDigest)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err == nil && (owner != p.Owner || published) {
		return errors.New("Dieser Entwurf kann nicht geändert werden.")
	}
	hash := sha256.Sum256(p.Data)
	digest := hex.EncodeToString(hash[:])
	key := oldKey
	upload := oldKey == "" || oldDigest != digest
	if upload {
		key = "meetings/" + rand.Text()
	}
	// Upload errors may leave partial objects. Always attempt compensation.
	committed := false
	defer func() {
		if !committed && upload {
			tx.Rollback()
			if cleanup := s.blobs.Delete(ctx, key); cleanup != nil {
				err = errors.Join(err, fmt.Errorf("Datei konnte nicht entfernt werden: %w", cleanup))
				_, queueErr := s.db.ExecContext(ctx, "INSERT OR IGNORE INTO meeting_blob_cleanup(blob_key) VALUES(?)", key)
				err = errors.Join(err, queueErr)
			}
		}
	}()
	if upload {
		if err = s.blobs.Upload(ctx, key, bytes.NewReader(p.Data), int64(len(p.Data)), "application/pdf"); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO meeting_publications(id,owner,date,number,kind,filename,blob_key,bytes,sha256)
	VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET date=excluded.date,number=excluded.number,kind=excluded.kind,
	filename=excluded.filename,blob_key=excluded.blob_key,bytes=excluded.bytes,sha256=excluded.sha256`, p.ID, p.Owner, p.Date, p.Number, p.Kind, p.Filename, key, len(p.Data), digest)
	if err != nil {
		return err
	}
	if upload && oldKey != "" {
		if _, err = tx.ExecContext(ctx, "INSERT INTO meeting_blob_cleanup(blob_key) VALUES(?)", oldKey); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	committed = true
	// Deletion failures remain durably queued. The saved draft is already committed.
	_ = s.CleanupBlobs(ctx)
	return nil
}

type publicationRow interface{ Scan(...any) error }

func (s *Store) readPublication(row publicationRow) (Publication, bool, error) {
	var p Publication
	var key, digest string
	var size int64
	err := row.Scan(&p.ID, &p.Owner, &p.Date, &p.Number, &p.Kind, &p.Filename, &key, &size, &digest, &p.Published)
	if errors.Is(err, sql.ErrNoRows) {
		return p, false, nil
	}
	if err != nil {
		return p, false, err
	}
	object, err := s.blobs.GetObject(context.Background(), key)
	if err != nil {
		return p, false, err
	}
	p.Data, err = io.ReadAll(io.LimitReader(object, MaxPDFBytes+1))
	err = errors.Join(err, object.Close())
	if err != nil {
		return p, false, err
	}
	hash := sha256.Sum256(p.Data)
	if int64(len(p.Data)) != size || hex.EncodeToString(hash[:]) != digest {
		return p, false, errors.New("Die gespeicherte PDF-Datei ist beschädigt.")
	}
	return p, true, nil
}

const publicationColumns = "id,owner,date,number,kind,filename,blob_key,bytes,sha256,published"

func (s *Store) getPersistent(id string) (Publication, bool, error) {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return Publication{}, false, err
	}
	defer tx.Rollback()
	// Serialize blob snapshots with replacement and cleanup across Store instances.
	if _, err := tx.Exec("UPDATE meeting_publications SET id=id WHERE id=?", id); err != nil {
		return Publication{}, false, err
	}
	return s.readPublication(tx.QueryRow("SELECT "+publicationColumns+" FROM meeting_publications WHERE id=?", id))
}

func (s *Store) publishPersistent(id, owner string) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "UPDATE meeting_publications SET id=id WHERE id=?", id); err != nil {
		return err
	}
	p, ok, err := s.readPublication(tx.QueryRowContext(ctx, "SELECT "+publicationColumns+" FROM meeting_publications WHERE id=?", id))
	if err != nil {
		return err
	}
	if !ok || p.Owner != owner {
		return errors.New("Dieser Entwurf wurde nicht gefunden.")
	}
	if p.Published {
		return nil
	}
	if len(ValidatePublication(p)) != 0 {
		return errors.New("Prüfe die Angaben und die PDF-Datei.")
	}
	date, _ := time.Parse("2006-01-02", p.Date)
	conflict := errors.New("Datum und Sitzungsnummer stimmen nicht mit der vorhandenen Sitzung überein.")
	for _, m := range All() {
		if m.Year == date.Year() && (m.Date == p.Date && m.Number != p.Number || m.Number == p.Number && m.Date != p.Date) {
			return conflict
		}
	}
	var existing string
	err = tx.QueryRowContext(ctx, "SELECT date FROM published_meetings WHERE year=? AND number=?", date.Year(), p.Number).Scan(&existing)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err == nil && existing != p.Date {
		return conflict
	}
	var number int
	err = tx.QueryRowContext(ctx, "SELECT number FROM published_meetings WHERE date=?", p.Date).Scan(&number)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err == nil && number != p.Number {
		return conflict
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO published_meetings(year,number,date) VALUES(?,?,?) ON CONFLICT DO NOTHING", date.Year(), p.Number, p.Date); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE meeting_publications SET published=1 WHERE id=? AND owner=?", id, owner); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) allPersistent() ([]Meeting, error) {
	rows, err := s.db.Query("SELECT id,date,number,kind,bytes FROM meeting_publications WHERE published=1 ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := All()
	for rows.Next() {
		var p Publication
		var size int
		if err := rows.Scan(&p.ID, &p.Date, &p.Number, &p.Kind, &size); err != nil {
			return nil, err
		}
		records = appendPublication(records, p, size)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sortMeetings(records)
	return records, nil
}

// CleanupBlobs retries at most 16 obsolete draft objects. Failed deletions remain
// queued across restarts. Saves also invoke it after committing their changes.
func (s *Store) CleanupBlobs(ctx context.Context) error {
	if s.db == nil {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Readers acquire the same SQLite writer lock until their bytes are buffered.
	if _, err := tx.ExecContext(ctx, "UPDATE meeting_blob_cleanup SET blob_key=blob_key WHERE 0"); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `SELECT blob_key FROM meeting_blob_cleanup
	WHERE NOT EXISTS (SELECT 1 FROM meeting_publications p WHERE p.blob_key=meeting_blob_cleanup.blob_key)
	ORDER BY attempts,blob_key LIMIT 16`)
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
		if err := s.blobs.Delete(ctx, key); err != nil {
			failures = errors.Join(failures, err)
			if _, err := tx.ExecContext(ctx, "UPDATE meeting_blob_cleanup SET attempts=attempts+1 WHERE blob_key=?", key); err != nil {
				return errors.Join(failures, err)
			}
			continue
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM meeting_blob_cleanup WHERE blob_key=?", key); err != nil {
			return err
		}
	}
	return errors.Join(failures, tx.Commit())
}
