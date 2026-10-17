// Command import-site-data explicitly imports preserved website content.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed content.sql documents/*.pdf
var source embed.FS

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: go run ./tools/import-site-data DATABASE BLOBS")
	}
	if err := importData(os.Args[1], os.Args[2]); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Imported organization settings and historical documents; existing records preserved.")
}

func importData(database, blobs string) error {
	info, err := os.Stat(database)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("database must be an existing file")
	}
	db, err := sql.Open("sqlite", database)
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`SELECT file FROM historical_meeting_documents LIMIT 0`); err != nil {
		return fmt.Errorf("database must have schema migration 13 applied: %w", err)
	}
	files, err := source.ReadDir("documents")
	if err != nil {
		return err
	}
	for _, file := range files {
		data, err := source.ReadFile("documents/" + file.Name())
		if err != nil {
			return err
		}
		digest := fmt.Sprintf("%x", sha256.Sum256(data))
		target := filepath.Join(blobs, "meetings", "historical", digest)
		if err := preserveObject(target, data); err != nil {
			return err
		}
	}
	script, err := source.ReadFile("content.sql")
	if err != nil {
		return err
	}
	// The SQL is also usable manually. Here the transaction remains open until
	// every published document has passed the byte-count and hash checks.
	statements := strings.TrimSuffix(strings.Replace(string(script), "BEGIN;\n", "", 1), "COMMIT;\n")
	if _, err := tx.Exec(statements); err != nil {
		return err
	}
	rows, err := tx.Query(`SELECT file,sha256,bytes,blob_key FROM historical_meeting_documents`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var file, digest, key string
		var size int64
		if err := rows.Scan(&file, &digest, &size, &key); err != nil {
			return err
		}
		hash, err := hex.DecodeString(digest)
		if err != nil || len(hash) != sha256.Size || key != "meetings/historical/"+digest {
			return fmt.Errorf("invalid document object key: %s", file)
		}
		data, err := os.ReadFile(filepath.Join(blobs, filepath.FromSlash(key)))
		if err != nil {
			return err
		}
		if int64(len(data)) != size || fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
			return fmt.Errorf("document integrity mismatch: %s", file)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	return tx.Commit()
}

func preserveObject(target string, data []byte) error {
	if existing, err := os.ReadFile(target); err == nil {
		if !bytes.Equal(existing, data) {
			return fmt.Errorf("existing object differs: %s", target)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		return err
	}
	// Publish complete bytes atomically without replacing an existing object.
	file, err := os.CreateTemp(filepath.Dir(target), ".import-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Link(file.Name(), target); err != nil {
		if !os.IsExist(err) {
			return err
		}
		existing, readErr := os.ReadFile(target)
		if readErr != nil {
			return readErr
		}
		if !bytes.Equal(existing, data) {
			return fmt.Errorf("existing object differs: %s", target)
		}
	}
	return nil
}
