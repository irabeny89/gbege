package main

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create dummy users table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY)`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (id) VALUES (1)`)
	if err != nil {
		t.Fatalf("Failed to insert dummy user: %v", err)
	}

	return db
}

func TestSessionLifecycle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := CreateSessionTable(db)
	if err != nil {
		t.Fatalf("CreateSessionTable failed: %v", err)
	}

	userId := 1

	err = SaveSession(db, userId)
	if err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	var sessionId []byte
	err = db.QueryRow("SELECT id FROM sessions WHERE user_id = ?", userId).Scan(&sessionId)
	if err != nil {
		t.Fatalf("Failed to query created session: %v", err)
	}

	session, err := GetSession(db, sessionId)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if session.UserId != int64(userId) {
		t.Errorf("Expected UserId %d, got %d", userId, session.UserId)
	}

	if session.ExpiresAt.IsZero() {
		t.Error("Expected ExpiresAt to be set")
	}

	err = DeleteSession(db, sessionId)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	_, err = GetSession(db, sessionId)
	if err == nil {
		t.Error("Expected error when getting deleted session, but got nil")
	}
}

func TestDeleteExpiredSessions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := CreateSessionTable(db)
	if err != nil {
		t.Fatalf("CreateSessionTable failed: %v", err)
	}

	userId := 1

	pastTime := time.Now().Add(-24 * time.Hour)
	futureTime := time.Now().Add(24 * time.Hour)

	expiredId := []byte("expired-id")
	validId := []byte("valid-id")

	_, err = db.Exec(
		`INSERT INTO sessions (id, user_id, expires_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		expiredId, userId, pastTime, pastTime, pastTime,
	)
	if err != nil {
		t.Fatalf("Failed to insert expired session: %v", err)
	}

	_, err = db.Exec(
		`INSERT INTO sessions (id, user_id, expires_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		validId, userId, futureTime, futureTime, futureTime,
	)
	if err != nil {
		t.Fatalf("Failed to insert valid session: %v", err)
	}

	err = DeleteExpiredSessions(db)
	if err != nil {
		t.Fatalf("DeleteExpiredSessions failed: %v", err)
	}

	_, err = GetSession(db, expiredId)
	if err == nil {
		t.Error("Expected expired session to be deleted, but it was found")
	}

	_, err = GetSession(db, validId)
	if err != nil {
		t.Errorf("Expected valid session to exist, but got error: %v", err)
	}
}
