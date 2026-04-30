package session

import (
	"context"
	"testing"
	"time"

	"github.com/irabeny89/gbege/internal/user"
	"github.com/irabeny89/gbege/migrate"
	"github.com/irabeny89/gosqlitex"
	_ "modernc.org/sqlite"
)

func setupSessionTestDB(t *testing.T) *gosqlitex.DbClient {
	dbPath := t.TempDir() + "/test.db"
	db, err := gosqlitex.Open(&gosqlitex.Config{
		DbPath: dbPath,
		Driver: "sqlite",
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create sessions and users tables via migrations
	err = migrate.RunMigrations(context.Background(), "../../migrations", "_", db)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Insert dummy user
	_, err = user.SaveUser(db, "test_user", "password")
	if err != nil {
		t.Fatalf("Failed to insert dummy user: %v", err)
	}

	return db
}

func TestSessionLifecycle(t *testing.T) {
	db := setupSessionTestDB(t)

	userId := 1

	session, err := SaveSession(db, userId)
	if err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	if session.UserId != int64(userId) {
		t.Errorf("Expected UserId %d, got %d", userId, session.UserId)
	}

	if session.ExpiresAt.IsZero() {
		t.Error("Expected ExpiresAt to be set")
	}

	retrievedSession, err := GetSession(db, session.Id)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if retrievedSession.UserId != int64(userId) {
		t.Errorf("Expected retrieved UserId %d, got %d", userId, retrievedSession.UserId)
	}

	err = DeleteSession(db, session.Id)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	_, err = GetSession(db, session.Id)
	if err == nil {
		t.Error("Expected error when getting deleted session, but got nil")
	}
}

func TestDeleteExpiredSessions(t *testing.T) {
	db := setupSessionTestDB(t)

	userId := 1

	pastTime := time.Now().Add(-24 * time.Hour)
	futureTime := time.Now().Add(24 * time.Hour)

	expiredId := []byte("expired-id")
	validId := []byte("valid-id")

	// We use the underlying db to insert sessions with specific times for testing
	// Note: In a real scenario, we might want a way to mock time or use a helper
	_, err := db.Exec(
		`INSERT INTO sessions (id, user_id, expires_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		expiredId, userId, pastTime, pastTime, pastTime,
	)
	if err != nil {
		// If it fails because of unique constraint on user_id (setupSessionTestDB inserts one), 
		// we should probably handle it. But setupSessionTestDB only inserts a user, not a session.
		// Wait, the sessions table has UNIQUE on user_id.
		t.Fatalf("Failed to insert expired session: %v", err)
	}

	// Insert another user for the second session to avoid unique constraint
	_, err = user.SaveUser(db, "test_user2", "password")
	if err != nil {
		t.Fatalf("Failed to insert dummy user 2: %v", err)
	}
	userId2 := 2

	_, err = db.Exec(
		`INSERT INTO sessions (id, user_id, expires_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		validId, userId2, futureTime, futureTime, futureTime,
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
