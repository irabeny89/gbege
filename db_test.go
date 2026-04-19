package main

import (
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func cleanupTestDb(db string) {
	os.Remove(db)
	os.Remove(db + "-shm")
	os.Remove(db + "-wal")
}

func TestNewDbClient(t *testing.T) {
	// Temporarily override dbPath to a test file
	originalDbPath := dbPath
	testDbPath := "test_new_dbclient.db"
	dbPath = testDbPath
	defer func() {
		dbPath = originalDbPath
		cleanupTestDb(testDbPath)
	}()

	client, err := NewDbClient()
	if err != nil {
		t.Fatalf("Failed to create DbClient: %v", err)
	}
	defer client.writePool.Close()
	defer client.readPool.Close()

	if client.readPool == nil {
		t.Error("Expected readPool to be initialized")
	}
	if client.writePool == nil {
		t.Error("Expected writePool to be initialized")
	}
}

func TestDbClient_ExecAndQuery(t *testing.T) {
	// Temporarily override dbPath to a test file
	originalDbPath := dbPath
	testDbPath := "test_exec_query.db"
	dbPath = testDbPath
	defer func() {
		dbPath = originalDbPath
		cleanupTestDb(testDbPath)
	}()

	client, err := NewDbClient()
	if err != nil {
		t.Fatalf("Failed to create DbClient: %v", err)
	}
	defer client.writePool.Close()
	defer client.readPool.Close()

	// 1. Test Exec (Write)
	_, err = client.Exec("CREATE TABLE IF NOT EXISTS test_users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Exec failed to create table: %v", err)
	}

	res, err := client.Exec("INSERT INTO test_users (name) VALUES (?)", "Alice")
	if err != nil {
		t.Fatalf("Exec failed to insert: %v", err)
	}
	
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get last insert id: %v", err)
	}

	// 2. Test QueryRow (Read)
	var name string
	err = client.QueryRow("SELECT name FROM test_users WHERE id = ?", id).Scan(&name)
	if err != nil {
		t.Fatalf("QueryRow failed: %v", err)
	}
	if name != "Alice" {
		t.Errorf("Expected name 'Alice', got '%s'", name)
	}

	// 3. Test Query (Read multi)
	client.Exec("INSERT INTO test_users (name) VALUES (?)", "Bob")
	
	rows, err := client.Query("SELECT name FROM test_users ORDER BY id ASC")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatalf("Rows scan failed: %v", err)
		}
		names = append(names, n)
	}

	if len(names) != 2 || names[0] != "Alice" || names[1] != "Bob" {
		t.Errorf("Expected ['Alice', 'Bob'], got %v", names)
	}
}
