// Package migrate handles migration using the migration directory
package migrate

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/irabeny89/gbege/internal/logger"
	"github.com/irabeny89/gosqlitex"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func createMigTable(db *gosqlitex.DbClient) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			query BLOB NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TRIGGER IF NOT EXISTS update_mig_updated_at 
		AFTER UPDATE ON migrations
		BEGIN
			UPDATE migrations SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
	`)
	return err
}

func validateFile(f os.DirEntry, sep string) error {
	if f.IsDir() {
		return errors.New("only files are allowed in migrations folder")
	}
	// split the filename on the first separator to get the timestamp.
	v, _, ok := strings.Cut(f.Name(), sep)
	if !ok {
		return errors.New("migration file name separator not found")
	}
	// check if the timestamp is a valid integer
	if _, err := strconv.Atoi(v); err != nil {
		return errors.New("migration file name prefix is not a number")
	}
	return nil
}

func updateDB(ctx context.Context, db *gosqlitex.DbClient, fn string, q []byte) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, string(q))
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
			INSERT INTO migrations (name, query) VALUES (?, ?)
		`, fn, q)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// RunMigrations applies all migrations in the given directory to the database.
//
// This function:
// - Creates the migrations table if it doesn't exist
//
// - Reads all files in the specified directory
//
// - Validates that each file is a valid migration file
//
// - Checks if the migration has already been applied
//
// - Applies the migration if it hasn't been applied
//
// - Records the migration in the migrations table
//
// - Rolls back the transaction if any error occurs
func RunMigrations(ctx context.Context, dir string, sep string, db *gosqlitex.DbClient) error {
	if err := createMigTable(db); err != nil {
		return err
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err = validateFile(f, sep); err != nil {
			return err
		}
		var query []byte
		err = db.QueryRowContext(ctx, `SELECT query FROM migrations WHERE name = ?`, f.Name()).Scan(&query)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		sqlBytes, errFile := os.ReadFile(filepath.Join(dir, f.Name()))
		if errFile != nil {
			return errFile
		}

		if err == nil {
			// Migration already applied
			logger.Log.Info("Migration already applied", "name", f.Name())
			dmp := diffmatchpatch.New()
			if !bytes.Equal(query, sqlBytes) {
				diff := dmp.DiffMain(string(query), string(sqlBytes), false)
				fmt.Printf("Migration mismatch for %s\n", f.Name())
				fmt.Println(dmp.DiffPrettyText(diff))
				return fmt.Errorf("migration content changed for %s", f.Name())
			}
			continue
		}

		// apply migration
		if err = updateDB(ctx, db, f.Name(), sqlBytes); err != nil {
			return err
		}
		logger.Log.Info("Applied migration", "name", f.Name())
	}
	return nil
}
