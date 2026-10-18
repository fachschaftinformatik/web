package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pressly/goose/v3"
)

func programYear(value string) (int, error) {
	value = strings.TrimSpace(strings.TrimPrefix(value, "PO"))
	year, err := strconv.Atoi(value)
	if err != nil || len(value) != 4 || year < 1000 || year > 9999 {
		return 0, fmt.Errorf("invalid examination regulation year %q", value)
	}
	return year, nil
}

func numericProgramVersions(ctx context.Context, tx *sql.Tx) error {
	for _, entry := range []struct{ table, column string }{{"program_versions", "name"}, {"archive", "version"}, {"archive_assignments", "version"}} {
		if err := numericVersionColumn(ctx, tx, entry.table, entry.column); err != nil {
			return err
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT user_id,versions FROM user_archive_preferences`)
	if err != nil {
		return err
	}
	type preference struct {
		userID   int64
		versions string
	}
	var preferences []preference
	for rows.Next() {
		var p preference
		if err := rows.Scan(&p.userID, &p.versions); err != nil {
			rows.Close()
			return err
		}
		preferences = append(preferences, p)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, p := range preferences {
		var values []string
		if err := json.Unmarshal([]byte(p.versions), &values); err != nil {
			return err
		}
		for i, value := range values {
			year, err := programYear(value)
			if err != nil {
				return err
			}
			values[i] = strconv.Itoa(year)
		}
		encoded, err := json.Marshal(values)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE user_archive_preferences SET versions=? WHERE user_id=?`, string(encoded), p.userID); err != nil {
			return err
		}
	}
	return nil
}

func numericVersionColumn(ctx context.Context, tx *sql.Tx, table, column string) error {
	name, field := quoteIdentifier(table), quoteIdentifier(column)
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT `+field+` FROM `+name)
	if err != nil {
		return err
	}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			rows.Close()
			return err
		}
		if _, err := programYear(value); err != nil {
			rows.Close()
			return fmt.Errorf("%s.%s: %w", table, column, err)
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	var schema string
	if err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_schema WHERE type='table' AND name=?`, table).Scan(&schema); err != nil {
		return err
	}
	// Keep every other column, constraint, index and trigger unchanged. Recreate
	// under the original name so dependent views and foreign keys stay intact.
	start := strings.Index(schema, "(")
	if start < 0 {
		return fmt.Errorf("missing table definition for %s", table)
	}
	definition := schema[start:]
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(column) + `\s+TEXT NOT NULL`)
	if !pattern.MatchString(definition) {
		return fmt.Errorf("unrecognized version column definition in %s", table)
	}
	definition = pattern.ReplaceAllString(definition, column+" INTEGER NOT NULL CHECK("+field+" BETWEEN 1000 AND 9999)")
	rows, err = tx.QueryContext(ctx, `SELECT sql FROM sqlite_schema WHERE tbl_name=? AND type IN ('index','trigger') AND sql IS NOT NULL ORDER BY type,name`, table)
	if err != nil {
		return err
	}
	var objects []string
	for rows.Next() {
		var statement string
		if err := rows.Scan(&statement); err != nil {
			rows.Close()
			return err
		}
		objects = append(objects, statement)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	rows, err = tx.QueryContext(ctx, `SELECT name FROM pragma_table_info(?) ORDER BY cid`, table)
	if err != nil {
		return err
	}
	var expressions []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			rows.Close()
			return err
		}
		quoted := quoteIdentifier(value)
		if value == column {
			quoted = `CAST(trim(CASE WHEN substr(` + field + `,1,2)='PO' THEN substr(` + field + `,3) ELSE ` + field + ` END) AS INTEGER)`
		}
		expressions = append(expressions, quoted)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	copyName := quoteIdentifier(table + "_numeric")
	if _, err := tx.ExecContext(ctx, `CREATE TABLE `+copyName+definition+`; INSERT INTO `+copyName+` SELECT `+strings.Join(expressions, ",")+` FROM `+name+`; DROP TABLE `+name+`; CREATE TABLE `+name+definition+`; INSERT INTO `+name+` SELECT * FROM `+copyName+`; DROP TABLE `+copyName); err != nil {
		return err
	}
	for _, statement := range objects {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func programVersionsMigration() *goose.Migration {
	return goose.NewGoMigration(20, &goose.GoFunc{RunTx: numericProgramVersions}, nil)
}
