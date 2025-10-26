package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func Migrate(dsn string, logger *log.Logger) error {
	logger.Println("Running database migrations...")
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("goose: failed to open DB: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("goose: failed to ping DB: %w", err)
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("goose: failed to set dialect: %w", err)
	}

	if err := goose.Up(db, "database/migrations"); err != nil {
		return fmt.Errorf("goose: up failed: %w", err)
	}

	logger.Println("Migrations applied successfully.")
	return nil
}
