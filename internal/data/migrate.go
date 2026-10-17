package data

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

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

	files, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("goose: embedded migrations: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, files)
	if err != nil {
		return fmt.Errorf("goose: create migration provider: %w", err)
	}
	if _, err := provider.Up(context.Background()); err != nil {
		return fmt.Errorf("goose: up failed: %w", err)
	}

	logger.Println("Migrations applied successfully.")
	return nil
}
