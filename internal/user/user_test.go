package user

import (
	"testing"

	"github.com/irabeny89/gosqlitex"
	_ "modernc.org/sqlite"
)

func setupUserTestDB(t *testing.T) *gosqlitex.DbClient {
	dbPath := t.TempDir() + "/test.db"
	db, err := gosqlitex.Open(&gosqlitex.Config{
		DbPath: dbPath,
		Driver: "sqlite",
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	return db
}

func TestUserLifecycle(t *testing.T) {
	db := setupUserTestDB(t)

	err := CreateUserTable(db)
	if err != nil {
		t.Fatalf("CreateUserTable failed: %v", err)
	}

	fullName := "John Doe"
	alias := "johndoe"
	plainPassword := "secretpassword"

	user, err := SaveUser(db, fullName, alias, plainPassword)
	if err != nil {
		t.Fatalf("SaveUser failed: %v", err)
	}

	if user.Name != fullName {
		t.Errorf("Expected name %s, got %s", fullName, user.Name)
	}

	userByAlias, err := GetUserByAlias(db, alias)
	if err != nil {
		t.Fatalf("GetUserByAlias failed: %v", err)
	}

	if userByAlias.Name != fullName {
		t.Errorf("Expected name %s, got %s", fullName, userByAlias.Name)
	}

	userById, err := GetUser(db, int(userByAlias.Id))
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if userById.Id != userByAlias.Id {
		t.Errorf("Expected id %d, got %d", userByAlias.Id, userById.Id)
	}

	// Test UpdateUserPhoto
	newPhoto := "https://example.com/photo.jpg"
	err = UpdateUserPhoto(db, int(userById.Id), newPhoto)
	if err != nil {
		t.Fatalf("UpdateUserPhoto failed: %v", err)
	}

	updatedUser, _ := GetUser(db, int(userById.Id))
	if updatedUser.Photo != newPhoto {
		t.Errorf("Expected photo %s, got %s", newPhoto, updatedUser.Photo)
	}

	// Test SoftDeleteUser
	err = SoftDeleteUser(db, int(userById.Id))
	if err != nil {
		t.Fatalf("SoftDeleteUser failed: %v", err)
	}

	deletedUser, _ := GetUser(db, int(userById.Id))
	if deletedUser.DeletedAt.IsZero() {
		t.Error("Expected DeletedAt to be set")
	}

	// Test RemoveUser
	err = RemoveUser(db, int(userById.Id))
	if err != nil {
		t.Fatalf("RemoveUser failed: %v", err)
	}

	_, err = GetUser(db, int(userById.Id))
	if err == nil {
		t.Error("Expected error when getting removed user, but got nil")
	}
}
