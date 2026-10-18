package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fachschaftinformatik/web/internal/publicid"
	"github.com/pressly/goose/v3"
)

var publicIDTables = []struct{ namespace, table, key string }{
	{"posts", "discussion_posts", "id"},
	{"comments", "discussion_comments", "id"},
	{"events", "calendar_events", "id"},
	{"albums", "events", "id"},
	{"media", "media", "id"},
	{"modules", "modules", "id"},
	{"programs", "programs", "id"},
	{"documents", "archive", "id"},
	{"submissions", "archive_requests", "id"},
	{"users", "users", "id"},
	{"members", "people", "id"},
	{"roles", "responsibilities", "id"},
	{"publications", "meeting_publications", "id"},
	{"historical_documents", "historical_meeting_documents", "file"},
}

func createPublicIDs(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `CREATE TABLE public_ids (
		namespace TEXT NOT NULL CHECK(length(namespace)>0),
		record_id TEXT NOT NULL CHECK(length(record_id)>0),
		public_id TEXT NOT NULL CHECK(length(public_id)=11 AND public_id NOT GLOB '*[^A-Za-z0-9_-]*' AND substr(public_id,11,1) GLOB '[AEIMQUYcgkosw048]'),
		retired INTEGER NOT NULL DEFAULT 0 CHECK(retired IN (0,1)),
		UNIQUE(namespace,record_id),
		UNIQUE(namespace,public_id)
	) STRICT;
	CREATE UNIQUE INDEX public_ids_document_path ON public_ids(public_id)
		WHERE namespace IN ('publications','historical_documents');
	CREATE TRIGGER public_ids_no_delete BEFORE DELETE ON public_ids BEGIN
		SELECT RAISE(ABORT,'public ID reservations are permanent');
	END;
	CREATE TRIGGER public_ids_no_replace BEFORE INSERT ON public_ids
	WHEN EXISTS(SELECT 1 FROM public_ids WHERE
		(namespace=NEW.namespace AND (record_id=NEW.record_id OR public_id=NEW.public_id)) OR
		(namespace IN ('publications','historical_documents') AND NEW.namespace IN ('publications','historical_documents') AND public_id=NEW.public_id)) BEGIN
		SELECT RAISE(IGNORE);
	END;
	CREATE TRIGGER public_ids_immutable BEFORE UPDATE ON public_ids
	WHEN NEW.namespace IS NOT OLD.namespace OR NEW.record_id IS NOT OLD.record_id OR NEW.public_id IS NOT OLD.public_id OR NEW.retired < OLD.retired BEGIN
		SELECT RAISE(ABORT,'public ID mappings are immutable');
	END;`)
	if err != nil {
		return err
	}
	for _, entry := range publicIDTables {
		rows, err := tx.QueryContext(ctx, `SELECT CAST(`+quoteIdentifier(entry.key)+` AS TEXT) FROM `+quoteIdentifier(entry.table))
		if err != nil {
			return err
		}
		var ids []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			ids = append(ids, id)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		for _, id := range ids {
			if _, err := publicid.Ensure(ctx, tx, entry.namespace, id); err != nil {
				return fmt.Errorf("backfill public IDs for %s: %w", entry.table, err)
			}
		}
		statement := `CREATE TRIGGER ` + quoteIdentifier("trg_public_ids_retire_"+entry.table) + ` AFTER DELETE ON ` + quoteIdentifier(entry.table) + ` BEGIN UPDATE public_ids SET retired=1 WHERE namespace='` + entry.namespace + `' AND record_id=CAST(OLD.` + quoteIdentifier(entry.key) + ` AS TEXT); END`
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func publicIDsMigration() *goose.Migration {
	// No down migration: rolling back must not release public ID reservations.
	return goose.NewGoMigration(17, &goose.GoFunc{RunTx: createPublicIDs}, nil)
}
