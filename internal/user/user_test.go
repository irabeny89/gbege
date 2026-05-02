package user

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/irabeny89/gosqlitex"
	_ "modernc.org/sqlite"
)

// MARK: - Test db
func setupUserTestDB(t *testing.T) *gosqlitex.DbClient {
	dbPath := t.TempDir() + "/test.db"
	db, err := gosqlitex.Open(&gosqlitex.Config{
		DbPath: dbPath,
		Driver: "sqlite",
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Run migrations to setup the schema
	err = db.RunMigrationsContext(context.Background(), "../../migrations", "_")
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

// MARK: - User lifecycle
func TestUserLifecycle(t *testing.T) {
	db := setupUserTestDB(t)

	username := "johndoe"
	plainPassword := "secretpassword"

	user, err := SaveUser(db, username, plainPassword)
	if err != nil {
		t.Fatalf("SaveUser failed: %v", err)
	}

	if user.Username != username {
		t.Errorf("Expected username %s, got %s", username, user.Username)
	}

	userByUsername, err := GetUserByUsername(db, username)
	if err != nil {
		t.Fatalf("GetUserByUsername failed: %v", err)
	}

	if userByUsername.Username != username {
		t.Errorf("Expected username %s, got %s", username, userByUsername.Username)
	}

	userById, err := GetUser(db, int(userByUsername.Id))
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if userById.Id != userByUsername.Id {
		t.Errorf("Expected id %d, got %d", userByUsername.Id, userById.Id)
	}

	// Test UpdateUserPhoto
	newPhoto := "https://example.com/photo.jpg"
	err = UpdateUserPhoto(db, int(userById.Id), newPhoto)
	if err != nil {
		t.Fatalf("UpdateUserPhoto failed: %v", err)
	}

	updatedUser, _ := GetUser(db, int(userById.Id))
	if updatedUser.Photo == nil || *updatedUser.Photo != newPhoto {
		t.Errorf("Expected photo %s, got %v", newPhoto, updatedUser.Photo)
	}

	// Test SoftDeleteUser
	err = SoftDeleteUser(db, int(userById.Id))
	if err != nil {
		t.Fatalf("SoftDeleteUser failed: %v", err)
	}

	deletedUser, _ := GetUser(db, int(userById.Id))
	if deletedUser != nil {
		t.Error("Expected GetUser to return nil for soft-deleted user")
	}

	// Verify DeletedAt directly in DB
	var deletedAt sql.NullTime
	err = db.QueryRow("SELECT deleted_at FROM users WHERE id = ?", userById.Id).Scan(&deletedAt)
	if err != nil {
		t.Fatalf("Failed to query deleted_at: %v", err)
	}
	if !deletedAt.Valid {
		t.Error("Expected deleted_at to be set in database")
	}

	// Test RemoveUser
	err = RemoveUser(db, int(userById.Id))
	if err != nil {
		t.Fatalf("RemoveUser failed: %v", err)
	}

	u, err := GetUser(db, int(userById.Id))
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if u != nil {
		t.Error("Expected GetUser to return nil for removed user, but it still exists")
	}
}

// MARK: - User cleanup
func TestCleanupDeletedUsers(t *testing.T) {
	var (
		db         = setupUserTestDB(t)
		oldUser    = "old_user"
		recentUser = "recent_user"
		password   = "password"
	)

	// Insert a user deleted more than 6 months ago
	oldDate := time.Now().AddDate(0, -7, 0).Format("2006-01-02 15:04:05")
	_, err := db.Exec(`
		INSERT INTO users (username, password, deleted_at)
		VALUES (?, ?, ?)
	`, oldUser, password, oldDate)
	if err != nil {
		t.Fatalf("Failed to insert old deleted user: %v", err)
	}

	// Insert a user deleted recently
	recentDate := time.Now().AddDate(0, -1, 0).Format("2006-01-02 15:04:05")
	_, err = db.Exec(`
		INSERT INTO users (username, password, deleted_at)
		VALUES (?, ?, ?)
	`, recentUser, password, recentDate)
	if err != nil {
		t.Fatalf("Failed to insert recent deleted user: %v", err)
	}

	err = CleanupDeletedUsers(db)
	if err != nil {
		t.Fatalf("CleanupDeletedUsers failed: %v", err)
	}

	// Verify old user is gone
	u_old, err := GetUserByUsername(db, oldUser)
	if err != nil {
		t.Fatalf("GetUserByUsername failed: %v", err)
	}
	if u_old != nil {
		t.Error("Expected old user to be cleaned up, but it still exists")
	}

	// Verify recent user still exists
	_, err = GetUserByUsername(db, recentUser)
	if err != nil {
		t.Errorf("Expected recent user to still exist, but got error: %v", err)
	}
}
