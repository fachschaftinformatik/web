package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/pressly/goose/v3"
)

func quoteIdentifier(value string) string { return `"` + strings.ReplaceAll(value, `"`, `""`) + `"` }

// utcRules covers every timestamp column in the schema, including nullable
// timestamps and legacy tables used by imports. Calendar-only event dates are
// allowed, but event date-times must obey the same UTC contract.
func utcRules(ctx context.Context, tx *sql.Tx) (map[string][]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT m.name,p.name FROM sqlite_schema m JOIN pragma_table_info(m.name) p WHERE m.type='table' AND m.name NOT LIKE 'sqlite_%' AND m.name!='goose_db_version' ORDER BY m.name,p.cid`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rules := map[string][]string{}
	for rows.Next() {
		var table, column string
		if err := rows.Scan(&table, &column); err != nil {
			return nil, err
		}
		if strings.HasSuffix(column, "_at") || column == "last_seen" || column == "password_reset_expires" || column == "event_date" {
			rules[table] = append(rules[table], column)
		}
	}
	return rules, rows.Err()
}

func validUTC(column string, event bool) string {
	// +0 seconds forces SQLite to normalise calendar overflow, so dates such
	// as 30 February and 24:00 cannot pass by merely matching the input shape.
	stamp := `(typeof(` + column + `)='text' AND length(` + column + `)=20 AND strftime('%Y-%m-%dT%H:%M:%SZ',` + column + `,'+0 seconds') IS ` + column + `)`
	if event {
		stamp += ` OR (typeof(` + column + `)='text' AND length(` + column + `)=10 AND strftime('%Y-%m-%d',` + column + `,'+0 days') IS ` + column + `)`
	}
	return `(` + column + ` IS NULL OR ` + column + `='' OR ` + stamp + `)`
}

func enforceUTC(ctx context.Context, tx *sql.Tx) error {
	rules, err := utcRules(ctx, tx)
	if err != nil {
		return err
	}
	tables := make([]string, 0, len(rules))
	for table := range rules {
		tables = append(tables, table)
	}
	sort.Strings(tables)
	for _, table := range tables {
		var invalid []string
		for _, column := range rules[table] {
			var count int
			if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM `+quoteIdentifier(table)+` WHERE NOT `+validUTC(quoteIdentifier(column), column == "event_date")).Scan(&count); err != nil {
				return err
			}
			if count != 0 {
				return fmt.Errorf("UTC enforcement: %s.%s contains %d non-canonical values", table, column, count)
			}
			invalid = append(invalid, `NOT `+validUTC(`NEW.`+quoteIdentifier(column), column == "event_date"))
		}
		for _, action := range []string{"INSERT", "UPDATE"} {
			name := quoteIdentifier("trg_utc_" + table + "_" + strings.ToLower(action))
			statement := `CREATE TRIGGER ` + name + ` BEFORE ` + action + ` ON ` + quoteIdentifier(table) + ` WHEN ` + strings.Join(invalid, " OR ") + ` BEGIN SELECT RAISE(ABORT,'timestamps must be UTC RFC 3339 ending in Z; calendar dates must be YYYY-MM-DD'); END`
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				return err
			}
		}
	}
	return nil
}

func removeUTCEnforcement(ctx context.Context, tx *sql.Tx) error {
	rules, err := utcRules(ctx, tx)
	if err != nil {
		return err
	}
	for table := range rules {
		for _, action := range []string{"insert", "update"} {
			if _, err := tx.ExecContext(ctx, `DROP TRIGGER IF EXISTS `+quoteIdentifier("trg_utc_"+table+"_"+action)); err != nil {
				return err
			}
		}
	}
	return nil
}

func utcEnforcementMigration() *goose.Migration {
	return goose.NewGoMigration(16, &goose.GoFunc{RunTx: enforceUTC}, &goose.GoFunc{RunTx: removeUTCEnforcement})
}
