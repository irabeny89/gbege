package main

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/irabeny89/gosqlitex"
)

func createMigTable(db *gosqlitex.DbClient) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
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
		return errors.New("Only files are allowed in migrations folder")
	}
	// split the filename on the first separator to get the timestamp.
	v, _, ok := strings.Cut(f.Name(), sep)
	if !ok {
		return errors.New("Migration file name separator not found")
	}
	// check if the timestamp is a valid integer
	if _, err := strconv.Atoi(v); err != nil {
		return errors.New("Migration file name prefix is not a number")
	}
	return nil
}

func updateDb(db *gosqlitex.DbClient, dir string, f os.DirEntry) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	sqlBytes, err := os.ReadFile(filepath.Join(dir, f.Name()))
	if err != nil {
		return err
	}
	_, err = tx.Exec(string(sqlBytes))
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
			INSERT INTO migrations (name) VALUES (?)
		`, f.Name())
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func RunMigrations(dir string, sep string, db *gosqlitex.DbClient) error {
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
		var id int64
		err = db.QueryRow(`SELECT id FROM migrations WHERE name = ?`, f.Name()).Scan(&id)
		if err != nil {
			return err
		}
		// ignore if migration has been run
		if id > 0 {
			continue
		}
		if err = updateDb(db, dir, f); err != nil {
			return err
		}
	}
	return nil
}
