package database

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func createPublicIDSequences(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `CREATE TABLE public_id_sequences (
		namespace TEXT PRIMARY KEY,
		last_id INTEGER NOT NULL CHECK(last_id>=0)
	) STRICT;
	CREATE TRIGGER public_id_sequences_no_delete BEFORE DELETE ON public_id_sequences BEGIN
		SELECT RAISE(ABORT,'integer sequence reservations are permanent');
	END;
	CREATE TRIGGER public_id_sequences_no_replace BEFORE INSERT ON public_id_sequences
	WHEN EXISTS(SELECT 1 FROM public_id_sequences WHERE namespace=NEW.namespace) BEGIN
		SELECT RAISE(IGNORE);
	END;
	CREATE TRIGGER public_id_sequences_monotonic BEFORE UPDATE ON public_id_sequences
	WHEN NEW.namespace IS NOT OLD.namespace OR NEW.last_id<OLD.last_id BEGIN
		SELECT RAISE(ABORT,'integer sequences cannot decrease');
	END;`); err != nil {
		return err
	}
	for _, entry := range publicIDTables {
		// AFTER INSERT fires for REPLACE but not for ON CONFLICT DO UPDATE.
		// Checking after the insert also handles implicit INTEGER rowid values.
		statement := `CREATE TRIGGER ` + quoteIdentifier("trg_public_ids_no_reuse_"+entry.table) + ` AFTER INSERT ON ` + quoteIdentifier(entry.table) + ` WHEN EXISTS(SELECT 1 FROM public_ids WHERE namespace='` + entry.namespace + `' AND record_id=CAST(NEW.` + quoteIdentifier(entry.key) + ` AS TEXT)) BEGIN SELECT RAISE(ABORT,'resource record identity is permanently reserved'); END`
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
		var integerKey bool
		if err := tx.QueryRowContext(ctx, `SELECT type='INTEGER' AND pk=1 FROM pragma_table_info(?) WHERE name=?`, entry.table, entry.key).Scan(&integerKey); err != nil {
			return err
		}
		if !integerKey {
			continue
		}
		// Include retired reservations as well as live rows: migration 17 may
		// already have run, and its highest resource may since have been deleted.
		statement = `INSERT INTO public_id_sequences(namespace,last_id) SELECT ?,max(0,COALESCE((SELECT max(` + quoteIdentifier(entry.key) + `) FROM ` + quoteIdentifier(entry.table) + `),0),COALESCE((SELECT max(CAST(record_id AS INTEGER)) FROM public_ids WHERE namespace=?),0))`
		if _, err := tx.ExecContext(ctx, statement, entry.namespace, entry.namespace); err != nil {
			return err
		}
		statement = `CREATE TRIGGER ` + quoteIdentifier("trg_public_ids_sequence_"+entry.table) + ` AFTER INSERT ON ` + quoteIdentifier(entry.table) + ` BEGIN UPDATE public_id_sequences SET last_id=max(last_id,NEW.` + quoteIdentifier(entry.key) + `) WHERE namespace='` + entry.namespace + `'; END`
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func publicIDSequencesMigration() *goose.Migration {
	return goose.NewGoMigration(18, &goose.GoFunc{RunTx: createPublicIDSequences}, nil)
}
