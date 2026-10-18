package database

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/fachschaftinformatik/web/internal/publicid"
	"github.com/pressly/goose/v3"
)

// normalizeProgramNames keeps study-program names separate from qualifications.
// programs is still the stable legacy relationship key; distinct degrees can
// therefore have the same name without changing any foreign or public keys.
func normalizeProgramNames(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `SELECT sql FROM sqlite_schema WHERE tbl_name='programs' AND type IN ('trigger','index') AND sql IS NOT NULL ORDER BY type,name`)
	if err != nil {
		return err
	}
	var definitions []string
	for rows.Next() {
		var definition string
		if err := rows.Scan(&definition); err != nil {
			rows.Close()
			return err
		}
		definitions = append(definitions, definition)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	// The migration connection has foreign_keys disabled. Copy before dropping,
	// then recreate the original triggers (UTC, reservations and sequences).
	if _, err := tx.ExecContext(ctx, `CREATE TABLE programs_normalized (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
		deleted_at TEXT
	) STRICT;
	INSERT INTO programs_normalized(id,name,created_at,updated_at,deleted_at)
	SELECT p.id,COALESCE(sp.name,p.name),p.created_at,p.updated_at,p.deleted_at
	FROM programs p
	LEFT JOIN program_qualifications pq ON pq.legacy_program_id=p.id
	LEFT JOIN study_programs sp ON sp.id=pq.study_program_id;
	DROP TABLE programs;
	ALTER TABLE programs_normalized RENAME TO programs;
	CREATE INDEX programs_name ON programs(name);`); err != nil {
		return err
	}
	for _, definition := range definitions {
		if _, err := tx.ExecContext(ctx, definition); err != nil {
			return err
		}
	}
	return registerStudyPrograms(ctx, tx)
}

func registerStudyPrograms(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `SELECT id FROM study_programs ORDER BY id`)
	if err != nil {
		return err
	}
	var ids []int64
	for rows.Next() {
		var id int64
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
		if _, err := publicid.Ensure(ctx, tx, "study_programs", strconv.FormatInt(id, 10)); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO public_id_sequences(namespace,last_id)
		SELECT 'study_programs',COALESCE(max(id),0) FROM study_programs;
	CREATE TRIGGER trg_public_ids_sequence_study_programs AFTER INSERT ON study_programs BEGIN
		UPDATE public_id_sequences SET last_id=max(last_id,NEW.id) WHERE namespace='study_programs';
	END;
	CREATE TRIGGER trg_public_ids_no_reuse_study_programs AFTER INSERT ON study_programs
	WHEN EXISTS(SELECT 1 FROM public_ids WHERE namespace='study_programs' AND record_id=CAST(NEW.id AS TEXT)) BEGIN
		SELECT RAISE(ABORT,'resource record identity is permanently reserved');
	END;
	CREATE TRIGGER trg_public_ids_retire_study_programs AFTER DELETE ON study_programs BEGIN
		UPDATE public_ids SET retired=1 WHERE namespace='study_programs' AND record_id=CAST(OLD.id AS TEXT);
	END;`)
	return err
}

func programNamesMigration() *goose.Migration {
	// Reintroducing degree suffixes would make names presentation-dependent
	// again and cannot faithfully reconstruct arbitrary original spelling.
	return goose.NewGoMigration(19, &goose.GoFunc{RunTx: normalizeProgramNames}, nil)
}
