package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/irabeny89/gosqlitex"
	_ "modernc.org/sqlite"
)

func setupMigrateTestDB(t *testing.T) *gosqlitex.DbClient {
	dbPath := t.TempDir() + "/test_migrate.db"
	db, err := gosqlitex.Open(&gosqlitex.Config{
		DbPath: dbPath,
		Driver: "sqlite",
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	return db
}

func TestValidateFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		isDir    bool
		sep      string
		wantErr  bool
	}{
		{"valid file", "1_init.sql", false, "_", false},
		{"invalid sep", "1-init.sql", false, "_", true},
		{"not a number", "a_init.sql", false, "_", true},
		{"is directory", "1_dir", true, "_", true},
		{"empty name", "", false, "_", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &mockDirEntry{name: tt.fileName, isDir: tt.isDir}
			err := validateFile(f, tt.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockDirEntry struct {
	name  string
	isDir bool
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return m.isDir }
func (m *mockDirEntry) Type() os.FileMode          { return 0 }
func (m *mockDirEntry) Info() (os.FileInfo, error) { return nil, nil }

func TestRunMigrations(t *testing.T) {
	db := setupMigrateTestDB(t)

	// Create a temporary migrations directory
	migDir := t.TempDir()

	// Create some migration files
	mig1 := "1_create_users.sql"
	mig2 := "2_add_email.sql"

	err := os.WriteFile(filepath.Join(migDir, mig1), []byte(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(migDir, mig2), []byte(`ALTER TABLE users ADD COLUMN email TEXT;`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Run migrations
	err = RunMigrations(migDir, "_", db)
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Verify table exists and has column
	_, err = db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", "test", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to insert into migrated table: %v", err)
	}

	// Verify migrations table recorded the entries
	var count int
	err = db.QueryRow("SELECT count(*) FROM migrations").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("Expected 2 migrations recorded, got %d", count)
	}

	// Run migrations again, should skip
	err = RunMigrations(migDir, "_", db)
	if err != nil {
		t.Fatalf("RunMigrations failed on second run: %v", err)
	}

	err = db.QueryRow("SELECT count(*) FROM migrations").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("Expected still 2 migrations recorded, got %d", count)
	}

	// Add a new migration
	mig3 := "3_new_table.sql"
	err = os.WriteFile(filepath.Join(migDir, mig3), []byte(`CREATE TABLE logs (id INTEGER PRIMARY KEY);`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = RunMigrations(migDir, "_", db)
	if err != nil {
		t.Fatalf("RunMigrations failed on third run: %v", err)
	}

	err = db.QueryRow("SELECT count(*) FROM migrations").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Errorf("Expected 3 migrations recorded, got %d", count)
	}
}

func TestRunMigrations_InvalidFile(t *testing.T) {
	db := setupMigrateTestDB(t)

	migDir := t.TempDir()
	err := os.WriteFile(filepath.Join(migDir, "invalid.sql"), []byte(`CREATE TABLE x (id);`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = RunMigrations(migDir, "_", db)
	if err == nil {
		t.Error("Expected error for invalid migration filename, got nil")
	}
}

func TestRunMigrations_FailingMigration(t *testing.T) {
	db := setupMigrateTestDB(t)

	migDir := t.TempDir()
	// Valid migration
	err := os.WriteFile(filepath.Join(migDir, "1_valid.sql"), []byte(`CREATE TABLE valid (id INTEGER);`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	// Failing migration (invalid SQL)
	err = os.WriteFile(filepath.Join(migDir, "2_invalid.sql"), []byte(`CREATE TABLE invalid (id INTEGER;`), 0644) // Missing closing paren
	if err != nil {
		t.Fatal(err)
	}

	err = RunMigrations(migDir, "_", db)
	if err == nil {
		t.Error("Expected error for failing migration, got nil")
	}

	// Verify that the first migration was applied (Wait, RunMigrations stops at the first failure)
	// But it should have applied the first one because it's a loop.
	// Actually, RunMigrations processes files one by one.
	
	var count int
	err = db.QueryRow("SELECT count(*) FROM migrations").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	// It should be 1 if it stopped at the second one
	if count != 1 {
		t.Errorf("Expected 1 migration to be recorded, got %d", count)
	}

	// Verify that the second table was NOT created
	_, err = db.Exec("INSERT INTO invalid (id) VALUES (1)")
	if err == nil {
		t.Error("Expected error inserting into table from failed migration, but table seems to exist")
	}
}
